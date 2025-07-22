package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unsafe"
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

	// Fetch initial conversation with selected friend
	err = fetchConversation(token, selectedFriend)
	if err != nil {
		fmt.Printf("Error fetching conversation: %v\n", err)
		os.Exit(1)
	}

	// Wait for CTRL+R input to refresh, CTRL+S to send message, or CTRL+C to exit
	fmt.Println("\nPress CTRL+R to refresh conversation, CTRL+S to send message, or CTRL+C to exit...")
	waitForCtrlRInReceiveMessage(token, selectedFriend)
	
	return nil
}

// waitForCtrlRInReceiveMessage waits for CTRL+R key combination to refresh conversation or CTRL+S to send message
func waitForCtrlRInReceiveMessage(token *TokenData, friend *Friend) {
	// Set terminal to raw mode to capture key combinations
	oldState, err := makeRawForReceiveMessage(int(os.Stdin.Fd()))
	if err != nil {
		fmt.Printf("Error setting terminal to raw mode: %v\n", err)
		return
	}
	defer restoreForReceiveMessage(int(os.Stdin.Fd()), oldState)

	buffer := make([]byte, 1)
	for {
		n, err := os.Stdin.Read(buffer)
		if err != nil {
			fmt.Printf("Error reading input: %v\n", err)
			return
		}
		
		if n > 0 {
			// Check for CTRL+R (ASCII 18)
			if buffer[0] == 18 {
				// Restore terminal before fetching conversation
				restoreForReceiveMessage(int(os.Stdin.Fd()), oldState)
				
				fmt.Println("\n🔄 Refreshing conversation...")
				err = fetchConversation(token, friend)
				if err != nil {
					fmt.Printf("Error refreshing conversation: %v\n", err)
				}
				
				fmt.Println("\nPress CTRL+R to refresh conversation, CTRL+S to send message, or CTRL+C to exit...")
				
				// Set terminal back to raw mode
				oldState, err = makeRawForReceiveMessage(int(os.Stdin.Fd()))
				if err != nil {
					fmt.Printf("Error setting terminal to raw mode: %v\n", err)
					return
				}
			}
			// Check for CTRL+S (ASCII 19)
			if buffer[0] == 19 {
				// Restore terminal before sending message
				restoreForReceiveMessage(int(os.Stdin.Fd()), oldState)
				
				fmt.Println("\n💬 Send Message Mode")
				err = handleSendMessage(token, friend)
				if err != nil {
					fmt.Printf("Error sending message: %v\n", err)
				}
				
				fmt.Println("\nPress CTRL+R to refresh conversation, CTRL+S to send message, or CTRL+C to exit...")
				
				// Set terminal back to raw mode
				oldState, err = makeRawForReceiveMessage(int(os.Stdin.Fd()))
				if err != nil {
					fmt.Printf("Error setting terminal to raw mode: %v\n", err)
					return
				}
			}
			// Check for CTRL+C (ASCII 3)
			if buffer[0] == 3 {
				fmt.Println("\nExiting...")
				os.Exit(0)
			}
		}
	}
}

// Terminal manipulation functions for Unix-like systems (for receive_message)
type termiosReceiveMessage struct {
	Iflag  uint32
	Oflag  uint32
	Cflag  uint32
	Lflag  uint32
	Cc     [20]uint8
	Ispeed uint32
	Ospeed uint32
}

func makeRawForReceiveMessage(fd int) (*termiosReceiveMessage, error) {
	var oldState termiosReceiveMessage
	
	// Get current terminal settings
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x5401, uintptr(unsafe.Pointer(&oldState)))
	if errno != 0 {
		return nil, errno
	}
	
	// Create new settings for raw mode
	newState := oldState
	newState.Lflag &^= syscall.ICANON | syscall.ECHO
	newState.Cc[syscall.VMIN] = 1
	newState.Cc[syscall.VTIME] = 0
	
	// Apply new settings
	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x5402, uintptr(unsafe.Pointer(&newState)))
	if errno != 0 {
		return nil, errno
	}
	
	return &oldState, nil
}

func restoreForReceiveMessage(fd int, oldState *termiosReceiveMessage) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), 0x5402, uintptr(unsafe.Pointer(oldState)))
	if errno != 0 {
		return errno
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
	url := fmt.Sprintf("https://wasalbackend-production.up.railway.app/auth/conversation/%s", friendUserID)
	
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
	
	// Clear screen for refresh (optional - uncomment if you want to clear screen on refresh)
	// fmt.Print("\033[2J\033[H")
	
	fmt.Printf("\n=== Conversation with %s ===\n", friendUsername)
	fmt.Printf("Total messages in conversation: %d\n", conversation.TotalMessages)
	fmt.Printf("Participants: %v\n", conversation.Participants)
	fmt.Printf("Last updated: %s\n", time.Now().Format("Jan 2, 2006 at 3:04 PM"))
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

// handleSendMessage handles the message sending flow
func handleSendMessage(token *TokenData, friend *Friend) error {
	friendUsername := friend.GetUsername()
	friendUserID := friend.GetUserID()
	
	fmt.Printf("Sending message to: %s\n", friendUsername)
	fmt.Print("Enter your message: ")
	
	// Read message from user
	reader := bufio.NewReader(os.Stdin)
	message, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("error reading message input: %v", err)
	}
	
	message = strings.TrimSpace(message)
	if message == "" {
		fmt.Println("Message cannot be empty. Message sending cancelled.")
		return nil
	}
	
	// Send the message using the API
	fmt.Println("📤 Sending message...")
	err = sendMessageToFriend(token.Token, message, friendUserID)
	if err != nil {
		return fmt.Errorf("failed to send message: %v", err)
	}
	
	fmt.Printf("✅ Message sent successfully to %s!\n", friendUsername)
	
	// Automatically refresh conversation to show the new message
	fmt.Println("🔄 Refreshing conversation to show your message...")
	err = fetchConversation(token, friend)
	if err != nil {
		fmt.Printf("Error refreshing conversation: %v\n", err)
	}
	
	return nil
}

// sendMessageToFriend sends a message using the API (from send_message.go logic)
func sendMessageToFriend(token, message, recipientUID string) error {
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
	req, err := http.NewRequest("POST", "https://wasalbackend-production.up.railway.app/auth/send_message", bytes.NewBuffer(jsonData))
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

	// Parse and display response (optional - for debugging)
	var messageResp MessageResponse
	err = json.Unmarshal(body, &messageResp)
	if err != nil {
		// If parsing fails, just return success since the message was sent
		return nil
	}

	return nil
}
