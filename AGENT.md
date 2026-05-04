# AI Agent Instructions - MCOP Storage Service

> Optimized for: Cursor, Windsurf, Claude, and other AI development assistants

## Project Quick Reference

**Type**: Production Go Microservice
**Purpose**: Object storage management with S3/MinIO compatibility
**Architecture**: Clean Architecture + Dependency Injection
**Go Version**: 1.25.3

## Tech Stack

### Core Technologies

- **Language**: Go 1.25.3
- **Web Framework**: Gin (HTTP/1.1 & HTTP/2)
- **Database**: PostgreSQL with Bun ORM
- **Object Storage**: MinIO (S3-compatible)
- **Message Queue**: Apache Kafka (IBM/sarama) with SASL/TLS
- **Cache**: Redis with JSON support
- **Observability**: OpenTelemetry (traces, metrics, logs)
- **Logging**: `log/slog` (structured logging with OpenTelemetry bridge)
- **Configuration**: Environment-based with reflection
- **Validation**: go-playground/validator with i18n
- **CLI**: Cobra for commands
- **Error Tracking**: Sentry

### Key Dependencies

```go
github.com/gin-gonic/gin         // HTTP framework
github.com/uptrace/bun           // SQL ORM
github.com/minio/minio-go/v7     // S3 client
github.com/IBM/sarama            // Kafka client
github.com/redis/go-redis/v9     // Redis client
go.opentelemetry.io/otel         // Observability
github.com/spf13/cobra           // CLI framework
github.com/google/uuid           // UUID generation
github.com/getsentry/sentry-go   // Error tracking
```

## Architecture Overview

### Directory Structure

```
/app                    # Application layer
  /console             # CLI commands (Cobra-based)
  /modules             # Business modules
    /entities          # Data entities (Bun models)
    /sentry            # Sentry error tracking
    /specs             # API specifications
  /utils               # Shared utilities
/config                # Configuration structs and initialization
  /i18n                # Internationalization files
/database              # SQL migrations
/internal              # DI Layer - Infrastructure services
  /cmd                 # Command implementations
  /config              # Configuration service
  /database            # Database connection management
  /http                # HTTP server setup (CORS, pprof)
  /kafka               # Kafka service wrapper
  /log                 # Logging service (Slog + OpenTelemetry)
  /otel                # OpenTelemetry collector
  /provider            # Lifecycle management interface
  /redis               # Redis service wrapper
/routes                # HTTP route registration
```

### DI Layer (`/internal`)

The **Dependency Injection Layer** provides infrastructure services that are injected into business modules:

- All services in `/internal` are **initialized once** at startup
- Business modules receive dependencies via constructors
- Centralized lifecycle management via `provider.Close` interface

## Development Guidelines

### 1. Module Pattern

Every module follows this structure:

```go
package mymodule

import (
    "mcop/internal/config"
    "go.opentelemetry.io/otel"
    "go.opentelemetry.io/otel/trace"
)

// Module struct
type Module struct {
    tracer trace.Tracer
    Svc    *Service      // Required: Business logic
    Ctl    *Controller   // Optional: HTTP handlers
}

// Configuration
type Config struct {
    Endpoint        string `conf:"required"`
    AccessKeyId     string `conf:"required"`
    SecretAccessKey string `conf:"required"`
    BucketName      string // Optional fields don't need tag
}

// Service options
type Options struct {
    Config *config.Config[Config]
    tracer trace.Tracer
    // Injected dependencies (use interfaces from entities/inf)
    objEnt entitiesinf.ObjectEntity
    s3     *s3.Service
}

type Service struct {
    *Options
}

// Controller
type Controller struct {
    tracer trace.Tracer
    svc    *Service
}

// Constructor
func New(conf *config.Config[Config], dependencies...) *Module {
    tracer := otel.Tracer("mcop.service_name.module_name")
    svc := newService(&Options{
        Config: conf,
        tracer: tracer,
        // ... dependencies
    })
    ctl := newController(tracer, svc)

    return &Module{
        tracer: tracer,
        Svc:    svc,
        Ctl:    ctl,
    }
}
```

