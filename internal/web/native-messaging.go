package web

import (
	"database/sql"
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"procguard/internal/data"
	"reflect"
	"strings"
	"time"
)

const (
	// pollInterval is the interval at which the web blocklist is polled for changes.
	pollInterval = 2 * time.Second
)

// WebMetadataPayload is the payload for the log_web_metadata message from the extension.
type WebMetadataPayload struct {
	Domain  string `json:"domain"`
	Title   string `json:"title"`
	IconURL string `json:"iconUrl"`
}

// Request is a message received from the browser extension.
type Request struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

// Response is a message sent to the browser extension.
type Response struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Run starts the native messaging host, which listens for messages from the browser extension.
func Run() {
	log := data.GetLogger()

	db, err := data.OpenDB()
	if err != nil {
		log.Fatalf("Native host failed to open database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Printf("Failed to close database: %v", err)
		}
	}()

	// Start a goroutine to poll the web blocklist and push updates to the extension.
	go pollWebBlocklist()

	// The main loop reads messages from stdin, which is connected to the browser extension.
	for {
		// The native messaging protocol prefixes each message with its length in bytes.
		var length uint32
		if err := binary.Read(os.Stdin, binary.LittleEndian, &length); err != nil {
			if err == io.EOF {
				log.Println("EOF received, exiting native messaging host.")
				break // Exit loop on EOF
			}
			log.Printf("Error reading message length: %v", err)
			continue
		}

		msg := make([]byte, length)
		if _, err := io.ReadFull(os.Stdin, msg); err != nil {
			log.Printf("Error reading message body: %v", err)
			continue
		}

		var req Request
		if err := json.Unmarshal(msg, &req); err != nil {
			log.Printf("Error unmarshalling message: %v", err)
			continue
		}

		// Handle the message based on its type.
		switch req.Type {
		case "ping":
			var payload string
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				log.Printf("Error unmarshalling ping payload: %v", err)
				continue
			}
			resp := Response{
				Type:    "echo",
				Payload: payload,
			}
			sendMessage(resp)
		case "log_url":
			var url string
			if err := json.Unmarshal(req.Payload, &url); err != nil {
				log.Printf("Error unmarshalling log_url payload: %v", err)
				continue
			}
			// Ignore logging the app's own GUI.
			if strings.HasPrefix(url, "http://127.0.0.1:58141") {
				continue
			}
			writeUrlToDatabase(db, url)
		case "log_web_metadata":
			var payload WebMetadataPayload
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				log.Printf("Error unmarshalling log_web_metadata payload: %v", err)
				continue
			}
			writeWebMetadataToDatabase(db, &payload)
		case "get_web_blocklist":
			list, err := data.LoadWebBlocklist()
			if err != nil {
				log.Printf("Error loading web blocklist: %v", err)
				continue
			}
			resp := Response{
				Type:    "web_blocklist",
				Payload: list,
			}
			sendMessage(resp)
		case "add_to_web_blocklist":
			var domain string
			if err := json.Unmarshal(req.Payload, &domain); err != nil {
				log.Printf("Error unmarshalling add_to_web_blocklist payload: %v", err)
				continue
			}
			if _, err := data.AddWebsiteToBlocklist(domain); err != nil {
				log.Printf("Error adding to web blocklist: %v", err)
			}
		default:
			// Optionally handle unknown message types.
		}
	}
}

// writeUrlToDatabase inserts a new web event into the database.
func writeUrlToDatabase(db *sql.DB, url string) {
	data.EnqueueWrite("INSERT INTO web_events (url, timestamp) VALUES (?, ?)", url, time.Now().Unix())
}

// writeWebMetadataToDatabase inserts or updates web metadata in the database.
func writeWebMetadataToDatabase(db *sql.DB, payload *WebMetadataPayload) {
	// Use an UPSERT operation to either insert a new row or update the existing one for the given domain.
	// This is useful to keep the metadata up-to-date.
	query := `
		INSERT INTO web_metadata (domain, title, icon_url, timestamp)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(domain) DO UPDATE SET
			title = excluded.title,
			icon_url = excluded.icon_url,
			timestamp = excluded.timestamp;
	`
	data.EnqueueWrite(query, payload.Domain, payload.Title, payload.IconURL, time.Now().Unix())
}

// pollWebBlocklist periodically checks for changes in the web blocklist and sends updates to the extension.
func pollWebBlocklist() {
	log := data.GetLogger()
	var lastBlocklist []string
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for range ticker.C {
		list, err := data.LoadWebBlocklist()
		if err != nil {
			log.Printf("Error loading web blocklist for polling: %v", err)
			continue
		}

		// Only send an update if the blocklist has changed.
		if !reflect.DeepEqual(list, lastBlocklist) {
			lastBlocklist = list
			resp := Response{
				Type:    "web_blocklist",
				Payload: list,
			}
			sendMessage(resp)
		}
	}
}

// sendMessage sends a message to the browser extension.
func sendMessage(resp Response) {
	log := data.GetLogger()
	b, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	// The native messaging protocol requires that the message length be sent first.
	if err := binary.Write(os.Stdout, binary.LittleEndian, uint32(len(b))); err != nil {
		log.Printf("Error writing message length: %v", err)
		return
	}

	// Then, the message body is sent.
	if _, err := os.Stdout.Write(b); err != nil {
		log.Printf("Error writing message body: %v", err)
		return
	}
}
