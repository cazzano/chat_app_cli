package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// Message represents a single message in the conversation
type Message struct {
	Direction   string `json:"direction"`
	IsRead      bool   `json:"is_read"`
	Message     string `json:"message"`
	MessageID   int    `json:"message_id"`
	Recipient   string `json:"recipient"`
	Sender      string `json:"sender"`
	Timestamp   string `json:"timestamp"`
}

// ConversationResponse represents the API response for conversation
type ConversationResponse struct {
	Conversation  []Message `json:"conversation"`
	Participants  []string  `json:"participants"`
	TotalMessages int       `json:"total_messages"`
}

func receive_message() error {
	// Read token from config file
	token, err := readTokenForReceiveMessage()
	if err != nil {
		fmt.Printf("Error reading token: %v\n", err)
		os.Exit(1)
	}

	// Fetch friends from API instead of local file for consistency
	friends, err := fetchFriendsFromAPI(token.Token)
	if err != nil {
		fmt.Printf("Error fetching friends: %v\n", err)
		os.Exit(1)
	}

	// Check if friends list is empty
	if len(friends.Friends) == 0 {
		fmt.Println("No friends found in your friends list.")
		os.Exit(1)
	}

	// Display friends and ask user to select
	selectedFriend, err := selectFriendForReceiveMessage(friends)
	if err != nil {
		fmt.Printf("Error selecting friend: %v\n", err)
		os.Exit(1)
	}

	// Fetch conversation with selected friend
	err = fetchConversation(token, selectedFriend)
	if err != nil {
		fmt.Printf("Error fetching conversation: %v\n", err)
		os.Exit(1)
	}
	return nil
}

// readTokenForReceiveMessage reads the token from ~/.config/chat_app/token.json
func readTokenForReceiveMessage() (*TokenData, error) {
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

// readFriendsForReceiveMessage reads the friends list from ~/.config/chat_app/friends.json
// Kept for backward compatibility but now also supports API fetching
func readFriendsForReceiveMessage() (*FriendsData, error) {
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

// selectFriendForReceiveMessage displays the friends list and asks user to select one
func selectFriendForReceiveMessage(friends *FriendsData) (*Friend, error) {
	fmt.Println("\n--- Your Friends ---")
	for i, friend := range friends.Friends {
		username := friend.GetUsername()
		userID := friend.GetUserID()
		fmt.Printf("%d. %s (ID: %s)\n", i+1, username, userID)
	}
	
	fmt.Print("\nEnter the number of the friend whose conversation you want to view: ")
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
	fmt.Printf("Selected: %s\n", selectedFriend.GetUsername())
	
	return selectedFriend, nil
}

// fetchConversation fetches and displays the conversation with the selected friend
func fetchConversation(token *TokenData, friend *Friend) error {
	// Build API URL using the appropriate user ID
	friendUserID := friend.GetUserID()
	url := fmt.Sprintf("http://localhost:2000/auth/conversation/%s", friendUserID)
	
	// Create HTTP request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %v", err)
	}

	// Set authorization header
	req.Header.Set("Authorization", "Bearer "+token.Token)

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

	// Parse response
	var conversation ConversationResponse
	err = json.Unmarshal(body, &conversation)
	if err != nil {
		return fmt.Errorf("failed to parse response: %v", err)
	}

	// Display conversation
	displayConversation(token, friend, &conversation)
	
	return nil
}

// displayConversation displays the filtered conversation between you and the selected friend
func displayConversation(token *TokenData, friend *Friend, conversation *ConversationResponse) {
	friendUsername := friend.GetUsername()
	friendUserID := friend.GetUserID()
	
	fmt.Printf("\n=== Conversation with %s ===\n", friendUsername)
	fmt.Printf("Total messages in conversation: %d\n", conversation.TotalMessages)
	fmt.Printf("Participants: %v\n", conversation.Participants)
	fmt.Println(strings.Repeat("=", 50))

	// Filter messages between you and the selected friend only
	var filteredMessages []Message
	for _, msg := range conversation.Conversation {
		// Only include messages where either:
		// - You sent to this friend (sender = your ID, recipient = friend ID)
		// - This friend sent to you (sender = friend ID, recipient = your ID)
		if (msg.Sender == token.UserID && msg.Recipient == friendUserID) ||
		   (msg.Sender == friendUserID && msg.Recipient == token.UserID) {
			filteredMessages = append(filteredMessages, msg)
		}
	}

	if len(filteredMessages) == 0 {
		fmt.Printf("No messages found between you and %s.\n", friendUsername)
		return
	}

	fmt.Printf("\nMessages between you and %s (%d messages):\n\n", friendUsername, len(filteredMessages))

	// Display filtered messages
	for _, msg := range filteredMessages {
		// Parse timestamp
		timestamp, err := time.Parse("2006-01-02 15:04:05", msg.Timestamp)
		var timeStr string
		if err != nil {
			timeStr = msg.Timestamp // Use original if parsing fails
		} else {
			timeStr = timestamp.Format("Jan 2, 2006 at 3:04 PM")
		}

		// Determine message direction and display accordingly
		if msg.Sender == token.UserID {
			// Message sent by you
			fmt.Printf("📤 [%s] You: %s\n", timeStr, msg.Message)
			if !msg.IsRead {
				fmt.Printf("   Status: Delivered\n")
			} else {
				fmt.Printf("   Status: Read\n")
			}
		} else {
			// Message received from friend
			fmt.Printf("📥 [%s] %s: %s\n", timeStr, friendUsername, msg.Message)
			if !msg.IsRead {
				fmt.Printf("   Status: Unread\n")
			} else {
				fmt.Printf("   Status: Read\n")
			}
		}
		
		fmt.Printf("   Message ID: %d\n", msg.MessageID)
		fmt.Println(strings.Repeat("-", 40))
	}

	fmt.Printf("\nEnd of conversation with %s\n", friendUsername)
}
