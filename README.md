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

Olake-UI is built on top of Olake CLI to execute commands via UI.

- [UI Readme](/olake_frontend/README.md)
- [Server Readme](/server/README.md)
- [UI Figma Design](https://www.figma.com/design/FwLnU97I8LjtYNREPyYofc/Olake-Design-Community?node-id=1-46&p=f&t=y3BIsLTUaXhHwYLG-0)
- [Contributor Guidlines](/CONTRIBUTING.md)
- [API Contracts](/api-contract.md)

## Contributing

We ❤️ contributions big or small check our [Bounty Program](https://olake.io/docs/community/issues-and-prs#goodies). As always, thanks to our amazing contributors!.

- To contribute to Olake-UI visit [CONTRIBUTING.md](CONTRIBUTING.md)
- To contribute to Olake Main Repo, visit [OLake Main Repository](https://github.com/datazip-inc/olake).
- To contribute to OLake website and documentation (olake.io), visit [Olake Docs Repository][https://github.com/datazip-inc/olake-docs/].

## Running with Docker Compose

This Docker Compose setup provides a comprehensive environment(UI, backend, Temporal worker, Temporal services, and dependencies) for demonstrating and exploring Olake's capabilities. This is the recommended way to get started for local development or evaluation.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed and running (Docker Desktop recommended for Mac/Windows)
- [Docker Compose](https://docs.docker.com/compose/) (comes with Docker Desktop)

### Quick Start

1. **Clone the repository:**

   ```bash
   git clone https://github.com/datazip-inc/olake-ui.git
   cd olake-ui
   ```

2. **Start all services:**

   ```bash
   docker compose up -d
   ```

3. **Check that everything is running:**

   ```bash
   docker compose ps
   ```

4. **Access the services:**

   - **Frontend UI:** [http://localhost:8000](http://localhost:8000)
   - **Default login:** Username: `admin`, Password: `password`
   - **Make sure port `8000` is exposed and accessible**, as both the frontend and backend run on this single port.

5. **Stopping the stack:**
   ```bash
   docker compose down
   ```

### Notes

- On first run, docker will pull all required images.
- Data and configuration are persisted in the directory set in `docker-compose.yml`.
- The Temporal worker requires access to the Docker socket to launch containers for jobs. This is handled by the volume mount in the compose file.

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

# 🛠️ Development Setup

This guide helps you set up the full Olake development environment including backend, frontend, and Temporal services.

---

## ✅ Prerequisites

- [Go](https://go.dev/doc/install) (≥ 1.20)
- [Node.js](https://nodejs.org/) and [pnpm](https://pnpm.io/installation)
- [Docker](https://www.docker.com/)
- [BeeGo CLI](https://beego.me/docs/install/bee.md)  
  Install it with:

  ```bash
  go install github.com/beego/bee/v2@latest
  ```

---

## 🌱 Environment Variables

Before running the project, you can optionally set the following environment variable for development:

```bash
export IS_DEV=true
```

- If `IS_DEV=true`, the backend will **proxy requests to the frontend dev server** at `http://localhost:5173`.
- If not set or false, the backend will serve the frontend from the built files (`/opt/frontend/dist`).

---

## 📦 1. Clone the Repository

```bash
git clone https://github.com/your-org/olake-frontend.git
cd olake-frontend
```

---

## 🐳 2. Start Temporal Services

Make sure Docker is running:

```bash
make start-temporal
```

> Starts `temporal`, `temporal-ui`, and `temporal-admin-tools` using Docker Compose inside the `server/` directory.

---

## ⚙️ 3. Start Temporal Worker (Go)

In a separate terminal tab/window:

```bash
make start-temporal-server
```

> Runs the Temporal worker from `server/cmd/temporal-worker/main.go`.

---

## 🔙 4. Start Backend Server (Beego)

```bash
make start-backend
```

> Runs the Beego backend server with live reload using `bee run`.

---

## 🌐 5. Start Frontend Dev Server (Vite + React)

```bash
make start-frontend
```

> Installs frontend dependencies using `pnpm` and runs the development server.

---

## 🔗 6. Access the Services

- **Frontend UI:** [http://localhost:5173](http://localhost:5173)
- **Backend API:** [http://localhost:8000](http://localhost:8000)
- **Temporal UI:** [http://localhost:8080](http://localhost:8080)

---
