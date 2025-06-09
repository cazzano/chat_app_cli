package main

import (
	"fmt"
	"os"
)

func main() {
	// Check if command is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <command>")
		fmt.Println("Available commands:")
		fmt.Println("  signup                   - User registration")
		return
	}

	command := os.Args[1]

	// Execute based on command
	switch command {
	case "signup":
		// Execute signup process
		err := ExecuteSignup()
		if err != nil {
			fmt.Printf("Signup failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Signup completed successfully!")
	
	case "search":
                                
		err := friend()
		if err != nil {
			fmt.Printf("Search failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Search completed successfully!")

	case "login":
                                
		err := login_()
		if err != nil {
			fmt.Printf("Login failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Login completed successfully!")

	case "send":
                                
		err := send_message()
		if err != nil {
			fmt.Printf("Message failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Message sent successfully!")

	case "receive":
                                
		err := receive_message()
		if err != nil {
			fmt.Printf("Message failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Message received successfully!")

       case "requests":
                                
		err := manageFriendRequests()
		if err != nil {
			fmt.Printf("Requests failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Requests received successfully!")





	default:
		fmt.Printf("Error: Unknown command '%s'\n", command)
		fmt.Println("Use 'go run main.go' to see available commands")
	}
}
