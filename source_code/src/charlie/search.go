package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// All types are now defined in types.go

func friend() error {
	// Get username from command line argument or prompt
	var username string
	if len(os.Args) > 2 && os.Args[2] != "" {
		username = os.Args[2]
	} else {
		fmt.Print("Enter username to search: ")
		fmt.Scanln(&username)
	}

	if username == "" {
		fmt.Println("Error: Username cannot be empty")
		os.Exit(1)
	}

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

	// Search for user via API
	userInfo, err := searchUser(username, token.Token)
	if err != nil {
		fmt.Printf("Error searching user: %v\n", err)
		os.Exit(1)
	}

	// Display user information
	fmt.Printf("User found: %s (ID: %s)\n", userInfo.UserData.Username, userInfo.UserData.UserID)
	fmt.Printf("Search performed by: %s\n", userInfo.SearchedBy)
	fmt.Printf("Search timestamp: %s\n", userInfo.Timestamp)
	
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
	req, err := http.NewRequest("GET", "http://localhost:2000/auth/search_user", nil)
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
