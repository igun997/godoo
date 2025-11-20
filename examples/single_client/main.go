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
	pool := odoo.NewPool(10, 5, 1*time.Minute)

	c, err := odoo.NewClient(&odoo.ClientConfig{
		Database: "your-database",
		Admin:    "admin",
		Password: "your-password",
		URL:      "https://your-odoo-instance.com",
		Pool:     pool,
		Timeout:  30, // Timeout in seconds (default: 30)
		// Protocol: odoo.ProtocolJSONRPC // This is now the default
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
	}
	defer c.Close()

	v, err := c.Version()
	if err != nil {
		log.Printf("error get version: %v", err)
	}

	log.Printf("data version %v", v.ProtocolVersion)
	log.Printf("Using protocol: %s", c.GetProtocol())

	resp := []Unit{}

	if err := c.Read("propertek.property.unit", []int64{1}, nil, &resp); err != nil {
		log.Printf("error getting units: %v", err)
	}

	log.Println(resp)
}
