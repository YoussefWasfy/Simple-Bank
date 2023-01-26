package token

import "time"

// An interface for managing tokens
type Maker interface {
	// Creates a new token for a specific username and duration
	CreateToken(username string, duration time.Duration) (string, error)

	// Check if the input token is valid or not
	VerifyToken(token string) (*Payload, error)
}
