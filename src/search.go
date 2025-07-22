package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// All types are now defined in types.go

// FriendRequestPayload represents the request payload for sending friend request
type FriendRequestPayload struct {
	Username string `json:"username"`
}

// FriendRequestResponse represents the API response for friend request
type FriendRequestResponse struct {
	Message   string `json:"message"`
	RequestID int    `json:"request_id,omitempty"`
	Recipient string `json:"recipient,omitempty"`
	Sender    string `json:"sender,omitempty"`
	Timestamp string `json:"timestamp,omitempty"`
}

var authToken string

func friend() error {
	// Get config directory path
	homeDir, err := os.UserHomeDir()
	if err != nil {
		fmt.Printf("Error getting home directory: %v\n", err)
		os.Exit(1)
	}

	configDir := filepath.Join(homeDir, ".config", "chat_app")
	tokenPath := filepath.Join(configDir, "token.json")

	// Read token from file
	token, err := readToken(tokenPath)
	if err != nil {
		fmt.Printf("Error reading token: %v\n", err)
		os.Exit(1)
	}
	authToken = token.Token

	fmt.Println("Chat App - User Search")
	fmt.Println("Commands:")
	fmt.Println("- Type username to search")
	fmt.Println("- Type 'quit' or 'exit' to exit")
	fmt.Println("----------------------------------------------------")

	// Main input loop
	for {
		fmt.Print("\nEnter username to search (or 'quit' to exit): ")
		
		var input string
		fmt.Scanln(&input)

		if input == "quit" || input == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		if input == "" {
			fmt.Println("Please enter a username")
			continue
		}

		// Search for user via API
		userInfo, err := searchUser(input, authToken)
		if err != nil {
			fmt.Printf("Error searching user: %v\n", err)
			continue
		}

		// Display user information
		fmt.Printf("\n✓ User found: %s (ID: %s)\n", userInfo.UserData.Username, userInfo.UserData.UserID)
		fmt.Printf("  Search performed by: %s\n", userInfo.SearchedBy)
		fmt.Printf("  Search timestamp: %s\n", userInfo.Timestamp)
		
		// Ask if user wants to send friend request
		fmt.Printf("\nDo you want to send a friend request to %s? (y/n): ", userInfo.UserData.Username)
		var choice string
		fmt.Scanln(&choice)
		
		if choice == "y" || choice == "Y" || choice == "yes" || choice == "Yes" || choice == "YES" {
			err := sendFriendRequest(userInfo.UserData.Username, authToken)
			if err != nil {
				fmt.Printf("❌ Error sending friend request: %v\n", err)
			}
		} else {
			fmt.Println("Friend request not sent.")
		}
	}
	
	return nil
}

func readToken(tokenPath string) (*TokenData, error) {
	file, err := os.Open(tokenPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open token file: %w", err)
	}
	defer file.Close()

	var tokenData TokenData
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&tokenData)
	if err != nil {
		return nil, fmt.Errorf("failed to decode token JSON: %w", err)
	}

	return &tokenData, nil
}

func searchUser(username, token string) (*APIResponse, error) {
	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request
	req, err := http.NewRequest("GET", "https://wasalbackend-production.up.railway.app/auth/search_user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("username", username)

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var apiResponse APIResponse
	decoder := json.NewDecoder(resp.Body)
	err = decoder.Decode(&apiResponse)
	if err != nil {
		return nil, fmt.Errorf("failed to decode API response: %w", err)
	}

	return &apiResponse, nil
}

func sendFriendRequest(username, token string) error {
	// Create HTTP client
	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	// Create request payload
	payload := FriendRequestPayload{
		Username: username,
	}

	// Convert payload to JSON
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal request payload: %w", err)
	}

	// Create request
	req, err := http.NewRequest("POST", "https://wasalbackend-production.up.railway.app/auth/send_friend_request", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	// Make request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response body: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var friendResponse FriendRequestResponse
	err = json.Unmarshal(body, &friendResponse)
	if err != nil {
		return fmt.Errorf("failed to decode API response: %w", err)
	}

	// Display success message
	fmt.Printf("✓ Friend request sent successfully to %s!\n", username)
	fmt.Printf("  Message: %s\n", friendResponse.Message)
	if friendResponse.RequestID != 0 {
		fmt.Printf("  Request ID: %d\n", friendResponse.RequestID)
	}
	if friendResponse.Timestamp != "" {
		fmt.Printf("  Timestamp: %s\n", friendResponse.Timestamp)
	}

	return nil
}
