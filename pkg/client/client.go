package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
)

const defaultBaseURL = "http://127.0.0.1:58141"

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
}

func New() *Client {
	return &Client{
		BaseURL:    defaultBaseURL,
		HTTPClient: &http.Client{},
	}
}

func (c *Client) Block(name string) error {
	requestBody, err := json.Marshal(map[string][]string{
		"names": {name},
	})
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/block", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to add to blocklist: %s", resp.Status)
	}

	return nil
}

func (c *Client) Unblock(name string) error {
	requestBody, err := json.Marshal(map[string][]string{
		"names": {name},
	})
	if err != nil {
		return err
	}

	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/unblock", "application/json", bytes.NewBuffer(requestBody))
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to remove from blocklist: %s", resp.Status)
	}

	return nil
}

func (c *Client) ClearBlocklist() error {
	resp, err := c.HTTPClient.Post(c.BaseURL+"/api/blocklist/clear", "application/json", nil)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to clear blocklist: %s", resp.Status)
	}

	return nil
}

func (c *Client) GetBlocklist() ([]string, error) {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/blocklist")
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get blocklist: %s", resp.Status)
	}

	var list []string
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return nil, err
	}

	return list, nil
}

func (c *Client) SaveBlocklist(dest string) error {
	resp, err := c.HTTPClient.Get(c.BaseURL + "/api/blocklist/save")
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to save blocklist: %s", resp.Status)
	}

	out, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer func() {
		if err := out.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
		}
	}()

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) LoadBlocklist(filePath string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer func() {
		if err := file.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
		}
	}()

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, err := writer.CreateFormFile("file", filePath)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, file)
	if err != nil {
		return err
	}
	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.BaseURL+"/api/blocklist/load", body)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer func() {
		if err := resp.Body.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "Error closing response body: %v\n", err)
		}
	}()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to load blocklist: %s", resp.Status)
	}

	return nil
}
