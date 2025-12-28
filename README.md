# MizbanCloud CLI

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-GPL--2.0-blue.svg)](LICENSE)
[![Platform](https://img.shields.io/badge/Platform-Linux%20%7C%20macOS%20%7C%20Windows-lightgrey)](https://github.com/mizbancloud/cli/releases)

A powerful, feature-rich command-line interface for managing MizbanCloud infrastructure services. Built with Go for cross-platform compatibility and optimal performance.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Authentication](#authentication)
- [Command Reference](#command-reference)
  - [Cloud (IaaS)](#cloud-iaas)
  - [CDN](#cdn)
  - [Support](#support)
- [Configuration](#configuration)
- [Output Formats](#output-formats)
- [Contributing](#contributing)
- [License](#license)

## Features

- **Cloud Infrastructure (IaaS)**
  - Virtual machine lifecycle management
  - Block storage volumes and snapshots
  - SSH key management
  - Security groups and firewall rules
  - Private network configuration

- **Content Delivery Network (CDN)**
  - Domain and DNS record management
  - SSL/TLS certificate provisioning (Let's Encrypt & custom)
  - Advanced caching controls with purge capabilities
  - Web Application Firewall (WAF) with rule management
  - DDoS protection with multiple defense layers
  - Rate limiting and IP/Geo blocking
  - Load balancer cluster configuration
  - Real-time log forwarding
  - Custom error pages

- **Support System**
  - Ticket creation and management
  - Department-based routing
  - Conversation threading

## Installation

### Ubuntu/Debian

```bash
# Download binary
curl -LO https://github.com/mizbancloud/cli/releases/latest/download/mizban-linux-amd64

# Make executable
chmod +x mizban-linux-amd64

# Move to system path
sudo mv mizban-linux-amd64 /usr/local/bin/mizban

# Verify installation
mizban --version
```

### macOS

```bash
# Intel Mac
curl -LO https://github.com/mizbancloud/cli/releases/latest/download/mizban-darwin-amd64
chmod +x mizban-darwin-amd64
sudo mv mizban-darwin-amd64 /usr/local/bin/mizban

# Apple Silicon (M1/M2/M3)
curl -LO https://github.com/mizbancloud/cli/releases/latest/download/mizban-darwin-arm64
chmod +x mizban-darwin-arm64
sudo mv mizban-darwin-arm64 /usr/local/bin/mizban

# Verify installation
mizban --version
```

### Windows

**PowerShell:**
```powershell
# Download binary
Invoke-WebRequest -Uri "https://github.com/mizbancloud/cli/releases/latest/download/mizban-windows-amd64.exe" -OutFile "mizban.exe"

# Move to a directory in your PATH (e.g., C:\Windows or create custom folder)
Move-Item mizban.exe C:\Windows\mizban.exe

# Verify installation
mizban --version
```

**Manual Installation:**
1. Download `mizban-windows-amd64.exe` from [Releases](https://github.com/mizbancloud/cli/releases)
2. Rename to `mizban.exe`
3. Move to a folder in your PATH (e.g., `C:\Program Files\mizban\`)
4. Or add the folder to your system PATH environment variable

### Build from Source

Requires Go 1.21 or later.

```bash
git clone https://github.com/mizbancloud/cli.git
cd cli

# Build for current platform
make build

# Platform-specific builds
make linux
make darwin
make windows
```

## Quick Start

```bash
# 1. Authenticate with your API token
mizban login --token YOUR_API_TOKEN

# 2. List your cloud servers
mizban server list

# 3. List your CDN domains
mizban domain list

# 4. Check your profile
mizban profile show
```

## Authentication

The CLI supports multiple authentication methods:

```bash
# Interactive login (prompts for token)
mizban login

# Direct token authentication
mizban login --token YOUR_API_TOKEN

# Environment variable
export MIZBAN_API_TOKEN=YOUR_API_TOKEN
mizban server list

# Logout and clear credentials
mizban logout
```

API tokens can be generated from the [MizbanCloud Dashboard](https://panel.mizbancloud.com/profile/api-keys).

## Command Reference

### Cloud (IaaS)

#### Server Management

```bash
# List all servers
mizban server list [--json]

# Create a new server
mizban server create \
  --name web-server \
  --os ubuntu-22.04 \
  --cpu 2 \
  --ram 2048 \
  --storage 40 \
  --datacenter tehran-1

# Get server details
mizban server get <server-id> [--json]

# Power operations
mizban server power on <server-id>
mizban server power off <server-id>
mizban server power reboot <server-id>

# Access VNC console
mizban server vnc <server-id>

# Resize server resources
mizban server resize <server-id> --cpu 4 --ram 4096

# Delete server
mizban server delete <server-id> [--force]
```

#### Volume Management

```bash
# List volumes
mizban volume list [--json]

# Create volume
mizban volume create --name data-vol --size 100

# Attach/Detach operations
mizban volume attach <volume-id> --server <server-id>
mizban volume detach <volume-id>

# Resize volume
mizban volume resize <volume-id> --size 200

# Delete volume
mizban volume delete <volume-id> [--force]
```

#### Snapshots

```bash
# List snapshots
mizban snapshot list [--json]

# Create snapshot
mizban snapshot create --name backup-$(date +%Y%m%d) --server <server-id>

# Restore from snapshot
mizban snapshot restore <snapshot-id> --server <server-id>

# Delete snapshot
mizban snapshot delete <snapshot-id>
```

#### SSH Keys

```bash
# List SSH keys
mizban ssh-key list

# Add existing key
mizban ssh-key add --name laptop --key "ssh-rsa AAAA..."

# Generate new key pair
mizban ssh-key generate --name production

# Delete key
mizban ssh-key delete <key-id>
```

#### Firewall (Security Groups)

```bash
# List firewalls
mizban firewall list

# Create firewall
mizban firewall create --name web-traffic

# Add rules
mizban firewall rule add --firewall <id> \
  --direction ingress \
  --protocol tcp \
  --port-min 80 \
  --port-max 80 \
  --remote 0.0.0.0/0

mizban firewall rule add --firewall <id> \
  --direction ingress \
  --protocol tcp \
  --port-min 443 \
  --port-max 443

# Attach to server
mizban firewall attach <firewall-id> --server <server-id>

# Detach from server
mizban firewall detach <firewall-id> --server <server-id>
```

#### Private Networks

```bash
# List networks
mizban network list

# Create network
mizban network create --name internal --cidr 10.0.0.0/24

# Attach server to network
mizban network attach <network-id> --server <server-id>

# Detach server
mizban network detach <network-id> --server <server-id>
```

---

### CDN

#### Domain Management

```bash
# List domains
mizban domain list [--json]

# Add domain
mizban domain add --domain example.com

# Get domain details (includes nameserver info)
mizban domain get <domain-id> [--json]

# Get domain WHOIS information
mizban domain whois <domain-id>

# Get traffic usage
mizban domain usage <domain-id> --period month

# Get traffic reports
mizban domain reports --domain <domain-id> --period week [--json]

# Set redirect mode (none/www/naked)
mizban domain redirect-mode --domain <domain-id> --mode www

# Delete domain
mizban domain delete <domain-id> [--force]
```

#### DNS Records

```bash
# List records
mizban dns list --domain <domain-id> [--json]

# Get single record
mizban dns get <record-id> --domain <domain-id>

# List proxiable records
mizban dns proxiable --domain <domain-id>

# Add records
mizban dns add --domain <domain-id> \
  --type A \
  --name @ \
  --destination 203.0.113.50 \
  --ttl 3600 \
  --proxy

mizban dns add --domain <domain-id> \
  --type CNAME \
  --name www \
  --destination example.com \
  --proxy

mizban dns add --domain <domain-id> \
  --type MX \
  --name @ \
  --destination mail.example.com \
  --priority 10

# Update record
mizban dns update --domain <domain-id> \
  --record <record-id> \
  --destination 203.0.113.100

# Delete record
mizban dns delete <record-id> --domain <domain-id>

# Import/Export zone files
mizban dns export --domain <domain-id> > zone.txt
mizban dns import --domain <domain-id> --zone "$(cat zone.txt)"

# Auto-fetch records from current nameservers
mizban dns fetch-records --domain <domain-id>

# Custom nameservers (vanity NS)
mizban dns custom-ns get --domain <domain-id>
mizban dns custom-ns set --domain <domain-id> --ns1 ns1.example.com --ns2 ns2.example.com
mizban dns custom-ns delete --domain <domain-id>

# DNSSEC management
mizban dns dnssec status --domain <domain-id>
mizban dns dnssec enable --domain <domain-id>
mizban dns dnssec disable --domain <domain-id>
```

#### SSL/TLS Certificates

```bash
# List certificates
mizban ssl list --domain <domain-id> [--json]

# Get SSL status and settings
mizban ssl status --domain <domain-id>

# Get certificate info
mizban ssl info --domain <domain-id>

# Request free Let's Encrypt certificate
mizban ssl request-free --domain <domain-id>

# Add custom certificate
mizban ssl add-custom --domain <domain-id> \
  --cert "$(cat cert.pem)" \
  --key "$(cat key.pem)" \
  --chain "$(cat chain.pem)"

# Attach certificate to DNS records
mizban ssl attach --domain <domain-id> --cert <cert-id> --records 1,2,3

# Detach certificate
mizban ssl detach --domain <domain-id> --records 1,2,3

# Use default MizbanCloud SSL
mizban ssl attach-default --domain <domain-id>
mizban ssl detach-default --domain <domain-id>

# Delete certificate
mizban ssl delete <cert-id> --domain <domain-id>

# SSL Settings
mizban ssl settings tls-version --domain <domain-id> --version 1.2
mizban ssl settings hsts --domain <domain-id> --enabled --max-age 31536000 --include-subdomains --preload
mizban ssl settings redirect --domain <domain-id> --enabled
mizban ssl settings backend-protocol --domain <domain-id> --protocol https
mizban ssl settings h3 --domain <domain-id> --enabled
mizban ssl settings csp-override --domain <domain-id> --enabled
```

#### Cache Management

```bash
# Get cache status
mizban cache status --domain <domain-id> [--json]

# Set cache mode (standard/aggressive/no-cache)
mizban cache mode --domain <domain-id> --mode aggressive

# Developer mode (bypass cache)
mizban cache dev-mode --domain <domain-id> --enabled

# Always online mode
mizban cache always-online --domain <domain-id> --enabled

# Cookie caching
mizban cache cache-cookies --domain <domain-id> --enabled

# Purge cache
mizban cache purge --domain <domain-id> --all
mizban cache purge --domain <domain-id> --url https://example.com/page.html

# Cache TTL settings
mizban cache settings ttl --domain <domain-id> --ttl 86400
mizban cache settings browser --domain <domain-id> --mode override --ttl 3600
mizban cache settings errors-ttl --domain <domain-id> --ttl 300

# Minification
mizban cache settings minify --domain <domain-id> --html --css --js

# Image optimization
mizban cache settings image webp --domain <domain-id> --enabled
mizban cache settings image resize --domain <domain-id> --enabled
```

#### Web Application Firewall (WAF)

```bash
# Get WAF status
mizban waf status --domain <domain-id> [--json]

# Enable/Disable WAF
mizban waf enable --domain <domain-id> --mode block
mizban waf disable --domain <domain-id>

# List WAF layers
mizban waf layers --domain <domain-id>

# List and manage rules
mizban waf rules list --domain <domain-id>
mizban waf rules disabled --domain <domain-id>
mizban waf rules toggle --domain <domain-id> --rule <rule-id> --enabled

# Toggle rule groups
mizban waf groups toggle --domain <domain-id> --group <group-id> --enabled

# IP/Country firewall (legacy - use access-rules instead)
mizban waf firewall block-ip --domain <domain-id> --ip 1.2.3.4 --action block
mizban waf firewall unblock-ip --domain <domain-id> --ip 1.2.3.4
mizban waf firewall block-country --domain <domain-id> --country CN
mizban waf firewall unblock-country --domain <domain-id> --country CN
```

#### Access Rules (IP/Geo Blocking)

```bash
# Get access rules status
mizban access-rules status --domain <domain-id> [--json]

# IP-based rules
mizban access-rules add-ip --domain <domain-id> --ip 192.168.1.0/24 --action allow
mizban access-rules add-ip --domain <domain-id> --ip 10.0.0.1 --action block
mizban access-rules add-ip --domain <domain-id> --ip 172.16.0.1 --action challenge
mizban access-rules remove-ip --domain <domain-id> --ip 10.0.0.1

# Country-based rules
mizban access-rules add-country --domain <domain-id> --country CN --action block
mizban access-rules add-country --domain <domain-id> --country IR --action allow
mizban access-rules remove-country --domain <domain-id> --country CN
```

#### DDoS Protection

```bash
# Get DDoS status
mizban ddos status --domain <domain-id> [--json]

# Set protection mode
mizban ddos mode --domain <domain-id> --mode normal      # Standard protection
mizban ddos mode --domain <domain-id> --mode high        # High protection
mizban ddos mode --domain <domain-id> --mode under_attack # Maximum protection
mizban ddos mode --domain <domain-id> --mode off         # Disable

# Configure captcha module
mizban ddos captcha --domain <domain-id> --module recaptcha
mizban ddos captcha --domain <domain-id> --module hcaptcha
mizban ddos captcha --domain <domain-id> --module turnstile

# Set challenge TTLs
mizban ddos ttl cookie --domain <domain-id> --ttl 3600
mizban ddos ttl js --domain <domain-id> --ttl 1800
mizban ddos ttl captcha --domain <domain-id> --ttl 7200
```

#### Rate Limiting

```bash
# Get rate limit status
mizban ratelimit status --domain <domain-id> [--json]

# Configure rate limiting
mizban ratelimit set --domain <domain-id> \
  --limit 100 \
  --window 60 \
  --block 300

# Enable/Disable
mizban ratelimit enable --domain <domain-id>
mizban ratelimit disable --domain <domain-id>
```

#### Load Balancer Clusters

```bash
# List clusters
mizban cluster list --domain <domain-id> [--json]

# List cluster assignments
mizban cluster assignments --domain <domain-id>

# Create cluster pool
mizban cluster add --domain <domain-id> \
  --name backend-pool \
  --port 443 \
  --method roundrobin \
  --error-reporting

# Update cluster
mizban cluster update --domain <domain-id> \
  --cluster <cluster-id> \
  --method leastconn

# Delete cluster
mizban cluster delete --domain <domain-id> --cluster <cluster-id> [--force]

# Server management
mizban cluster server add --domain <domain-id> \
  --cluster <cluster-id> \
  --address 10.0.0.1 \
  --port 8080 \
  --weight 100 \
  --protocol HTTPS

mizban cluster server delete --domain <domain-id> \
  --cluster <cluster-id> \
  --server <server-id> [--force]

# Assign cluster to path
mizban cluster assign --domain <domain-id> --cluster <cluster-id> --path <path-id>
mizban cluster unassign --domain <domain-id> --cluster <cluster-id> --path <path-id>
```

#### Page Rules

```bash
# List page rules/paths
mizban page-rules list --domain <domain-id> [--json]
mizban page-rules list --domain <domain-id> --type waf
mizban page-rules list --domain <domain-id> --type ratelimit

# Add path
mizban page-rules add-path --domain <domain-id> --path "/api/*" --priority 1

# Set rule for path
mizban page-rules set-rule --domain <domain-id> \
  --path <path-id> \
  --type cache \
  --settings '{"ttl": 3600, "mode": "aggressive"}'

# Delete rule from path
mizban page-rules delete-rule --domain <domain-id> --path <path-id> --type cache

# Delete path
mizban page-rules delete-path <path-id> --domain <domain-id> [--force]
```

#### Custom Error Pages

```bash
# Get custom pages
mizban custom-pages get --domain <domain-id> [--json]

# Set custom error page
mizban custom-pages set --domain <domain-id> \
  --code 503 \
  --content "<html><body><h1>Maintenance</h1></body></html>"

# Delete custom page
mizban custom-pages delete --domain <domain-id> --code 503
```

#### Log Forwarding

```bash
# List log forwarders
mizban log-forwarder list --domain <domain-id> [--json]

# Add log forwarder
mizban log-forwarder add --domain <domain-id> \
  --name production-logs \
  --type elasticsearch \
  --endpoint https://es.example.com:9200 \
  --enabled \
  --config '{"index": "cdn-logs", "username": "elastic"}'

# Update forwarder
mizban log-forwarder update --domain <domain-id> \
  --forwarder <forwarder-id> \
  --enabled=false

# Delete forwarder
mizban log-forwarder delete <forwarder-id> --domain <domain-id> [--force]
```

#### CDN Plans

```bash
# List available plans
mizban plan list [--json]
```

---

### Support

#### Ticket Management

```bash
# List tickets
mizban ticket list [--json]
mizban ticket list --status open
mizban ticket list --status closed

# List departments
mizban ticket departments

# Create ticket
mizban ticket create \
  --subject "Technical Issue" \
  --message "Detailed description..." \
  --department support \
  --priority high

# Get ticket details
mizban ticket get <ticket-id> [--json]

# Reply to ticket
mizban ticket reply <ticket-id> --message "Follow-up message..."

# Close ticket
mizban ticket close <ticket-id>
```

---

### Profile & Settings

```bash
# View profile
mizban profile show [--json]

# Update profile
mizban profile update --name "John Doe" --phone "+1234567890"

# Manage API keys
mizban profile api-keys list
mizban profile api-keys create --name "CI/CD Pipeline"
mizban profile api-keys delete <key-id>
```

## Configuration

The CLI stores configuration in `~/.mizbancloud/config.yaml`:

```yaml
api_token: your-api-token-here
base_url: https://auth.mizbancloud.com/api
```

### Environment Variables

| Variable | Description |
|----------|-------------|
| `MIZBAN_API_TOKEN` | API authentication token |
| `MIZBAN_BASE_URL` | API base URL (optional) |
| `MIZBAN_CONFIG_PATH` | Custom config file path |

## Output Formats

All list and get commands support JSON output for scripting:

```bash
# JSON output
mizban server list --json
mizban domain get 123 --json | jq '.name'

# Use with other tools
mizban dns list --domain 1 --json | jq '.[] | select(.type == "A")'
```

## Exit Codes

| Code | Description |
|------|-------------|
| 0 | Success |
| 1 | General error |
| 2 | Authentication error |
| 3 | Network error |
| 4 | Resource not found |

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## Security

If you discover a security vulnerability, please email security@mizbancloud.com instead of opening a public issue.

## License

This project is licensed under the GNU General Public License v2.0 - see the [LICENSE](LICENSE) file for details.

---

<p align="center">
  <strong>MizbanCloud CLI</strong> - Cloud Infrastructure at Your Fingertips
  <br>
  <a href="https://mizbancloud.com">Website</a> •
  <a href="https://docs.mizbancloud.com">Documentation</a> •
  <a href="https://panel.mizbancloud.com">Dashboard</a>
</p>
