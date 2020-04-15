package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/labbcb/rnnr/models"
)

const contentType = "application/json"

// ListTasks retrieves all tasks from server.
func ListTasks(host string) (*models.ListTasksResponse, error) {
	resp, err := http.Get(host + "/tasks")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, raiseHTTPError(resp)
	}

	var r models.ListTasksResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return nil, err
	}
	return &r, nil
}

// GetTask retrieves task information given its ID.
func GetTask(host, id string) (*models.Task, error) {
	resp, err := http.Get(host + "/tasks/" + id)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, raiseHTTPError(resp)
	}

	var t models.Task
	if err := json.NewDecoder(resp.Body).Decode(&t); err != nil {
		return nil, err
	}
	return &t, nil
}

// CancelTask cancels task by its ID.
func CancelTask(host, id string) error {
	resp, err := http.Post(host+"/tasks/"+id+":cancel", "application/json", nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return raiseHTTPError(resp)
	}

	return nil
}

// EnableNode subscribes or enables worker node.
func EnableNode(host string, n *models.Node) (id string, err error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(n); err != nil {
		return "", fmt.Errorf("encoding node to json: %w", err)
	}

	resp, err := http.Post(host+"/nodes", contentType, &b)
	if err != nil {
		return "", err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusCreated {
		return "", raiseHTTPError(resp)
	}

	var res map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&res); err != nil {
		return "", fmt.Errorf("decoding node id from json: %w", err)
	}

	return res["host"], nil
}

// DisableNode disables worker node.
// Cancel tells server to cancel running tasks in node and enqueue those tasks.
func DisableNode(host, id string, cancel bool) error {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cancel); err != nil {
		return fmt.Errorf("encoding cancel option to json: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/nodes/%s:disable", host, id), contentType, &b)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return raiseHTTPError(resp)
	}
	return nil
}

// ListNodes retrieves all worker nodes.
func ListNodes(host string) ([]*models.Node, error) {
	resp, err := http.Get(host + "/nodes")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			log.Fatal(err)
		}
	}()
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

// new error with 'HTTP Status (Status Code): Body'
// it doesn't close resp.Body reader
func raiseHTTPError(resp *http.Response) error {
	var b bytes.Buffer
	_, err := b.ReadFrom(resp.Body)
	if err != nil {
		return err
	}
	return fmt.Errorf("%s: %s", resp.Status, b.String())
}
