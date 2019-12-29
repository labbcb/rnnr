package client

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/labbcb/rnnr/models"
)

const contentType = "application/json"

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

func CreateTask(host string, t *models.Task) (string, error) {
	var b bytes.Buffer
	if err := json.NewEncoder(&b).Encode(t); err != nil {
		return "", err
	}

	resp, err := http.Post(host+"/tasks", "application/json", &b)
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

	var r models.CreateTaskResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		return "", err
	}
	return r.ID, nil
}

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

func DisableNode(host, id string) error {
	c := http.Client{}
	req, err := http.NewRequest(http.MethodDelete, fmt.Sprintf("%s/nodes/%s", host, id), nil)
	if err != nil {
		return fmt.Errorf("creating node deactivation request: %w", err)
	}
	resp, err := c.Do(req)
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
	return errors.New(fmt.Sprintf("%s: %s", resp.Status, b.String()))
}
