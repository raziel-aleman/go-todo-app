package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	GetAll() ([]Todo, error)

	MarkDone(int64) (int, error)

	Create(string, string) (int, error)

	Edit(int, string, string) (int, error)
}

type service struct {
	db *sql.DB
}

var (
	dburl      = os.Getenv("DB_URL")
	dbInstance *service
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
	Body  string `json:"body"`
}

func New() Service {
	// Reuse Connection
	if dbInstance != nil {
		return dbInstance
	}

	db, err := sql.Open("sqlite3", dburl)
	if err != nil {
		// This will not be a connection error, but a DSN parse error or
		// another initialization error.
		log.Fatal(err)
	}

	// Table initializaiton query if it does not exist
	const createTable string = `CREATE TABLE IF NOT EXISTS todos (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description	TEXT NOT NULL,
		done INTEGER NOT NULL DEFAULT 0
	);`

	// Execute initialization query
	if _, err := db.Exec(createTable); err != nil {
		fmt.Println("Error creating todos table")
		log.Fatal(err)
	}

	dbInstance = &service{
		db: db,
	}
	return dbInstance
}

// Health checks the health of the database connection by pinging the database.
// It returns a map with keys indicating various health statistics.
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	stats := make(map[string]string)

	// Ping the database
	err := s.db.PingContext(ctx)
	if err != nil {
		stats["status"] = "down"
		stats["error"] = fmt.Sprintf("db down: %v", err)
		log.Fatalf(fmt.Sprintf("db down: %v", err)) // Log the error and terminate the program
		return stats
	}

	// Database is up, add more statistics
	stats["status"] = "up"
	stats["message"] = "It's healthy"

	// Get database stats (like open connections, in use, idle, etc.)
	dbStats := s.db.Stats()
	stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
	stats["in_use"] = strconv.Itoa(dbStats.InUse)
	stats["idle"] = strconv.Itoa(dbStats.Idle)
	stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
	stats["wait_duration"] = dbStats.WaitDuration.String()
	stats["max_idle_closed"] = strconv.FormatInt(dbStats.MaxIdleClosed, 10)
	stats["max_lifetime_closed"] = strconv.FormatInt(dbStats.MaxLifetimeClosed, 10)

	// Evaluate stats to provide a health message
	if dbStats.OpenConnections > 40 { // Assuming 50 is the max for this example
		stats["message"] = "The database is experiencing heavy load."
	}

	if dbStats.WaitCount > 1000 {
		stats["message"] = "The database has a high number of wait events, indicating potential bottlenecks."
	}

	if dbStats.MaxIdleClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many idle connections are being closed, consider revising the connection pool settings."
	}

	if dbStats.MaxLifetimeClosed > int64(dbStats.OpenConnections)/2 {
		stats["message"] = "Many connections are being closed due to max lifetime, consider increasing max lifetime or revising the connection usage pattern."
	}

	return stats
}

// Close closes the database connection.
// It logs a message indicating the disconnection from the specific database.
// If the connection is successfully closed, it returns nil.
// If an error occurs while closing the connection, it returns the error.
func (s *service) Close() error {
	log.Printf("Disconnected from database: %s", dburl)
	return s.db.Close()
}

// Retrieves all todos
func (s *service) GetAll() ([]Todo, error) {
	todos := []Todo{}
	rows, err := s.db.Query("SELECT * FROM todos")
	if err != nil {
		log.Fatal("Error selecting todos from database")
		return nil, nil
	}
	defer rows.Close()
	for rows.Next() {
		todo := Todo{}
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Done)
		if err != nil {
			log.Fatal("Error scanning todos from select")
			return nil, nil
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

// Marks todo as done
func (s *service) MarkDone(id int64) (int, error) {
	currentValue := 0
	row := s.db.QueryRow("SELECT Done FROM todos WHERE id=?;", id)
	row.Scan(&currentValue)

	if currentValue == 0 {
		_, err := s.db.Exec("UPDATE todos SET done=? WHERE id=?;", 1, id)
		if err != nil {
			return 0, err
		}
	} else {
		_, err := s.db.Exec("UPDATE todos SET done=? WHERE id=?;", 0, id)
		if err != nil {
			return 0, err
		}
	}
	return int(id), nil
}

// Creates new todo
func (s *service) Create(title string, description string) (int, error) {
	res, err := s.db.Exec("INSERT INTO todos VALUES(NULL,?,?,?);", title, description, 0)

	if err != nil {
		return -1, err
	}

	var id int64

	if id, err = res.LastInsertId(); err != nil {
		return -1, err
	}

	return int(id), nil
}

// Edit todo
func (s *service) Edit(id int, newTitle string, newDescription string) (int, error) {
	_, err := s.db.Exec("UPDATE todos SET title=?, description=? WHERE id=?;", newTitle, newDescription, int64(id))
	if err != nil {
		return 0, err
	}

	return int(id), nil
}
