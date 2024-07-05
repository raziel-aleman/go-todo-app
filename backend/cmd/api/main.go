// package main

// import (
// 	"fmt"
// 	"os"
// 	"strconv"

// 	"github.com/raziel-aleman/go-todo-app/internal/server"

// 	"github.com/gofiber/fiber/v2/middleware/cors"
// 	_ "github.com/joho/godotenv/autoload"
// )

// type ProviderIndex struct {
// 	Providers    []string
// 	ProvidersMap map[string]string
// }

// func main() {

// 	server := server.New()

// 	server.App.Use(cors.New(cors.Config{
// 		AllowOrigins: "http://localhost:3000",
// 		AllowHeaders: "Origin, Content-Type, Accept",
// 	}))

// 	server.RegisterFiberRoutes()

// 	port, _ := strconv.Atoi(os.Getenv("PORT"))
// 	err := server.Listen(fmt.Sprintf(":%d", port))
// 	if err != nil {
// 		panic(fmt.Sprintf("cannot start server: %s", err))
// 	}
// }

package main

import (
	"fmt"

	"github.com/raziel-aleman/go-todo-app/internal/auth"
	"github.com/raziel-aleman/go-todo-app/internal/server"
)

func main() {

	auth.NewAuth()

	server := server.NewServer()

	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
