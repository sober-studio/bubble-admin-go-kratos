# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Bubble Admin is a Go-based admin dashboard backend built on [Kratos](https://go-kratos.dev/) microservices framework. It provides JWT authentication, RBAC permissions, file storage (OSS), SMS/email services, scheduled jobs, and WebSocket support.

## Common Commands

```bash
# Start infrastructure (PostgreSQL + Redis)
make dev

# Generate all code (proto, gorm, wire)
make all

# Generate only proto files
make api

# Generate GORM models and queries
make gormgen

# Generate wire dependency injection
make generate

# Build the application
make build

# Run the service (requires -conf flag for config path)
go run ./cmd/bubble-admin-go-kratos -conf configs
```

The service runs on:
- HTTP: `8000`
- gRPC: `9000`

## Architecture

This project follows Clean Architecture with Kratos layout:

```
├── api/              # Protobuf API definitions
├── cmd/              # Application entrypoints
├── configs/          # Configuration files (config.yaml)
└── internal/
    ├── biz/          # Business logic layer (domain models, interfaces)
    ├── conf/         # Configuration definitions (protobuf)
    ├── data/         # Data access layer (DB/Redis implementations)
    ├── pkg/          # Shared packages (auth, casbin, oss, sms, etc.)
    ├── server/       # HTTP/gRPC server configuration
    └── service/      # Application services (DTO conversion, biz calls)
```

### Dependency Flow

```
API (proto) → Service → Biz → Data → Database/Redis
                ↑              ↑
            Wire DI      GORM Queries
```

### Key Components

- **Auth**: JWT-based authentication with token revocation via Redis
- **Casbin**: RBAC permission middleware for API authorization
- **Data Scope**: Tenant-based data isolation using GORM hooks
- **Provider Pattern**: PermissionProvider loads API permissions from database

## Database

- Uses PostgreSQL (configurable to MySQL)
- GORM for ORM with GORM Gen for type-safe queries
- Models in `internal/data/model/`
- AutoMigrate runs on startup in development

## Configuration

Main configuration in `configs/config.yaml`. Key sections:
- `data.database`: PostgreSQL/MySQL connection
- `data.redis`: Redis connection
- `app.auth`: JWT settings, public paths, token expiration
- `app.enable_multi_tenant`: Enable/disable multi-tenant mode

## Database Initialization

When working on features that require initial data (permissions, roles, etc.), **always check the database first** using the PostgreSQL MCP tool to understand the existing data structure before providing SQL insert statements.

### Workflow for Missing Data Issues

1. Use `mcp__postgres__query` to query existing data (e.g., `sys_permission`, `sys_role_permission` tables)
2. Compare with code to identify missing records
3. Provide targeted INSERT statements for the user to execute

### Example Queries

```sql
-- Check existing permissions
SELECT id, parent_id, name, code, type, api_path FROM sys_permission ORDER BY id;

-- Check role-permission bindings
SELECT rp.role_id, rp.permission_id, r.code as role_code, p.code as perm_code
FROM sys_role_permission rp
LEFT JOIN sys_role r ON rp.role_id = r.id
LEFT JOIN sys_permission p ON rp.permission_id = p.id;

-- Check table structure
SELECT column_name, data_type FROM information_schema.columns WHERE table_name = 'table_name';
```
