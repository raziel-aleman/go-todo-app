package server

import (
	"github.com/gofiber/fiber/v2"

	"go-blueprint/internal/database"
)

type FiberServer struct {
	*fiber.App

	db database.Service
}

func New() *FiberServer {
	server := &FiberServer{
		App: fiber.New(fiber.Config{
			ServerHeader: "go-blueprint",
			AppName:      "go-blueprint",
		}),

		db: database.New(),
	}

	return server
}
