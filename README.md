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

- [Docker](https://docs.docker.com/get-docker/) installed (Docker Desktop recommended for Mac/Windows)
- [Docker Compose](https://docs.docker.com/compose/) (comes with Docker Desktop)

### Quick Start

1. **Clone the repository:**

   ```bash
   git clone https://github.com/datazip-inc/olake-ui.git
   cd olake-ui
   ```

2. **Edit persistence/config paths (required):**

   - The docker-compose.yml uses `/your/chosen/host/path/olake-data` as a placeholder for the host directory where Olake's persistent data and configuration will be stored. You **must** replace this with an actual path on your system before starting the services. You can change this by editing the `x-app-defaults` section at the top of `docker-compose.yml`:
     ```yaml
     x-app-defaults:
       host_persistence_path: &hostPersistencePath /your/host/path
     ```
   - Make sure the directory exists and is writable by the user running Docker (see how to change [file permissions for Linux/macOS](https://wiki.archlinux.org/title/File_permissions_and_attributes#Changing_permissions)).

3. **Customizing Admin User (optional):**

   The stack automatically creates an initial admin user on first startup. The default credentials are:

   - Username: "admin"
   - Password: "password"
   - Email: "test@example.com"

   To change these defaults, edit the `x-signup-defaults` section in your `docker-compose.yml`:

   ```yaml
   x-signup-defaults:
   username: &defaultUsername "your-custom-username"
   password: &defaultPassword "your-secure-password"
   email: &defaultEmail "your-email@example.com"
   ```

4. **Start all services:**

   ```bash
   docker compose up -d
   ```

5. **Check that everything is running:**

   ```bash
   docker compose ps
   ```

6. **Access the services:**

   - **Frontend UI:** [http://localhost:8000](http://localhost:8000)

7. **Stopping the stack:**
   ```bash
   docker compose down
   ```

### Notes

- The first time you run, Docker will pull all required images.
- Data and configuration are persisted in the directory you set in `docker-compose.yml` at Step 2.
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