package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// manageFriendRequests is the main function that handles friend request management
func manageFriendRequests() error {
	// Read token from config file
	token, err := readTokenForFriendRequests()
	if err != nil {
		fmt.Printf("Error reading token: %v\n", err)
		os.Exit(1)
	}

	// Display menu and get user choice
	choice, err := displayFriendRequestMenu()
	if err != nil {
		fmt.Printf("Error getting user choice: %v\n", err)
		os.Exit(1)
	}

	// Handle user choice
	switch choice {
	case 1:
		err = handleIncomingRequests(token)
	case 2:
		err = handleOutgoingRequests(token)
	default:
		fmt.Println("Invalid choice. Please try again.")
		return manageFriendRequests()
	}

	if err != nil {
		fmt.Printf("Error handling friend requests: %v\n", err)
		os.Exit(1)
	}

	return nil
}

// readTokenForFriendRequests reads the token from ~/.config/chat_app/token.json
func readTokenForFriendRequests() (*TokenData, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	tokenPath := filepath.Join(homeDir, ".config", "chat_app", "token.json")
	
	file, err := os.Open(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open token file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %v", err)
	}

	var tokenData TokenData
	err = json.Unmarshal(data, &tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse token file: %v", err)
	}

	return &tokenData, nil
}

// displayFriendRequestMenu displays the menu and returns user choice
func displayFriendRequestMenu() (int, error) {
	fmt.Println("\n=== Friend Requests Management ===")
	fmt.Println("1. View Incoming Friend Requests")
	fmt.Println("2. View Outgoing Friend Requests")
	fmt.Print("\nEnter your choice (1 or 2): ")

	var choice string
	fmt.Scanln(&choice)

	choiceNum, err := strconv.Atoi(choice)
	if err != nil {
		return 0, fmt.Errorf("invalid choice: please enter a number")
	}

	if choiceNum < 1 || choiceNum > 2 {
		return 0, fmt.Errorf("invalid choice: please select 1 or 2")
	}

	return choiceNum, nil
}

// handleIncomingRequests fetches and displays incoming friend requests
func handleIncomingRequests(token *TokenData) error {
	fmt.Println("\n📥 Fetching incoming friend requests...")
	
	url := "http://localhost:2000/auth/get_incoming_friend_requests"
	
	requests, err := fetchIncomingFriendRequests(token, url)
	if err != nil {
		return fmt.Errorf("failed to fetch incoming requests: %v", err)
	}

	displayIncomingFriendRequests(requests)
	return nil
}

// handleOutgoingRequests fetches and displays outgoing friend requests
func handleOutgoingRequests(token *TokenData) error {
	fmt.Println("\n📤 Fetching outgoing friend requests...")
	
	url := "http://localhost:2000/auth/get_outgoing_friend_requests"
	
	requests, err := fetchOutgoingFriendRequests(token, url)
	if err != nil {
		return fmt.Errorf("failed to fetch outgoing requests: %v", err)
	}

	displayOutgoingFriendRequests(requests)
	return nil
}

// fetchIncomingFriendRequests makes HTTP request to fetch incoming friend requests
func fetchIncomingFriendRequests(token *TokenData, url string) (*IncomingFriendRequestsResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token.Token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var incomingResponse IncomingFriendRequestsResponse
	err = json.Unmarshal(body, &incomingResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &incomingResponse, nil
}

// fetchOutgoingFriendRequests makes HTTP request to fetch outgoing friend requests
func fetchOutgoingFriendRequests(token *TokenData, url string) (*OutgoingFriendRequestsResponse, error) {
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token.Token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse response
	var outgoingResponse OutgoingFriendRequestsResponse
	err = json.Unmarshal(body, &outgoingResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	return &outgoingResponse, nil
}

// displayIncomingFriendRequests displays the incoming friend requests in a formatted way
func displayIncomingFriendRequests(response *IncomingFriendRequestsResponse) {
	fmt.Printf("\n=== Incoming Friend Requests ===\n")
	
	if response.Message != "" {
		fmt.Printf("Message: %s\n", response.Message)
	}
	
	fmt.Printf("Total incoming requests: %d\n", response.TotalIncoming)
	fmt.Printf("Your User ID: %s\n", response.UserID)
	
	if len(response.IncomingRequests) == 0 {
		fmt.Printf("No incoming friend requests found.\n")
		return
	}

	fmt.Println("─────────────────────────────────────────────────")
	
	for i, request := range response.IncomingRequests {
		fmt.Printf("\n%d. Request ID: %d\n", i+1, request.RequestID)
		fmt.Printf("   From: %s (ID: %s)\n", request.SenderUsername, request.SenderUserID)
		fmt.Printf("   To: %s (ID: %s)\n", request.RecipientUsername, request.RecipientUserID)
		fmt.Printf("   Status: %s\n", request.Status)
		fmt.Printf("   Timestamp: %s\n", request.Timestamp)
		fmt.Printf("   Request Data: %s\n", request.RequestData)
		
		if i < len(response.IncomingRequests)-1 {
			fmt.Println("   ─────────────────────────────")
		}
	}
	
	fmt.Printf("\n=== End of Incoming Friend Requests ===\n")
}

// displayOutgoingFriendRequests displays the outgoing friend requests in a formatted way
func displayOutgoingFriendRequests(response *OutgoingFriendRequestsResponse) {
	fmt.Printf("\n=== Outgoing Friend Requests ===\n")
	
	if response.Message != "" {
		fmt.Printf("Message: %s\n", response.Message)
	}
	
	fmt.Printf("Total outgoing requests: %d\n", response.TotalOutgoing)
	fmt.Printf("Your User ID: %s\n", response.UserID)
	
	if len(response.OutgoingRequests) == 0 {
		fmt.Printf("No outgoing friend requests found.\n")
		return
	}

	fmt.Println("─────────────────────────────────────────────────")
	
	for i, request := range response.OutgoingRequests {
		fmt.Printf("\n%d. Request ID: %d\n", i+1, request.RequestID)
		fmt.Printf("   From: %s (ID: %s)\n", request.SenderUsername, request.SenderUserID)
		fmt.Printf("   To: %s (ID: %s)\n", request.RecipientUsername, request.RecipientUserID)
		fmt.Printf("   Status: %s\n", request.Status)
		fmt.Printf("   Timestamp: %s\n", request.Timestamp)
		fmt.Printf("   Request Data: %s\n", request.RequestData)
		
		if i < len(response.OutgoingRequests)-1 {
			fmt.Println("   ─────────────────────────────")
		}
	}
	
	fmt.Printf("\n=== End of Outgoing Friend Requests ===\n")
}
