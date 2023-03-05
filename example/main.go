package main

import (
	"log"
	"time"

	odoo "repo.nusatek.id/sugeng/godoo"
)

func main() {
	pool := odoo.NewPool(10, 5, 1*time.Minute)

	c, err := odoo.NewClient(&odoo.ClientConfig{
		Database: "beta",
		Admin:    "admin",
		Password: "admin",
		URL:      "https://beta.propertek.id",
		Pool:     pool,
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

	i, _ := c.Count("note.note", odoo.NewCriteria(), &odoo.Options{})

	log.Println(i)

	k, err := c.ExecuteKw("get_units", "propertek.property.unit", []interface{}{1, 0}, nil)
	if err != nil {
		log.Printf("error getting units: %v", err)
	}

	log.Println(k)
}
