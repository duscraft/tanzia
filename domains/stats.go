package domains

import (
	"log"
	"net"
	"net/http"
	"os"
	"sync"
	"time"
)

var (
	statsMu         sync.Mutex
	userConnections = make(map[string]int) // key: "website" or "app"
	ipLogFile       = "user_ips.log"
)

// LogUserConnection increments the connection count and logs the IP address.
// appType should be "website" or "app".
func LogUserConnection(w http.ResponseWriter, r *http.Request, appType string) {
	statsMu.Lock()
	defer statsMu.Unlock()

	userConnections[appType]++

	ip := getIP(r)
	entry := time.Now().Format(time.RFC3339) + " [" + appType + "] " + ip + "\n"
	f, err := os.OpenFile(ipLogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		_, _ = f.WriteString(entry)
		_ = f.Close()
	} else {
		log.Printf("Failed to log IP: %v", err)
	}
}

// GetUserConnections returns the current connection counts.
func GetUserConnections() map[string]int {
	statsMu.Lock()
	defer statsMu.Unlock()
	cpy := make(map[string]int)
	for k, v := range userConnections {
		cpy[k] = v
	}
	return cpy
}

// getIP extracts the real client IP address.
func getIP(r *http.Request) string {
	// Check X-Forwarded-For header (if behind proxy)
	xff := r.Header.Get("X-Forwarded-For")
	if xff != "" {
		ips := net.ParseIP(xff)
		if ips != nil {
			return ips.String()
		}
		// If multiple IPs, take the first
		return xff
	}
	ip, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return ip
}