### 2. Configuration Management

**Environment Variable Pattern**: `{MODULE}_{FIELD}` with `__` for nesting

```bash
# Example
DATABASE_SQL__HOST=localhost
```

**In Code**:

```go
type Config struct {
    Endpoint        string `conf:"required"` // S3__ENDPOINT required
    AccessKeyId     string `conf:"required"` // S3__ACCESS_KEY_ID required
    SecretAccessKey string `conf:"required"` // S3__SECRET_ACCESS_KEY required
    BucketName      string                   // S3__BUCKET_NAME optional
}

// Usage in module
func New(conf *config.Config[Config]) *Module {
    // Access config: conf.Value.Endpoint
}
```

### 3. Database Operations

- Use Bun ORM for database operations
- **Avoid raw SQL queries** - use Bun ORM methods instead
- Implement repository pattern through interfaces in `entities/inf`
- Always use context for operations
- Implement proper transaction handling
- Use soft deletes with `deleted_at` timestamps

**Entity Interface Pattern** (Repository in `entities/inf`):

```go
type ObjectEntity interface {
    CreateObject(ctx context.Context, obj *ent.Object) (*ent.Object, error)
    GetObjectByID(ctx context.Context, id uuid.UUID) (*ent.Object, error)
    UpdateObjectStatusByID(ctx context.Context, id uuid.UUID, status ent.ObjectStatus) (*ent.Object, error)
    SoftDeleteObjectByID(ctx context.Context, id uuid.UUID) error
}
```

**Bun Model**:

```go
type Object struct {
    bun.BaseModel `bun:"table:objects,alias:obj"`

    ID        uuid.UUID      `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
    CreatedAt time.Time      `bun:"created_at,nullzero,notnull,default:current_timestamp"`
    UpdatedAt time.Time      `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
    DeletedAt *time.Time     `bun:"deleted_at,soft_delete"`

    Status    ObjectStatus   `bun:"status,notnull"`
}
```

**Raw SQL Query Guidelines:**

```go
// ❌ Bad: Raw SQL with string concatenation (SQL injection risk)
query := fmt.Sprintf("DELETE FROM objects WHERE id = '%s'", id)
db.Exec(query)

// ❌ Bad: Raw SQL without parameters
db.Exec("UPDATE objects SET status = 'active' WHERE user_id = " + userID)

// ✅ Good: Use Bun ORM methods
db.NewDelete().
    Model(&Object{}).
    Where("id = ?", id).
    Exec(ctx)

db.NewUpdate().
    Model(&Object{}).
    Set("status = ?", "active").
    Where("user_id = ?", userID).
    Exec(ctx)

// ✅ Acceptable: Complex queries with proper parameter binding
db.NewRaw(`
    SELECT o.*, u.name
    FROM objects o
    JOIN users u ON o.user_id = u.id
    WHERE o.status = ? AND o.created_at > ?
`, status, startDate).
    Scan(ctx, &results)
```

**Why avoid raw SQL?**

- Bun ORM prevents SQL injection automatically
- Type-safe queries reduce runtime errors
- Better maintainability and readability
- Automatic query building and validation

### 4. HTTP Handlers

```go
func (c *Controller) Create(ctx *gin.Context) {
    span, log := utils.LogSpanFromGin(ctx)
    span.SetAttributes(attribute.String("operation", "create_object"))

    var req CreateObjectRequest
    if err := ctx.ShouldBindJSON(&req); err != nil {
        log.Errf("Failed to bind request: %v", err)
        ctx.JSON(400, gin.H{"error": err.Error()})
        return
    }

    result, err := c.svc.CreateObject(ctx.Request.Context(), &req)
    if err != nil {
        log.Errf("Failed to create object: %v", err)
        ctx.JSON(500, gin.H{"error": err.Error()})
        return
    }

    log.Infof("Object created successfully")
    ctx.JSON(201, result)
}
```

