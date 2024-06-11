package server

import (
	"encoding/json"
	"fmt"

	"github.com/gofiber/fiber/v2"
)

type NewTodo struct {
	Title       string `json:"title"`
	Description string `json:"body"`
}

func (s *FiberServer) RegisterFiberRoutes() {
	s.App.Get("/", s.HelloWorldHandler)

	s.App.Get("/health", s.healthHandler)

	s.App.Get("/healthcheck", s.healthCheckHandler)

	s.App.Get("/api/todos", s.getAllHandler)

	s.App.Patch("/api/todos/:id/done", s.markDoneHandler)

	s.App.Post("/api/todos/", s.createHandler)

	s.App.Patch("/api/todos/:id/edit", s.editHandler)
}

func (s *FiberServer) HelloWorldHandler(c *fiber.Ctx) error {
	resp := fiber.Map{
		"message": "Hello World!",
	}

	return c.JSON(resp)
}

func (s *FiberServer) healthHandler(c *fiber.Ctx) error {
	return c.JSON(s.db.Health())
}

func (s *FiberServer) healthCheckHandler(c *fiber.Ctx) error {
	return c.SendString("OK")
}

func (s *FiberServer) getAllHandler(c *fiber.Ctx) error {
	rows, _ := s.db.GetAll()

	return c.JSON(rows)
}

func (s *FiberServer) markDoneHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")

	if err != nil {
		return c.Status(401).SendString("Invalid id")
	}

	s.db.MarkDone(int64(id))

	rows, _ := s.db.GetAll()

	return c.JSON(rows)
}

func (s *FiberServer) createHandler(c *fiber.Ctx) error {
	var body NewTodo
	err := json.Unmarshal(c.Body(), &body)
	if err != nil {
		return err
	}

	fmt.Println(body)

	id, err := s.db.Create(body.Title, body.Description)

	fmt.Println(id)
	if err != nil {
		return err
	}

	rows, _ := s.db.GetAll()

	return c.JSON(rows)
}

func (s *FiberServer) editHandler(c *fiber.Ctx) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(401).SendString("Invalid id")
	}

	var body NewTodo
	bodyErr := json.Unmarshal(c.Body(), &body)
	if bodyErr != nil {
		return c.Status(500).SendString(("Internal server error"))
	}

	fmt.Println(body)

	id, editErr := s.db.Edit(id, body.Title, body.Description)

	fmt.Println(id)

	if editErr != nil {
		return editErr
	}

	rows, _ := s.db.GetAll()

	return c.JSON(rows)
}
