# WebRTC Signaling Server

A scalable WebRTC signaling server implementation in Go with STUN/TURN support, designed for high-performance real-time communication applications.

## Features

- **WebSocket-based Signaling**: Real-time bidirectional communication
- **Multi-room Support**: Users can join different rooms with up to 10 participants each
- **Session Management**: Cookie-based user identification
- **STUN/TURN Server**: Integrated coturn for NAT traversal
- **Redis Integration**: Distributed state management for horizontal scaling
- **Kubernetes Ready**: Complete K8s deployment configurations
- **Docker Support**: Containerized deployment with Docker Compose
- **Auto-scaling**: Horizontal Pod Autoscaler configuration
- **Health Checks**: Built-in health and readiness endpoints
- **Load Balancing**: Session affinity for WebSocket connections

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Client 1      │    │   Client 2      │    │   Client N      │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │    Load Balancer          │
                    └─────────────┬─────────────┘
                                 │
          ┌──────────────────────┼──────────────────────┐
          │                      │                      │
┌─────────┴───────┐    ┌─────────┴───────┐    ┌─────────┴───────┐
│ Signaling Pod 1 │    │ Signaling Pod 2 │    │ Signaling Pod N │
└─────────┬───────┘    └─────────┬───────┘    └─────────┬───────┘
          │                      │                      │
          └──────────────────────┼──────────────────────┘
                                 │
                    ┌─────────────┴─────────────┐
                    │      Redis Cluster        │
                    └───────────────────────────┘

                    ┌───────────────────────────┐
                    │    STUN/TURN Server       │
                    │       (coturn)            │
                    └───────────────────────────┘
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Kubernetes cluster (optional)
- Make (optional, for convenience commands)

### Development Setup

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd signaling
   ```

2. **Set up development environment**
   ```bash
   make dev-setup
   # or
   ./scripts/dev-setup.sh
   ```

3. **Start development environment**
   ```bash
   make dev
   # or
   cd deployments/docker-compose && docker-compose up --build
   ```

4. **Access the application**
   - Open http://localhost:8080 in your browser
   - Try opening multiple tabs to test multi-user functionality

### Local Development (without Docker)

1. **Start Redis**
   ```bash
   docker run -d -p 6379:6379 redis:7-alpine
   ```

2. **Run the signaling server**
   ```bash
   make run
   # or
   go run ./cmd/signaling
   ```

## Deployment

### Docker Compose (Recommended for Development)

```bash
# Start all services
make dev

# View logs
make docker-compose-logs

# Stop services
make stop
```

### Kubernetes (Production)

1. **Build and deploy**
   ```bash
   make full-deploy
   ```

2. **Or step by step**
   ```bash
   # Build Docker images
   make docker-build

   # Deploy to Kubernetes
   make k8s-deploy
   ```

3. **Check deployment status**
   ```bash
   make k8s-status
   ```

4. **Access via port forwarding**
   ```bash
   make k8s-port-forward
   ```

### Minikube (Local Kubernetes)

```bash
# Start minikube
make minikube-start

# Build and deploy
make minikube-deploy

# Access the service
make minikube-service
```

## Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `SERVER_HOST` | `0.0.0.0` | Server bind address |
| `SERVER_PORT` | `8080` | Server port |
| `REDIS_HOST` | `localhost` | Redis host |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | `` | Redis password |
| `REDIS_DB` | `0` | Redis database number |
| `STUN_URL` | `stun:localhost:3478` | STUN server URL |
| `TURN_URL` | `turn:localhost:3478` | TURN server URL |
| `READ_TIMEOUT` | `60` | WebSocket read timeout (seconds) |
| `WRITE_TIMEOUT` | `60` | WebSocket write timeout (seconds) |

### STUN/TURN Configuration

The coturn server is configured with:
- **STUN Port**: 3478 (UDP/TCP)
- **TURN Port**: 3478 (UDP/TCP)
- **TURNS Port**: 5349 (UDP/TCP)
- **Default Credentials**: `webrtc:password123`

For production, update the credentials in:
- `deployments/docker/coturn/turnserver.conf`
- `deployments/kubernetes/secret.yaml`

## API Reference

### WebSocket Endpoints

- **`/ws`**: Main WebSocket endpoint for signaling

### HTTP Endpoints

- **`GET /health`**: Health check endpoint
- **`GET /ready`**: Readiness check endpoint
- **`GET /`**: Static file server (test interface)

### WebSocket Message Types

#### Client to Server

```json
// Join a room
{
  "type": "join_room",
  "data": "{\"room_id\": \"room-123\"}"
}

// Leave current room
{
  "type": "leave_room"
}

