package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"syscall"

	"golang.org/x/term"
)

// RegistrationResponse represents the API response structure
type RegistrationResponse struct {
	Message string `json:"message,omitempty"`
	UserID  string `json:"user_id,omitempty"`
	Error   string `json:"error,omitempty"`
}

func main() {
	fmt.Println("=== User Registration ===")
	
	// Get username
	username, err := getUsername()
	if err != nil {
		fmt.Printf("Error getting username: %v\n", err)
		os.Exit(1)
	}

	// Get password
	password, err := getPassword()
	if err != nil {
		fmt.Printf("Error getting password: %v\n", err)
		os.Exit(1)
	}

	// Register user
	err = registerUser(username, password)
	if err != nil {
		fmt.Printf("Registration failed: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Registration completed successfully!")
}

// getUsername prompts for and validates username input
func getUsername() (string, error) {
	fmt.Print("Enter username: ")
	var username string
	fmt.Scanln(&username)

	// Validate username
	username = strings.TrimSpace(username)
	if username == "" {
		return "", fmt.Errorf("username cannot be empty")
	}

	if len(username) < 3 {
		return "", fmt.Errorf("username must be at least 3 characters long")
	}

	// Check for invalid characters (basic validation)
	if strings.Contains(username, " ") {
		return "", fmt.Errorf("username cannot contain spaces")
	}

	return username, nil
}

// getPassword prompts for password input (hidden input)
func getPassword() (string, error) {
	fmt.Print("Enter password: ")
	
	// Hide password input
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}
	
	fmt.Println() // Print newline after hidden input

	password := string(bytePassword)
	password = strings.TrimSpace(password)

	// Validate password
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	if len(password) < 6 {
		return "", fmt.Errorf("password must be at least 6 characters long")
	}

	// Confirm password
	fmt.Print("Confirm password: ")
	byteConfirmPassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password confirmation: %v", err)
	}
	
	fmt.Println() // Print newline after hidden input

	confirmPassword := string(byteConfirmPassword)
	confirmPassword = strings.TrimSpace(confirmPassword)

	if password != confirmPassword {
		return "", fmt.Errorf("passwords do not match")
	}

	return password, nil
}

// registerUser sends registration request to the API
func registerUser(username, password string) error {
	// API endpoint
	url := "http://localhost:5000/register"

	// Create HTTP request
	req, err := http.NewRequest("POST", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set headers with username and password
	req.Header.Set("username", username)
	req.Header.Set("password", password)
	req.Header.Set("Content-Type", "application/json")

	// Display request details (for debugging)
	fmt.Printf("\nSending registration request...\n")
	fmt.Printf("URL: %s\n", url)
	fmt.Printf("Username: %s\n", username)
	fmt.Printf("Password: %s\n", strings.Repeat("*", len(password)))

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

	// Display response details
	fmt.Printf("\n--- API Response ---\n")
	fmt.Printf("Status Code: %d\n", resp.StatusCode)
	fmt.Printf("Status: %s\n", resp.Status)

	// Handle different response codes
	switch resp.StatusCode {
	case http.StatusOK, http.StatusCreated:
		fmt.Printf("✅ Registration successful!\n")
		fmt.Printf("Response: %s\n", string(body))
		return nil
	case http.StatusBadRequest:
		fmt.Printf("❌ Bad Request: %s\n", string(body))
		return fmt.Errorf("registration failed - bad request: %s", string(body))
	case http.StatusConflict:
		fmt.Printf("❌ Username already exists: %s\n", string(body))
		return fmt.Errorf("username '%s' is already taken", username)
	case http.StatusInternalServerError:
		fmt.Printf("❌ Server Error: %s\n", string(body))
		return fmt.Errorf("server error occurred: %s", string(body))
	default:
		fmt.Printf("❌ Unexpected response: %s\n", string(body))
		return fmt.Errorf("registration failed with status %d: %s", resp.StatusCode, string(body))
	}
}
