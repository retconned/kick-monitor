# Kick Monitor

**Disclaimer:** This project began as a small side project for personal learning and exploration. While the full stack is functional, certain parts are still under active development or awaiting further design polish. Contributions are very welcome to help mature this project!

Kick Monitor is a full-stack application born from a passion to enable more transparency about livestreaming metrics and numbers, especially in an environment where botting and artificial number boosting are prevalent. This project works solely based on public-facing data, cleaning it up and presenting it in a more insightful way. Currently, it's designed for self-hosting and local use but has the potential to be upgraded into a full-fledged platform usable by multiple real users.

## Current Project Status

- **Frontend:** Mostly functional. The dashboard displays data, but its design is not yet finalized. The contact page is designed but not yet implemented.
- **Backend:** Core data collection, processing, and API endpoints are functional.
- **Data Lifecycle (To Be Implemented):** Chat messages are collected. A cleanup mechanism to delete chat messages after stream reports are generated is planned to ensure efficient data management.

## Features

**Backend (Go)**
Kick Monitor's Go backend tracks specified channels, fetches real-time and historical data, and processes this data for analytics and reporting, leveraging PostgreSQL, GORM, Go routines, and concurrency for optimized performance.

- **Channel Management:** Add channels to monitor via HTTP POST requests.
- **Periodic Data Fetching:** Automatically fetches current channel and livestream data from the Kick API every 2 minutes for active monitored channels.
- **Real-time Chat Monitoring:** Connects to Kick's WebSocket chat servers to capture and process live chat messages.
- **Database Persistence:** Stores all collected data (monitored channels, historical channel data, livestream snapshots) in a PostgreSQL database.
    - `monitored_channels`: Stores the list of channels being monitored.
    - `channel_data`: Historical snapshots of channel information.
    - `livestream_data`: Historical snapshots of livestream details (viewer count, title, etc.).
- **Optimized Performance:** Utilizes Go routines and channels for highly concurrent and efficient data processing, especially for high-volume chat messages.

**Frontend (React)**
The frontend is a modern React application providing an intuitive user interface to interact with the Kick Monitor backend.

- **Technologies Used:** Built with React 19, Vite, React Router, React Query, Zod for schema validation, TailwindCSS for styling, and Shadcn/UI for accessible UI components.
- **Cloudflare Optimized:** Designed and targeted for deployment on Cloudflare Pages (or similar static site hosting) for production, leveraging CDN for speed and reliability.

## Full Stack Architecture

The Kick Monitor application utilizes a containerized microservice-like architecture orchestrated by Docker Compose.

- **Frontend Service (`kick-monitor-frontend`):** A lightweight Nginx container serves the static React application files.
- **Backend Service (`kick-monitor-api`):** The Go application handles all business logic, data processing, and API endpoints.
- **Nginx Reverse Proxy (`nginx`):** A central Nginx instance acts as the public entry point. It intelligently routes traffic:
    - Requests for static assets (e.g., `/`, `/index.html`, `/assets/*`) are proxied to the `kick-monitor-frontend` service.
    - Requests for API endpoints (e.g., `/api/*`, `/api/protected/*`) are proxied to the `kick-monitor-api` service.
- **Database (`db`):** A PostgreSQL instance for persistent data storage.
- **Proxy (`flaresolverr`):** An external proxy (FlareSolverr) to handle potential anti-bot measures when fetching data from Kick.com.

This setup ensures optimal performance, clear separation of concerns, and ease of deployment.

## Getting Started (Local Development)

### Prerequisites

- Go (version 1.22 or higher recommended)
- Node.js (version 20 or higher recommended, for frontend development)
- Docker and Docker Compose
- A PostgreSQL client (e.g., `psql`)
- `air` for Go hot-reloading (`go install github.com/air-verse/air@latest`)

### 1. Clone the Repository

```bash
git clone https://github.com/retconned/kick-monitor
cd kick-monitor
```

### 2. Configure Environment Variables

The application relies on environment variables for database connection, proxy URL, and JWT secret. Copy `.env.example` to `.env` and fill in the values.

```bash
cp .env.example .env
# Open .env and populate JWT_SECRET, etc.
```

### 3. Build and Run the Full Stack with Docker Compose (Recommended)

This command will build your Go backend, build your React frontend, set up the Nginx proxy, and start all services, including PostgreSQL and Flaresolverr.

```bash
# Build the Docker images for all services and start them in detached mode
docker-compose build --no-cache && docker-compose up -d

# To stop and remove containers later
docker-compose down
```

Once running, access the application:

- **Frontend:** Open your browser to `http://localhost`.
- **Backend API (through Nginx):** `http://localhost/api/health` (or other API endpoints).

### 4. Running Backend Locally (for faster iteration)

While Docker Compose runs the full stack, you might want to run the Go backend locally for faster development cycles with hot-reloading.

**First, ensure Docker Compose services for `db`, `flaresolverr`, and `nginx` are running:**

```bash
docker-compose up -d db flaresolverr nginx
```

Then, you can run the Go backend:

