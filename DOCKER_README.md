# Docker Setup for Clove APIs

This Docker setup includes the following services:
- **PostgreSQL**: Database server
- **api-auth**: Authentication service (HTTP: 8081, gRPC: 50051)
- **api-portal-users**: User management service (HTTP: 8082)
- **api-portal-organizations**: Organization management service (HTTP: 8083)
- **nginx**: Reverse proxy for routing (HTTP: 80)

## Prerequisites

- Docker
- Docker Compose

## Quick Start

1. Build and start all services:
```bash
docker-compose up --build
```

2. Start services in detached mode:
```bash
docker-compose up -d
```

3. View logs:
```bash
docker-compose logs -f
```

4. Stop services:
```bash
docker-compose down
```

5. Stop services and remove volumes:
```bash
docker-compose down -v
```

## Accessing the APIs

All APIs are accessible through nginx at `http://localhost`:

- **Auth API**: `http://localhost/api/auth`
- **Users API**: `http://localhost/api/users`
- **Organizations API**: `http://localhost/api/organizations`

### Direct Access (without nginx)

You can also access the APIs directly:

- **Auth API**: `http://localhost:8081/api/auth`
- **Users API**: `http://localhost:8082/api/users`
- **Organizations API**: `http://localhost:8083/api/organizations`

### gRPC Access

The auth service gRPC endpoint is available at:
- `localhost:50051`

## Database Connection

PostgreSQL is configured with:
- **Host**: `postgres` (within Docker network) or `localhost:5432` (from host machine)
- **Database**: `clovedb`
- **Username**: `cloveuser`
- **Password**: `clovepassword`

Connection string used by services:
```
postgres://cloveuser:clovepassword@postgres:5432/clovedb?sslmode=disable
```

## Database Migrations

Migrations are automatically run when the postgres container is first created. The migrations are located in `shared/db/migrations/` and are mounted to the postgres container.

**Important**: Migrations only run on initial database creation. If you need to run migrations on an existing database:

1. Stop and remove the postgres container and volume:
```bash
docker-compose down -v
```

2. Start fresh (migrations will run automatically):
```bash
docker-compose up --build
```

### Manually Running Migrations

If you need to run migrations manually on an existing database:

```bash
docker-compose exec postgres psql -U cloveuser -d clovedb -f /docker-entrypoint-initdb.d/0001_init.sql
```

Or to run all migration files:
```bash
for file in ./shared/db/migrations/*.sql; do
  docker-compose exec -T postgres psql -U cloveuser -d clovedb < "$file"
done
```

## Configuration

### Environment Variables

Each service has a `.env` file in its directory with the following variables:

**api-auth/.env**:
```
DATABASE_URL=postgres://cloveuser:clovepassword@postgres:5432/clovedb?sslmode=disable
SIGV4_MASTER_SECRET=ac983692bf1eb2fb4d32d41d5d518f9f55b867b45aad1df9a7519d7b633f00be
```

**Note**: The `SIGV4_MASTER_SECRET` is a 256-bit (32 byte) hex-encoded value used to deterministically derive SigV4 secret keys from access keys using HMAC-SHA256. This enables stateless credential verification without storing secrets in the database.

**api-portal-users/.env**:
```
DATABASE_URL=postgres://cloveuser:clovepassword@postgres:5432/clovedb?sslmode=disable
AUTH_CONNECTION=api-auth:50051
```

**api-portal-organizations/.env**:
```
DATABASE_URL=postgres://cloveuser:clovepassword@postgres:5432/clovedb?sslmode=disable
AUTH_CONNECTION=api-auth:50051
```

### Modifying Database Credentials

To change database credentials, update:
1. The `postgres` service environment variables in `docker-compose.yml`
2. The `DATABASE_URL` in each API's environment section in `docker-compose.yml`
3. The `.env` files in each API directory

## Development

### Rebuilding a Specific Service

```bash
docker-compose up -d --build api-auth
docker-compose up -d --build api-portal-users
docker-compose up -d --build api-portal-organizations
```

### Viewing Logs for a Specific Service

```bash
docker-compose logs -f api-auth
docker-compose logs -f api-portal-users
docker-compose logs -f api-portal-organizations
docker-compose logs -f postgres
docker-compose logs -f nginx
```

### Executing Commands in a Container

```bash
docker-compose exec api-auth sh
docker-compose exec postgres psql -U cloveuser -d clovedb
```

### Adding New Migrations

To add new migration files:

1. Create a new `.sql` file in `shared/db/migrations/` with a numbered prefix (e.g., `0002_add_users_table.sql`)
2. For existing databases, run the migration manually:
```bash
docker-compose exec postgres psql -U cloveuser -d clovedb -f /docker-entrypoint-initdb.d/0002_add_users_table.sql
```
3. For fresh databases, the migration will run automatically on first startup

## Troubleshooting

### Services Not Starting

Check logs:
```bash
docker-compose logs
```

### Database Connection Issues

1. Ensure postgres is healthy:
```bash
docker-compose ps
```

2. Check postgres logs:
```bash
docker-compose logs postgres
```

3. Test database connection:
```bash
docker-compose exec postgres psql -U cloveuser -d clovedb
```

### Port Conflicts

If you have port conflicts, modify the port mappings in `docker-compose.yml`:
```yaml
ports:
  - "8081:8080"  # Change 8081 to another available port
```

### Rebuilding from Scratch

```bash
docker-compose down -v
docker-compose build --no-cache
docker-compose up
```

## Network Architecture

All services communicate through a Docker bridge network called `clove-network`. This allows:
- Services to resolve each other by container name
- Isolated network environment
- Secure inter-service communication

## Health Checks

- **Postgres**: Includes a health check that verifies database connectivity
- **nginx**: Provides a `/health` endpoint that returns "healthy"

Test nginx health:
```bash
curl http://localhost/health
```
