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

// TokenData represents the structure of token.json
type TokenData struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
}

// APIResponse represents the API response structure
type APIResponse struct {
	Message    string `json:"message"`
	SearchedBy string `json:"searched_by"`
	Timestamp  string `json:"timestamp"`
	UserData   struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
	} `json:"user_data"`
}

// Friend represents a friend entry
type Friend struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	AddedAt  string `json:"added_at"`
}

// FriendsData represents the structure of friends.json
type FriendsData struct {
	Friends []Friend `json:"friends"`
}

func main() {
	// Get username from command line argument or prompt
	var username string
	if len(os.Args) > 1 {
		username = os.Args[1]
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
	friendsPath := filepath.Join(configDir, "friends.json")

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

	fmt.Printf("User found: %s (ID: %s)\n", userInfo.UserData.Username, userInfo.UserData.UserID)

	// Add user to friends list
	err = addFriend(friendsPath, userInfo.UserData.UserID, userInfo.UserData.Username)
	if err != nil {
		fmt.Printf("Error adding friend: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Successfully added %s to friends list!\n", userInfo.UserData.Username)
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

func addFriend(friendsPath, userID, username string) error {
	// Create directory if it doesn't exist
	dir := filepath.Dir(friendsPath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}

	// Read existing friends data
	var friendsData FriendsData
	if file, err := os.Open(friendsPath); err == nil {
		defer file.Close()
		decoder := json.NewDecoder(file)
		decoder.Decode(&friendsData) // Ignore error for empty/invalid file
	}

	// Check if friend already exists
	for _, friend := range friendsData.Friends {
		if friend.UserID == userID {
			return fmt.Errorf("user %s is already in friends list", username)
		}
	}

	// Add new friend
	newFriend := Friend{
		UserID:   userID,
		Username: username,
		AddedAt:  time.Now().Format(time.RFC3339),
	}
	friendsData.Friends = append(friendsData.Friends, newFriend)

	// Write updated friends data
	file, err := os.Create(friendsPath)
	if err != nil {
		return fmt.Errorf("failed to create friends file: %w", err)
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	err = encoder.Encode(friendsData)
	if err != nil {
		return fmt.Errorf("failed to encode friends JSON: %w", err)
	}

	return nil
}