```bash
# With hot-reloading using `air`
make dev-air

# Without `air`
make dev
```

When running the backend locally, your frontend (if also run locally with `npm run dev` in `web/`) will communicate with the backend via the Nginx proxy.

### 5. Running Frontend Locally (for faster iteration)

To develop the frontend with Vite's dev server, navigate to the `web` directory and run the development server.

**First, ensure your Docker Compose services for `db`, `flaresolverr`, `kick-monitor-api`, and `nginx` are running:**

```bash
docker-compose up -d db flaresolverr kick-monitor-api nginx
```

Then, in a separate terminal:

```bash
cd web
pnpm install # Only if you haven't done it before or deps changed
pnpm dev

# Or
make dev-web
```

The Vite dev server will typically start on `http://localhost:5173` (or similar). Open your browser to this address. Vite's proxy configuration (in `web/vite.config.js`) will forward `/api/*` requests to your running `kick-monitor-api` service.

## API Endpoints

Once the backend is running (either via Docker Compose or locally), it exposes the following (and more) API endpoints via the Nginx proxy:

- **`GET /api/health`**: Checks the health status of the backend API.
- **`GET /api/status`**: Provides a general status message from the backend.
- **`POST /api/auth/login`**: Authenticates a user and returns a JWT token.
    - **Body (x-www-form-urlencoded or JSON if your handler accepts):** `username=admin&password=password`
- **`POST /api/add_channel`**
    - **Body (JSON):** `{"username": "xqc", "is_active": true}`
    - Adds or updates a channel in `monitored_channels`. If active, it starts monitoring API and WebSocket data.

## Deploying Frontend to Cloudflare Pages

The frontend (located in the `web/` directory) is built to be a static single-page application (SPA), making it ideal for deployment on Cloudflare Pages.

### 1. Prepare Your Project

Ensure your `web/vite.config.js` has `base: './'` for relative paths in the build output. Ensure `web/package.json` has a `build` script (`"build": "vite build"`).

### 2. Connect to Cloudflare Pages

1.  **Commit your code:** Push your frontend code to a Git repository (GitHub, GitLab, Bitbucket, etc.).
2.  **Log in to Cloudflare:** Go to your Cloudflare dashboard.
3.  **Navigate to Pages:** Select "Workers & Pages" -> "Pages" and click "Create application".
4.  **Connect Git:** Choose "Connect to Git" and select your repository.
5.  **Configure Build Settings:**
    - **Project name:** `kick-monitor-frontend` (or your preferred name)
    - **Production branch:** `main` (or your default branch)
    - **Build command:** `pnpm run build`
    - **Build directory:** `web/dist` (This tells Cloudflare where to find your built frontend files after the `npm run build` command runs from within the `web` directory)
    - **Root directory:** `web` (This is crucial, it tells Cloudflare where your `package.json` and frontend source code are located within your repo)
6.  **Environment Variables (Optional):** If your frontend build process needs any environment variables, define them here.
7.  **Save and Deploy:** Click "Save and Deploy". Cloudflare Pages will automatically fetch your code, run the build command, and deploy your static site.

### 3. Configure API Endpoint in Frontend (for Production)

When deployed on Cloudflare Pages, your frontend will likely be served from `your-app-name.pages.dev`. It will need to know where your _backend API_ is hosted.

**Methods:**

1.  **Runtime Environment Variables (Cloudflare Pages):**

    - On Cloudflare Pages, you can define environment variables that are available to your frontend _at runtime_ (e.g., in `process.env.REACT_APP_API_URL` or similar).
    - In your frontend code, you would use this variable:

        ```javascript
        // Example: api.js in your React app
        const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || "/api"; // Use /api locally, full URL in prod

        export const fetchHealth = () =>
            fetch(`${API_BASE_URL}/health`).then((res) => res.json());
        ```

    - In Cloudflare Pages settings, define a variable like `VITE_API_BASE_URL` with the full URL to your _deployed backend service_.
        - **Example:** If your Go backend is deployed to `api.yourdomain.com`, set `VITE_API_BASE_URL=https://api.yourdomain.com/api`.
        - If your backend is behind a Cloudflare Worker or Gateway that proxies `yourdomain.com/api` to your backend, you might set `VITE_API_BASE_URL=https://yourdomain.com/api`.

2.  **Cloudflare Workers/Gateways for Rewrites:**
    A common pattern is to deploy your backend to a server (VM, Kubernetes, Cloud Run, etc.) and then use a Cloudflare Worker or a Custom Hostname with an origin rule to proxy requests from `/api/*` on your Cloudflare Pages domain to your backend's actual URL. This way, your frontend can simply call `/api/health`, and Cloudflare handles the proxying.

**Important:** Your frontend's `vite.config.js` proxy only applies to `npm run dev`. For the production build, you need to rely on environment variables or Cloudflare routing.

## Contributing

Contributions are very welcome to help enhance and complete this project! Feel free to open issues, submit pull requests, or discuss new features. Your help is appreciated in evolving Kick Monitor.
