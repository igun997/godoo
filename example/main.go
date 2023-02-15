package main

import (
	"log"

	odoo "repo.nusatek.id/sugeng/godoo"
)

func main() {
	c, err := odoo.NewClient(&odoo.ClientConfig{
		Database: "beta",
		Admin:    "admin",
		Password: "admin",
		URL:      "https://beta.propertek.id",
	})
	if err != nil {
		log.Printf("Error creating client: %v", err)
	}

	v, err := c.Version()
	if err != nil {
		log.Printf("error get version: %v", err)
	}

	log.Printf("data version %v", v.ProtocolVersion)

	i, _ := c.Count("note.note", odoo.NewCriteria(), &odoo.Options{})

	log.Println(i)
}

// 78a6d15e2b85b63b985d46cf5a8ee091f73c6cd3
