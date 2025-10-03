package main

import (
	"fmt"
	"procguard/services/api"
)

func main() {
	const defaultPort = "58141"
	addr := "127.0.0.1:" + defaultPort
	fmt.Println("Starting API server on http://" + addr)
	api.StartWebServer(addr)
}
