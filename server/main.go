package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/gofiber/contrib/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/lib/pq"
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

type Section struct {
	ID       string  `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	RowStart int     `json:"row_start"`
	RowEnd   int     `json:"row_end"`
	ColStart int     `json:"col_start"`
	ColEnd   int     `json:"col_end"`
}

type VenueLayout struct {
	Sections []Section `json:"sections"`
}

type Hub struct {
	clients    map[*websocket.Conn]bool
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	broadcast  chan []byte
}

var rdb *redis.Client
var db *sql.DB

func main() {
	rdb = redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("❌ Failed to connect to Redis: %v", err)
	}
	log.Println("✅ Connected to Redis")

	// Connect to Postgres
	connStr := "user=user password=password dbname=ticketmain sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}
	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS orders (
			id SERIAL PRIMARY KEY,
			seat_id TEXT NOT NULL,
			user_id TEXT NOT NULL,
			amount INTEGER NOT NULL,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("✅ Connected to Postgres and ensured table exists")

	// Parse venue layout
	venueData, err := os.ReadFile("venue_layout.json")
	if err != nil {
		log.Fatalf("Failed to load venue data: %v", err)
	}

	var venueLayout VenueLayout
	if err := json.Unmarshal(venueData, &venueLayout); err != nil {
		log.Fatalf("Failed to parse venue data: %v", err)
	}

	hub := Hub{
		clients:    make(map[*websocket.Conn]bool),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
		broadcast:  make(chan []byte),
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

	go func() {
		for {
			select {
			case conn := <-hub.register:
				hub.clients[conn] = true
				log.Println("Client connected")
			case conn := <-hub.unregister:
				if _, ok := hub.clients[conn]; ok {
					delete(hub.clients, conn)
					conn.Close()
					log.Println("Client disconnected")
				}
			case message := <-hub.broadcast:
				for conn := range hub.clients {
					if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
						log.Println("Write error:", err)
						conn.Close()
						delete(hub.clients, conn)
					}
				}
			}
		}
	}()

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return c.Status(fiber.StatusUpgradeRequired).SendString("Upgrade Required")
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		hub.register <- c
		defer func() {
			hub.unregister <- c
		}()

		for {
			// Keep connection alive by reading messages, even if not processed
			// If client closes connection, ReadMessage will return an error
			_, _, err := c.ReadMessage()
			if err != nil {
				break
			}
		}
	}))

	app.Post("/api/purchase", func(c *fiber.Ctx) error {
		var req HoldRequest
		if err := c.BodyParser(&req); err != nil {
			return c.Status(400).SendString("Invalid body")
		}

		lockKey := "seat:" + req.SeatID + ":lock"
		ctx := c.Context()

		val, err := rdb.Get(ctx, lockKey).Result()
		if err == redis.Nil {
			return c.Status(400).JSON(fiber.Map{"status": "fail", "message": "Seat not held"})
		} else if err != nil {
			return c.Status(500).SendString("Internal Server Error")
		}

		if val != req.UserID {
			return c.Status(403).JSON(fiber.Map{"status": "fail", "message": "You do not own this seat"})
		}

		// Calculate price
		var row, col int
		_, err = fmt.Sscanf(req.SeatID, "r%d-c%d", &row, &col)
		if err != nil {
			return c.Status(400).SendString("Invalid seat ID format")
		}

		price := 0
		for _, section := range venueLayout.Sections {
			if row >= section.RowStart && row <= section.RowEnd && col >= section.ColStart && col <= section.ColEnd {
				price = int(section.Price)
				break
			}
		}

		if price == 0 {
			price = 100
		}

		_, err = db.Exec("INSERT INTO orders (seat_id, user_id, amount) VALUES ($1, $2, $3)", req.SeatID, req.UserID, price)
		if err != nil {
			log.Printf("DB error: %v", err)
			return c.Status(500).SendString("Failed to process order")
		}

		rdb.Del(ctx, lockKey)

		update := map[string]interface{}{
			"type": "SEAT_UPDATE",
			"payload": map[string]string{
				"seatId":  req.SeatID,
				"status":  "SOLD",
				"ownerId": req.UserID,
			},
		}
		msg, _ := json.Marshal(update)
		hub.broadcast <- msg

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Purchase successful",
			"seatId":  req.SeatID,
		})
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

		update := map[string]interface{}{
			"type": "SEAT_UPDATE",
			"payload": map[string]string{
				"seatId":  req.SeatID,
				"status":  "HELD",
				"ownerId": req.UserID,
			},
		}
		msg, _ := json.Marshal(update)
		hub.broadcast <- msg

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

		// Lua script to check ownership and delete atomically
		script := `
			if redis.call("get", KEYS[1]) == ARGV[1] then
				return redis.call("del", KEYS[1])
			else
				return 0
			end
		`

		result, err := rdb.Eval(ctx, script, []string{lockKey}, req.UserID).Result()
		if err != nil {
			log.Printf("Redis Eval error: %v", err)
			return c.Status(500).SendString("Internal Server Error")
		}

		if result.(int64) == 0 {
			return c.Status(403).JSON(fiber.Map{
				"status":  "fail",
				"message": "You do not own this seat or it expired",
			})
		}

		update := map[string]interface{}{
			"type": "SEAT_UPDATE",
			"payload": map[string]string{
				"seatId":  req.SeatID,
				"status":  "AVAILABLE",
				"ownerId": "",
			},
		}
		msg, _ := json.Marshal(update)
		hub.broadcast <- msg

		return c.JSON(fiber.Map{
			"status":  "success",
			"message": "Seat released successfully",
			"seatId":  req.SeatID,
		})
	})

	app.Get("/api/seats", func(c *fiber.Ctx) error {
		ctx := c.Context()

		var keys []string
		iter := rdb.Scan(ctx, 0, "seat:*:lock", 0).Iterator()
		for iter.Next(ctx) {
			keys = append(keys, iter.Val())
		}
		if err := iter.Err(); err != nil {
			log.Printf("Redis scan error: %v", err)
			return c.Status(500).SendString("Failed to fetch seats")
		}

		if len(keys) == 0 {
			return c.JSON([]SeatState{})
		}

		pipe := rdb.Pipeline()
		cmds := make(map[string]*redis.StringCmd)

		for _, key := range keys {
			cmds[key] = pipe.Get(ctx, key)
		}

		_, err := pipe.Exec(ctx)
		if err != nil && err != redis.Nil {
			log.Printf("Pipeline error: %v", err)
			return c.Status(500).SendString("Failed to fetch seat details")
		}

		seatsMap := make(map[string]SeatState)
		for key, cmd := range cmds {
			userID, err := cmd.Result()
			if err == nil {
				seatID := key[5 : len(key)-5]
				seatsMap[seatID] = SeatState{
					SeatID:  seatID,
					Status:  "HELD",
					OwnerID: userID,
				}
			}
		}

		rows, err := db.Query("SELECT seat_id, user_id FROM orders")
		if err != nil {
			log.Printf("DB query error: %v", err)
		} else {
			defer rows.Close()
			for rows.Next() {
				var seatID, userID string
				if err := rows.Scan(&seatID, &userID); err == nil {
					seatsMap[seatID] = SeatState{
						SeatID:  seatID,
						Status:  "SOLD",
						OwnerID: userID,
					}
				}
			}
		}

		var seats []SeatState
		for _, seat := range seatsMap {
			seats = append(seats, seat)
		}

		return c.JSON(seats)
	})

	log.Fatal(app.Listen(":3000"))
}
