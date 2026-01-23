# Tic-Tac-Toe PWA CI/CD PoC

A proof-of-concept demonstrating a dual deployment pipeline for a Progressive Web App, producing both a Windows desktop executable and a cloud-hosted web application from a single codebase.

## Project Structure

```
├── cmd/server/          # Go backend entry point
├── internal/            # Go backend packages
│   ├── api/            # HTTP handlers
│   ├── auth/           # Authentication service
│   ├── config/         # Configuration
│   ├── database/       # Database connections
│   ├── game/           # Game logic
│   ├── matchmaking/    # Matchmaking service
│   ├── middleware/     # HTTP middleware
│   └── models/         # Data models
├── frontend/           # React TypeScript frontend
├── desktop/            # Tauri desktop application
├── docker/             # Docker configurations
├── infrastructure/
│   ├── terraform/      # AWS infrastructure as code
│   └── k8s/           # Kubernetes manifests
└── .github/workflows/  # CI/CD pipelines
```

## Tech Stack

### Backend
- Go 1.22+ with Echo framework
- PostgreSQL for persistent storage
- Redis for session management
- Server-Sent Events (SSE) for real-time updates

### Frontend
- React 18 with TypeScript
- Vite for build tooling
- Tailwind CSS for styling

### Desktop
- Tauri (Rust) for native packaging
- Embedded Go backend

### Infrastructure
- AWS (EKS, RDS, ElastiCache, ECR)
- Terraform for infrastructure as code
- Kubernetes for container orchestration

## Local Development

### Prerequisites
- Go 1.22+
- Node.js 20+
- Docker and Docker Compose

### Running with Docker Compose

```bash
# Start all services
docker compose up

# Or start only databases for local development
docker compose -f docker-compose.dev.yml up
```

### Running locally

```bash
# Start databases
docker compose -f docker-compose.dev.yml up -d

# Run backend
go run ./cmd/server

# Run frontend (in another terminal)
cd frontend
npm install
npm run dev
```

## CI/CD Pipelines

| Workflow | Trigger | Description |
|----------|---------|-------------|
| CI | PR to develop/main | Lint, test, build |
| CD Staging | Push to develop | Deploy to staging |
| CD Production | Push to main | Deploy to production |
| Manual Deploy | Manual | Deploy any branch to any environment |
| Build Desktop | Manual/Release | Build Windows installer |
| Release | Push to main | Create release from conventional commits |

## Environment Setup

### GitHub Secrets Required
- `AWS_ROLE_ARN`: IAM role ARN for OIDC authentication
- `AWS_ACCOUNT_ID`: AWS account ID

### GitHub Environments
- `staging`: Staging deployment environment
- `production`: Production deployment environment (with required reviewers)

## Deployment

### Initial Infrastructure Setup

```bash
# Bootstrap Terraform state backend
cd infrastructure/terraform/bootstrap
terraform init
terraform apply

# Deploy staging environment
cd ../environments/staging
terraform init
terraform apply

# Deploy production environment
cd ../environments/production
terraform init
terraform apply
```

### Manual Deployment

Use the "Manual Deploy" workflow in GitHub Actions to deploy any branch to any environment.

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | /api/v1/auth/register | Register new user |
| POST | /api/v1/auth/login | Login |
| POST | /api/v1/auth/logout | Logout |
| GET | /api/v1/auth/me | Get current user |
| POST | /api/v1/matchmaking/join | Join matchmaking queue |
| POST | /api/v1/matchmaking/leave | Leave matchmaking queue |
| POST | /api/v1/game/ready | Mark ready for game |
| POST | /api/v1/game/move | Make a move |
| POST | /api/v1/game/forfeit | Forfeit game |
| GET | /api/v1/game/state | Get current game state |
| GET | /api/v1/stats | Get user statistics |
| GET | /api/v1/events | SSE event stream |
| GET | /health | Health check |

## License

MIT