// WebRTC Offer
{
  "type": "offer",
  "target_id": "user-456",
  "data": "{\"sdp\": \"...\", \"type\": \"offer\"}"
}

// WebRTC Answer
{
  "type": "answer",
  "target_id": "user-456",
  "data": "{\"sdp\": \"...\", \"type\": \"answer\"}"
}

// ICE Candidate
{
  "type": "ice_candidate",
  "target_id": "user-456",
  "data": "{\"candidate\": \"...\", \"sdpMid\": \"...\", \"sdpMLineIndex\": 0}"
}
```

#### Server to Client

```json
// STUN/TURN Configuration
{
  "type": "stun_config",
  "data": {
    "iceServers": [{"urls": ["stun:server:3478"]}]
  }
}

// User joined room
{
  "type": "user_joined",
  "user_id": "user-123",
  "room_id": "room-456",
  "data": "{\"user_id\": \"user-123\", \"users\": [\"user-123\", \"user-456\"]}"
}

// User left room
{
  "type": "user_left",
  "user_id": "user-123",
  "room_id": "room-456"
}

// Room is full
{
  "type": "room_full",
  "room_id": "room-456"
}

// Error message
{
  "type": "error",
  "data": "{\"code\": 400, \"message\": \"Invalid request\"}"
}
```

## Scaling

### Horizontal Scaling

The application supports horizontal scaling through:

1. **Stateless Design**: All state is stored in Redis
2. **Session Affinity**: WebSocket connections are sticky to pods
3. **Redis Pub/Sub**: Cross-pod communication for room events
4. **Auto-scaling**: HPA configuration based on CPU/memory usage

### Performance Tuning

- **Connection Limits**: Adjust `maxRoomUsers` in `internal/model/room.go`
- **Redis Configuration**: Tune Redis for your workload
- **Resource Limits**: Adjust Kubernetes resource requests/limits
- **Load Balancer**: Configure appropriate session affinity

## Monitoring

### Health Checks

```bash
# Application health
curl http://localhost:8080/health

# Application readiness
curl http://localhost:8080/ready
```

### Kubernetes Monitoring

```bash
# Pod status
kubectl get pods -n webrtc-signaling

# Service status
kubectl get services -n webrtc-signaling

# HPA status
kubectl get hpa -n webrtc-signaling

# Logs
kubectl logs -f deployment/signaling-deployment -n webrtc-signaling
```

## Development

### Project Structure

```
signaling/
├── cmd/signaling/           # Application entry point
├── internal/
│   ├── config/             # Configuration management
│   ├── handler/            # HTTP/WebSocket handlers
│   ├── middleware/         # HTTP middleware
│   ├── model/              # Data models
│   ├── repository/         # Data access layer
│   └── service/            # Business logic
├── pkg/logger/             # Logging utilities
├── web/static/             # Test frontend
├── deployments/            # Deployment configurations
│   ├── docker/            # Docker configurations
│   ├── docker-compose/    # Docker Compose files
│   └── kubernetes/        # Kubernetes manifests
└── scripts/               # Utility scripts
```

### Available Make Commands

```bash
make help                   # Show all available commands
make dev-setup             # Set up development environment
make build                 # Build the application
make run                   # Run locally
make test                  # Run tests
make dev                   # Start development environment
make docker-build          # Build Docker images
make k8s-deploy           # Deploy to Kubernetes
make minikube-deploy      # Deploy to minikube
```

### Testing

The project includes a web-based test interface accessible at the root URL. Open multiple browser tabs to test multi-user scenarios.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Troubleshooting

### Common Issues

1. **WebSocket Connection Failed**
   - Check if the server is running on the correct port
   - Verify firewall settings
   - Check browser console for errors

2. **Redis Connection Failed**
   - Ensure Redis is running and accessible
   - Check Redis host/port configuration
   - Verify network connectivity

3. **STUN/TURN Not Working**
   - Check coturn server logs
   - Verify port accessibility (3478, 5349)
   - Update firewall rules for UDP ports

4. **Kubernetes Deployment Issues**
   - Check pod logs: `kubectl logs -f deployment/signaling-deployment -n webrtc-signaling`
   - Verify resource availability
   - Check service endpoints: `kubectl get endpoints -n webrtc-signaling`

### Debug Mode

Enable verbose logging by setting environment variables:
```bash
export LOG_LEVEL=debug
```

For more detailed troubleshooting, check the logs:
```bash
# Docker Compose
docker-compose logs -f signaling

# Kubernetes
kubectl logs -f deployment/signaling-deployment -n webrtc-signaling
# cline-meet-3
