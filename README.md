# Poker Web Application

A real-time multiplayer poker web application built with Go and React. Features WebSocket communication for live gameplay, hot reload development, Docker support, and comprehensive testing.

## Features

- **Real-time Communication** - WebSocket-based bidirectional messaging
- **Live Gameplay** - Play poker with other players in real-time
- **Hot Reload Development** - Automatic code recompilation on changes
- **Docker Support** - Full Docker and docker-compose setup for consistent development
- **Comprehensive Testing** - 40+ tests (Go backend + React frontend)
- **Type-Safe** - TypeScript frontend and Go backend with strong typing
- **Structured Logging** - JSON-formatted structured logs for debugging and monitoring
- **Production Ready** - Multi-stage Docker builds, proper error handling, graceful shutdown

## Tech Stack

### Backend
- **Go 1.24** - High-performance backend
- **Chi Router** - Lightweight HTTP routing
- **Gorilla WebSocket** - WebSocket communication
- **log/slog** - Structured logging

### Frontend
- **React 19** - Modern UI framework
- **TypeScript** - Type-safe JavaScript
- **Vite** - Lightning-fast build tool and dev server
- **Vitest + React Testing Library** - Testing framework
- **CSS Modules** - Component-scoped styling

### DevOps
- **Docker 20.10+** - Containerization
- **docker-compose 2.0+** - Multi-container orchestration
- **Air** - Go hot reload
- **ESLint + Prettier** - Code quality and formatting

## Prerequisites

### Option A: Docker (Recommended)

```bash
Docker 20.10+
docker-compose 2.0+
```

No need to install Go or Node.js locally.

### Option B: Local Development

```bash
Go 1.24+
Node.js 20+ (with npm)
Git
```

## Quick Start

### Using Docker (Recommended)

```bash
# Clone the repository
git clone https://github.com/robinr2/poker.git
cd poker

# Start the development environment with hot reload
make docker-dev

# Access the application
# Backend: http://localhost:8080
# Frontend: http://localhost:5173
```

The Docker environment includes:
- Go backend with hot reload via Air
- React frontend with Vite hot reload
- Pre-configured networking and volumes

**Stop the environment:**
```bash
make docker-down
```

### Local Development

**Terminal 1 - Start Backend:**
```bash
make dev-backend
# or
go run cmd/server/main.go
```

Backend runs on `http://localhost:8080`

**Terminal 2 - Start Frontend:**
```bash
make dev-frontend
# or
cd frontend && npm run dev
```

Frontend runs on `http://localhost:5173`

**Install dependencies (one-time):**
```bash
make install-tools
```

## Project Structure

```
poker/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── frontend/                     # React frontend (separate npm project)
│   ├── src/
│   │   ├── App.tsx              # Main App component
│   │   ├── main.tsx             # Entry point
│   │   ├── hooks/
│   │   │   └── useWebSocket.ts  # WebSocket hook
│   │   └── services/
│   │       └── WebSocketService.ts  # WebSocket client
│   └── package.json             # Frontend dependencies
├── internal/
│   └── server/
│       ├── server.go            # HTTP server setup
│       ├── handlers.go          # HTTP request handlers
│       └── websocket.go         # WebSocket management
├── docs/
│   ├── ARCHITECTURE.md          # System architecture
│   ├── DEVELOPMENT.md           # Development guide
│   └── WEBSOCKET_PROTOCOL.md    # WebSocket API reference
├── scripts/
│   ├── test.sh                  # Run all tests
│   ├── lint.sh                  # Run linters
│   ├── build.sh                 # Build frontend and backend
│   └── docker-dev.sh            # Docker development setup
├── Makefile                     # Build automation
├── docker-compose.yml           # Development environment
├── Dockerfile                   # Production image
├── .env.example                 # Environment variables template
└── README.md                    # This file
```

## Development

### Available Commands

