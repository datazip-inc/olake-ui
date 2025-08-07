

# Contributing to Olake-UI

Thanks for taking the time and for your help in improving this project!

## Table of Contents
- [Olake-UI Contributor Agreement](#olake-ui-contributor-agreement)
- [How You Can Contribute to Olake-UI](#how-you-can-contribute-to-olake-ui)
- [Step-by-Step Guide](#step-by-step-guide)
- [Submitting a Pull Request](#submitting-a-pull-request)
- [Committing](#committing)
- [Installing and Setting Up Olake-UI](#installing-and-setting-up-olake-ui)
- [Development Setup](#development-setup)
- [Getting Help](#getting-help)

## Olake-UI Contributor Agreement

To contribute to this project, we need you to sign the [**Contributor License Agreement (“CLA”)**][CLA] for the first commit you make. By agreeing to the [**CLA**][CLA], we can add you to the list of approved contributors and review the changes proposed by you.

## How You Can Contribute to Olake-UI

You can contribute to the open-source Olake-UI project in several ways. View our [**Issues Page**](https://github.com/datazip-inc/olake-ui/issues) to see all open issues. If you encounter a bug or have an improvement suggestion, you can [**submit an issue**](https://github.com/datazip-inc/olake-ui/issues/new) describing your proposed change.

Contributions can include creating UI components, Server APIs, and Temporal Integrations. For more information on the different ways to contribute, join our [**Slack**](https://join.slack.com/t/getolake/shared_invite/zt-2usyz3i6r-8I8c9MtfcQUINQbR7vNtCQ) channel to chat with us.

## Step-by-Step Guide

1. **Fork the repository**  
   Go to [https://github.com/datazip-inc/olake-ui](https://github.com/datazip-inc/olake-ui) and click **"Fork"**.

2. **Clone your fork locally**  
   ```bash
   git clone https://github.com/<your-username>/olake-ui.git
   cd olake-ui
   ```

3. **Create a new branch for your changes**  
   ```bash
   git checkout -b feat/my-feature-name
   ```

4. **Make your changes**  
   - Backend: Work in the `server/` folder.  
   - Frontend: Work in the `ui/` folder.

5. **Run pre-commit checks**  
   ```bash
   make pre-commit
   ```

6. **Commit your changes**  
   ```bash
   git add .
   git commit -m "feat: add <your-feature>"
   ```

7. **Push to your fork**  
   ```bash
   git push origin feat/my-feature-name
   ```

8. **Open a Pull Request**  
   - Go to your fork on GitHub.  
   - Click "Compare & pull request."  
   - Set the target branch to `main`.  
   - Add a clear title and description.  
   - Submit!

## Submitting a Pull Request

The type of change you make will dictate which repositories you need to make pull requests for. If you have questions, reach out on our [**Slack**](https://join.slack.com/t/getolake/shared_invite/zt-2usyz3i6r-8I8c9MtfcQUINQbR7vNtCQ) channel.

To contribute a new UI component, for example, create a pull request (PR) with the following steps:  
- Provide a clear and concise PR title.  
- Write a detailed and descriptive PR description.  
- Request a code review from the maintainers.

## Committing

We prefer squash or rebase commits so that all changes from a branch are committed to `main` as a single commit. All pull requests are squashed when merged, but rebasing prior to merge gives you better control over the commit message. Run the `make pre-commit` command before committing.

## Installing and Setting Up Olake-UI

To contribute to this project, you may need to install Olake-UI on your machine. Follow our [README](/README.md) to set up Olake-UI quickly.

## Development Setup

To contribute code, you’ll need the full development environment running locally. For complete instructions, see the [README](/README.md). Below is a quick summary:

### Quick Start Summary
**Required Configurations:**  
- In the frontend `.env` file: `VITE_IS_DEV=true`  
- In the backend config (`server/conf/app.conf`): `runmode = dev`  

➡️ To set both configurations automatically, run:

```bash
make init-config
```
>This will create the required .env and app.conf files with default values, including a local PostgreSQL connection string.

### Prerequisites
- **Go** ≥ 1.20  
- **Node.js** and **pnpm**  
- **Docker** & **Docker Compose**  
- **Beego CLI**  
  Install Beego CLI:  
  ```bash
  go install github.com/beego/bee/v2@latest
  ```
## Start Temporal Services
Make sure Docker is running:

```bash
make start-temporal
```

> Starts `temporal`, `temporal-ui`, and `temporal-admin-tools` using Docker Compose inside the `server/` directory.

---

##  Start Temporal Worker (Go)
In a separate terminal tab/window:
```bash
make start-temporal-server
```
> Runs the Temporal worker from `server/cmd/temporal-worker/main.go`.

---

## Start Backend Server (Beego)
In a separate terminal tab/window:
```bash
make start-backend
```

> Runs the Beego backend server with live reload using `bee run`.

---

## Start Frontend Dev Server (Vite + React)
In a separate terminal tab/window:
```bash
make start-frontend
```

## Sign up
```bash
make create-user username=admin password=admin123 email=admin@example.com
```

> Installs frontend dependencies using `pnpm` and runs the development server.

### Access Local Services
- **Frontend**: http://localhost:5173  
- **Backend API**: http://localhost:8000  
- **Temporal UI**: http://localhost:8080  

## Getting Help

For any questions, concerns, or queries, start by asking on our [**Slack**](https://join.slack.com/t/getolake/shared_invite/zt-2usyz3i6r-8I8c9MtfcQUINQbR7vNtCQ) channel.

### We look forward to your feedback on improving this project!

[CLA]: https://docs.google.com/forms/d/e/1FAIpQLSdze2q6gn81fmbIp2bW5cIpAXcpv7Y5OQjQyXflNvoYWiO4OQ/viewform

