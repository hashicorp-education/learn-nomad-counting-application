# Consul DNS Deployment Guide - Counting Dashboard

This guide explains how to deploy the counting-dashboard application to HashiCorp Nomad using Consul DNS for service discovery.

## 📋 Prerequisites

- HashiCorp Nomad cluster
- HashiCorp Consul cluster (running and accessible)
- Docker installed on Nomad client nodes
- Consul DNS configured on port 8600
- Docker images built and available

## 🏗️ Architecture

The Nomad job deploys two service groups with Consul DNS integration:

```
┌─────────────────────────────────────────────────────┐
│                  Consul Service Registry             │
│  ┌──────────────────┐    ┌──────────────────┐      │
│  │ counting-service │    │ dashboard-service │      │
│  │  .service.consul │    │  .service.consul  │      │
│  └──────────────────┘    └──────────────────┘      │
└─────────────────────────────────────────────────────┘
                    │
                    │ DNS Resolution
                    ▼
┌─────────────────────────────────────────────────────┐
│              Nomad Allocations                       │
│  ┌──────────────────┐    ┌──────────────────┐      │
│  │  Dashboard       │───▶│  Counting        │      │
│  │  Service         │    │  Service         │      │
│  │  (Nginx)         │    │  (Go)            │      │
│  │  Port: 8080      │    │  Port: 9001      │      │
│  └──────────────────┘    └──────────────────┘      │
└─────────────────────────────────────────────────────┘
```

## 🔍 Key Differences from Nomad Native Service Discovery

### Service Registration
Services register with **Consul** instead of Nomad:

```hcl
service {
  name = "counting-service"
  port = "http"
  # No provider = "nomad" - defaults to Consul
}
```

### DNS Configuration
Dashboard container is configured to use Consul DNS:

```hcl
config {
  dns_servers = ["${attr.unique.network.ip-address}"]
  dns_search_domains = ["service.consul"]
}
```

### Service Discovery
Uses Consul's `service` function instead of `nomadService`:

```hcl
template {
  data = <<EOH
window.COUNTING_SERVICE_URL = '{{ range service "counting-service" }}http://counting-service.service.consul:{{ .Port }}{{ end }}';
EOH
}
```

## 🚀 Quick Start

### 1. Verify Consul is Running

```bash
# Check Consul status
consul members

# Verify Consul DNS is working
dig @127.0.0.1 -p 8600 consul.service.consul

# Check Consul UI (if enabled)
# http://localhost:8500
```

### 2. Build Docker Images

```bash
# Build counting service
cd ~/Dev/github/aimeeu/counting-dashboard/counting-service
docker build -t aimeeu/counting-service:latest .

# Build dashboard service
cd ../dashboard-service
docker build -t aimeeu/dashboard-service:latest .
```

### 3. Deploy to Nomad

```bash
# From the counting-dashboard directory
cd ~/Dev/github/aimeeu/counting-dashboard
nomad job run counting-dashboard-consul-dns.nomad.hcl
```

### 4. Verify Deployment

```bash
# Check job status
nomad job status counting-dashboard-consul-dns

# Check Consul service registrations
consul catalog services
consul catalog nodes -service=counting-service
consul catalog nodes -service=dashboard-service

# Check service health in Consul
consul health service counting-service
consul health service dashboard-service
```

### 5. Test DNS Resolution

```bash
# From a Nomad client node
dig @127.0.0.1 -p 8600 counting-service.service.consul
dig @127.0.0.1 -p 8600 dashboard-service.service.consul

# Test with curl
curl http://counting-service.service.consul:9001/health
```

### 6. Access the Application

- **Dashboard**: http://localhost:8080 (or your node IP)
- **Counting API**: http://counting-service.service.consul:9001
- **Consul UI**: http://localhost:8500 (if enabled)

## 🔧 Configuration

### Consul DNS Settings

The dashboard service is configured to use Consul DNS:

```hcl
config {
  # Use the node's IP for Consul DNS
  dns_servers = ["${attr.unique.network.ip-address}"]
  
  # Automatically append .service.consul to DNS queries
  dns_search_domains = ["service.consul"]
}
```

This allows the dashboard to resolve `counting-service.service.consul` to the actual service IP and port.

### Service Discovery Template

The template queries Consul for healthy service instances:

```hcl
template {
  data = <<EOH
window.COUNTING_SERVICE_URL = '{{ range service "counting-service" }}{{ if .Status | eq "passing" }}http://counting-service.service.consul:{{ .Port }}{{ end }}{{ end }}';
EOH
  destination = "local/config.js"
  change_mode = "restart"
}
```

### Health Checks

Services register health checks with Consul:

```hcl
check {
  type     = "http"
  path     = "/health"
  interval = "10s"
  timeout  = "2s"
}
```

## 📊 Monitoring with Consul

### View Services in Consul

```bash
# List all services
consul catalog services

# Get service details
consul catalog nodes -service=counting-service -detailed

# Check service health
consul health service counting-service
consul health service dashboard-service
```

### Consul UI

If Consul UI is enabled, access it at http://localhost:8500 to:
- View all registered services
- Check service health status
- View service instances and their metadata
- Monitor service checks

### Query Consul DNS

