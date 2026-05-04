# Copilot Instructions

## Architecture & Tech Stack

### Core Technologies

- **Language**: Go 1.25.3
- **Web Framework**: Gin (HTTP/1.1 & HTTP/2 support)
- **Database**: PostgreSQL with Bun ORM
- **Message Queue**: Apache Kafka with SASL/TLS (IBM/sarama)
- **Cache**: Redis with JSON support
- **Object Storage**: MinIO (S3-compatible)
- **Observability**: OpenTelemetry (traces, metrics, logs)
- **Logging**: `log/slog` with structured logging and OpenTelemetry bridge
- **Configuration**: Environment-based with reflection
- **Validation**: go-playground/validator with i18n support
- **Build**: Docker multi-stage builds
- **Testing**: Built-in test framework

### Project Structure

```
/app                    # Application layer
  /console             # CLI commands (Cobra-based)
  /modules             # Business modules
    /entities          # Data entities (Bun models)
    /sentry            # Sentry error tracking
    /specs             # API specifications
  /utils               # Shared utilities
/config                # Configuration definitions
  /i18n                # Internationalization files
/database              # Database migrations
/internal              # Internal services (DI layer)
  /cmd                 # Command definitions
  /config              # Configuration service
  /database            # Database service
  /http                # HTTP server (with CORS, pprof)
  /kafka               # Kafka service
  /log                 # Logging service (Slog + OpenTelemetry)
  /otel                # OpenTelemetry collector
  /provider            # Provider interface (lifecycle management)
  /redis               # Redis service
/routes                # HTTP routes definition
```

## Development Guidelines

### 1. Module Development Pattern

When creating new modules, follow the established pattern:

```go
// Module structure
type Module struct {
    tracer trace.Tracer
    Svc    *Service      // Business logic service
    Ctl    *Controller   // HTTP controllers (optional)
}

// Service with dependencies and options
type Options struct {
    Config *config.Config[Config]
    tracer trace.Tracer
    objEnt entitiesinf.ObjectEntity  // Database interface
    s3     *s3.Service                // External service dependencies
}

type Service struct {
    *Options
}

// Controller with service dependency
type Controller struct {
    tracer trace.Tracer
    svc    *Service
}

// Module constructor
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

### 2. Lifecycle Management (Provider Interface)

Modules that need cleanup should implement the `provider.Close` interface:

```go
import "mcop/internal/provider"

var _ provider.Close = (*Service)(nil)

func (s *Service) Close(ctx context.Context) error {
    // Cleanup logic (close connections, flush buffers, etc.)
    return nil
}
```

The provider automatically calls `Close()` in reverse order of module registration during shutdown.

### 3. Code Generation Commands

Use the built-in generators for consistency:

```bash
# HTTP Server
go run . http        # Start server (HTTP/1.1 & HTTP/2)

# Database migrations
go run . db init     # Initialize migration tables
go run . db create   # Create new migration
go run . db migrate  # Apply migrations
go run . db status   # Check migration status

# Initialize project dependencies
go mod vendor        # Vendor dependencies
```

### 4. Configuration Management

- Use environment variables with structured configuration
- Follow the naming pattern: `{MODULE}_{FIELD}` → `S3__ENDPOINT`
- Support nested structures with double underscores `__`
- Use struct tags for validation: `conf:"required"`
- Configuration is loaded via reflection and supports defaults

```go
type Config struct {
    Endpoint        string `conf:"required"`
    AccessKeyId     string `conf:"required"`
    SecretAccessKey string `conf:"required"`
    BucketName      string
}

// Usage in module
func New(conf *config.Config[Config]) *Module {
    // Access config: conf.Value.Endpoint
}
```

### 5. Database Patterns

- Use Bun ORM for database operations
- **Avoid raw SQL queries** - use Bun ORM methods instead
- Implement repository pattern through interfaces in `entities/inf`
- Support multiple database connections
- Always use context for operations
- Implement proper transaction handling
- Use soft deletes with `deleted_at` timestamps

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

```go
// Entity interface pattern
type ObjectEntity interface {
    CreateObject(ctx context.Context, obj *ent.Object) (*ent.Object, error)
    GetObjectByID(ctx context.Context, id uuid.UUID) (*ent.Object, error)
    UpdateObjectStatusByID(ctx context.Context, id uuid.UUID, status ent.ObjectStatus) (*ent.Object, error)
    SoftDeleteObjectByID(ctx context.Context, id uuid.UUID) error
}

