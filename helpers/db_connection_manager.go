package helpers

import (
	"database/sql"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
)

var instance *ConnectionManager

type Connection struct {
	db     *sql.DB
	driver string
}

type ConnectionManager struct {
	connections map[string]Connection
}

func GetConnectionManager() *ConnectionManager {
	if instance == nil {
		instance = &ConnectionManager{
			connections: map[string]Connection{},
		}

		return instance
	}

	return instance
}

func (connManager *ConnectionManager) GetConnection(driver string, username string) *sql.DB {
	conn, ok := connManager.connections[username]

	if ok {
		return conn.db
	} else {
		return connManager.AddConnection(driver, username)
	}
}

func (connManager *ConnectionManager) AddConnection(driver string, username string) *sql.DB {
	db, err := sql.Open(driver, getUrl(driver, username))
	if err != nil {
		fmt.Println(err)
	}

	connManager.connections[username] = Connection{
		db,
		driver,
	}

	if username != "auth" {
		_, _ = db.Exec("CREATE TABLE IF NOT EXISTS persons (name TEXT, tantieme INTEGER)")
		_, _ = db.Exec("CREATE TABLE IF NOT EXISTS bills (label TEXT, amount INTEGER, billingDate TEXT)")
	}

	return db
}

func (connManager *ConnectionManager) CloseConnection(username string) {
	conn, ok := connManager.connections[username]
	if ok {
		conn.db.Close()
		delete(connManager.connections, username)
	}
}

func getUrl(driver string, username string) string {
	if driver == "sqlite3" {
		return fmt.Sprintf("file:%s.db", username)
	} else {
		panic(fmt.Sprintf("driver %s is not implemented.", driver))
	}
}
