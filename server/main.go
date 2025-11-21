package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/redis/go-redis/v9"
)

type HoldRequest struct {
	SeatID string `json:"seatId"`
	UserID string `json:"userId"`
}

type SeatState struct {
	SeatID  string `json:"seatId"`
	Status  string `json:"status"`
	OwnerID string `json:"ownerId,omitempty"`
}

var rdb *redis.Client

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	venueData, err := os.ReadFile("venue_layout.json")
	if err != nil {
		log.Fatalf("Failed to load venue data: %v", err)
	}

	app := fiber.New()
	app.Use(cors.New())

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendString("Ok")
	})

	app.Get("/api/venue", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/json")
		return c.Send(venueData)
	})

	app.Post("/api/hold", func(c *fiber.Ctx) error {
		var req HoldRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid body")
		}

		lockKey := "seat:" + req.SeatID + ":lock"

		success, err := rdb.SetNX(c.Context(), lockKey, req.UserID, 60*time.Second).Result()
		if err != nil {
			log.Printf("Redis error: %v", err)
			return c.Status(500).SendString("Internal Server Error")
		}

		if !success {
			return c.Status(409).JSON(fiber.Map{
				"status":  "fail",
				"message": "Seat is already held by another user",
			})
		}

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Seat held successfully",
			"seatId":  req.SeatID,
		})
	})

	app.Delete("/api/hold", func(c *fiber.Ctx) error {
		var req HoldRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid body")
		}

		lockKey := "seat:" + req.SeatID + ":lock"
		ctx := c.Context()

		val, err := rdb.Get(ctx, lockKey).Result()
		if err == redis.Nil {
			return c.JSON(fiber.Map{"status": "success", "message": "Seat already released"})
		} else if err != nil {
			return c.Status(500).SendString("Internal Server Error")
		}

		if val != req.UserID {
			return c.Status(403).JSON(fiber.Map{
				"status":  "fail",
				"message": "You do not own this seat",
			})
		}

		rdb.Del(ctx, lockKey)

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Seat released successfully",
			"seatId":  req.SeatID,
		})
	})

	app.Get("/api/seats", func(c *fiber.Ctx) error {
		ctx := c.Context()

		iter := rdb.Scan(ctx, 0, "seat:*:lock", 0).Iterator()

		var seats []SeatState

		for iter.Next(ctx) {
			key := iter.Val()

			userID, err := rdb.Get(ctx, key).Result()
			if err != nil {
				continue
			}

			seatID := key[5 : len(key)-5]

			seats = append(seats, SeatState{
				SeatID:  seatID,
				Status:  "HELD",
				OwnerID: userID,
			})
		}

		if err := iter.Err(); err != nil {
			log.Printf("Redis scan error: %v", err)
			return c.Status(500).SendString("Failed to fetch seats")
		}

		if seats == nil {
			seats = []SeatState{}
		}

		return c.JSON(seats)
	})

	log.Fatal(app.Listen(":3000"))
}
