# CLAUDE.md

请忽略 ./data 文件，这个是我分离出来打包的内容。
src/web 是通过 src/ui 打包出来的内容，为了本地启动方便。你忽略/src/web中的内容。
调试需要在 环境中增加全局配置：export CGO_CPPFLAGS=-I/opt/homebrew/include;CGO_LDFLAGS=-L/opt/homebrew/lib -lssl -lcrypto

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

BlueKing CMDB (蓝鲸配置平台) is an enterprise-level configuration management platform for assets and applications. It's a distributed microservice system built with Go backend services and Vue.js frontend.

## Build and Development Commands

### Building the Project

```bash
# Build entire project (backend + frontend)
cd src && make

# Build for Linux (cross-compile)
cd src && make linux

# Build only backend services
cd src && make server

# Build only frontend
cd src && make ui

# Build debug version
cd src && make debug
```

### Frontend Development

```bash
# Install dependencies
cd src/ui && npm install

# Development server
cd src/ui && npm run dev

# Production build
cd src/ui && npm run build

# Linting
cd src/ui && npm run lint
cd src/ui && npm run lint-fix
```

### Testing

```bash
# Run Go tests with coverage
cd scripts && bash gotest.sh

# Run individual Go tests
go test -v ./src/...
```

### Other Commands

```bash
# Package for distribution
cd src && make package

# Clean build artifacts
cd src && make clean
cd src && make cleanall

# Enterprise packaging
cd src && make enterprise
```

## Architecture Overview

The system follows a layered microservice architecture:

1. **Resource Layer**: Storage (MongoDB), message queues, and caching systems
2. **Service Layer**: Split into two modules:
   - **Resource Management**: Atomic interface services for different resource types
   - **Business Scenario**: Application-specific services built on resource management
3. **API Layer**: API gateway service (apiserver)
4. **Web Layer**: Web interface service (webserver)

### Key Microservices

- **adminserver**: System configuration refresh and initialization
- **authserver**: Authentication and authorization
- **cloudserver**: Cloud resource management
- **coreservice**: Core atomic operations for resources
- **eventserver**: Event subscription and publishing
- **hostserver**: Host management
- **procserver**: Process management
- **toposerver**: Topology and model management
- **webserver**: Web interface and static assets

## Code Structure

### Main Directories

- `src/`: All source code
  - `apiserver/`: API gateway service
  - `scene_server/`: Business scenario microservices
  - `source_controller/`: Resource management services (coreservice, cacheservice)
  - `web_server/`: Web interface service
  - `ui/`: Vue.js frontend application
  - `common/`: Shared libraries and utilities
  - `apimachinery/`: API clients and abstractions
  - `storage/`: Database and storage abstractions
  - `thirdparty/`: Third-party service integrations

### Key Technologies

- **Backend**: Go 1.20, go-restful framework, Gin web framework
- **Frontend**: Vue.js 2.7, Webpack, SCSS
- **Database**: MongoDB (with mgo driver)
- **Service Discovery**: Zookeeper
- **Message Queue**: Kafka
- **Cache**: Redis

## Development Workflow

1. Each microservice has its own Makefile in its directory
2. Services are built using the main build script `scripts/build.sh`
3. Configuration is managed through Zookeeper
4. Service discovery is handled via Zookeeper
5. All services support graceful shutdown and health checks

## Important Notes

- The project uses Go modules (go.mod at root level)
- Build requires Python for configuration generation
- Frontend requires Node.js >= 18.18.0 and npm >= 9.8.1
- Services communicate via REST APIs
- All services support horizontal scaling
- Database migrations are handled through admin scripts