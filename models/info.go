package models

// ServiceInfo describes information about the service, such as storage details, resource availability, and other documentation.
type ServiceInfo struct {
	// Returns the name of the service
	Name string `json:"name,omitempty"`
	// Returns a documentation string
	Doc string `json:"doc,omitempty"`
	// Lists some, but not necessarily all, storage locations supported by the service.
	Storage []string `json:"storage,omitempty"`
}
