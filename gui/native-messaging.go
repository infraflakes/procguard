package gui

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

// Request is the message received from the extension.
type Request struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Response is the message sent to the extension.
type Response struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

// Run starts the native messaging host.
func Run() {
	log := data.GetLogger()
	log.Println("Native messaging host started")

	db, err := data.OpenDB()
	if err != nil {
		log.Fatalf("Native host failed to open database: %v", err)
	}
	defer db.Close()

	go pollWebBlocklist()

	for {
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

		log.Printf("Received message: Type=%s", req.Type)

		// Handle the message based on its type.
		switch req.Type {
		case "ping":
			// For now, just echo the request back.
			resp := Response{
				Type:    "echo",
				Payload: req.Payload,
			}
			sendMessage(resp)
		case "log_url":
			// Ignore logging the app's own GUI
			if strings.HasPrefix(req.Payload, "http://127.0.0.1:58141") {
				continue
			}
			writeUrlToDatabase(db, req.Payload)
		case "get_web_blocklist":
			list, err := data.LoadWeb()
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
			if _, err := data.AddWeb(req.Payload); err != nil {
				log.Printf("Error adding to web blocklist: %v", err)
			}
		default:
			// Optionally handle unknown message types
		}
	}
}

func writeUrlToDatabase(db *sql.DB, url string) {
	_, err := db.Exec("INSERT INTO web_events (url, timestamp) VALUES (?, ?)", url, time.Now().Unix())
	if err != nil {
		data.GetLogger().Printf("Failed to insert web event: %v", err)
	}
}

func pollWebBlocklist() {
	log := data.GetLogger()
	var lastBlocklist []string
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		list, err := data.LoadWeb()
		if err != nil {
			log.Printf("Error loading web blocklist for polling: %v", err)
			continue
		}

		if !reflect.DeepEqual(list, lastBlocklist) {
			log.Println("Web blocklist has changed, sending update to extension.")
			lastBlocklist = list
			resp := Response{
				Type:    "web_blocklist",
				Payload: list,
			}
			sendMessage(resp)
		}
	}
}

func sendMessage(resp Response) {
	log := data.GetLogger()
	b, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error marshalling response: %v", err)
		return
	}

	if err := binary.Write(os.Stdout, binary.LittleEndian, uint32(len(b))); err != nil {
		log.Printf("Error writing message length: %v", err)
		return
	}

	if _, err := os.Stdout.Write(b); err != nil {
		log.Printf("Error writing message body: %v", err)
		return
	}
}