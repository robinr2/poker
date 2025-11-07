## Plan: Bootstrap 2D Poker Web Application

Bootstrap a complete, production-ready poker application with Go backend, React frontend, Docker containerization, and hot reload for optimal AI-assisted development workflow.

**Phases: 6 phases**

1. **Phase 1: Backend Foundation**
    - **Objective:** Set up Go project structure, dependencies, and core server scaffolding
    - **Files/Functions to Create:**
      - `go.mod`, `go.sum` - Go module configuration (github.com/robinr2/poker) with gorilla/websocket, go-chi/chi
      - `cmd/server/main.go` - Application entry point with HTTP server, structured logging (log/slog)
      - `internal/server/server.go` - HTTP server struct with WebSocket upgrade support
      - `internal/server/websocket.go` - WebSocket connection manager and client registry
      - `internal/server/handlers.go` - HTTP handlers for health check and API endpoints
      - `.gitignore` - Ignore patterns for Go binaries, build artifacts, node_modules, Docker volumes
    - **Tests to Write:**
      - `internal/server/server_test.go` - Test server initialization and graceful shutdown
      - `internal/server/websocket_test.go` - Test WebSocket connection/disconnection handling
      - `internal/server/handlers_test.go` - Test HTTP handlers
    - **Steps:**
        1. Write tests for server initialization, WebSocket handling, and HTTP handlers (should fail)
        2. Initialize go.mod with module path github.com/robinr2/poker
        3. Add required dependencies (gorilla/websocket, go-chi/chi)
        4. Create directory structure (cmd/server/, internal/server/)
        5. Implement HTTP server with chi router, WebSocket upgrade capability, structured logging
        6. Implement health check endpoint
        7. Run tests to verify they pass

2. **Phase 2: Frontend Foundation**
    - **Objective:** Set up React + Vite + TypeScript project with WebSocket client
    - **Files/Functions to Create:**
      - `frontend/package.json` - npm dependencies (React 19, Vite 5, TypeScript, React Testing Library)
      - `frontend/vite.config.ts` - Vite config with proxy to backend (port 8080), build output to ../web/static/
      - `frontend/tsconfig.json` - TypeScript strict mode configuration
      - `frontend/tsconfig.node.json` - TypeScript config for Vite config file
      - `frontend/index.html` - HTML entry point
      - `frontend/src/main.tsx` - React application entry point
      - `frontend/src/App.tsx` - Root component with connection status
      - `frontend/src/services/WebSocketService.ts` - WebSocket client with auto-reconnect
      - `frontend/src/hooks/useWebSocket.ts` - React hook for WebSocket connection management
      - `frontend/src/vite-env.d.ts` - Vite type definitions
    - **Tests to Write:**
      - `frontend/src/services/WebSocketService.test.ts` - Test WebSocket connect/disconnect/reconnect
      - `frontend/src/hooks/useWebSocket.test.ts` - Test React hook behavior
    - **Steps:**
        1. Write tests for WebSocket service and hook (should fail)
        2. Initialize npm project with `npm create vite@latest frontend -- --template react-ts`
        3. Install additional dependencies (React Testing Library, vitest)
        4. Configure Vite to build to ../web/static/ and proxy /ws and /api to localhost:8080
        5. Configure TypeScript with strict mode
        6. Implement WebSocket service with connection management and auto-reconnect
        7. Implement useWebSocket hook
        8. Create minimal App component showing connection status
        9. Run tests to verify they pass

3. **Phase 3: Integration & Static Serving**
    - **Objective:** Connect frontend and backend, configure static file serving, create build tooling
    - **Files/Functions to Modify/Create:**
      - `internal/server/server.go` - Add static file serving middleware for /web/static/
      - `cmd/server/main.go` - Add environment variable configuration (PORT, LOG_LEVEL)
      - `Makefile` - Build automation: dev, build, test, clean, install-tools
      - `web/.gitkeep` - Placeholder for web directory
      - `scripts/build.sh` - Production build script (frontend + backend)
    - **Tests to Write:**
      - `internal/server/static_test.go` - Test static file serving
      - Integration test: backend serves index.html at root
    - **Steps:**
        1. Write tests for static file serving (should fail)
        2. Create web/ directory structure
        3. Modify server.go to serve static files with SPA fallback to index.html
        4. Add environment variable parsing in main.go (PORT defaults to 8080)
        5. Create Makefile with targets: dev-backend, dev-frontend, build, test, clean
        6. Create build.sh script that builds frontend then backend
        7. Run tests to verify static serving works

