package handlers

// ErrorResponse is the standard error body returned on 4xx / 5xx responses.
//
// Example: {"error": "invalid payload"}
type ErrorResponse struct {
	Error string `json:"error"`
}

