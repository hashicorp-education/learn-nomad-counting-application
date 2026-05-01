# Nomad Deployment Guide - Counting Dashboard

This guide explains how to deploy the counting-dashboard application to HashiCorp Nomad using Nomad native service discovery.

## 📋 Prerequisites

- HashiCorp Nomad cluster (or Nomad dev mode for local testing)
- Docker installed on Nomad client nodes
- Docker images built and available locally or in a registry

## 🏗️ Architecture

The Nomad job deploys two service groups:

1. **counting-service** (Backend)
   - Go-based counting API
   - Port: 9001
   - Registers with Nomad service discovery as `counting-service`

2. **dashboard-service** (Frontend)
   - Nginx-based web dashboard
   - Port: 8080
   - Discovers counting-service via Nomad service discovery
   - Registers as `dashboard-service`

## 🚀 Quick Start

### 1. Build Docker Images

```bash
# Build counting service
cd counting-service
docker build -t counting-service:latest .

# Build dashboard service
cd ../dashboard-service
docker build -t dashboard-service:latest .
```

### 2. Deploy to Nomad

```bash
# From the counting-dashboard directory
nomad job run counting-dashboard.nomad.hcl
```

### 3. Verify Deployment

```bash
# Check job status
nomad job status counting-dashboard

# Check service registrations
nomad service list

# View service details
nomad service info counting-service
nomad service info dashboard-service
```

### 4. Access the Application

- **Dashboard**: http://localhost:8080
- **Counting API**: http://localhost:9001
- **Health Checks**:
  - Counting: http://localhost:9001/health
  - Dashboard: http://localhost:8080/health

## 📊 Service Discovery

The job uses **Nomad native service discovery** (no Consul required):

### Counting Service Registration

```hcl
service {
  provider = "nomad"
  name = "counting-service"
  port = "http"
  
  check {
    type     = "http"
    path     = "/health"
    interval = "10s"
    timeout  = "2s"
  }
}
```

### Dashboard Service Discovery

The dashboard automatically discovers the counting service using Nomad templates:

```hcl
template {
  data = <<EOH
window.COUNTING_SERVICE_URL = '{{ range nomadService "counting-service" }}http://{{ .Address }}:{{ .Port }}{{ end }}';
EOH
  destination = "local/service-config.html"
  change_mode = "restart"
}
```

## 🔧 Configuration Options

### Scaling

To run multiple instances, modify the `count` parameter:

```hcl
group "counting-service" {
  count = 3  # Run 3 instances
  ...
}
```

### Resource Allocation

Adjust resources based on your needs:

```hcl
resources {
  cpu    = 500  # MHz
  memory = 256  # MB
}
```

### Multi-Datacenter Deployment

Update the datacenters list:

```hcl
job "counting-dashboard" {
  datacenters = ["dc1", "dc2", "dc3"]
  ...
}
```

### Using Docker Registry

If using a Docker registry instead of local images:

```hcl
config {
  image = "your-registry/counting-service:latest"
  force_pull = true  # Always pull latest
}
```

## 🔍 Monitoring

### View Logs

```bash
# Get allocation IDs
nomad job status counting-dashboard

# View counting service logs
nomad alloc logs <alloc-id> counting-service

# View dashboard service logs
nomad alloc logs <alloc-id> dashboard-service

# Follow logs in real-time
nomad alloc logs -f <alloc-id> counting-service
```

### Check Service Health

```bash
# List all services
nomad service list

# Get service details with health status
nomad service info counting-service
nomad service info dashboard-service

# Check allocation health
nomad alloc status <alloc-id>
```

### Test Endpoints

```bash
# Test counting service
curl http://localhost:9001/health
curl http://localhost:9001

# Test dashboard service
curl http://localhost:8080/health
curl -I http://localhost:8080
```

## 🔄 Updates and Rollbacks

### Update the Application

```bash
# Make changes to your code
# Rebuild Docker images
docker build -t counting-service:latest ./counting-service
docker build -t dashboard-service:latest ./dashboard-service

# Deploy update
nomad job run counting-dashboard.nomad.hcl
```

### Rollback to Previous Version

```bash
# View job versions
nomad job history counting-dashboard

# Revert to previous version
nomad job revert counting-dashboard <version-number>
```

### Stop the Job

```bash
# Stop all allocations
nomad job stop counting-dashboard

# Stop and purge (removes from history)
nomad job stop -purge counting-dashboard
```

## 🐛 Troubleshooting

### Services Not Starting

1. Check allocation status:
```bash
nomad alloc status <alloc-id>
```

2. View allocation events:
```bash
nomad alloc status -verbose <alloc-id>
```

3. Check Docker images are available:
```bash
docker images | grep counting
```

### Service Discovery Not Working

1. Verify services are registered:
```bash
nomad service list
```

2. Check service health:
```bash
nomad service info counting-service
```

3. View template rendering:
```bash
nomad alloc fs <alloc-id> local/service-config.html
```

### Port Conflicts

If ports 8080 or 9001 are in use, modify the job spec:

```hcl
port "http" {
  static = 9002  # Change to available port
  to     = 9001
}
```

### Health Checks Failing

1. Check if services are responding:
```bash
curl http://localhost:9001/health
curl http://localhost:8080/health
```

2. View health check configuration:
```bash
nomad job inspect counting-dashboard | jq '.Job.TaskGroups[].Services[].Checks'
```

3. Increase health check timeout if needed:
```hcl
check {
  timeout  = "5s"  # Increase from 2s
}
```

## 📈 Production Considerations

### High Availability

For production, run multiple instances with constraints:

```hcl
group "counting-service" {
  count = 3
  
  # Don't run multiple instances on same node
  constraint {
    operator = "distinct_hosts"
    value    = "true"
  }
  
  # Spread across datacenters
  spread {
    attribute = "${node.datacenter}"
    weight    = 100
  }
}
```

### Update Strategy

Add canary deployments and auto-revert:

```hcl
update {
  max_parallel      = 1
  min_healthy_time  = "30s"
  healthy_deadline  = "5m"
  auto_revert       = true
  auto_promote      = true
  canary            = 1
}
```

### Resource Limits

Set appropriate resource limits:

```hcl
resources {
  cpu    = 1000  # MHz
  memory = 512   # MB
  
  # Optional: Set memory limit
  memory_max = 1024  # MB
}
```

### Persistent Storage

If you need persistent data:

```hcl
ephemeral_disk {
  size    = 1000  # MB
  migrate = true
}
```

## 🔐 Security

### Network Isolation

Use bridge networking for better isolation:

```hcl
network {
  mode = "bridge"
  
  port "http" {
    to = 9001
  }
}
```

### Service Mesh

For mTLS and advanced networking, consider using Consul Connect:

```hcl
service {
  provider = "consul"
  name = "counting-service"
  
  connect {
    sidecar_service {}
  }
}
```

## 📚 Additional Resources

- [Nomad Service Discovery](https://developer.hashicorp.com/nomad/docs/service-discovery)
- [Nomad Job Specification](https://developer.hashicorp.com/nomad/docs/job-specification)
- [Nomad Templates](https://developer.hashicorp.com/nomad/docs/job-specification/template)
- [Docker Driver](https://developer.hashicorp.com/nomad/docs/drivers/docker)

## 🎯 Next Steps

1. Set up monitoring with Prometheus/Grafana
2. Configure log aggregation (ELK, Loki)
3. Implement CI/CD pipeline
4. Add load balancing (Traefik, Fabio)
5. Set up alerting for service health

---

For more information, see the main [README.md](README.md) file.