### 5. Logging (slog)

**Structured logging** with `log/slog`:

```go
import "log/slog"
import "mcop/internal/log"

// Get logger with structured fields (standalone usage)
logger := log.With(slog.String("module", "mymodule"))

// Formatted logging
logger.Infof("Processing object: %s", objectID)
logger.Warnf("Retry attempt: %d", retryCount)
logger.Errf("Failed to upload: %v", err)

// With context for automatic trace correlation
logger.WithCtx(ctx).Infof("Request processed successfully")

// In Controllers and Services, use the logger from utils/otel
// Controller: span, log := utils.LogSpanFromGin(ctx)
// Service: ctx, span, log := otel.NewLogSpan(ctx, s.tracer, "OperationName")
```

**Log levels**: DEBUG, INFO, WARN, ERROR

### 6. OpenTelemetry Tracing

Add tracing to **all business operations**:

```go
func (s *Service) SomeOperation(ctx context.Context, userID uuid.UUID) error {
    ctx, span, log := otel.NewLogSpan(ctx, s.tracer, "SomeOperation")
    defer span.End()

    span.SetAttributes(
        attribute.String("operation.type", "business_logic"),
        attribute.String("user.id", userID.String()),
    )

    log.Infof("Starting operation for user: %s", userID)

    // Operation logic...
    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
        log.Errf("Operation failed: %v", err)
        return err
    }

    log.Infof("Operation completed successfully")
    return nil
}
```

### 7. Error Handling

- Use Sentry for production error tracking
- Implement custom error types for better categorization
- Add context to errors for better debugging
- Use appropriate HTTP status codes

```go
// Error handling with OpenTelemetry
if err != nil {
    span.RecordError(err)
    return fmt.Errorf("failed to process: %w", err)
}

// HTTP status codes
400 - Bad Request (validation errors)
404 - Not Found
500 - Internal Server Error
201 - Created (for POST)
```

### 8. Lifecycle Management

Modules needing cleanup implement `provider.Close`:

```go
import "mcop/internal/provider"

var _ provider.Close = (*Service)(nil)

func (s *Service) Close(ctx context.Context) error {
    // Cleanup: close connections, flush buffers, etc.
    return nil
}
```

**Automatic cleanup** happens in reverse order during shutdown.

### 9. Kafka Integration

- Use the built-in Kafka service (IBM/sarama)
- Implement proper serialization (JSON)
- Handle consumer groups with work handlers
- Use SSL/TLS for production (certificate-based)
- Implement proper error handling and retries

```go
// Producer pattern
err := s.kafka.ProduceJSON(ctx, "topic-name", key, messageData)

// Consumer pattern with worker
handler := func(ctx context.Context, message *sarama.ConsumerMessage) error {
    // Process message
    var payload ObjectEvent
    if err := json.Unmarshal(message.Value, &payload); err != nil {
        return err
    }

    return s.HandleEvent(ctx, &payload)
}

closeFn, err := s.kafka.ConsumerGroup(ctx, "group-id", []string{"topic"}, handler)
defer closeFn(ctx)
```

## Best Practices

### Code Style

- Follow Go conventions (gofmt, golint)
- Use meaningful names: `CreateObject` not `Create`
- Keep functions focused (single responsibility)
- Use dependency injection via constructors
- Implement interfaces for testability
- Error wrapping: `fmt.Errorf("context: %w", err)`

### Security

- Use SSL/TLS for all external communications
- Validate all inputs with go-playground/validator
- Use proper authentication/authorization (Keycloak ready)
- **Avoid raw SQL queries** - use Bun ORM methods to prevent SQL injection
- If raw queries are necessary, always use parameter binding (`?` placeholders)
- Never log sensitive data (passwords, tokens, keys)
- Use environment variables for secrets

### Performance