4. **Phase 4: Docker Containerization**
    - **Objective:** Create Docker setup for consistent development and deployment
    - **Files/Functions to Create:**
      - `Dockerfile` - Multi-stage build (build stage + runtime stage)
      - `docker-compose.yml` - Services for backend with hot reload, frontend dev server
      - `.dockerignore` - Ignore files for Docker builds
      - `.air.toml` - Air configuration for Go hot reload in container
      - `scripts/docker-dev.sh` - Helper script to start Docker environment
    - **Tests to Write:**
      - Test that `docker-compose up` starts both services
      - Test hot reload works for both backend and frontend
    - **Steps:**
        1. Write tests for Docker setup (should fail - verify services start and hot reload works)
        2. Create Dockerfile with multi-stage build (Go build + minimal runtime)
        3. Create docker-compose.yml with backend (with volume mounts for hot reload) and frontend services
        4. Create .dockerignore to exclude unnecessary files
        5. Install and configure air for Go hot reload
        6. Create .air.toml with proper configuration for Docker
        7. Test docker-compose up starts both services
        8. Verify hot reload: change backend code, should rebuild automatically
        9. Verify hot reload: change frontend code, should reflect immediately
        10. Run tests to verify everything works

5. **Phase 5: Development Tooling**
    - **Objective:** Set up linting, formatting, and testing infrastructure
    - **Files/Functions to Create:**
      - `frontend/.eslintrc.json` - ESLint configuration with TypeScript and React rules
      - `frontend/.prettierrc` - Prettier code formatting rules
      - `frontend/vitest.config.ts` - Vitest test configuration
      - `.editorconfig` - Editor configuration for consistent formatting
      - `scripts/lint.sh` - Run all linters (Go + Frontend)
      - `scripts/test.sh` - Run all tests (Go + Frontend)
    - **Tests to Write:**
      - Test that lint scripts execute without errors on clean code
      - Test that test scripts execute and report results
    - **Steps:**
        1. Write tests that lint and test scripts work (should fail initially)
        2. Add ESLint with TypeScript and React plugins to frontend
        3. Add Prettier to frontend
        4. Configure Vitest for frontend testing
        5. Create .editorconfig for consistent editor settings
        6. Create lint.sh script that runs gofmt, go vet, staticcheck, eslint
        7. Create test.sh script that runs Go tests and Vitest
        8. Add npm scripts for lint and format in package.json
        9. Run tests to verify all tooling works

6. **Phase 6: Documentation & Environment Setup**
    - **Objective:** Create comprehensive documentation and environment templates
    - **Files/Functions to Create:**
      - `README.md` - Project overview, quick start, development guide, architecture
      - `.env.example` - Environment variable template
      - `docs/ARCHITECTURE.md` - System architecture and design decisions
      - `docs/DEVELOPMENT.md` - Development workflow and best practices
      - `docs/WEBSOCKET_PROTOCOL.md` - WebSocket message format documentation
      - `CONTRIBUTING.md` - Contributing guidelines
    - **Tests to Write:**
      - Validation checklist: all README commands execute successfully
    - **Steps:**
        1. Create validation checklist for all setup commands in README (should pass when docs are correct)
        2. Write README with: project description, prerequisites, quick start (Docker), local development, testing, building
        3. Create .env.example with PORT, LOG_LEVEL, etc.
        4. Document architecture: Go backend with WebSocket, React frontend, in-memory state
        5. Document development workflow: Docker for dev, hot reload, testing, linting
        6. Document WebSocket protocol basics (ready for game features)
        7. Create contributing guidelines
        8. Execute all README commands to verify they work
        9. Run validation checklist to confirm documentation is accurate

**Key Decisions:**
- Module path: github.com/robinr2/poker
- Ports: Backend 8080, Frontend dev 5173
- Testing: Standard Go testing + React Testing Library + Vitest
- Hot Reload: Yes, using air for Go backend
- Docker: Yes, full Docker Compose setup with hot reload support
- Environment: .env support for flexible configuration
