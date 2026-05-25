# Terraform Provider for Dokploy

Fork of [ahmedali6/terraform-provider-dokploy](https://github.com/ahmedali6/terraform-provider-dokploy) with bug fixes and new resources.

A Terraform provider for [Dokploy](https://dokploy.com/), allowing you to manage Dokploy resources using Infrastructure as Code.

## Changes from upstream

- **`previewLabels`** — fixed type mismatch (string → json.RawMessage, handles null/array)
- **`cpuLimit`/`cpuReservation`** — changed from Int64 to String (supports float values like "0.5")
- **`traefik_config`** — now Computed, never wiped on apply, always read after create
- **`dockerfile_path`/`docker_context_path`** — removed forced defaults that broke Dokploy build paths
- **GitHub fields on import** — fixed null vs unknown check so fields populate correctly
- **Preview deployment fields** — now included in application update payload
- **`dokploy_schedule`** — new resource for managing Dokploy scheduled tasks

## Features

### Resources
- **Projects** - Organize your infrastructure
- **Environments** - Manage deployment environments
- **Applications** - Deploy applications from Git (GitHub, GitLab, Bitbucket, Gitea, custom Git)
- **Databases** - Provision databases (PostgreSQL, MySQL, MongoDB, MariaDB, Redis)
- **Compose** - Deploy Docker Compose stacks
- **Domains** - Configure domains and routing
- **Environment Variables** - Manage application configuration (bulk key-value map)
- **Schedules** - Manage cron-based scheduled tasks
- **SSH Keys** - Handle Git repository authentication
- **Mounts** - Configure volume, bind, and file mounts
- **Ports** - Manage port mappings for non-HTTP services
- **Redirects** - Set up URL redirects and rewrites
- **Registry** - Configure Docker registry credentials

### Data Sources
- **GitHub Providers** - Query configured GitHub integrations
- **Servers** - Retrieve information about Dokploy servers

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- A [Dokploy](https://dokploy.com/) instance with API access

## Using the Provider

### Installation

```hcl
terraform {
  required_providers {
    dokploy = {
      source  = "automindz-solutions/dokploy"
      version = "~> 0.8"
    }
  }
}

provider "dokploy" {
  host    = "https://your-dokploy-instance.com/api"
  api_key = var.dokploy_api_key
}
```

**Note:** The `host` must include `/api` — the provider does not append it automatically.

### Quick Example

```hcl
resource "dokploy_project" "main" {
  name        = "my-project"
  description = "My application infrastructure"
}

resource "dokploy_environment" "production" {
  name       = "production"
  project_id = dokploy_project.main.id
}

resource "dokploy_application" "web" {
  name           = "web-app"
  environment_id = dokploy_environment.production.id

  source_type       = "github"
  github_id         = var.github_provider_id
  repository        = "my-repo"
  owner             = "my-org"
  branch            = "main"
  build_path        = "/app/"
  github_repository = "my-repo"
  github_owner      = "my-org"
  github_branch     = "main"
  github_build_path = "/app/"

  build_type      = "dockerfile"
  auto_deploy     = true
  create_env_file = true
}

resource "dokploy_environment_variables" "web" {
  application_id = dokploy_application.web.id
  variables = {
    DATABASE_URL = var.database_url
    API_KEY      = var.api_key
  }
}

resource "dokploy_domain" "web" {
  application_id   = dokploy_application.web.id
  host             = "app.example.com"
  port             = 3000
  https            = true
  certificate_type = "letsencrypt"
}

resource "dokploy_schedule" "weekly_job" {
  name            = "Weekly cleanup"
  cron_expression = "0 2 * * 0"
  command         = "cd /app && python cleanup.py"
  shell_type      = "bash"
  schedule_type   = "application"
  application_id  = dokploy_application.web.id
  enabled         = true
  timezone        = "UTC"
}
```

### Importing Existing Resources

```bash
terraform import dokploy_application.web <application-id>
terraform import dokploy_domain.web "application:<app-id>:<domain-id>"
terraform import dokploy_schedule.weekly_job <schedule-id>
```

**Important:** Set `github_repository`, `github_owner`, `github_branch`, `github_build_path` in your config alongside `repository`, `owner`, `branch`, `build_path` — the provider uses both field sets.

## Building The Provider

```shell
git clone https://github.com/automindz-solutions/terraform-provider-dokploy.git
cd terraform-provider-dokploy
go build .
```

## License

This provider is published under the same license as the original project.
