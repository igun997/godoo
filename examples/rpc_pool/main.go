package main

import (
	"fmt"
	"log"
	"time"

	"github.com/wongpinter/godoo"
)

func main() {
	// Define configurations for multiple Odoo instances.
	serverConfigs := map[string]*godoo.ClientConfig{
		"odoo_instance_1": {
			Database: "db1",
			Admin:    "admin",
			Password: "pass1",
			URL:      "http://localhost:8069",
			Timeout:  30, // Timeout in seconds (default: 30)
			// Protocol: godoo.ProtocolJSONRPC // This is now the default
		},
		"odoo_instance_2": {
			Database: "db2",
			Admin:    "admin",
			Password: "pass2",
			URL:      "http://localhost:8070",
			Timeout:  45, // Custom timeout for slower instance
			// Protocol: godoo.ProtocolJSONRPC // This is now the default
		},
		"failed_instance": {
			Database: "db3",
			Admin:    "admin",
			Password: "wrong_password",
			URL:      "http://localhost:8071", // Assume this one will fail initially
			Timeout:  30,                      // Timeout in seconds (default: 30)
			// Protocol: godoo.ProtocolJSONRPC // This is now the default
		},
	}

	// Create a new connection pool.
	// It will try to connect to all instances and retry failed ones in the background.
	log.Println("Initializing connection pool...")
	pool, err := godoo.NewConnectionPool(5*time.Second, 3, serverConfigs)
	if err != nil {
		log.Fatalf("Failed to create connection pool: %v", err)
	}
	defer pool.Close()

	log.Printf("Initial number of successful connections: %d", pool.NumConnections())

	// --- Use a client from the pool ---
	log.Println("\n--- Retrieving client for 'odoo_instance_1' ---")
	client1, err := pool.GetClient("odoo_instance_1")
	if err != nil {
		log.Printf("Failed to get client 'odoo_instance_1': %v", err)
	} else {
		// Ping the server to check the connection
		if err := client1.Ping(); err != nil {
			log.Printf("Failed to ping 'odoo_instance_1': %v", err)
		} else {
			version, _ := client1.Version()
			fmt.Printf("Successfully pinged 'odoo_instance_1'. Server version: %s (Protocol: %s)\n", version.ServerVersion, client1.GetProtocol())
		}
	}

	// --- Wait to see if the failed instance reconnects ---
	log.Println("\n--- Waiting for background reconnection attempts for 'failed_instance' ---")
	time.Sleep(7 * time.Second) // Wait longer than the retry interval

	log.Printf("Number of connections after waiting: %d", pool.NumConnections())
	failedClient, err := pool.GetClient("failed_instance")
	if err != nil {
		log.Printf("Could not retrieve 'failed_instance' after retries: %v", err)
	} else {
		log.Printf("Successfully reconnected to 'failed_instance' in the background! Client is now available.")
		version, _ := failedClient.Version()
		fmt.Printf("Successfully pinged 'failed_instance'. Server version: %s (Protocol: %s)\n", version.ServerVersion, failedClient.GetProtocol())
	}

	// --- Periodically check connections ---
	log.Println("\n--- Running periodic connection check ---")
	// In a real application, you might run this in a ticker.
	pool.CheckConnections()
	log.Printf("Number of connections after check: %d", pool.NumConnections())

	log.Println("\nExample finished. Closing pool.")
}
