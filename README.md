# Requirements:
Implement simple microservice (preferably in Go). The service will provide two REST API (accept and provide JSON) endpoints with following definition:

1. storing data into DB
```bash
   POST /save -d '{
   "id": "<uuid>"
   "name": "some name",
   "email": "email@email.com",
   "date_of_birth": "2020-01-01T12:12:34+00:00"
   }' 
```

2. receiving data from DB
```bash
GET /{id}
   response:
   {
   "id": "<uuid>"
   "name": "some name",
   "email": "email@email.com",
   "date_of_birth": "2020-01-01T12:12:34+00:00"
   } 
```

# Solution
Personally according the REST API definition, I would like the "creating" and "receiving" data to/from DB endpoints to name like this:

`POST /v1/client`

`GET /v1/client/{id}`

Because that's why we have `POST/GET/PUT/DELETE` HTTP methods, to create, read, update and delete resources so we don't have to specify it in the name of the endpoint.

## PostgreSQL
*DB model*:

client
```sql
CREATE TABLE client (
    id BIGSERIAL PRIMARY KEY,
    uuid UUID UNIQUE NOT NULL,
    name TEXT NOT NULL,
    email TEXT UNIQUE NOT NULL,
    date_of_birth TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);
```

# ENV Variables
```bash
CONFIG_HTTP_LISTEN_PORT: 3000
APP_ENV: local
CONFIG_ALLOW_ORIGINS: "*"
CONFIG_HTTP_READ_TIMEOUT: 5s
CONFIG_HTTP_WRITE_TIMEOUT: 15s
CONFIG_HTTP_SHUTDOWN_TIMEOUT: 30s
CONFIG_HEALTH_CHECK_TIMEOUT: 5s
CONFIG_TIMEZONE: Europe/Warsaw
CONFIG_APP_NAME: whalebone_clients

# LOGGER
CONFIG_LOG_LEVEL: debug
CONFIG_LOG_DEVEL_MODE: true

# POSTGRESQL
CONFIG_DATABASE_HOST: postgres
CONFIG_DATABASE_PORT: 5432
CONFIG_DATABASE_USER: postgres
CONFIG_DATABASE_PASSWORD: postgres
CONFIG_DATABASE_NAME: whalebone-clients
CONFIG_DATABASE_POOL_MAX_CONN_LIFETIME: 50s
CONFIG_DATABASE_POOL_MAX_CONN_IDLE_TIME: 50s
CONFIG_DATABASE_QUERY_TIMEOUT: 30s
CONFIG_DATABASE_POOL_MAX_CONNS: 100
CONFIG_DATABASE_POOL_MIN_CONNS: 1
CONFIG_DATABASE_POOL_HEALTH_CHECK_PERIOD: 5s
```

# Run App locally

Every dependency (postgres, go app) is dockerized with its health check and with exposed ports to be accessible from outside. Check [docker-compose](./docker-compose.yaml).

For accessing the Go App REST API, check the swagger UI [API Docs](http://localhost:59110/api/indexlhtml).

Before using the REST API you need to run migrations first, see [migrations up](#run-migrations-up)

## Migrations
### Create New Migration
```makefile
make create-migration name=<migration_name>
```

### [Run Migrations Up](#run-migrations-up)
```makefile
make migration-up
```

### Run All Migrations Down
```makefile
make migration-down-all
```

### Run Migration Down By One
```makefile
make migration-down-by-one
```

## API Docs
- implemented with Swagger UI
- [API Docs](http://localhost:59110/api/indexlhtml)

## Observability and Health Checks
- [Health Check Readiness Probe](http://localhost:59110/health/readiness)
- [Health Check Liveness Probe](http://localhost:59110/health/liveness)
- [Metrics](http://localhost:59110/metrics)
