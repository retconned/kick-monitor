# Kick Monitor

Kick Monitor is a Go backend application designed to monitor live streams on Kick.com. It tracks specified channels, fetches real-time and historical data, and processes this data for analytics and reporting, leveraging PostgreSQL, GORM, Go routines, and concurrency for optimized performance.

## Features

- **Channel Management:** Add channels to monitor via a simple HTTP POST request.
- **Periodic Data Fetching:** Automatically fetches current channel and livestream data from the Kick API every 2 minutes for active monitored channels.
- **Real-time Chat Monitoring:** Connects to Kick's WebSocket chat servers to capture and process live chat messages.
- **Database Persistence:** Stores all collected data (monitored channels, historical channel data, livestream snapshots) in a PostgreSQL database.
    - `monitored_channels`: Stores the list of channels being monitored.
    - `channel_data`: Historical snapshots of channel information.
    - `livestream_data`: Historical snapshots of livestream details (viewer count, title, etc.).
- **Optimized Performance:** Utilizes Go routines and channels for highly concurrent and efficient data processing, especially for high-volume chat messages.
- **Docker Support:** Containerized with Docker and Docker Compose for easy setup and deployment.
- **Development Workflow:** Includes a `Makefile` for streamlined development tasks like building, running, and hot-reloading with `air`.

## Getting Started

### Prerequisites

- Go (version 1.22 or higher recommended)
- Docker and Docker Compose
- A PostgreSQL client (e.g., `psql`)
- A local or remote proxy capable of handling the `request.get` format (e.g., [FlareSolverr](https://github.com/FlareSolverr/FlareSolverr))

### 1. Clone the Repository

```bash
git clone https://github.com/retconned/kick-monitor
cd kick-monitor
```

### 2. Configure Environment Variables

The application relies on environment variables for database connection and proxy URL. These are defined in `.env`.

### 3. Build and Run with Docker Compose (Recommended)

This sets up both the application and a PostgreSQL database in Docker containers.

```bash
# Build the Docker image and start the containers
make docker-run

# To stop and remove containers later
make docker-down
```

### 5. Running Locally

Ensure you have `air` installed (`go install github.com/air-verse/air@latest`).

```bash
make dev-air
```

If you wanna run it without `air`

```bash
make dev
```

## API Endpoints

Once the application is running (e.g., on `http://localhost:8080`):

- **`POST /add_channel`**
    - **Body:** `{"username": "xqc", "is_active": true}`
    - Adds or updates a channel in `monitored_channels`. If active, it starts monitoring API and WebSocket data.

## Contributing

Feel free to open issues or submit pull requests.
