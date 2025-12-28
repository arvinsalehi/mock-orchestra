# Mock Project

A distributed testing and device orchestration system built with FastAPI, Go, MQTT, and React. This project enables real-time session management, device command execution, and live status monitoring across multiple test targets.

## Architecture Overview

The project uses a microservices architecture with Docker Compose for orchestration:

- **Backend (Python/FastAPI)**: REST API and WebSocket server for session management
- **Frontend (React/TypeScript)**: Real-time dashboard for monitoring and test execution
- **MQTT Service (Go)**: Orchestrates device commands via MQTT protocol
- **Message Broker (Mosquitto)**: MQTT broker for device communication
- **Redis**: In-memory cache for session data
- **MongoDB**: Persistent storage for test results and device mappings

## Prerequisites

- Docker & Docker Compose
- Node.js 18+ (for local frontend development)
- Python 3.11+ (for local backend development)
- Go 1.21+ (for local MQTT service development)

## Quick Start

### Using Docker Compose

1. **Clone and navigate to project directory**:
   ```bash
   git clone https://github.com/arvinsalehi/mock-orchestra.git
   cd /path/to/mock
   ```

2. **Start all services**:
   ```bash
   docker-compose up --build
   ```

3. **Access services**:
   - Frontend: http://localhost:3000
   - Backend API: http://localhost:8000
   - API Docs: http://localhost:8000/docs
   - MQTT Broker: localhost:1883

### Local Development

#### Backend Setup

```bash
cd backend
python -m venv venv
source venv/bin/activate
pip install -r requirements.txt
uvicorn app.main:app --host 0.0.0.0 --port 8000 --reload
```

#### Frontend Setup

```bash
cd frontend
npm install
npm run dev
```

#### MQTT Service Setup

```bash
cd mqtt-service
go mod download
go run main.go
```

## Project Structure

```
.
├── docker-compose.yml          # Service orchestration
├── backend/                    # Python FastAPI application
│   ├── Dockerfile
│   ├── requirements.txt
│   └── app/
│       ├── main.py            # FastAPI app setup
│       ├── routes.py          # API endpoints
│       ├── models.py          # Data models
│       ├── database.py        # Database connection
│       ├── connection.py      # WebSocket manager
│       ├── mqtt_handler.py    # MQTT integration
│       └── config.py          # Configuration
├── frontend/                   # React TypeScript application
│   ├── Dockerfile
│   ├── package.json
│   ├── vite.config.ts
│   └── src/
│       ├── main.tsx           # React entry point
│       ├── App.tsx            # Main component
│       ├── api.ts             # API client
│       ├── types.ts           # TypeScript types
│       └── components/
│           ├── LoginForm.tsx
│           ├── SessionDashboard.tsx
│           └── TestTable.tsx
├── mqtt-service/               # Go MQTT orchestrator
│   ├── Dockerfile
│   ├── main.go                # Device orchestration logic
│   └── go.mod
├── mosquitto/                  # MQTT Broker configuration
│   └── config/
│       └── mosquitto.conf
└── scripts/                    # Utility scripts
    ├── auto_test.sh           # Automated testing
    └── mock_live_stream.sh     # Test data streaming
```

## Key Features

### 1. Session Management
- Create and manage test sessions via REST API
- Real-time session state tracking
- WebSocket support for live updates

### 2. Device Orchestration
- Target device groups by name (e.g., "production_line_1", "lab_bench_A")
- Send commands to multiple devices simultaneously
- Monitor device health and test progress

### 3. Real-time Communication
- MQTT for device-to-cloud messaging
- WebSocket for frontend-to-backend updates
- Publish/Subscribe pattern for scalability

### 4. Data Persistence
- MongoDB for long-term storage
- Redis for caching and session state
- Structured logging for debugging

## API Endpoints

### Health & Info
- `GET /api/health` - Service health check

### Session Management
- `POST /api/sessions` - Create a new test session
- `GET /api/sessions/{session_id}` - Get session details
- `GET /api/sessions` - List all sessions

### Device Management
- `GET /api/devices` - List all registered devices (Not implemented)
- `POST /api/devices/{device_id}/command` - Send command to device (Not implemented)

