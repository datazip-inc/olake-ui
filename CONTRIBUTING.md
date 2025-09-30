# Contributing to Olake-UI

## 🎃 Hacktoberfest 2025 @ OLake

OLake is officially open for Hacktoberfest contributions! 🚀

If you're participating in Hacktoberfest, look out for any issues labeled:

- `hacktoberfest`
- `good first issue`

These are designed to help new contributors get started quickly.
We welcome everything — bug fixes, documentation updates, tests, or feature enhancements.

👉 Check our open issues [here](https://github.com/datazip-inc/olake-ui/issues)

Let's hack, learn, and grow together this Hacktoberfest. Happy contributing & happy engineering!

---

Thank you for your interest in contributing to Olake-UI — we appreciate your support in making this project better!

## Olake-UI Contributor Agreement

To contribute to this project, we need you to sign the [**Contributor License Agreement (“CLA”)**][CLA] on the first commit you make. By agreeing to the [**CLA**][CLA], we can add you to the list of approved contributors and review the changes proposed.

## How to Contribute

- Check the [**Issues Page**](https://github.com/datazip-inc/olake-ui/issues) for open issues.
- Submit bugs or suggestions via [**New Issue**](https://github.com/datazip-inc/olake-ui/issues/new).
- Contribute UI components, APIs, or integrations.
- Join our [**Slack**](https://join.slack.com/t/getolake/shared_invite/zt-2usyz3i6r-8I8c9MtfcQUINQbR7vNtCQ) for questions.

## Contribution Steps

1. **Fork the Repository**  
   Fork at [https://github.com/datazip-inc/olake-ui](https://github.com/datazip-inc/olake-ui).

2. **Clone Your Fork**  
   ```bash
   git clone https://github.com/<your-username>/olake-ui.git
   cd olake-ui
   ```

3. **Create a Branch**  
   ```bash
   git checkout -b feat/my-feature
   ```

4. **Make Changes**  
   - Backend: Edit in `server/`.  
   - Frontend: Edit in `ui/`.

5. **Run Pre-Commit Checks**  
   ```bash
   make pre-commit
   ```

6. **Commit Changes**  
   ```bash
   git add .
   git commit -m "feat: add <feature>"
   ```

7. **Push to Your Fork**  
   ```bash
   git push origin feat/my-feature
   ```

8. **Submit a Pull Request**  
   - Go to your fork on GitHub.  
   - Click "Compare & pull request."  
   - Target the `staging` branch.  
   - Add a clear title and description.  
   - Submit.

## Pull Request Guidelines

- Use a clear PR title and description.
- Request a review from maintainers.
- Commits are squashed on merge, but you can rebase for better control.

## Development Setup

### Quick Start Summary
**Required Configurations:**  
If you want to make changes to environment variables or configuration values, you can modify them in the Makefile. Otherwise, they are set to default values suitable for local development.
You can also change the local PostgreSQL connection or any other settings based on your specific requirements.

### Prerequisites
- **Go** ≥ 1.20  
- **Node.js** and **pnpm**  
- **Docker** and **Docker Compose**  
- **Beego CLI**:  
  ```bash
  go install github.com/beego/bee/v2@latest
  ```

### Start Services
1. **Temporal Services**  
Make sure Docker is running:
   ```bash
   make start-temporal
   ```

2. **Temporal Worker**  
In a separate terminal tab/window:
   ```bash
   make start-temporal-server
   ```
>>Runs the Temporal worker from server/cmd/temporal-worker/main.go.

3. **Backend Server**  
In a separate terminal tab/window:
   ```bash
   make start-backend
   ```
>>Runs the backend server from server/main.go.

4. **Frontend Server**  
In a separate terminal tab/window:
   ```bash
   make start-frontend
   ```
>>Installs frontend dependencies using pnpm and runs the development server.

5. **Create User**  
   ```bash
   make create-user username=olake password=olake123 email=olake@example.com
   ```
>>Creates a new user with the specified username, password, and email.

### Access Services
- Frontend: http://localhost:5173  
- Backend API: http://localhost:8000  
- Temporal UI: http://localhost:8081  

## Getting Help

Ask questions on our [**Slack**](https://join.slack.com/t/getolake/shared_invite/zt-2usyz3i6r-8I8c9MtfcQUINQbR7vNtCQ).

[CLA]: https://docs.google.com/forms/d/e/1FAIpQLSdze2q6gn81fmbIp2bW5cIpAXcpv7Y5OQjQyXflNvoYWiO4OQ/viewform

</xaiArtifact>