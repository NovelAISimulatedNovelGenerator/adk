# ADK API Server Docker Deployment Guide

## Quick Start

### Prerequisites
- Docker 20.10+
- Docker Compose 2.0+
- Go 1.24+ (only for local development)

### 1. Configure Application

```bash
# Copy example configuration and edit as needed
cp config.example.yaml config.yaml
# Edit config.yaml with your settings
```

### 2. Build Docker Image

```bash
# Build the image
docker build -t adk-apiserver .

# Verify the image
docker images | grep adk-apiserver
```

### 3. Run with Docker Compose (Recommended for Development)

```bash
# Ensure you have plugins directory (create if empty)
mkdir -p plugins

# Build and start services
docker-compose up --build

# Run in background
docker-compose up -d

# View logs
docker-compose logs -f apiserver

# Stop services
docker-compose down
```

### 4. Run with Docker Directly

```bash
# Ensure plugins directory exists
mkdir -p plugins

# Run container
docker run -d \
  --name adk-apiserver \
  -p 8080:8080 \
  -v $(pwd)/config.yaml:/app/config.yaml:ro \
  -v $(pwd)/plugins:/app/plugins:rw \
  adk-apiserver

# View logs
docker logs -f adk-apiserver

# Stop container
docker stop adk-apiserver
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `ADK_CONFIG` | Configuration file path | `/app/config.yaml` |
| `ADDR` | Server listen address | `:8080` |
| `LOG_LEVEL` | Log level (debug/info/warn/error) | `info` |

### Volume Mounts

- `/app/config.yaml`: Application configuration file (read-only)
- `/app/plugins`: Directory for flow plugins (read-write)

## Production Deployment

### 1. Multi-Architecture Build

```bash
# Build for multiple architectures
docker buildx build --platform linux/amd64,linux/arm64 -t adk-apiserver:prod .
```

### 2. Security Hardening

```dockerfile
# Alternative production Dockerfile using distroless
FROM gcr.io/distroless/static:nonroot AS runtime
COPY --from=builder /app/apiserver /apiserver
COPY --from=builder /app/config.example.yaml /config.example.yaml
USER nonroot:nonroot
EXPOSE 8080
ENTRYPOINT ["/apiserver"]
```

### 3. Kubernetes Manifest

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: adk-apiserver
spec:
  replicas: 3
  selector:
    matchLabels:
      app: adk-apiserver
  template:
    metadata:
      labels:
        app: adk-apiserver
    spec:
      containers:
      - name: apiserver
        image: adk-apiserver:latest
        ports:
        - containerPort: 8080
        env:
        - name: LOG_LEVEL
          value: "info"
        volumeMounts:
        - name: config
          mountPath: /app/config.yaml
          subPath: config.yaml
        - name: plugins
          mountPath: /app/plugins
        resources:
          requests:
            cpu: 100m
            memory: 128Mi
          limits:
            cpu: 500m
            memory: 512Mi
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
      volumes:
      - name: config
        configMap:
          name: adk-config
      - name: plugins
        persistentVolumeClaim:
          claimName: adk-plugins-pvc
```

## Development Workflow

### Building Plugins

```bash
# Build flow plugins
go build -buildmode=plugin -o plugins/novel_flow_v4.so ./flows/novel_v4
go build -buildmode=plugin -o plugins/search_assistant.so ./examples/search_assistant

# Run with local plugins
docker-compose up --build
```

### Debugging

```bash
# Get shell access
docker exec -it adk-apiserver /bin/sh

# Check health endpoint
curl http://localhost:8080/health

# View environment
docker exec adk-apiserver env

# Check running processes
docker exec adk-apiserver ps aux
```

## Monitoring

### Health Checks
The image includes a health check that performs HTTP requests to the health endpoint every 30 seconds.

### Metrics
Standard container metrics are exposed via Docker stats:

```bash
docker stats adk-apiserver
```

### Logging
All logs are sent to stdout/stderr for proper Docker logging:

```bash
# View logs in real-time
docker-compose logs -f apiserver

# Filter logs by time
docker-compose logs --since=1h apiserver

# Show only errors
docker-compose logs apiserver | grep ERROR
```

## Updates and Rollbacks

### Rolling Updates

```bash
# Pull new image
docker pull adk-apiserver:latest

# Recreate containers
docker-compose up -d --force-recreate
```

### Rollback

```bash
# Use specific tag
docker run -d --name adk-apiserver \
  adk-apiserver:v1.0.0
```

## Troubleshooting

### Container Won't Start
Check logs:
```bash
docker logs adk-apiserver
```
Check config format:
```bash
# Validate config before running
docker run --rm -v $(pwd)/config.yaml:/app/config.yaml:ro \
  adk-apiserver --config /app/config.yaml --test-config
```

### Performance Issues
Monitor resources:
```bash
docker stats --no-stream
docker top adk-apiserver
```

### Network Issues
Check connectivity:
```bash
docker network ls
docker inspect adk-apiserver
```