### WebSocket
- `WS /ws/{session_hash}` - Real-time session updates

## MQTT Topics

### Published Topics (From Devices)
- `devices/{device_id}/status` - Device heartbeats and status updates

### Subscribed Topics (To Devices)
- `hub/session/start` - Start test session commands
- `devices/{device_id}/command` - Device-specific commands

## Environment Variables

### Backend
```
REDIS_URL=redis://redis:6379
MONGO_URL=mongodb://mongo:27017
MQTT_HOST=mosquitto
MQTT_PORT=1883
```

### MQTT Service
```
MQTT_BROKER_URL=tcp://mosquitto:1883
```

### Frontend
```
VITE_API_URL=http://localhost:8000
```

## Dependencies

### Backend
- **fastapi** - Modern Python web framework
- **uvicorn** - ASGI server
- **fastapi-mqtt** - MQTT integration
- **motor** - Async MongoDB driver
- **redis** - Redis client
- **pydantic** - Data validation
- **websockets** - WebSocket support

### Frontend
- **react** - UI framework
- **react-dom** - React DOM rendering
- **@tanstack/react-query** - Data fetching
- **react-use-websocket** - WebSocket client
- **tailwindcss** - Styling
- **vite** - Build tool

### MQTT Service
- **paho.mqtt.golang** - MQTT client for Go

## Testing

### Prerequisites
Ensure Docker Compose services are running:
```bash
docker-compose up -d
```

### Automated Test Scripts

Run the provided testing scripts:

### 1. Auto-Create a Session (Database Init)
Instead of clicking through the UI, generate a valid session via the API:
```python
python3 scripts/init_session.py "Build-2025-RC1"
```

**Output**: Returns a Session UUID and the direct URL to the UI.
```bash
chmod +x ./scripts/test_ui_update.sh
./scripts/test_ui_update.sh <PASTE_SESSION_UUID_HERE>
```

**Expected Behavior**:
1. Row "Battery Check" appears as **Running** (Yellow).
2. Row "Battery Check" updates to **Passed** (Green).
3. Row "WiFi Connectivity" appears as **Failed** (Red).

### 2. Simulate Real-Time Test Results
Verify the React Table updates dynamically by injecting mock MQTT messages:

### Frontend Dev

#### Run Development Server
```bash
cd frontend
npm install
npm run dev
```

#### Build for Production
```bash
cd frontend
npm run build
```

### End-to-End Testing

1. **Start all services**:
   ```bash
   docker-compose up -d
   ```

2. **Open Frontend**:
   - Navigate to http://localhost:3000
   - Login with test credentials

3. **Create a Test Session**:
   - Use the SessionDashboard component
   - Select a target group

4. **Monitor Live Updates**:
   - Watch real-time device status
   - Verify WebSocket connection

5. **Verify Data Flow**:
   - Check backend logs: `docker-compose logs backend`
   - Monitor MQTT topics: `mosquitto_sub -h localhost -t "#"`
   - Verify database entries in MongoDB

### Performance Testing

#### Load Testing Backend
```bash
# Using Apache Bench
ab -n 1000 -c 10 http://localhost:8000/api/health

# Using wrk (if installed)
wrk -t4 -c100 -d30s http://localhost:8000/api/health
```

### Debugging

#### View Container Logs
```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f backend
docker-compose logs -f mqtt-service
docker-compose logs -f mosquitto
```

#### Interactive Shell in Container
```bash
# Backend container
docker-compose exec backend bash

# MQTT service container
docker-compose exec mqtt-service /bin/sh

# MongoDB
docker-compose exec mongo mongosh
```

#### Check Service Health
```bash
# View running containers
docker-compose ps

# Inspect service network
docker-compose exec backend ping redis
docker-compose exec backend ping mongo
docker-compose exec backend ping mosquitto
```


## Troubleshooting

### Connection Issues
- Verify all containers are running: `docker-compose ps`
- Check network: `docker-compose logs`
- Ensure MQTT broker is accessible

### Backend Errors
- Check logs: `docker-compose logs backend`
- Verify database connections
- Check environment variables

### Frontend Issues
- Clear browser cache and reinstall dependencies
- Check API URL configuration
- Verify WebSocket connection
