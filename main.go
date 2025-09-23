package main

//go:generate go run github.com/akavel/rsrc -manifest procguard.manifest -o rsrc.syso

import "procguard/cmd"

func main() { cmd.Execute() }
