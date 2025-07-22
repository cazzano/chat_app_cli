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

func login_() error {
	// Get username and password from command line arguments or user input
	var username, password string
	
	if len(os.Args) == 3 {
		username = os.Args[1]
		password = os.Args[2]
	} else {
		fmt.Print("Enter username: ")
		fmt.Scanln(&username)
		fmt.Print("Enter password: ")
		fmt.Scanln(&password)
	}

	// Create login request
	loginReq := LoginRequest{
		Username: username,
		Password: password,
	}

	// Convert to JSON
	jsonData, err := json.Marshal(loginReq)
	if err != nil {
		return fmt.Errorf("error marshaling JSON: %v", err)
	}

	fmt.Printf("Sending JSON: %s\n", string(jsonData))

	// Create HTTP client with more detailed request
	client := &http.Client{}
	
	// Create request
	req, err := http.NewRequest("POST", "https://wasalbackend-production.up.railway.app/login", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("error creating request: %v", err)
	}

	// Set headers exactly as in curl
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Go-http-client/1.1")

	fmt.Printf("Making request to: %s\n", req.URL.String())
	fmt.Printf("Headers: %+v\n", req.Header)

	// Make the request
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("error making request: %v", err)
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("error reading response: %v", err)
	}

	fmt.Printf("Response status: %d\n", resp.StatusCode)
	fmt.Printf("Response headers: %+v\n", resp.Header)
	fmt.Printf("Response body: %s\n", string(body))

	// Check if request was successful
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("login failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var loginResp LoginResponse
	if err := json.Unmarshal(body, &loginResp); err != nil {
		return fmt.Errorf("error parsing response: %v", err)
	}

	fmt.Printf("Login successful: %s\n", loginResp.Message)

	// Prepare token data to save
	tokenData := TokenData{
		Token:     loginResp.Token,
		ExpiresIn: loginResp.ExpiresIn,
		UserID:    loginResp.UserID,
		Username:  loginResp.Username,
	}

	// Save token to file
	if err := saveToken(tokenData); err != nil {
		return fmt.Errorf("error saving token: %v", err)
	}

	fmt.Println("Token saved successfully to ~/.config/chat_app/token.json")
	return nil
}

func saveToken(tokenData TokenData) error {
	// Get home directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("error getting home directory: %v", err)
	}

	// Create the config directory path
	configDir := filepath.Join(homeDir, ".config", "chat_app")
	
	// Create directories if they don't exist
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("error creating config directory: %v", err)
	}

	// Create the token file path
	tokenFile := filepath.Join(configDir, "token.json")

	// Convert token data to JSON
	jsonData, err := json.MarshalIndent(tokenData, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling token data: %v", err)
	}

	// Write to file
	if err := os.WriteFile(tokenFile, jsonData, 0600); err != nil {
		return fmt.Errorf("error writing token file: %v", err)
	}

	return nil
}
