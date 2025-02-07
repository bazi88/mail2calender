package main

import (
	"fmt"

	"mail2calendar/internal/server"
)

func main() {
	s := server.New()
	s.InitDomains()

	// In ra các routes đã đăng ký
	fmt.Print("Registered Routes:\n\n")
	fmt.Printf("Server is running on %s:%s\n", s.Config().Api.Host, s.Config().Api.Port)
}
