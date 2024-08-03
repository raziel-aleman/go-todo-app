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
	"github.com/markbates/goth"
	_ "github.com/mattn/go-sqlite3"
	m "github.com/raziel-aleman/go-todo-app/internal/models"
)

// Service represents a service that interacts with a database.
type Service interface {
	// Health returns a map of health status information.
	// The keys and values in the map are service-specific.
	Health() map[string]string

	// Close terminates the database connection.
	// It returns an error if the connection cannot be closed.
	Close() error

	GetAll(string) ([]m.Todo, error)

	MarkDone(int64) error

	Create(m.NewTodo, string) (int, error)

	Edit(int, m.NewTodo, string) error

	SaveUser(goth.User, string) (string, error)

	IsSessionIdValid(string) (string, error)
}

type service struct {
	db *sql.DB
}

var (
	// db url parameters for WAL mode, timeout for concurrent writes, and for foreing key checking
	dburl      = os.Getenv("DB_URL") + "?_journal=WAL&_timeout=5000&_fk=true"
	dbInstance *service
)

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

	// Users table initialization query if it does not exist
	const createUsersTable string = `CREATE TABLE IF NOT EXISTS users (
		id TEXT NOT NULL PRIMARY KEY,
		name TEXT NOT NULL,
		email	TEXT NOT NULL,
		avatarUrl DATE NOT NULL,
		accessToken TEXT NOT NULL,
		expiresAt DATE NOT NULL
	);`

	// Execute initialization query
	if _, err := db.Exec(createUsersTable); err != nil {
		log.Println("Error creating User table")
		log.Fatal(err)
	}

	// Todos table initializaiton query if it does not exist
	const createTodosTable string = `CREATE TABLE IF NOT EXISTS todos (
		id INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
		title TEXT NOT NULL,
		description	TEXT NOT NULL,
		done INTEGER NOT NULL DEFAULT 0,
		userId TEXT NOT NULL,
		FOREIGN KEY (userId) REFERENCES users (id) ON DELETE CASCADE
	);`

	// Execute initialization query
	if _, err := db.Exec(createTodosTable); err != nil {
		log.Println("Error creating Todo table")
		log.Fatal(err)
	}

	// Todos table initializaiton query if it does not exist
	const createSessionsTable string = `CREATE TABLE IF NOT EXISTS sessions (
		id TEXT NOT NULL PRIMARY KEY,
		expiresAt DATE NOT NULL,
		userId TEXT NOT NULL,
		FOREIGN KEY (userId) REFERENCES users (id) ON DELETE CASCADE
	);`

	// Execute initialization query
	if _, err := db.Exec(createSessionsTable); err != nil {
		log.Println("Error creating Sessions table")
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

/* Retrieves all todos. Takes the userId (string) and returns an array of Todos ([]m.Todo) and an error. */
func (s *service) GetAll(userId string) ([]m.Todo, error) {
	todos := []m.Todo{}
	rows, err := s.db.Query("SELECT id, title, description, done FROM todos WHERE userId=?", userId)
	if err != nil {
		log.Fatal("Error selecting todos from database")
		return []m.Todo{}, nil
	}
	defer rows.Close()
	for rows.Next() {
		todo := m.Todo{}
		err := rows.Scan(&todo.ID, &todo.Title, &todo.Body, &todo.Done)
		if err != nil {
			log.Fatal("Error scanning todos from select")
			return []m.Todo{}, nil
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

/* Marks todo as done. Takes the Todo id (int64) and returns the Todo id (int) and an error. */
func (s *service) MarkDone(id int64) error {
	var currentValue int
	var isDone int
	row := s.db.QueryRow("SELECT Done FROM todos WHERE id=?;", id)
	row.Scan(&currentValue)

	if currentValue == 0 {
		isDone = 1
	} else {
		isDone = 0
	}

	// Mark Todo as done or not done depending on current status
	_, err := s.db.Exec("UPDATE todos SET done=? WHERE id=?;", isDone, id)
	if err != nil {
		return err
	}

	return nil
}

/* Creates new Todo. Takes a Todo struct and returns an id (int) and an error. */
func (s *service) Create(todo m.NewTodo, userId string) (int, error) {
	res, err := s.db.Exec("INSERT INTO todos VALUES(NULL,?,?,?,?);", todo.Title, todo.Description, 0, userId)

	if err != nil {
		log.Println("error trying to insert new todo into database")
		return -1, err
	}

	var id int64

	if id, err = res.LastInsertId(); err != nil {
		log.Println("error retreiving last inserted id")
		return -1, err
	}

	return int(id), nil
}

/* Edit Todo. Takes an EditedTodo struct and returnds an id (int) and an error. */
func (s *service) Edit(id int, newData m.NewTodo, userId string) error {
	_, err := s.db.Exec("UPDATE todos SET title=?, description=? WHERE id=? AND userId=?;", newData.Title, newData.Description, int64(id), userId)
	if err != nil {
		return err
	}

	return nil
}

/* Save user to database upon successful login and creates new session. */
func (s *service) SaveUser(user goth.User, sessionId string) (string, error) {
	var userId string

	// If userId does not exists, insert user and session in database
	if err := s.db.QueryRow("SELECT id FROM users WHERE id = ?;",
		user.UserID).Scan(&userId); err != nil {
		if err == sql.ErrNoRows {
			_, err := s.db.Exec("INSERT INTO users VALUES(?,?,?,?,?,?);",
				user.UserID,
				user.Name,
				user.Email,
				user.AvatarURL,
				user.AccessToken,
				user.ExpiresAt)

			if err != nil {
				log.Println("could not insert new user to database")
				return string(0), err
			}

			_, err = s.db.Exec("INSERT INTO sessions VALUES(?,date('now','+14 day'),?);",
				sessionId,
				user.UserID)

			if err != nil {
				log.Println("could not insert session to database")
				return string(0), err
			}

			return sessionId, nil
		}
	}

	// If userId exists, expire previous active session if they exist and insert new session for user upon re-login
	_, err := s.db.Exec("UPDATE sessions SET expiresAt=date('now','-1 day') WHERE expiresAt>=date('now') AND userId=?;", user.UserID)
	if err != nil {
		log.Println("an error ocurred when trying to expire previous active sessions")
		return string(0), err
	}

	_, err = s.db.Exec("INSERT INTO sessions VALUES(?,date('now','+14 day'),?);",
		sessionId,
		user.UserID)

	if err != nil {
		log.Println("an error ocurred when trying to insert new session to database")
		return string(0), err
	}

	return sessionId, nil
}

/* Validates session. Takes sessionId (string) and returns userId (string) if valid and an error */
func (s *service) IsSessionIdValid(sessionId string) (string, error) {
	var userId string
	if err := s.db.QueryRow("SELECT userId FROM sessions WHERE id = ? AND expiresAt > date('now');",
		sessionId).Scan(&userId); err != nil {
		if err == sql.ErrNoRows {
			log.Println("no valid session exists in the database, please login")
			return string(0), err
		}
	}

	return userId, nil
}
