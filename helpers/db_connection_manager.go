package helpers

import (
	"database/sql"
	"fmt"
	"os"
	"sync"

	_ "github.com/lib/pq"
)

var (
	instance *ConnectionManager
	once     sync.Once
)

type Connection struct {
	db     *sql.DB
	driver string
}

type ConnectionManager struct {
	connection Connection
	mu         sync.Mutex
}

func GetConnectionManager() *ConnectionManager {
	once.Do(func() {
		instance = &ConnectionManager{}
	})
	return instance
}

func (connManager *ConnectionManager) GetConnection(driver string) (*sql.DB, error) {
	connManager.mu.Lock()
	defer connManager.mu.Unlock()

	if connManager.connection.db != nil {
		return connManager.connection.db, nil
	}

	return connManager.AddConnection(driver)
}

func (connManager *ConnectionManager) AddConnection(driver string) (*sql.DB, error) {
	db, err := sql.Open(driver, getUrl(driver))
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	connManager.connection = Connection{
		db:     db,
		driver: driver,
	}

	if err := createTables(db); err != nil {
		return nil, fmt.Errorf("failed to create tables: %w", err)
	}

	return db, nil
}

func (connManager *ConnectionManager) CloseConnection() error {
	connManager.mu.Lock()
	defer connManager.mu.Unlock()

	if connManager.connection.db != nil {
		if err := connManager.connection.db.Close(); err != nil {
			return fmt.Errorf("failed to close database connection: %w", err)
		}
		connManager.connection.db = nil
	}
	return nil
}

func getUrl(driver string) string {
	if driver == "postgres" {
		hostname := os.Getenv("PG_HOSTNAME")
		port := os.Getenv("PG_PORT")
		user := os.Getenv("PG_USERNAME")
		password := os.Getenv("PG_PASSWORD")
		dbname := os.Getenv("PG_DBNAME")
		return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			hostname, port, user, password, dbname)
	}
	return ""
}

func createTables(db *sql.DB) error {
	queries := []string{
		"CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT, password TEXT)",
		"CREATE TABLE IF NOT EXISTS persons (name TEXT, tantieme INTEGER, userId INTEGER REFERENCES users(id))",
		"CREATE TABLE IF NOT EXISTS bills (label TEXT, amount FLOAT, userId INTEGER REFERENCES users(id))",
		"CREATE TABLE IF NOT EXISTS provisions (label TEXT, amount FLOAT, userId INTEGER REFERENCES users(id))",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}
	return nil
}
