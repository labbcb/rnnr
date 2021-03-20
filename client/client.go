// Package client implements Task Execution Service API for requesting servers.
package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/labbcb/rnnr/models"
)

const contentType = "application/json"

// ListTasks retrieves tasks from server that matches worker nodes and task states.
// Pagination is done via pageSize and pageToken parameters.
// view defines task fields to be returned.
//
// Minimal returns only task ID and state.
//
// Basic returns all fields except Logs.ExecutorLogs.Stdout, Logs.ExecutorLogs.Stderr, Inputs.Content and Logs.SystemLogs.
//
// Full returns all fields.
func ListTasks(host string, pageSize int, pageToken string, view models.View, nodes []string, states []models.State) (*models.ListTasksResponse, error) {
	u, err := url.Parse(host + "/v1/tasks")
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("pageSize", fmt.Sprint(pageSize))
	v.Set("pageToken", pageToken)
	v.Set("view", string(view))
	for _, state := range states {
		v.Add("state", string(state))
	}
	for _, node := range nodes {
		v.Add("node", string(node))
	}
	u.RawQuery = v.Encode()

	resp, err := http.Get(u.String())
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
	resp, err := http.Get(host + "/v1/tasks/" + id)
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
	resp, err := http.Post(host+"/v1/tasks/"+id+":cancel", "application/json", nil)
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

// EnableNode enables worker node at main server returning its ID.
func EnableNode(host string, n *models.Node) (id string, err error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(n); err != nil {
		return "", fmt.Errorf("encoding node to json: %w", err)
	}

	resp, err := http.Post(host+"/v1/nodes", contentType, &b)
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
// cancel tells server to cancel running tasks in the worker node and enqueue those tasks.
func DisableNode(host, id string, cancel bool) error {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(cancel); err != nil {
		return fmt.Errorf("encoding cancel option to json: %w", err)
	}

	resp, err := http.Post(fmt.Sprintf("%s/v1/nodes/%s:disable", host, id), contentType, &b)
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
func ListNodes(host string, onlyActive bool) ([]*models.Node, error) {
	u, err := url.Parse(host + "/v1/nodes")
	if err != nil {
		return nil, err
	}

	v := url.Values{}
	v.Set("active", strconv.FormatBool(onlyActive))
	u.RawQuery = v.Encode()

	resp, err := http.Get(u.String())
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
