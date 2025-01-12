package main

import (
	"fmt"

	"mono-golang/internal/server"
)

func main() {
	s := server.New()
	s.InitDomains()
	fmt.Print("Registered Routes:\n\n")
	s.PrintAllRegisteredRoutes()
}
