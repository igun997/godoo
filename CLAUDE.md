# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Godoo is a Go library that provides a wrapper around the Odoo RPC API, supporting both JSON-RPC (default) and XML-RPC protocols. It enables Go applications to interact with Odoo instances with features like connection pooling, automatic authentication, and type-safe data conversion.

## Development Commands

### Building and Testing
```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestClientProtocolMethods

# Build the library
go build

# Run examples
go run examples/single_client/main.go
go run examples/jsonrpc_client/main.go
go run examples/rpc_pool/main.go
```

### Module Management
```bash
# Update dependencies
go mod tidy

# Verify dependencies
go mod verify
```

## Architecture

### Layered Design

The codebase follows a clear layered architecture:

1. **Client Layer** (`client.go`) - High-level API for Odoo operations (Create, Read, Update, Delete, Search, Count, FieldsGet)
2. **Protocol Adapters** - `JSONRPCClient` (jsonrpc_client.go) and XML-RPC client via `kolo/xmlrpc`
3. **Connection Management** - Two-tier pooling system:
   - `Pool` (pool.go) - Low-level TCP connection pooling
   - `ConnectionPool` (rpc_pool.go) - High-level client pooling for multiple Odoo instances

### Key Components

**Client (`client.go`)**
- Main entry point for all Odoo operations
- Handles authentication and session management
- Supports dual protocols (JSON-RPC default, XML-RPC for legacy)
- Methods: `ExecuteKw()`, `Create()`, `Read()`, `Update()`, `Delete()`, `Search()`, `SearchRead()`, `Count()`, `FieldsGet()`, `Version()`, `Ping()`

**JSON-RPC Client (`jsonrpc_client.go`)**
- Default protocol implementation for modern Odoo
- HTTP-based communication with JSON-RPC 2.0
- Session management with UID and session_id tracking
- Context management for Odoo requests

**Type System (`types.go`)**
- Type-safe wrappers: `String`, `Int`, `Bool`, `Float`, `Time`, `Many2One`, `Relation`, `Selection`
- All types implement `Get()` methods with nil-safety
- Used to represent Odoo field types in Go structs

**Type Conversion (`conversion.go`)**
- Bidirectional conversion between Go structs and Odoo data
- `convertFromStaticToDynamic()` - Go → Odoo format
- `convertFromDynamicToStatic()` - Odoo → Go structs
- Reflection-based mapping using `xmlrpc` struct tags
- Handles date/time formatting (ISO 8601)

**Connection Pools**
- `Pool` - Manages idle TCP connections with configurable timeout and max connections
- `ConnectionPool` - Manages multiple Client instances with automatic retry and health checking

### Data Flow

```
User Code → Client → ExecuteKw() → Protocol Adapter (JSON-RPC/XML-RPC) 
→ HTTP/TCP Request → Odoo Response → Type Conversion → Go Structs
```

## Important Patterns

### Protocol Selection

JSON-RPC is the default protocol. Specify explicitly only for XML-RPC:

```go
// JSON-RPC (default) - no need to specify
cfg := &godoo.ClientConfig{
    Database: "db",
    Admin:    "admin",
    Password: "password",
    URL:      "https://odoo.example.com",
    Pool:     pool,
}

// XML-RPC (legacy) - must specify explicitly
cfg := &godoo.ClientConfig{
    Database: "db",
    Admin:    "admin",
    Password: "password",
    URL:      "https://odoo.example.com",
    Pool:     pool,
    Protocol: godoo.ProtocolXMLRPC,
}
```

### Struct Tag Mapping

Use `xmlrpc` tags to map Go struct fields to Odoo model fields:

```go
type Product struct {
    ID   *godoo.Int    `xmlrpc:"id"`
    Name *godoo.String `xmlrpc:"name"`
    Price *godoo.Float `xmlrpc:"list_price"`
}
```

### Type Safety with Nil Handling

Always use the wrapper types' `Get()` methods to safely access values:

```go
name := product.Name.Get() // Returns empty string if nil
id := product.ID.Get()     // Returns 0 if nil
```

### Thread Safety

- Connection pools use `sync.RWMutex` for concurrent access
- Background goroutines handle connection cleanup and retry logic
- All client operations are goroutine-safe

## Testing Approach

Tests are unit tests that validate:
- Protocol selection and client initialization
- Configuration validation
- Type conversions and struct mapping
- JSON-RPC request/response structures
- Connection pool behavior

Tests avoid network calls by using mock clients where possible.

## Code Organization Notes

- Root level contains core library code (client, protocols, types, conversion)
- `logr/` contains lightweight logging utility
- `examples/` contains three usage examples (single_client, jsonrpc_client, rpc_pool)
- Use reflection carefully in type conversion - it's central to the library's dynamic mapping capability
- Error handling uses custom error variables (errClientConfigurationInvalid, errClientNotAuthenticate, etc.)

## Protocol Details

**JSON-RPC 2.0** (Default)
- Endpoint: `/jsonrpc`
- Authentication: Returns `uid` and `session_id`
- ExecuteKw: Uses `call` method with service context

**XML-RPC** (Legacy)
- Endpoints: `/xmlrpc/2/common` (auth), `/xmlrpc/2/object` (operations)
- Authentication: Returns `uid` only
- ExecuteKw: Direct method invocation

## Critical Implementation Details

1. **Authentication Flow**: Always authenticate before calling any Odoo methods. The client handles this automatically via `NewClient()`.

2. **Context Management**: JSON-RPC requests include Odoo context parameters (lang, tz, uid) in the service context.

3. **Connection Reuse**: The pool maintains idle connections and cleans them up after `idleTimeout` (default 1 minute).

4. **Date/Time Handling**: Odoo uses ISO 8601 format without timezone (`2006-01-02 15:04:05`). The `Time` type handles conversion.

5. **Relational Fields**: 
   - `Many2One` represents many-to-one relations (returns `[id, "name"]` from Odoo)
   - `Relation` represents one2many/many2many (returns list of IDs)

6. **Error Responses**: JSON-RPC errors include detailed `message`, `code`, and `data` fields for debugging.
