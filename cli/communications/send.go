package communications

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/UnifyEM/UnifyEM/cli/global"
)

// sendRequest is a lower level function that sends HTTP requests
func (c *Communications) sendRequest(method, endpoint string, payload []byte) (int, []byte, error) {

	// Build the request URL
	url := fmt.Sprintf("%s%s", global.ServerURL, endpoint)

	// Create a new HTTP request
	httpReq, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return 0, nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	// Set the Authorization header if a token is present
	if c.token != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.token))
	}

	// Set the appropriate headers
	if method == "POST" || method == "PUT" {
		httpReq.Header.Set("Content-Type", "application/json")
	}

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to send HTTP request: %w", err)
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)

	// Read the response body
	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, fmt.Errorf("failed to read response body: %w", err)
	}

	return resp.StatusCode, responseBody.Bytes(), nil
}
