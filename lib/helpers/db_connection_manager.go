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
	db, err := sql.Open(driver, getURL(driver))
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

func getURL(driver string) string {
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
		"CREATE TABLE IF NOT EXISTS users (id SERIAL PRIMARY KEY, name TEXT, email TEXT UNIQUE, password TEXT, is_premium BOOLEAN DEFAULT FALSE, stripe_customer_id TEXT, needs_password_reset BOOLEAN DEFAULT FALSE)",
		"CREATE TABLE IF NOT EXISTS persons (name TEXT, tantieme INTEGER, userId INTEGER REFERENCES users(id))",
		"CREATE TABLE IF NOT EXISTS bills (label TEXT, amount FLOAT, userId INTEGER REFERENCES users(id))",
		"CREATE TABLE IF NOT EXISTS provisions (label TEXT, amount FLOAT, userId INTEGER REFERENCES users(id))",
	}

	for _, query := range queries {
		if _, err := db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute query: %w", err)
		}
	}

	migrations := []string{
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name='users' AND column_name='stripe_customer_id'
			) THEN
				ALTER TABLE users ADD COLUMN stripe_customer_id TEXT;
			END IF;
		END $$;`,
		`DO $$
		BEGIN
			IF NOT EXISTS (
				SELECT 1 FROM information_schema.columns 
				WHERE table_name='users' AND column_name='needs_password_reset'
			) THEN
				ALTER TABLE users ADD COLUMN needs_password_reset BOOLEAN DEFAULT FALSE;
			END IF;
		END $$;`,
	}

	for _, migration := range migrations {
		if _, err := db.Exec(migration); err != nil {
			return fmt.Errorf("failed to run migration: %w", err)
		}
	}

	return nil
}
