package main

import (
	backend "server/notes"
	"server/router"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
)

func main() {
	app := fiber.New()
	err := backend.Init()
	if err != nil {
		log.Error(err)
		return
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5500",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Content-Type, Authorization",
	}))
	router.Route(app)
	log.Fatal(app.Listen(":8080"))
}
