package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	venueData, err := os.ReadFile("venue_layout.json")
	if err != nil {
		log.Fatalf("Failed to load venue data: %v", err)
	}

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})

	app.Get("/api/venue", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Send(venueData)
	})

	log.Fatal(app.Listen(":3000"))
}
