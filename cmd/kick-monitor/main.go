package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/retconned/kick-monitor/internal/api"
	"github.com/retconned/kick-monitor/internal/auth"
	"github.com/retconned/kick-monitor/internal/db"
	"github.com/retconned/kick-monitor/internal/models"
	"github.com/retconned/kick-monitor/internal/monitor"
	"github.com/retconned/kick-monitor/internal/util"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
	"golang.org/x/time/rate"
	"gorm.io/gorm"
)

func main() {
	db.Init()

	// Initialize JWT Secret for authentication
	auth.InitAuth()

	proxyURLEnv := os.Getenv("PROXY_URL")
	if proxyURLEnv == "" {
		log.Fatal("PROXY_URL environment variable is not set. Please set it in your environment or docker-compose.yml.")
	}
	log.Printf("Resolved PROXY_URL from environment: %s", proxyURLEnv)

	e := echo.New()

	monitor.SetProxyURL(proxyURLEnv)
	e.Logger.Print("Proxy URL successfully configured.")

	// Start monitoring Go routines for active channels
	var activeChannels []models.MonitoredChannel
	if err := db.DB.Where("is_active = ?", true).Find(&activeChannels).Error; err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Fatalf("Failed to load active channels: %v", err)
		}
		e.Logger.Print("No active channels found in the database on startup.")
	}

	for _, channel := range activeChannels {
		go monitor.StartMonitoringChannel(&channel)
	}

	e.Logger.SetLevel(log.INFO) // (INFO, DEBUG, WARN, ERROR, OFF)

	// --- Custom Error Handler ---
	e.HTTPErrorHandler = util.CustomHTTPErrorHandler

	// Logger middleware (using Echo's default for requests)
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: `{"time":"${time_rfc3339}","id":"${id}","remote_ip":"${remote_ip}","host":"${host}",` +
			`"method":"${method}","uri":"${uri}","user_agent":"${user_agent}",` +
			`"status":${status},"error":"${error}","latency":"${latency_human}"` +
			`,"bytes_in":${bytes_in},"bytes_out":${bytes_out}}` + "\n",
		CustomTimeFormat: "2006-01-02 15:04:05.000",
		Output:           os.Stdout,
	}))

	e.Use(middleware.Recover())   // Recovers from panics and serves a 500 error
	e.Use(middleware.RequestID()) // Assigns a unique ID to each request (useful for tracing logs)
	e.Use(middleware.Secure())

	// CORS middleware (configure carefully for production)
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"}, //TODO: switch it to  "https://yourfrontend.com" when its production

		AllowMethods:     []string{http.MethodGet, http.MethodHead, http.MethodPut, http.MethodPatch, http.MethodPost, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization, echo.HeaderXCSRFToken},
		AllowCredentials: true,
		MaxAge:           300, // Max age for preflight requests in seconds
	}))

	// CSRF middleware (optional, for form submissions)
	// Needs careful implementation with frontend to send CSRF token with requests
	// e.Use(middleware.CSRFWithConfig(middleware.CSRFConfig{
	// 	TokenLookup: "header:X-CSRF-Token", // or "form:_csrf" etc.
	// 	CookiePath:  "/",
	// 	CookieHTTPOnly: true,
	// 	CookieSameSite: http.SameSiteLaxMode, // Adjust for cross-domain if needed
	// 	CookieSecure: true, // Set to true in HTTPS production
	// }))

	// Rate Limiter middleware
	config := middleware.RateLimiterConfig{
		Skipper: middleware.DefaultSkipper,
		Store: middleware.NewRateLimiterMemoryStoreWithConfig(
			middleware.RateLimiterMemoryStoreConfig{Rate: rate.Limit(10), Burst: 30, ExpiresIn: 3 * time.Minute},
		),
		IdentifierExtractor: func(ctx echo.Context) (string, error) {
			id := ctx.RealIP()
			return id, nil
		},
		ErrorHandler: func(context echo.Context, err error) error {
			return context.JSON(http.StatusForbidden, nil)
		},
		DenyHandler: func(context echo.Context, identifier string, err error) error {
			return context.JSON(http.StatusTooManyRequests, nil)
		},
	}

	e.Use(middleware.RateLimiterWithConfig(config))

	// health endpoint
	e.GET("/health", api.HealthCheckHandler)

	// public routes start here
	e.POST("/register", auth.RegisterHandler)
	e.POST("/login", auth.LoginHandler)

	e.POST("/process_livestream_report", api.ProcessLivestreamReportHandler) // This is asynchronous, can be public

	// Reports API
	// Group these routes with common prefixes
	// e.GET("/reports/:reportUUID", api.GetReportByUUIDHandler)                        // /reports/uuid-string
	// e.GET("/channels/:channelID/reports", api.GetReportsByChannelIDHandler)          // /channels/id/reports
	// e.GET("/livestreams/:livestreamID/reports", api.GetReportsByLivestreamIDHandler) // /livestreams/id/reports

	// Channels Info API
	e.GET("/profile/:username", api.GetStreamerProfileHandler) // /channels/id/profile (aggregated profile)

	// proeteced routes start here
	r := e.Group("/protected")
	r.Use(auth.AuthMiddleware())
	r.POST("/add_channel", api.AddChannelHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Start server in a goroutine
	go func() {
		if err := e.Start(":" + port); err != nil && !errors.Is(err, http.ErrServerClosed) {
			e.Logger.Fatalf("shutting down the server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, os.Kill)
	<-quit // Blocks until signal is received

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // 10 seconds timeout for shutdown
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}
	e.Logger.Print("Server shut down gracefully.")
}
