package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/labbcb/rnnr/models"
)

const contentType = "application/json"

// ListTasks gets all tasks
func ListTasks(host string) (*models.ListTasksResponse, error) {
	resp, err := http.Get(host + "/ga4gh/tes/v1/tasks")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, raiseHTTPError(resp)
	}

	var r models.ListTasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetTask gets a task by its ID
func GetTask(host, id string) (*models.Task, error) {
	resp, err := http.Get(host + "/ga4gh/tes/v1/tasks/" + id)
	if err != nil {
		return nil, &NetworkError{err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, raiseHTTPError(resp)
	}

	var t models.Task
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// CreateTask submits a task to be executed and return its ID
func CreateTask(host string, t *models.Task) (string, error) {
	// encode models to json
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(t); err != nil {
		return "", err
	}

	// post request to TES endpoint
	resp, err := http.Post(host+"/ga4gh/tes/v1/tasks", "application/json", &b)
	if err != nil {
		return "", &NetworkError{err}
	}
	defer resp.Body.Close()

	// check status code
	if resp.StatusCode != http.StatusCreated {
		return "", raiseHTTPError(resp)
	}

	// decode json from response body
	var r models.CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.ID, nil
}

// CancelTask cancels a task
func CancelTask(host, id string) error {
	resp, err := http.Post(host+"/ga4gh/tes/v1/tasks/"+id+":cancel", "application/json", nil)
	if err != nil {
		return &NetworkError{err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return raiseHTTPError(resp)
	}

	return nil
}

// EnableNode activates a computing node on master server.
func EnableNode(host string, n *models.Node) (id string, err error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(n); err != nil {
		return "", fmt.Errorf("encoding node to json: %w", err)
	}

	resp, err := http.Post(host+"/nodes", contentType, &b)
	if err != nil {
		return "", &NetworkError{err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		return "", raiseHTTPError(resp)
	}

	var res map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decoding node id from json: %w", err)
	}

	return res["id"], nil
}

// DisableNode deactivates a computing note on master server
func DisableNode(host, id string) error {
	// delete node by its id
	c := http.Client{}
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/nodes/%s", host, id), nil)
	if err != nil {
		return fmt.Errorf("creating node deactivation request: %w", err)
	}
	resp, err := c.Do(req)
	if err != nil {
		return &NetworkError{err}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return raiseHTTPError(resp)
	}
	return nil
}

// ListNodes retrieves list of registered nodes on master server
func ListNodes(host string) ([]*models.Node, error) {
	resp, err := http.Get(host + "/nodes")
	if err != nil {
		return nil, &NetworkError{err}
	}
	defer resp.Body.Close()
	// check status code
	if resp.StatusCode != http.StatusOK {
		return nil, raiseHTTPError(resp)
	}
	// decode response body
	var ns []*models.Node
	if err := json.NewDecoder(resp.Body).Decode(&ns); err != nil {
		return nil, err
	}
	return ns, nil
}

// GetNodeInfo retrieves server information
func GetNodeInfo(host string) (*models.Info, error) {
	resp, err := http.Get(host + "/info")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, raiseHTTPError(resp)
	}

	var i models.Info
	if err := json.NewDecoder(resp.Body).Decode(&i); err != nil {
		return nil, fmt.Errorf("parsing info from json: %w", err)
	}
	return &i, nil
}

// new error with 'HTTP Status (Status Code): Body'
// it doesn't close resp.Body reader
func raiseHTTPError(resp *http.Response) error {
	var b bytes.Buffer
	_, err := b.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	return errors.New(fmt.Sprintf("%s: %s", resp.Status, b.String()))
}