```bash
# Development
make dev-backend       # Run backend server
make dev-frontend      # Run frontend dev server

# Building
make build            # Build both frontend and backend
make build-backend    # Build backend binary to bin/poker
make build-frontend   # Build frontend to frontend/dist/

# Testing
make test             # Run all tests
make test-integration # Run integration tests (validates setup)

# Linting and Formatting
make lint             # Run all linters
make lint-fix         # Fix linting issues
make format           # Format code with Prettier
make format-check     # Check formatting without changes

# Docker
make docker-dev       # Start Docker development environment
make docker-down      # Stop Docker environment
make docker-build     # Build production Docker image
make docker-test      # Run tests in Docker
make docker-clean     # Clean Docker resources
make docker-logs      # View Docker logs

# Cleanup
make clean            # Remove build artifacts
make install-tools    # Install development dependencies

# Help
make help             # Display all available commands
```

### Testing

**Run all tests:**
```bash
make test
```

**Backend tests only:**
```bash
go test ./internal/... ./cmd/... -v
go test ./internal/... ./cmd/... -race  # With race detector
```

**Frontend tests only:**
```bash
cd frontend && npm test
cd frontend && npm test -- --watch  # Watch mode
cd frontend && npm test -- --ui     # UI mode
```

**Integration tests:**
```bash
./scripts/test-integration.sh
```

**Test coverage:**
```bash
go test ./internal/... -cover
cd frontend && npm test -- --coverage
```

### Linting and Formatting

**Check all code quality:**
```bash
make lint
```

**Auto-fix issues:**
```bash
make lint-fix
make format
```

**Check without fixing:**
```bash
make format-check
```

## Environment Variables

### Configuration

Create a `.env` file (copy from `.env.example`):

```bash
cp .env.example .env
```

**Backend Variables:**
```bash
PORT=8080                    # Server port (default: 8080)
LOG_LEVEL=info              # Log level: debug, info, warn, error (default: info)
```

**Frontend Variables:**
```bash
NODE_ENV=development        # Environment: development, production
VITE_WS_URL=ws://localhost:8080/ws  # WebSocket URL
```

## Building

### Frontend Build

```bash
cd frontend
npm run build

# Output: frontend/dist/
```

### Backend Build

```bash
make build-backend

# Output: bin/poker
```

### Production Docker Build

```bash
make docker-build

# Creates: poker:latest image
# Run with: docker run -p 8080:8080 poker:latest
```

## Docker

### Development Environment

```bash
# Start with hot reload
make docker-dev

# Access services
# Backend: http://localhost:8080
# Frontend: http://localhost:5173

# View logs
make docker-logs

# Stop
make docker-down
```

### Production Build

```bash
# Build image
make docker-build

# Run container
docker run -p 8080:8080 -e PORT=8080 -e LOG_LEVEL=info poker:latest
```

## Architecture

For detailed system architecture, see [docs/ARCHITECTURE.md](./docs/ARCHITECTURE.md).

### High-Level Overview

```
┌─────────────────────┐
│   React Frontend    │
│  (React 19 + TS)    │
└──────────┬──────────┘
           │ WebSocket
           │ ws://localhost:8080/ws
           │
┌──────────▼──────────┐
│   Go Backend        │
│  HTTP + WebSocket   │
│  Chi Router         │
└─────────────────────┘
```

- **Frontend**: React SPA served by backend, communicates via WebSocket
- **Backend**: HTTP server with WebSocket support, manages client connections
- **Communication**: JSON messages over WebSocket protocol
- **State**: In-memory (ready for database integration)

## WebSocket API

The application communicates via WebSocket at `ws://localhost:8080/ws`.

**Message Format:**
```json
{
  "type": "message_type",
  "payload": {}
}
```

**Basic Messages:**
- `ping` / `pong` - Heartbeat messages
- `error` - Error notifications

**Future Game Messages:**
- `join_game` - Join a game room
- `place_bet` - Place a bet
- `game_state` - Game state updates

For complete documentation, see [docs/WEBSOCKET_PROTOCOL.md](./docs/WEBSOCKET_PROTOCOL.md).