// Entity model with Bun
type Object struct {
    bun.BaseModel `bun:"table:objects,alias:obj"`

    ID        uuid.UUID      `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
    CreatedAt time.Time      `bun:"created_at,nullzero,notnull,default:current_timestamp"`
    UpdatedAt time.Time      `bun:"updated_at,nullzero,notnull,default:current_timestamp"`
    DeletedAt *time.Time     `bun:"deleted_at,soft_delete"`

    Status    ObjectStatus   `bun:"status,notnull"`
}
```

### 6. HTTP API Conventions

- Use Gin for routing
- Implement proper request/response DTOs
- Add OpenTelemetry tracing to all endpoints
- Use middleware for cross-cutting concerns (CORS, pprof, error handling)
- Follow RESTful conventions
- Support both `snake_case` and `camelCase` via configuration (`HTTP_JSON_NAMING`)

```go
// Controller pattern
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

### 7. Observability & Logging

**OpenTelemetry Tracing**: Add tracing to all business operations

**Logging**: Use `log/slog` for structured logging with automatic trace correlation

```go
// Service tracing pattern
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

**Logging with slog**:

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

### 8. Error Handling

- Use Sentry for production error tracking
- Implement custom error types for better categorization
- Add context to errors for better debugging
- Use appropriate HTTP status codes

```go
// Error handling with Otel

if err != nil {
    span.RecordError(err)
    return fmt.Errorf("failed to process: %w", err)
}
```

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

- Follow Go conventions and idioms
- Use meaningful variable and function names
- Keep functions small and focused (single responsibility)
- Use dependency injection through constructors
- Implement interfaces for testability
- Use proper error wrapping with `fmt.Errorf` and `%w`

### Security

- Use SSL/TLS for all external communications
- Validate all inputs with go-playground/validator
- Use proper authentication/authorization (Keycloak ready)
- **Avoid raw SQL queries** - use Bun ORM methods to prevent SQL injection
- If raw queries are necessary, always use parameter binding (`?` placeholders)
- Never log sensitive data (passwords, tokens, keys)
- Use environment variables for secrets

### Performance

- Use connection pooling for databases and Redis
- Implement proper caching strategies with Redis
- Use background contexts appropriately
- Handle graceful shutdowns with provider interface
- Use buffered channels for concurrent operations
- Implement proper timeouts

### Testing

- Write unit tests for business logic
- Use dependency injection for testability
- Mock external dependencies (database, S3, Kafka)
- Test error scenarios and edge cases
- Use table-driven tests for multiple scenarios

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
DATABASE_SQL__SSL_MODE=disable

# Redis (optional)
DATABASE_REDIS__ADDR=localhost:6379
DATABASE_REDIS__PASSWORD=
DATABASE_REDIS__DB=0

# S3/MinIO
S3__ENDPOINT=localhost:9000
S3__ACCESS_KEY_ID=minioadmin
S3__SECRET_ACCESS_KEY=minioadmin
S3__BUCKET_NAME=default

# Object Storage Module
OBJECT__PRIVATE_BUCKET=private-objects
OBJECT__PUBLIC_BUCKET=public-objects

# Kafka
KAFKA_BROKERS=localhost:9092
KAFKA_CA_PATH=path/to/ca.crt
KAFKA_CERT_PATH=path/to/cert.crt
KAFKA_KEY_PATH=path/to/key.key
KAFKA_SECTION=section-name

# Sentry
SENTRY__DSN=https://your-sentry-dsn

# OpenTelemetry
OTEL_ENABLE=true
OTEL_COLLECTOR_ENDPOINT=localhost:4317
OTEL_TRACE_MODE=grpc|stdout
OTEL_METRIC_MODE=grpc|stdout
OTEL_LOG_MODE=grpc|stdout
OTEL_LOG_LEVEL=debug|info|warn|error
OTEL_TRACE_RATIO=1.0


# HTTP
HTTP_JSON_NAMING=snake_case|camelCase
```

## Common Tasks

### Adding New Dependencies

1. Add to `go.mod` with `go get package@version`
2. Update vendor with `go mod vendor`
3. Import and use in your modules

### Creating New Module

1. Create module directory under `/app/modules`
2. Implement `Module` struct with `Svc` and optionally `Ctl`
3. Create `New()` constructor with dependencies
4. Register module in `/app/modules/modules.go`
5. Add configuration struct if needed
6. Implement `provider.Close` if cleanup needed

### Adding HTTP Routes

1. Create controller methods in module
2. Register routes in `/routes/routes.go`
3. Use middleware for cross-cutting concerns
4. Add OpenTelemetry tracing

### Creating Database Migration

1. Run `go run . db create migration_name`
2. Edit generated migration files in `/database/migrations`
3. Apply with `go run . db migrate`
4. Check status with `go run . db status`

## Troubleshooting

- **Check logs**: Structured logs with trace IDs for correlation
- **Use trace IDs**: Track requests across services with OpenTelemetry
- **Verify config**: Ensure all required environment variables are set
- **Database migrations**: Check status and apply pending migrations
- **Kafka connectivity**: Verify broker addresses and SSL certificates
- **S3 access**: Verify endpoint, credentials, and bucket permissions
- **Structured logs**: Check logs with trace IDs for request correlation
- **Health checks**: Implement health endpoints for monitoring

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

This storage service provides a production-ready foundation for object storage management with comprehensive observability, security, and scalability built-in.
