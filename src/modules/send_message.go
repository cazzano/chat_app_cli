package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
)

// TokenData represents the structure of the token file
type TokenData struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
}

// Friend represents a friend in the friends list
type Friend struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	AddedAt  string `json:"added_at"`
}

// FriendsData represents the structure of the friends file
type FriendsData struct {
	Friends []Friend `json:"friends"`
}

// MessageRequest represents the request payload for sending a message
type MessageRequest struct {
	Message           string `json:"message"`
	RecipientUserID   string `json:"recipient_user_id"`
}

// MessageResponse represents the API response
type MessageResponse struct {
	Message     string `json:"message"`
	MessageID   int    `json:"message_id"`
	Recipient   string `json:"recipient"`
	Sender      string `json:"sender"`
	Timestamp   string `json:"timestamp"`
}

func main() {
	// Check if message argument is provided
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go \"Your message here\"")
		os.Exit(1)
	}

	message := os.Args[1]

	// Read token from config file
	token, err := readTokenFromConfig()
	if err != nil {
		fmt.Printf("Error reading token: %v\n", err)
		os.Exit(1)
	}

	// Read friends from config file
	friends, err := readFriendsFromConfig()
	if err != nil {
		fmt.Printf("Error reading friends: %v\n", err)
		os.Exit(1)
	}

	// Check if friends list is empty
	if len(friends.Friends) == 0 {
		fmt.Println("No friends found in your friends list.")
		os.Exit(1)
	}

	// Display friends and ask user to select
	selectedFriend, err := selectFriend(friends)
	if err != nil {
		fmt.Printf("Error selecting friend: %v\n", err)
		os.Exit(1)
	}

	// Send message to selected friend
	err = sendMessage(token.Token, message, selectedFriend.UserID)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Message sent successfully to %s!\n", selectedFriend.Username)
}

// readTokenFromConfig reads the token from ~/.config/chat_app/token.json
func readTokenFromConfig() (*TokenData, error) {
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

// readFriendsFromConfig reads the friends list from ~/.config/chat_app/friends.json
func readFriendsFromConfig() (*FriendsData, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get home directory: %v", err)
	}

	friendsPath := filepath.Join(homeDir, ".config", "chat_app", "friends.json")
	
	file, err := os.Open(friendsPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open friends file: %v", err)
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read friends file: %v", err)
	}

	var friendsData FriendsData
	err = json.Unmarshal(data, &friendsData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse friends file: %v", err)
	}

	return &friendsData, nil
}

// selectFriend displays the friends list and asks user to select one
func selectFriend(friends *FriendsData) (*Friend, error) {
	fmt.Println("\n--- Your Friends ---")
	for i, friend := range friends.Friends {
		fmt.Printf("%d. %s (ID: %s)\n", i+1, friend.Username, friend.UserID)
	}
	
	fmt.Print("\nEnter the number of the friend you want to send the message to: ")
	var choice string
	fmt.Scanln(&choice)
	
	// Convert choice to integer
	choiceNum, err := strconv.Atoi(choice)
	if err != nil {
		return nil, fmt.Errorf("invalid choice: please enter a number")
	}
	
	// Validate choice
	if choiceNum < 1 || choiceNum > len(friends.Friends) {
		return nil, fmt.Errorf("invalid choice: please select a number between 1 and %d", len(friends.Friends))
	}
	
	// Return selected friend (subtract 1 for 0-based indexing)
	selectedFriend := &friends.Friends[choiceNum-1]
	fmt.Printf("Selected: %s\n", selectedFriend.Username)
	
	return selectedFriend, nil
}

// sendMessage sends a message using the API
func sendMessage(token, message, recipientUID string) error {
	// Prepare request payload
	messageReq := MessageRequest{
		Message:         message,
		RecipientUserID: recipientUID,
	}

	jsonData, err := json.Marshal(messageReq)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %v", err)
	}

	// Create HTTP request
	req, err := http.NewRequest("POST", "http://localhost:2000/auth/send_message", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)

	// Send request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Parse and display response
	var messageResp MessageResponse
	err = json.Unmarshal(body, &messageResp)
	if err != nil {
		fmt.Printf("Response: %s\n", string(body))
		return nil
	}

	// Display formatted response
	fmt.Printf("\n--- Message Details ---\n")
	fmt.Printf("Message ID: %d\n", messageResp.MessageID)
	fmt.Printf("From: %s\n", messageResp.Sender)
	fmt.Printf("To: %s\n", messageResp.Recipient)
	fmt.Printf("Timestamp: %s\n", messageResp.Timestamp)
	fmt.Printf("Status: %s\n", messageResp.Message)

	return nil
}
