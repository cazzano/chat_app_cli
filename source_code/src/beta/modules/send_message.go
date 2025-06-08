package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// TokenData represents the structure of the token file
type TokenData struct {
	Token     string `json:"token"`
	ExpiresIn string `json:"expires_in"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
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

	// Ask for recipient UID
	fmt.Print("Enter recipient's UID: ")
	var recipientUID string
	fmt.Scanln(&recipientUID)

	if recipientUID == "" {
		fmt.Println("Recipient UID cannot be empty")
		os.Exit(1)
	}

	// Send message
	err = sendMessage(token.Token, message, recipientUID)
	if err != nil {
		fmt.Printf("Error sending message: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Message sent successfully!")
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
