# Contributing to Olake Frontend

Thank you for your interest in contributing to Olake Frontend! This document provides guidelines and instructions for contributing to this project.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Setting Up the Development Environment](#setting-up-the-development-environment)
- [Development Workflow](#development-workflow)
  - [Branching Strategy](#branching-strategy)
  - [Commit Guidelines](#commit-guidelines)
  - [Pull Request Process](#pull-request-process)
- [Coding Standards](#coding-standards)
  - [TypeScript Guidelines](#typescript-guidelines)
  - [React Best Practices](#react-best-practices)
  - [Styling Guidelines](#styling-guidelines)
- [Testing](#testing)
- [Documentation](#documentation)
- [Issue Reporting](#issue-reporting)
- [Feature Requests](#feature-requests)

## Code of Conduct

We expect all contributors to follow our Code of Conduct. Please be respectful and considerate of others when contributing to this project.

## Getting Started

### Prerequisites

- Node.js (LTS version)
- pnpm (for package management)
- Git

### Setting Up the Development Environment


1. Clone the repository:
   ```bash
   git clone https://github.com/datazip-inc/olake-frontend.git
   cd olake_frontend
   ```

#### Common Steps

After setting up your local repository using either method above:

4. Install dependencies:
   ```bash
   pnpm install
   ```
5. Start the development server:
   ```bash
   pnpm dev
   ```

## Development Workflow

### Branching Strategy

- `master` - Production-ready code
- Feature branches - Named as `feat/your-feature-name`
- Bug fix branches - Named as `fix/bug-description`

### Commit Guidelines

We follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat:` - A new feature
- `fix:` - A bug fix
- `docs:` - Documentation changes
- `style:` - Code style changes (formatting, etc.)
- `refactor:` - Code changes that neither fix bugs nor add features
- `perf:` - Performance improvements
- `test:` - Adding or fixing tests
- `chore:` - Changes to the build process or auxiliary tools

Example:

```
feat: add user authentication component
```

### Pull Request Process

1. Update your feature branch with the latest changes from `staging`:
   ```bash
   git checkout staging
   git pull upstream staging
   git checkout feature/your-feature-name
   git rebase staging
   ```
2. Ensure your code passes all linting and formatting checks:
   ```bash
   pnpm lint
   pnpm format
   ```
3. Push your branch to your fork:
   ```bash
   git push origin feature/your-feature-name
   ```
4. Create a pull request against the `staging` branch of the original repository
5. Ensure your PR description clearly describes the changes and references any related issues
6. Wait for code review and address any feedback

## Coding Standards

### TypeScript Guidelines

- Use TypeScript for all new code
- Define proper interfaces and types for all components, functions, and variables
- Avoid using `any` type when possible
- Use type inference where appropriate

### React Best Practices

- Use functional components with hooks
- Keep components small and focused on a single responsibility
- Use proper component composition
- Follow the React hooks rules
- Use React Router for navigation
- Implement proper error handling

### Styling Guidelines

- Use Tailwind CSS for styling
- Follow the project's design system
- Ensure responsive design for all components
- Use Ant Design components when appropriate

## Code Quality

Before submitting your code, ensure it meets our quality standards:

1. Format your code:
   ```bash
   pnpm format
   ```
2. Check for linting issues:
   ```bash
   pnpm lint
   ```
3. Fix linting issues:
   ```bash
   pnpm lint:fix
   ```

## Testing

- Write tests for new features and bug fixes
- Ensure all tests pass before submitting a PR
- Aim for good test coverage

## Documentation

- Document components, functions, and complex logic
- Update the README.md if necessary
- Add JSDoc comments to functions and components

## Issue Reporting

When reporting issues, please use the issue template and include:

- A clear and descriptive title
- Steps to reproduce the issue
- Expected behavior
- Actual behavior
- Screenshots if applicable
- Environment information (browser, OS, etc.)

## Feature Requests

For feature requests, please provide:

- A clear and descriptive title
- Detailed description of the feature
- Any relevant mockups or examples
- Rationale for why this feature would be beneficial

Thank you for contributing to Olake Frontend!