- Use connection pooling (DB, Redis, HTTP clients)
- Implement proper caching strategies with Redis
- Use background contexts appropriately
- Handle graceful shutdowns with provider interface
- Use buffered channels for concurrent operations
- Implement proper timeouts

### Testing

- Write unit tests for business logic
- Use dependency injection for testability
- Mock external dependencies (DB, S3, Kafka)
- Test error scenarios and edge cases
- Use table-driven tests for multiple scenarios

## Common Tasks

### Start Development Server

```bash
go run . http              # Start server (HTTP/1.1 & HTTP/2)
```

### Database Migrations

```bash
go run . db init           # Initialize migration tables
go run . db create name    # Create new migration
go run . db migrate        # Apply pending migrations
go run . db status         # Check migration status
go run . db rollback       # Rollback last migration
```

### Dependencies

```bash
go mod tidy                # Clean up dependencies
go mod vendor              # Vendor dependencies
go get package@version     # Add new dependency
```

### Create New Module

1. Create directory: `/app/modules/mymodule/`
2. Create files: `mymodule.mod.go`, `mymodule.svc.go`, `mymodule.ctl.go`
3. Implement Module struct with constructor
4. Register in `/app/modules/modules.go`
5. Add to Modules struct
6. Register routes in `/routes/routes.go` if needed
7. Add configuration struct if needed
8. Implement `provider.Close` if cleanup needed

### Add HTTP Routes

```go
// In /routes/routes.go
func Register(r *gin.Engine, modules *appmod.Modules) {
    api := r.Group("/api/v1")
    {
        objects := api.Group("/objects")
        {
            objects.POST("", modules.Object.Ctl.Create)
            objects.GET("/:id", modules.Object.Ctl.Get)
            objects.PATCH("/:id", modules.Object.Ctl.Update)
            objects.DELETE("/:id", modules.Object.Ctl.Delete)
        }
    }
}
```

## Environment Variables

Key environment variables to configure:

```bash
# Application
APP_NAME=mcop-storage-service
APP_KEY=your-app-key
ENVIRONMENT=local|development|uat|production
PORT=8080
DEBUG=true|false

# Database (PostgreSQL)
DATABASE_SQL__HOST=localhost
DATABASE_SQL__PORT=5432
DATABASE_SQL__DATABASE=storage
DATABASE_SQL__USERNAME=postgres
DATABASE_SQL__PASSWORD=password
DATABASE_SQL__SSL_MODE=disable|require

# Redis (optional)
DATABASE_REDIS__ADDR=localhost:6379
DATABASE_REDIS__PASSWORD=
DATABASE_REDIS__DB=0

# S3/MinIO
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin
S3_BUCKET_NAME=default

# Object Storage Module
OBJECT_PRIVATE_BUCKET=private-objects
OBJECT_PUBLIC_BUCKET=public-objects

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_CA_PATH=path/to/ca.crt
KAFKA_CERT_PATH=path/to/cert.crt
KAFKA_KEY_PATH=path/to/key.key
KAFKA_SECTION=section-name

# Sentry
SENTRY_DSN=https://your-sentry-dsn

# OpenTelemetry
OTEL_ENABLE=true|false
OTEL_COLLECTOR_ENDPOINT=localhost:4317
OTEL_TRACE_MODE=grpc|stdout
OTEL_METRIC_MODE=grpc|stdout
OTEL_LOG_MODE=grpc|stdout
OTEL_LOG_LEVEL=debug|info|warn|error
OTEL_TRACE_RATIO=1.0

# HTTP
HTTP_JSON_NAMING=snake_case|camelCase
```

## Troubleshooting

### Issue: Environment Variables Not Loading

```bash
# Check .env file exists
ls -la .env

# Verify format (no spaces around =)
KEY=value  # ✓ Correct
KEY = value  # ✗ Wrong

# Check required fields have conf:"required" tag
type Config struct {
    Required string `conf:"required"`
}
```

### Issue: Database Connection Failed

