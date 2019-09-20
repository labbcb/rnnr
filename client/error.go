package client

import "fmt"

// NetworkError hapens when something went wrong with network connection.
type NetworkError struct {
	error
}

type APIError struct {
	Status     string
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s (%d): %s", e.Status, e.StatusCode, e.Body)
}
