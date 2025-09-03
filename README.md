# Olake-UI

<h1 align="center" style="border-bottom: none">
    <a href="https://datazip.io/olake" target="_blank">
        <img alt="olake" src="https://github.com/user-attachments/assets/d204f25f-5289-423c-b3f2-44b2194bdeaf" width="100" height="100"/>
    </a>
    <br>OLake
</h1>

<p align="center">Fastest open-source tool for replicating Databases to Apache Iceberg or Data Lakehouse. ⚡ Efficient, quick and scalable data ingestion for real-time analytics. Starting with MongoDB. Visit <a href="https://olake.io/" target="_blank">olake.io/docs</a> for the full documentation, and benchmarks</p>

<p align="center">
    <a href="https://github.com/datazip-inc/olake-ui/issues"><img alt="GitHub issues" src="https://img.shields.io/github/issues/datazip-inc/olake"/></a> <a href="https://olake.io/docs"><img alt="Documentation" height="22" src="https://img.shields.io/badge/view-Documentation-blue?style=for-the-badge"/></a>
    <a href="https://join.slack.com/t/getolake/shared_invite/zt-2utw44do6-g4XuKKeqBghBMy2~LcJ4ag"><img alt="slack" src="https://img.shields.io/badge/Join%20Our%20Community-Slack-blue"/></a>
</p>

## Overview

Olake-UI offers an intuitive web interface to configure, monitor, and manage your data replication jobs.

- [UI Readme](/olake_frontend/README.md)
- [Server Readme](/server/README.md)
- [UI Figma Design](https://www.figma.com/design/FwLnU97I8LjtYNREPyYofc/Olake-Design-Community?node-id=1-46&p=f&t=y3BIsLTUaXhHwYLG-0)
- [API Contracts](/api-contract.md)

## Running with Docker Compose

This Docker Compose setup provides a comprehensive environment(OLake UI, Temporal worker, Temporal service, and dependencies) for demonstrating and exploring Olake's capabilities. This is the recommended way to get started for local development or evaluation.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed and running (Docker Desktop recommended for Mac/Windows)
- [Docker Compose](https://docs.docker.com/compose/) (comes with Docker Desktop)
- Make sure port `8000` is available, as OLake UI is exposed on that port.

### Quick Start

1. **One-Command Setup:**

```sh
curl -sSL https://raw.githubusercontent.com/datazip-inc/olake-ui/master/docker-compose.yml | docker compose -f - up -d
```

2. **Access the services:**

   - **OLake UI:** [http://localhost:8000](http://localhost:8000)
   - **Default login:** Username: `admin`, Password: `password`

### Optional Configuration

**Custom Admin User:**

The stack automatically creates an initial admin user on first startup. To change the default credentials, edit the `x-signup-defaults` section in `docker-compose.yml`:

```yaml
x-signup-defaults:
username: &defaultUsername "your-custom-username"
password: &defaultPassword "your-secure-password"
email: &defaultEmail "your-email@example.com"
```

**Custom Data Directory:**

By default, data is stored in `${PWD}/olake-data` directory. To use a different location, edit the `x-app-defaults` section in `docker-compose.yml`:

```yaml
x-app-defaults:
  host_persistence_path: &hostPersistencePath /your/host/path
```

Make sure the directory exists and is writable.

**Encryption Modes:**

Configure encryption in `docker-compose.yml`:

```yaml
x-encryption:
  # 1. For AWS KMS (starts with 'arn:aws:kms:'):
  key: &encryptionKey "arn:aws:kms:..."
  
  # 2. For local AES-256 (any other non-empty string):
  # key: &encryptionKey "secret-key"  # Auto-hashed to 256-bit key
  
  # 3. For no encryption (not recommended for production):
  # key: &encryptionKey ""
```

- KMS: Uses AWS Key Management Service for encryption/decryption
- Local: Uses AES-256-GCM with key derived from your passphrase
- Empty: No encryption (for development only)

### Troubleshooting

- If there are any file permission error, ensure the host persistence/config directory is writable by Docker.
- For complete logs, use:
  ```bash
  docker-compose logs -f
  ```
- For logs specific to a service, use:
  ```bash
  docker compose logs -f <service_name>
  ```

### Reset Everything
The stack can be completely reset with:
```bash
docker compose up -d    # Fresh start
```

## Contributing

We ❤️ contributions big or small check our [Bounty Program](https://olake.io/docs/community/issues-and-prs#goodies). As always, thanks to our amazing contributors!.

- To contribute to Olake-UI visit [CONTRIBUTING.md](CONTRIBUTING.md)
- To contribute to Olake Main Repo, visit [OLake Main Repository](https://github.com/datazip-inc/olake).
- To contribute to OLake website and [documentation](https://olake.io/docs), visit [Olake Docs Repository](https://github.com/datazip-inc/olake-docs/).
