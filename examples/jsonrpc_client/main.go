package main

import (
	"log"
	"time"

	odoo "github.com/wongpinter/godoo"
)

type Unit struct {
	Name *odoo.String `xmlrpc:"name"`
}

func main() {
	// Create a new connection pool.
	pool := odoo.NewPool(10, 5, 1*time.Minute)

	// Create a new JSON-RPC client.
	c, err := odoo.NewClient(&odoo.ClientConfig{
		Database: "your-database",
		Admin:    "admin",
		Password: "your-password",
		URL:      "https://your-odoo-instance.com",
		Pool:     pool,
		Protocol: odoo.ProtocolJSONRPC,
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
	}
	defer c.Close()

	// Verify we're using JSON-RPC
	log.Printf("Using protocol: %s", c.GetProtocol())
	log.Printf("Is JSON-RPC: %v", c.IsJSONRPC())

	// Get the Odoo version.
	v, err := c.Version()
	if err != nil {
		log.Printf("error get version: %v", err)
	}

	log.Printf("data version %v", v.ProtocolVersion)

	// Read data from a model.
	resp := []Unit{}
	if err := c.Read("propertek.property.unit", []int64{1}, nil, &resp); err != nil {
		log.Printf("error getting units: %v", err)
	}

	log.Println(resp)

	// Search and read with criteria
	criteria := odoo.NewCriteria().Add("active", "=", true)
	options := odoo.NewOptions().Limit(10).FetchFields("name", "active")
	
	var units []Unit
	if err := c.SearchRead("propertek.property.unit", criteria, options, &units); err != nil {
		log.Printf("error search and read: %v", err)
	}
	
	log.Printf("Found %d units", len(units))
	for i, unit := range units {
		log.Printf("Unit %d: %s", i+1, unit.Name.Get())
	}
}