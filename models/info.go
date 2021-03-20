package models

import "time"

// ServiceInfo describes information about the service, such as storage details, resource availability, and other documentation.
type ServiceInfo struct {
	ID               string        `json:"id"`
	Name             string        `json:"name"`
	Type             *ServiceType  `json:"type"`
	Description      string        `json:"description"`
	Organization     *Organization `json:"organization"`
	ContactURL       string        `json:"contactUrl"`
	DocumentationURL string        `json:"documentationUrl,omitempty"`
	Storage          []string      `json:"storage,omitempty"`
	CreatedAt        time.Time     `json:"createdAt"`
	UpdatedAt        time.Time     `json:"updatedAt"`
	Environment      string        `json:"environment"`
	Version          string        `json:"version"`
}

type ServiceType struct {
	Group    string `json:"group"`
	Artifact string `json:"artifact"`
	Version  string `json:"version"`
}

type Organization struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}
