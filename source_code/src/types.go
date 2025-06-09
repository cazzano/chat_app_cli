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

// Friend represents a friend entry (unified structure for compatibility)
type Friend struct {
	// API response fields
	FriendID       string `json:"friend_id"`
	FriendUsername string `json:"friend_username"`
	FriendshipDate string `json:"friendship_date"`
	FriendshipID   int    `json:"friendship_id"`
	
	// Legacy/alternative fields for backward compatibility
	UserID   string `json:"user_id,omitempty"`
	Username string `json:"username,omitempty"`
}

// GetUserID returns the appropriate user ID field
func (f *Friend) GetUserID() string {
	if f.FriendID != "" {
		return f.FriendID
	}
	return f.UserID
}

// GetUsername returns the appropriate username field
func (f *Friend) GetUsername() string {
	if f.FriendUsername != "" {
		return f.FriendUsername
	}
	return f.Username
}

// FriendsAPIResponse represents the API response for get_friends endpoint
type FriendsAPIResponse struct {
	Friends      []Friend `json:"friends"`
	TotalFriends int      `json:"total_friends"`
	UserID       string   `json:"user_id"`
	Username     string   `json:"username"`
}

// FriendsData represents the structure of friends.json (kept for backward compatibility)
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

// IncomingFriendRequest represents a single incoming friend request
type IncomingFriendRequest struct {
	RecipientUserID   string `json:"recipient_user_id"`
	RecipientUsername string `json:"recipient_username"`
	RequestData       string `json:"request_data"`
	RequestID         int    `json:"request_id"`
	SenderUserID      string `json:"sender_user_id"`
	SenderUsername    string `json:"sender_username"`
	Status            string `json:"status"`
	Timestamp         string `json:"timestamp"`
}

// OutgoingFriendRequest represents a single outgoing friend request
type OutgoingFriendRequest struct {
	RecipientUserID   string `json:"recipient_user_id"`
	RecipientUsername string `json:"recipient_username"`
	RequestData       string `json:"request_data"`
	RequestID         int    `json:"request_id"`
	SenderUserID      string `json:"sender_user_id"`
	SenderUsername    string `json:"sender_username"`
	Status            string `json:"status"`
	Timestamp         string `json:"timestamp"`
}

// IncomingFriendRequestsResponse represents the API response for incoming friend requests
type IncomingFriendRequestsResponse struct {
	IncomingRequests []IncomingFriendRequest `json:"incoming_requests"`
	TotalIncoming    int                     `json:"total_incoming"`
	Message          string                  `json:"message"`
	UserID           string                  `json:"user_id"`
}

// OutgoingFriendRequestsResponse represents the API response for outgoing friend requests
type OutgoingFriendRequestsResponse struct {
	OutgoingRequests []OutgoingFriendRequest `json:"outgoing_requests"`
	TotalOutgoing    int                     `json:"total_outgoing"`
	Message          string                  `json:"message"`
	UserID           string                  `json:"user_id"`
}
