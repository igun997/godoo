# Godoo - Odoo RPC Client for Go

[![Go Report Card](https://goreportcard.com/badge/github.com/wongpinter/godoo)](https://goreportcard.com/report/github.com/wongpinter/godoo)
[![GoDoc](https://godoc.org/github.com/wongpinter/godoo?status.svg)](https://godoc.org/github.com/wongpinter/godoo)

Godoo is a Go library that provides a wrapper around the Odoo RPC API, supporting both JSON-RPC and XML-RPC protocols, making it easier to interact with Odoo from your Go applications. **JSON-RPC is now the default protocol** for compatibility with modern Odoo versions.

## Features

- Support for both JSON-RPC (default) and XML-RPC protocols
- Connection pooling to efficiently manage connections to your Odoo instance.
- A simple and intuitive API for calling Odoo methods.
- Type-safe wrappers for common Odoo data types.
- Automatic handling of authentication for both protocols.
- Robust connection management with automatic reconnection.
- Full compatibility with modern Odoo versions

## Installation

To install Godoo, use `go get`:

```bash
go get github.com/wongpinter/godoo
```

## Usage

This project includes three examples in the `examples` directory:

1.  **`single_client`**: A simple example demonstrating how to connect to a single Odoo instance using JSON-RPC (default) and read data.
2.  **`jsonrpc_client`**: An example showing how to explicitly use JSON-RPC protocol to connect to Odoo.
3.  **`rpc_pool`**: A more advanced example showing how to use the connection pool to manage connections to multiple Odoo instances using JSON-RPC, including automatic reconnection.

### Single Client Example

To run the single client example, navigate to the directory and run the `main.go` file:

```bash
cd examples/single_client
go run main.go
```

Here is the code from `examples/single_client/main.go` showing the default JSON-RPC usage:

```go
package main

import (
	"log"
	"time"

	"github.com/wongpinter/godoo"
)

type Unit struct {
	Name *godoo.String `xmlrpc:"name"`
}

func main() {
	// Create a new connection pool.
	pool := godoo.NewPool(10, 5, 1*time.Minute)

	// Create a new client (JSON-RPC is used by default).
	c, err := godoo.NewClient(&godoo.ClientConfig{
		Database: "your-database",
		Admin:    "admin",
		Password: "your-password",
		URL:      "https://your-odoo-instance.com",
		Pool:     pool,
		// Protocol: godoo.ProtocolJSONRPC // This is now the default
	})
	if err != nil {
		log.Fatalf("Error creating client: %v", err)
	}
	defer c.Close()

	// Get the Odoo version.
	v, err := c.Version()
	if err != nil {
		log.Fatalf("Error getting version: %v", err)
	}
	log.Printf("Odoo version: %s", v.ServerVersion)

	// Read data from a model.
	var units []Unit
	err = c.Read("propertek.property.unit", []int64{1}, nil, &units)
	if err != nil {
		log.Fatalf("Error getting units: %v", err)
	}

	log.Printf("Units: %+v", units)
}
```

### Protocol Selection

You can choose between JSON-RPC (default) and XML-RPC by setting the `Protocol` field in `ClientConfig`:

```go
// JSON-RPC (default)
cfg := &godoo.ClientConfig{
    Database: "your-database",
    Admin:    "admin",
    Password: "your-password",
    URL:      "https://your-odoo-instance.com",
    Pool:     pool,
    // Protocol: godoo.ProtocolJSONRPC // This is the default, can be omitted
}

// XML-RPC (for legacy Odoo versions)
cfg := &godoo.ClientConfig{
    Database: "your-database",
    Admin:    "admin",
    Password: "your-password",
    URL:      "https://your-odoo-instance.com",
    Pool:     pool,
    Protocol: godoo.ProtocolXMLRPC,
}
```

### XML-RPC Client Example

If you need to use XML-RPC for legacy Odoo versions, you can explicitly specify it:

```go
// Create a new XML-RPC client.
c, err := godoo.NewClient(&godoo.ClientConfig{
    Database: "your-database",
    Admin:    "admin",
    Password: "your-password",
    URL:      "https://your-odoo-instance.com",
    Pool:     pool,
    Protocol: godoo.ProtocolXMLRPC, // Explicitly use XML-RPC
})
```

### RPC Pool Example

To run the RPC pool example, navigate to the directory and run the `main.go` file:

```bash
cd examples/rpc_pool
go run main.go
```

This example demonstrates how to manage connections to multiple Odoo instances. See the code in `examples/rpc_pool/main.go` for details.

## Contributing

Contributions are welcome! Please feel free to submit a pull request or open an issue if you find a bug or have a feature request.

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
