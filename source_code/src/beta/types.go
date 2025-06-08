package main

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

// LoginRequest represents the request payload for login
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// LoginResponse represents the response from the login API
type LoginResponse struct {
	ExpiresIn string `json:"expires_in"`
	Message   string `json:"message"`
	Token     string `json:"token"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
}

// MessageRequest represents the request payload for sending a message
type MessageRequest struct {
	Message         string `json:"message"`
	RecipientUserID string `json:"recipient_user_id"`
}

// MessageResponse represents the API response
type MessageResponse struct {
	Message   string `json:"message"`
	MessageID int    `json:"message_id"`
	Recipient string `json:"recipient"`
	Sender    string `json:"sender"`
	Timestamp string `json:"timestamp"`
}
