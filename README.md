# Counting Dashboard - Multi-Architecture Microservices

A simple microservices application demonstrating service-to-service communication with support for **AMD64**, **ARM64**, **Darwin**, and **Windows** architectures. Created with IBM Bob, based on the [HashiCorp demo-consul-101 counting service example application](https://github.com/hashicorp/demo-consul-101).

## 🏗️ Architecture

```
┌─────────────────────┐         ┌─────────────────────┐
│                     │         │                     │
│  Dashboard Service  │────────▶│  Counting Service   │
│   (Nginx/HTML/JS)   │  HTTP   │      (Go)           │
│    Port: 8080       │         │    Port: 9001       │
│                     │         │                     │
└─────────────────────┘         └─────────────────────┘
```

## 📦 Components

### Counting Service (Backend)
- **Technology**: Go (Golang)
- **Port**: 9001
- **Features**:
  - Simple counter API that increments on each request
  - Health check endpoint
  - Returns system information (architecture, platform, Go version)
  - Thread-safe counter using mutex
  - Minimal Docker image using scratch base

### Dashboard Service (Frontend)
- **Technology**: HTML, CSS, JavaScript with Nginx
- **Port**: 8080
- **Features**:
  - Modern, responsive UI
  - Real-time count display
  - System information display
  - Auto-refresh capabilities
  - Health monitoring

## 🚀 Quick Start

### Prerequisites
- Docker installed
- Docker Compose installed (optional, for easier deployment)

### Option 1: Using Docker Compose (Recommended)

```bash
# Navigate to the counting-dashboard directory
cd counting-dashboard

# Build and start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop all services
docker-compose down
```

Access the application:
- **Dashboard**: http://localhost:8080
- **Counting API**: http://localhost:9001
- **Health Check**: http://localhost:9001/health

### Option 2: Using Docker Commands

```bash
# Create a network
docker network create counting-network

# Build and run counting service
cd counting-service
docker build -t counting-service:latest .
docker run -d \
  --name counting-service \
  --network counting-network \
  -p 9001:9001 \
  counting-service:latest

# Build and run dashboard service
cd ../dashboard-service
docker build -t dashboard-service:latest .
docker run -d \
  --name dashboard-service \
  --network counting-network \
  -p 8080:8080 \
  dashboard-service:latest
```

## 🏗️ Building Multi-Architecture Images

Both services support multi-architecture builds for AMD64, ARM64, Darwin, and Windows.

### Setup Docker Buildx

```bash
# Create a new builder instance
docker buildx create --name multiarch --use

# Verify the builder
docker buildx inspect --bootstrap
```

### Build Multi-Arch Images

```bash
# Build counting service for multiple platforms
cd counting-service
docker buildx build \
  --platform linux/amd64,linux/arm64,windows/amd64 \
  -t your-registry/counting-service:latest \
  --push \
  .

# Build dashboard service for multiple platforms
cd ../dashboard-service
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t your-registry/dashboard-service:latest \
  --push \
  .
```

### Build for Specific Platform

```bash
# For ARM64 (Apple Silicon, Raspberry Pi, etc.)
docker buildx build --platform linux/arm64 -t counting-service:arm64 .

# For AMD64 (Intel/AMD processors)
docker buildx build --platform linux/amd64 -t counting-service:amd64 .

# For Windows
docker buildx build --platform windows/amd64 -t counting-service:windows .
```

## 📡 API Endpoints

### Counting Service

#### Get Count (Increment)
```
GET /
```
Increments the counter and returns the current count with system information.

**Response:**
```json
{
  "count": 42,
  "timestamp": "2024-01-15T10:30:00Z",
  "architecture": "arm64",
  "platform": "linux",
  "go_version": "go1.21.0"
}
```

#### Health Check
```
GET /health
```
Returns service health status and system information.

**Response:**
```json
{
  "status": "healthy",
  "timestamp": "2024-01-15T10:30:00Z",
  "architecture": "arm64",
  "platform": "linux",
  "go_version": "go1.21.0"
}
```

### Dashboard Service

#### Main Dashboard
```
GET /
```
Serves the HTML dashboard interface.

#### Health Check
```
GET /health
```
Returns "healthy" status for the dashboard service.

## 🧪 Testing

### Test Counting Service Directly

```bash
# Health check
curl http://localhost:9001/health

# Increment counter
curl http://localhost:9001

# Multiple increments
for i in {1..10}; do curl http://localhost:9001; echo; done
```

### Test Dashboard

Open your browser and navigate to:
```
http://localhost:8080
```

Click the "Increment Count" button to increment the counter, or "Refresh" to check service health.

## 🔍 Monitoring

### View Container Logs

```bash
# Using Docker Compose
docker-compose logs -f counting-service
docker-compose logs -f dashboard-service

# Using Docker
docker logs -f counting-service
docker logs -f dashboard-service
```

### Check Container Health

```bash
# Using Docker Compose
docker-compose ps

# Using Docker
docker ps
docker inspect counting-service --format='{{.State.Health.Status}}'
docker inspect dashboard-service --format='{{.State.Health.Status}}'
```

## 🛠️ Development

### Counting Service Development

```bash
cd counting-service

# Run locally (requires Go 1.21+)
go run main.go

# Build binary
go build -o counting-service

# Run tests (if you add them)
go test ./...
```

### Dashboard Service Development

For dashboard development, you can use any local web server:

```bash
cd dashboard-service

# Using Python
python3 -m http.server 8080

# Using Node.js http-server
npx http-server -p 8080

# Or just open index.html in a browser
# Note: You'll need to update the COUNTING_SERVICE_URL in index.html
```

## 📁 Project Structure

```
counting-dashboard/
├── counting-service/
│   ├── Dockerfile
│   ├── go.mod
│   └── main.go
├── dashboard-service/
│   ├── Dockerfile
│   ├── nginx.conf
│   └── index.html
├── docker-compose.yml
└── README.md
```

## 🔧 Configuration

### Counting Service Environment Variables

- `PORT`: Server port (default: 9001)

### Dashboard Service Configuration

The dashboard automatically detects the counting service URL:
- **Localhost**: Uses `http://localhost:9001`
- **Docker**: Uses `http://counting-service:9001`
- **Custom**: Set `window.COUNTING_SERVICE_URL` in the HTML

## 🐛 Troubleshooting

### Dashboard can't connect to counting service

1. Ensure both containers are on the same network:
```bash
docker network inspect counting-network
```

2. Check if counting service is running:
```bash
docker ps | grep counting-service
curl http://localhost:9001/health
```

3. Check browser console for errors (F12 in most browsers)

### Port already in use

Change the port mapping in `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # Change 8080 to 8081
```

### Architecture mismatch

Verify your system architecture:
```bash
docker version | grep Arch
uname -m
```

Build for your specific architecture:
```bash
docker build --platform linux/arm64 -t counting-service:latest ./counting-service
```

### Go module issues

If you encounter Go module errors:
```bash
cd counting-service
go mod tidy
go mod download
```

## 🎯 Deployment Options

This application can be deployed in multiple ways:

1. **Docker Compose** - Local development and testing (this guide)
2. **HashiCorp Nomad** - Production orchestration
3. **Kubernetes** - Alternative orchestration platform
4. **Docker Swarm** - Docker-native orchestration

### Example Nomad Job Spec

See the parent directory for Nomad job specifications that deploy this application with:
- Multi-architecture support
- Service discovery (Consul DNS or Nomad native)
- Health checks
- Rolling updates
- High availability

## 🌟 Features

- ✅ Multi-architecture support (AMD64, ARM64, Windows)
- ✅ Minimal Docker images (Go service uses scratch base)
- ✅ Health checks for both services
- ✅ Modern, responsive UI
- ✅ Real-time updates
- ✅ System information display
- ✅ CORS enabled for development
- ✅ Production-ready Nginx configuration
- ✅ Thread-safe counter implementation

## 📝 License

MIT License - feel free to use this project for learning and development.

## 🤝 Contributing

Contributions are welcome! Feel free to submit issues and pull requests.

## 📚 Additional Resources

- [Docker Documentation](https://docs.docker.com/)
- [Docker Buildx Documentation](https://docs.docker.com/buildx/working-with-buildx/)
- [Go Documentation](https://golang.org/doc/)
- [Nginx Documentation](https://nginx.org/en/docs/)
- [HashiCorp Consul Demo](https://github.com/hashicorp/demo-consul-101)

## 🎓 Learning Objectives

This project demonstrates:
- Microservices architecture
- Service-to-service communication
- Multi-architecture Docker builds
- Go web service development
- Frontend-backend integration
- Container networking
- Health checks and monitoring
- Production-ready configurations

---

Built with ❤️ for multi-architecture deployment
