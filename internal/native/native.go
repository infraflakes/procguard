package native

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"procguard/internal/blocklist/webblocklist"
	"procguard/internal/logger"
	"reflect"
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
	log := logger.GetWebLogger()
	log.Println("Native messaging host started")

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

		log.Printf("Received message: Type=%s, Payload=%s", req.Type, req.Payload)

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
			log.Printf("URL: %s", req.Payload)
		case "get_web_blocklist":
			list, err := webblocklist.Load()
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
			if _, err := webblocklist.Add(req.Payload); err != nil {
				log.Printf("Error adding to web blocklist: %v", err)
			}
		default:
			// Optionally handle unknown message types
		}
	}
}

func pollWebBlocklist() {
	log := logger.GetWebLogger()
	var lastBlocklist []string
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		list, err := webblocklist.Load()
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
	log := logger.GetWebLogger()
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
