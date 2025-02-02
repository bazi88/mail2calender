# CI/CD Setup Guide

## Overview
This document outlines the CI/CD setup for the Mail2Calendar project using GitHub Actions.

## Current Issues and Improvements Needed

1. Path Configuration Issues:
   - Golangci-lint path prefix is incorrectly set to "go8"
   - Codecov upload path is incorrect

2. Missing Features:
   - Environment variables validation
   - Proper caching strategy
   - Comprehensive testing coverage

## Required GitHub Secrets

Configure these in your repository settings (Settings > Secrets and variables > Actions):

```
DOCKERHUB_USERNAME: Your Docker Hub username
DOCKERHUB_TOKEN: Your Docker Hub access token
```

## Pipeline Stages

### 1. Lint Stage
```yaml
lint:
  name: Lint Code
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v3
    - name: Set up Go
      uses: actions/setup-go@v4
      with:
        go-version: '1.21'
    - name: Run golangci-lint
      uses: golangci/golangci-lint-action@v3
      with:
        version: latest
        args: --timeout=5m
        working-directory: .
```

### 2. Test Stage
```yaml
test:
  name: Run Tests
  runs-on: ubuntu-latest
  services:
    postgres:
      image: postgres:16
      env:
        POSTGRES_USER: postgres
        POSTGRES_PASSWORD: password
        POSTGRES_DB: mail2calendar_test
      ports:
        - 5432:5432
      options: >-
        --health-cmd pg_isready
        --health-interval 10s
        --health-timeout 5s
        --health-retries 5
```

### 3. Build Stage
```yaml
build:
  name: Build and Push Docker Image
  needs: [lint, test]
  runs-on: ubuntu-latest
  if: github.event_name == 'push' && github.ref == 'refs/heads/main'
```

## Recommended Improvements

1. **Enhanced Testing**
   Add integration and e2e tests:
   ```yaml
   - name: Run Integration Tests
     run: go test -tags=integration ./...
   
   - name: Run E2E Tests
     run: cd e2e && go test -v ./...
   ```

2. **Proper Caching**
   Implement better caching strategy:
   ```yaml
   - uses: actions/cache@v3
     with:
       path: |
         ~/.cache/go-build
         ~/go/pkg/mod
       key: ${{ runner.os }}-go-${{ hashFiles('**/go.sum') }}
   ```

## Setup Instructions

1. Create `.github/workflows` directory if it doesn't exist:
   ```bash
   mkdir -p .github/workflows
   ```

2. Copy the CI/CD configuration into `.github/workflows/ci-cd.yml`

3. Configure GitHub Secrets:
   - Go to repository Settings > Secrets and variables > Actions
   - Add all required secrets listed above

4. Configure Branch Protection:
   - Go to Settings > Branches
   - Add rule for `main` branch
   - Enable required status checks
   - Enable required reviews

## Best Practices

1. **Version Control**
   - Use semantic versioning for releases
   - Tag Docker images with both version and latest tags
   - Keep consistent branch naming (feature/, bugfix/, etc.)

2. **Security**
   - Regular dependency updates
   - Scan Docker images for vulnerabilities
   - Use minimal base images

3. **Monitoring**
   - Add status badges to README
   - Monitor build success rates
   - Track test coverage
   - Alert on pipeline failures

## Troubleshooting

Common issues and solutions:

1. **Docker Build Fails**
   - Check Dockerfile syntax
   - Verify build context
   - Review build cache configuration

2. **Test Failures**
   - Check database connection settings
   - Verify environment variables
   - Review test logs for timeout issues

## Maintenance

Regular maintenance tasks:

1. Update GitHub Actions versions
2. Review and update dependencies
3. Validate security configurations
4. Clean up old Docker images
5. Review and optimize pipeline performance