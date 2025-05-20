# Olake-App

<h1 align="center" style="border-bottom: none">
    <a href="https://datazip.io/olake" target="_blank">
        <img alt="olake" src="https://github.com/user-attachments/assets/d204f25f-5289-423c-b3f2-44b2194bdeaf" width="100" height="100"/>
    </a>
    <br>OLake
</h1>

<p align="center">Fastest open-source tool for replicating Databases to Apache Iceberg or Data Lakehouse. ⚡ Efficient, quick and scalable data ingestion for real-time analytics. Starting with MongoDB. Visit <a href="https://olake.io/" target="_blank">olake.io/docs</a> for the full documentation, and benchmarks</p>

<p align="center">
    <a href="https://github.com/datazip-inc/olake-app/issues"><img alt="GitHub issues" src="https://img.shields.io/github/issues/datazip-inc/olake"/></a> <a href="https://olake.io/docs"><img alt="Documentation" height="22" src="https://img.shields.io/badge/view-Documentation-blue?style=for-the-badge"/></a>
    <a href="https://join.slack.com/t/getolake/shared_invite/zt-2utw44do6-g4XuKKeqBghBMy2~LcJ4ag"><img alt="slack" src="https://img.shields.io/badge/Join%20Our%20Community-Slack-blue"/></a>
</p>

## Overview

Olake-App is built on top of Olake CLI to execute commands via UI.

- [UI Readme](/olake_frontend/README.md)
- [Server Readme](/server/README.md)
- [UI Figma Design](https://www.figma.com/design/FwLnU97I8LjtYNREPyYofc/Olake-Design-Community?node-id=1-46&p=f&t=y3BIsLTUaXhHwYLG-0)
- [Contributor Guidlines](/CONTRIBUTING.md)
- [API Contracts](/api-contract.md)

## Contributing

We ❤️ contributions big or small check our [Bounty Program](https://olake.io/docs/community/issues-and-prs#goodies). As always, thanks to our amazing contributors!.

- To contribute to Olake-App visit [CONTRIBUTING.md](CONTRIBUTING.md)
- To contribute to Olake Main Repo, visit [OLake Main Repository](https://github.com/datazip-inc/olake).
- To contribute to OLake website and documentation (olake.io), visit [Olake Docs Repository][https://github.com/datazip-inc/olake-docs/].

## Running with Docker Compose

You can run the entire Olake stack (UI, backend, Temporal worker, Temporal services, and dependencies) using Docker Compose. This is the recommended way to get started for local development or evaluation.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed (Docker Desktop recommended for Mac/Windows)
- [Docker Compose](https://docs.docker.com/compose/) (comes with Docker Desktop)

### Configuration: `app.conf`

You **must** provide an `app.conf` file with your backend configuration.  
This file should be placed in the directory specified by the `app_config_path` variable in your `docker-compose.yml` (see the `x-app-defaults` section).

**Example `app.conf`:**

```ini
appname = olake-server
httpport = 8080
runmode = dev
copyrequestbody = true
postgresdb = postgres://temporal:temporal@postgresql:5432/temporal
logsdir = ./logger/logs
sessionon = true
TEMPORAL_ADDRESS=temporal:7233
```

### Quick Start

1. **Clone the repository:**

   ```bash
   git clone https://github.com/datazip-inc/olake-app.git
   cd olake-app
   ```

2. **Edit persistence/config paths (optional):**

   - By default, the `docker-compose.yml` uses `/Users/macos/temp` for persistent data and config.  
     You can change this to any directory on your host by editing the `x-app-defaults` section at the top of `docker-compose.yml`:
     ```yaml
     x-app-defaults:
       host_persistence_path: &hostPersistencePath /your/host/path
       app_config_path: &appConfigVolumeDetails /your/host/path
     ```
   - Make sure the directory exists and is writable.

3. **Start all services:**

   ```bash
   docker compose up -d
   ```

4. **Check that everything is running:**

   ```bash
   docker compose ps
   ```

5. **Access the services:**

   - **Frontend UI:** [http://localhost:8000](http://localhost:8000)

6. **Stopping the stack:**
   ```bash
   docker compose down
   ```

### Notes

- The first time you run, Docker will pull all required images.
- Data and configuration are persisted in the directory you set in `docker-compose.yml`.
- The Temporal worker requires access to the Docker socket to launch containers for jobs. This is handled by the volume mount in the compose file.

### Troubleshooting

- If you see errors about file permissions, ensure your host persistence/config directory is writable by Docker.
- For more logs, use:
  ```bash
  docker-compose logs -f
  ```
- If you change the code or configuration, you may need to rebuild images:
  ```bash
  docker-compose build
  docker-compose up -d
  ```