## Development Workflow

### Recommended Workflow

1. **Start development environment:**
   ```bash
   make docker-dev
   # or locally
   make dev-backend    # Terminal 1
   make dev-frontend   # Terminal 2
   ```

2. **Create feature branch:**
   ```bash
   git checkout -b feature/feature-name
   ```

3. **Write code and tests:**
   - Tests guide development (TDD)
   - Follow code style guidelines
   - Keep changes focused

4. **Verify quality:**
   ```bash
   make test
   make lint
   ```

5. **Commit and push:**
   ```bash
   git add .
   git commit -m "feat: describe your feature"
   git push origin feature/feature-name
   ```

6. **Create pull request**

For detailed guidelines, see [CONTRIBUTING.md](./CONTRIBUTING.md).

## Code Style

### Backend (Go)

```bash
go fmt ./internal/... ./cmd/...
go vet ./internal/... ./cmd/...
```

**Guidelines:**
- Use meaningful names
- Add comments for exported functions
- Handle errors explicitly
- Keep functions small and focused

### Frontend (TypeScript/React)

```bash
cd frontend && npm run lint:fix
cd frontend && npm run format
```

**Guidelines:**
- Use TypeScript for type safety
- Functional components with hooks
- One component per file
- Use CSS modules for styling

## Troubleshooting

### Common Issues

**Port 8080/5173 already in use:**
```bash
# Find and kill process on port 8080
lsof -i :8080 | grep -v PID | awk '{print $2}' | xargs kill -9

# Change port
PORT=8081 go run cmd/server/main.go
```

**Docker container won't start:**
```bash
# Check logs
make docker-logs

# Rebuild
docker-compose up --build --force-recreate
```

**Node modules issues:**
```bash
cd frontend
rm -rf node_modules package-lock.json
npm install
```

**Go module errors:**
```bash
go mod tidy
go mod download
```

See [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md) for more troubleshooting tips.

## Testing Checklist

Before submitting changes:

```bash
# Run all tests
make test

# Check linting
make lint

# Check formatting
make format-check

# Build successfully
make build

# Try Docker build
make docker-build
```

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for:
- How to report bugs
- How to request features
- Development process
- Code style guidelines
- Pull request process

**Quick start for contributors:**

```bash
# Fork and clone repository
git clone https://github.com/YOUR-USERNAME/poker.git
cd poker

# Create feature branch
git checkout -b feature/your-feature

# Make changes, run tests
make test && make lint

# Commit and push
git commit -m "feat: describe your changes"
git push origin feature/your-feature

# Create pull request on GitHub
```

## License

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## Project Status

**Current Phase:** Phase 6 - Documentation & Environment Setup (Complete)

**Completed:**
- ✅ Phase 1: Bootstrap & Tooling
- ✅ Phase 2: Frontend & Backend Setup
- ✅ Phase 3: WebSocket Communication
- ✅ Phase 4: Testing & Linting
- ✅ Phase 5: Docker & Deployment
- ✅ Phase 6: Documentation & Environment

**Future Phases:**
- Game logic and rules engine
- Player authentication and profiles
- Database integration (PostgreSQL)
- Lobby and room management
- Advanced gameplay features

## Support

- **Documentation**: Check [docs/](./docs/) folder
- **Issues**: Create an issue on GitHub
- **Discussions**: Use GitHub Discussions
- **Development Guide**: See [docs/DEVELOPMENT.md](./docs/DEVELOPMENT.md)

## Acknowledgments

Built with:
- Go and the amazing Go community
- React and the React community
- All open-source contributors whose tools make this possible

## Quick Links

- [Architecture Overview](./docs/ARCHITECTURE.md)
- [Development Guide](./docs/DEVELOPMENT.md)
- [WebSocket Protocol](./docs/WEBSOCKET_PROTOCOL.md)
- [Contributing Guidelines](./CONTRIBUTING.md)
- [Environment Template](./.env.example)

---

**Happy coding! 🎮**
