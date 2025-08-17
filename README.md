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

## 🧊 TL;DR: OLake UI — Simplify Data Replication

> **OLake UI** is an intuitive web interface for managing data replication from databases like **PostgreSQL, MySQL, MongoDB, Oracle, and Kafka** to **Apache Iceberg** and **Amazon S3**. Built on top of the powerful OLake CLI, it offers a self-serve experience to configure, monitor, and schedule jobs in minutes.

## 🧪 Quickstart (Docker Compose)

OLake UI provides a web-based interface for managing OLake jobs, sources, destinations, and configurations. Run the entire stack (UI, backend, Temporal worker, and dependencies) using Docker Compose for a quick start.

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/) installed
- 4GB RAM and 2 CPU cores recommended
- Port `8000` available

### Steps

1. **Clone the repository:**

   ```bash
   git clone https://github.com/datazip-inc/olake-ui.git
   cd olake-ui
   ```

2. **Start the stack:**

   ```bash
   docker compose up -d
   ```

3. **Check services:**

   ```bash
   docker compose ps
   ```

4. **Access the services:**

   - **Frontend UI:** [http://localhost:8000](http://localhost:8000)
   - **Default login:** Username: `admin`, Password: `password`
   - **Make sure port `8000` is exposed and accessible**, as both the frontend and backend run on this single port.

5. **Stop the stack:**
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


### 🛠️ Creating Your First Job
1. Navigate to Jobs: Open the Jobs tab in the UI.
2. Configure Source: Add a source (e.g., PostgreSQL, MySQL) with connection details.
3. Configure Destination: Set up Apache Iceberg or Parquet with your preferred catalog (e.g., Glue, Hive).
4. Select Streams: Choose tables to sync and set the mode (Full Refresh or CDC).
5. Run Job: Name your job, set a schedule, and click Create Job.

### 🛠️ Troubleshooting

* UI not loading? Ensure port 8000 is free and Docker is running.
* Permission errors? Verify the host persistence directory is writable.
* Authentication issues? Check default credentials or custom settings in docker-compose.yml.
* View logs:

```bash
docker compose logs -f
```

* More help: Troubleshooting Guide(#troubleshooting)

### Contributing

We ❤️ contributions! Join our community and help shape OLake UI.
* Contribute to UI: See [CONTRIBUTING.md](CONTRIBUTING.md)
* Contribute to CLI: Visit [OLake Main Repository](https://github.com/datazip-inc/olake)
* Contribute to Docs: Visit [OLake Docs Repository](https://github.com/datazip-inc/olake-docs)
* Community: Join our [Slack](https://join.slack.com/t/getolake/shared_invite/zt-2utw44do6-g4XuKKeqBghBMy2~LcJ4ag) or [GitHub Discussions](https://github.com/datazip-inc/olake-ui/discussions)
