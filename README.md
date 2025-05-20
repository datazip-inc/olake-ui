# olake-frontend
Frontend &amp; BFF (Backend for frontend) for Olake. This includes the UI code and backend code for storing the configuration of sync and orchestrating it.

## Contributing
Checkout - [frontend initial setup branch](https://github.com/datazip-inc/olake-frontend/tree/feat/frontend_initial_setup_and_BFF) to see the work in progress and the related issues [here.](https://github.com/datazip-inc/olake-frontend/issues)
Also checkout [CONTRIBUTING.MD](https://github.com/zriyanshdz/olake-frontend/blob/feat/frontend_initial_setup_and_BFF/olake_frontend/CONTRIBUTING.md) to get the guidelines.


The changes will then get merged to staging branch and then to master branch.

Checkout the [Figma Design](https://www.figma.com/design/FwLnU97I8LjtYNREPyYofc/Olake%2FDesign%2FCommunity?m=auto&t=3T4OEwuQNOxoE3zm-1) for OLake frontend that is being developed, for you to get better sense and contribute to the issues we have created.
 
## Running with Docker Compose

You can run the entire Olake stack (UI, Backend, Temporal worker, Temporal services, and dependencies) using Docker Compose. This is the recommended way to get started for local development or evaluation.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) installed and running (Docker Desktop recommended for Mac/Windows)
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

2. **Edit persistence/config paths (required):**
    - By default, the `docker-compose.yml` uses `/Users/macos/temp` for persistent data and config.  
      You can change this to any directory on your host by editing the `x-app-defaults` section at the top of `docker-compose.yml`:
      ```yaml
      x-app-defaults:
        host_persistence_path: &hostPersistencePath /Users/macos/temp
        app_config_path: &appConfigVolumeDetails    /Users/macos/temp
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
    - **Frontend UI:** [http://localhost:5173](http://localhost:5173)
    - **Backend API:** [http://localhost:8080](http://localhost:8080)
    - **Temporal UI:** [http://localhost:8081](http://localhost:8081)

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