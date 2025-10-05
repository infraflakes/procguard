package native

import (
	"encoding/binary"
	"encoding/json"
	"io"
	"os"
	"procguard/internal/logger"
)

// Request is the message received from the extension.
type Request struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Response is the message sent to the extension.
type Response struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

// Run starts the native messaging host.
func Run() {
	log := logger.GetWebLogger()
	log.Println("Native messaging host started")

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
		default:
			// Optionally handle unknown message types
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
