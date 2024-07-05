// package tests

// import (
// 	"io"
// 	"net/http"
// 	"testing"

// 	"github.com/gofiber/fiber/v2"
// 	"github.com/raziel-aleman/go-todo-app/internal/server"
// )

// func TestHandler(t *testing.T) {
// 	// Create a Fiber app for testing
// 	app := fiber.New()
// 	// Inject the Fiber app into the server
// 	s := &server.FiberServer{App: app}
// 	// Define a route in the Fiber app
// 	app.Get("/", s.HelloWorldHandler)
// 	// Create a test HTTP request
// 	req, err := http.NewRequest("GET", "/", nil)
// 	if err != nil {
// 		t.Fatalf("error creating request. Err: %v", err)
// 	}
// 	// Perform the request
// 	resp, err := app.Test(req)
// 	if err != nil {
// 		t.Fatalf("error making request to server. Err: %v", err)
// 	}
// 	// Your test assertions...
// 	if resp.StatusCode != http.StatusOK {
// 		t.Errorf("expected status OK; got %v", resp.Status)
// 	}
// 	expected := "{\"message\":\"Hello World!\"}"
// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		t.Fatalf("error reading response body. Err: %v", err)
// 	}
// 	if expected != string(body) {
// 		t.Errorf("expected response body to be %v; got %v", expected, string(body))
// 	}
// }

package tests

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/raziel-aleman/go-todo-app/internal/server"
)

func TestHandler(t *testing.T) {
	s := &server.Server{}
	server := httptest.NewServer(http.HandlerFunc(s.HelloWorldHandler))
	defer server.Close()
	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}
	defer resp.Body.Close()
	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
	expected := "{\"message\":\"Hello World\"}"
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading response body. Err: %v", err)
	}
	if expected != string(body) {
		t.Errorf("expected response body to be %v; got %v", expected, string(body))
	}
}