```bash
# Verify PostgreSQL is running
psql -h localhost -U postgres -d storage

# Check environment variables
DATABASE_SQL__HOST=localhost
DATABASE_SQL__PORT=5432
DATABASE_SQL__DATABASE=storage
DATABASE_SQL__USERNAME=postgres
DATABASE_SQL__PASSWORD=yourpassword
DATABASE_SQL__SSL_MODE=disable

# Check migrations are applied
go run . db status
```

### Issue: Kafka Connection Failed

```bash
# Verify brokers
KAFKA_BROKERS=localhost:9092

# For SSL/TLS, check certificate paths exist
ls -la path/to/ca.crt
ls -la path/to/cert.crt
ls -la path/to/key.key

# Verify certificates are valid
openssl x509 -in cert.crt -text -noout
```

### Issue: S3/MinIO Access Denied

```bash
# Check credentials
S3_ENDPOINT=localhost:9000
S3_ACCESS_KEY_ID=minioadmin
S3_SECRET_ACCESS_KEY=minioadmin

# Verify bucket exists
mc ls myminio/bucket-name

# Check bucket policy
mc policy get myminio/bucket-name
```

### Issue: Trace/Span Not Appearing

```bash
# Verify OpenTelemetry is enabled
OTEL_ENABLE=true
OTEL_COLLECTOR_ENDPOINT=localhost:4317

# Check tracer is initialized with correct naming
tracer := otel.Tracer("mcop.service_name.module_name")

# Ensure context is passed through
ctx, span := s.tracer.Start(ctx, "operation")
defer span.End()
```

### Issue: Logs Not Showing Trace Correlation

```bash

# Use context-aware logging for trace correlation
logger.WithCtx(ctx).Infof("Processing request")

# Verify OpenTelemetry bridge is enabled
OTEL_LOG_MODE=grpc|stdout
```

### Common Error Patterns

**Error: "dial tcp: lookup postgres: no such host"**
→ Database host not reachable. Check `DATABASE_SQL__HOST`

**Error: "context deadline exceeded"**
→ Operation timeout. Check network/service availability

**Error: "panic: runtime error: invalid memory address"**
→ Nil pointer. Check dependency injection in constructor

**Error: "bind: address already in use"**
→ Port conflict. Change `PORT` or stop existing process

**Error: "kafka: client has run out of available brokers"**
→ Kafka broker unreachable. Check `KAFKA_BROKERS` and SSL certificates

## Module-Specific Notes

### Object Module

- Core module for managing object storage
- Handles file upload, download, and lifecycle
- Supports public and private buckets
- Integrates with S3 service and database
- Implements Kafka event handling for async operations
- Status transitions: pending → active → obsolete

### S3 Module

- Wrapper around MinIO client
- Provides simplified interface for object operations
- Handles connection management and retry logic
- Supports multiple buckets

### Logging Module

- Primary logging system: `log/slog` (Go standard library)
- Automatic trace correlation (TraceID/SpanID injected into logs)
- OpenTelemetry bridge for unified observability
- Formatted logging methods: `Infof`, `Warnf`, `Errf`, `Debugf`
- Context-aware logging: `WithCtx(ctx)` for trace propagation

## Quick Command Reference

```bash
# Development
go run . http                    # Start server (HTTP/1.1 & HTTP/2)
go mod vendor                    # Update dependencies
go test ./...                    # Run tests

# Database
go run . db migrate              # Apply migrations
go run . db rollback             # Rollback last
go run . db create migration_name # New migration
go run . db status               # Check status

# Build
go build -o app .                # Build binary
docker build -t app:latest .     # Build Docker image

# Lint & Format
gofmt -w .                       # Format code
go vet ./...                     # Run go vet
```

---

**Key Principle**: This is a **DI-based microservice**. Infrastructure services (`/internal`) are initialized once and injected into business modules (`/app/modules`). Always use interfaces for dependencies to maintain testability and loose coupling.