```bash
# Get all instances of counting-service
dig @127.0.0.1 -p 8600 counting-service.service.consul

# Get SRV records (includes port information)
dig @127.0.0.1 -p 8600 counting-service.service.consul SRV

# Query specific datacenter
dig @127.0.0.1 -p 8600 counting-service.service.dc1.consul
```

## 🔄 Updates and Rollbacks

### Update the Application

```bash
# Rebuild images
docker build -t aimeeu/counting-service:latest ./counting-service
docker build -t aimeeu/dashboard-service:latest ./dashboard-service

# Deploy update
nomad job run counting-dashboard-consul-dns.nomad.hcl

# Monitor rollout
nomad job status counting-dashboard-consul-dns
consul watch -type=service -service=counting-service
```

### Rollback

```bash
# View job versions
nomad job history counting-dashboard-consul-dns

# Revert to previous version
nomad job revert counting-dashboard-consul-dns <version-number>
```

### Stop the Job

```bash
# Stop all allocations
nomad job stop counting-dashboard-consul-dns

# This will also deregister services from Consul
```

## 🐛 Troubleshooting

### Services Not Registering in Consul

1. Check Consul agent is running:
```bash
consul members
systemctl status consul
```

2. Verify Nomad can reach Consul:
```bash
nomad agent-info | grep consul
```

3. Check allocation logs:
```bash
nomad alloc logs <alloc-id> counting-service
```

### DNS Resolution Not Working

1. Verify Consul DNS is listening:
```bash
netstat -tulpn | grep 8600
dig @127.0.0.1 -p 8600 consul.service.consul
```

2. Check Docker DNS configuration:
```bash
nomad alloc exec <alloc-id> cat /etc/resolv.conf
```

3. Test DNS from inside container:
```bash
nomad alloc exec <alloc-id> nslookup counting-service.service.consul
```

### Service Health Checks Failing

1. Check service health in Consul:
```bash
consul health service counting-service
```

2. View detailed check information:
```bash
consul catalog nodes -service=counting-service -detailed
```

3. Test health endpoint directly:
```bash
curl http://<service-ip>:<port>/health
```

### Dashboard Can't Connect to Counting Service

1. Verify counting service is registered and healthy:
```bash
consul health service counting-service
```

2. Check DNS resolution from dashboard container:
```bash
nomad alloc exec <dashboard-alloc-id> nslookup counting-service.service.consul
```

3. View generated config.js:
```bash
nomad alloc fs <dashboard-alloc-id> local/config.js
```

4. Check dashboard logs:
```bash
nomad alloc logs <dashboard-alloc-id> dashboard-service
```

## 🔐 Security Considerations

### Consul ACLs

If using Consul ACLs, ensure Nomad has appropriate tokens:

```hcl
# In Nomad client configuration
consul {
  token = "your-consul-token"
}
```

### Service-to-Service Communication

For production, consider using Consul Connect for mTLS:

```hcl
service {
  name = "counting-service"
  
  connect {
    sidecar_service {}
  }
}
```

### Network Policies

Restrict network access between services:

```hcl
network {
  mode = "bridge"
}
```

## 📈 Production Considerations

### High Availability

Run multiple instances with health checks:

```hcl
group "counting-service" {
  count = 3
  
  constraint {
    operator = "distinct_hosts"
    value    = "true"
  }
}
```

### Load Balancing

Consul DNS automatically load balances across healthy instances:

```bash
# Returns different IPs on each query
dig @127.0.0.1 -p 8600 counting-service.service.consul +short
```

### Monitoring

Integrate with monitoring tools:
- Prometheus: Use Consul service discovery
- Grafana: Visualize service metrics
- Consul: Built-in health monitoring

### Backup and Recovery

```bash
# Backup Consul data
consul snapshot save backup.snap

# Restore Consul data
consul snapshot restore backup.snap
```

## 🆚 Comparison: Consul DNS vs Nomad Native

| Feature | Consul DNS | Nomad Native |
|---------|-----------|--------------|
| **Service Registry** | Consul | Nomad |
| **DNS Resolution** | .service.consul | Not available |
| **Health Checks** | Consul agent | Nomad |
| **Multi-DC** | Native support | Limited |
| **Service Mesh** | Consul Connect | Not available |
| **UI** | Consul UI | Nomad UI |
| **Complexity** | Higher (2 systems) | Lower (1 system) |
| **Features** | More advanced | Basic |

## 📚 Additional Resources

- [Consul Service Discovery](https://developer.hashicorp.com/consul/docs/discovery/services)
- [Consul DNS Interface](https://developer.hashicorp.com/consul/docs/discovery/dns)
- [Nomad Consul Integration](https://developer.hashicorp.com/nomad/docs/integrations/consul)
- [Consul Health Checks](https://developer.hashicorp.com/consul/docs/services/usage/checks)

## 🎯 Next Steps

1. Enable Consul Connect for service mesh
2. Implement Consul ACLs for security
3. Set up Consul monitoring and alerting
4. Configure multi-datacenter replication
5. Integrate with external load balancers

---

For more information, see the main [README.md](README.md) and [NOMAD-DEPLOYMENT.md](NOMAD-DEPLOYMENT.md) files.