package api

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/duscraft/tanzia/lib/helpers"

	"github.com/go-session/redis/v3"
	"github.com/go-session/session/v3"
)

func main() {
	redisUrl := os.Getenv("REDIS_URL")
	if len(redisUrl) == 0 {
		redisUrl = "127.0.0.1"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if len(redisPort) == 0 {
		redisPort = "6379"
	}
	session.InitManager(
		session.SetStore(redis.NewRedisStore(&redis.Options{
			Addr: fmt.Sprintf("%s:%s", redisUrl, redisPort),
			DB:   0,
		})),
	)

	connManager := helpers.GetConnectionManager()

	_, _ = connManager.AddConnection("postgres")
	defer func() {
		if err := connManager.CloseConnection(); err != nil {
			log.Printf("Error closing database connection: %v", err)
		}
	}()

	port := os.Getenv("PORT")
	if len(port) == 0 {
		port = "8080"
	}

	_, _ = fmt.Fprintf(os.Stdout, "Listening on port %s...", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
