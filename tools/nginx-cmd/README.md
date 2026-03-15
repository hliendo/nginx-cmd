# Nginx All-In-One Engine (v1.0.0-kaizen)

`nginx-cmd` is a complete, self-managed Nginx ecosystem. It integrates **Nginx**, **Certbot**, **Cron**, and a powerful **Go-based Orchestrator** into a single containerized solution.

## 🚀 Key Features

- **Domain Lifecycle**: Easily add, remove, and list Nginx domains.
- **SSL Bootstrap**: Built-in Internal PKI to issue instant certificates, allowing Nginx to start even before Let's Encrypt is active.
- **Auto-Escalation**: Automatically attempts to upgrade Bootstrap certificates to Let's Encrypt production certificates.
- **Industrial Integrity**: Periodic audits (`check` command) to ensure configuration and certificate consistency.
- **Trust Store Integration**: Install internal Root CAs into the system trust store with a single command.

## 🛠️ Usage

### Installation
The tool is designed to be compiled during your Nginx Docker build or used as a standalone binary in `/usr/local/bin/nginx-cmd`.

### Commands
- `nginx-cmd add [domain] [spa|proxy]`: Add a new domain with instant SSL.
- `nginx-cmd renew`: Intelligently renew or upgrade certificates.
- `nginx-cmd list`: View all active domains and their SSL status.
- `nginx-cmd check`: Audit system integrity.
- `nginx-cmd trust`: Install the internal CA into the container's trust store.

## ⚙️ Configuration (White-Labeling)

`nginx-cmd` is highly portable. You can customize its branding and paths via environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `NGINX_CMD_PKI_ROOT_CN` | Common Name for the Internal Root CA | `Internal Root CA` |
| `NGINX_CMD_PKI_INTER_CN` | Common Name for the Intermediate CA | `Internal Intermediate CA` |
| `NGINX_CMD_CERT_EMAIL` | Email for Certbot registration | `admin@example.com` |
| `NGINX_CMD_NGINX_PATH` | Path to Nginx root | `/opt/nginx` |
| `NGINX_CMD_CERTS_PATH` | Path to SSL certificates | `/etc/letsencrypt` |
| `NGINX_CMD_SITE_PATH` | Path to static sites | `/opt/site` |
| `NGINX_CMD_TRUST_PATH` | Path to system trust store | `/usr/local/share/ca-certificates` |

## 🏗️ The All-In-One Architecture

Unlike standard Nginx images, this engine is designed for zero-maintenance:
- **Nginx & Brotli**: High-performance engine with modern compression.
- **Certbot & ACME**: Built-in certificate management.
- **Cron Scheduler**: Automatic daily renewals and integrity checks.
- **Go Orchestrator**: The `nginx-cmd` tool simplifies complex ops into single commands.

## ⚖️ License

Distributed under the **MIT License**. See `LICENSE` for more information.

---
*Developed for Industrial Truth - 2026*
