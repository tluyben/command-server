package commands

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type FetchCommand struct{}

func (c *FetchCommand) Execute(args map[string]interface{}, writer ResponseWriter) error {
    // Extract arguments
    method, ok := args["method"].(string)
    if !ok {
        return fmt.Errorf("method must be a string")
    }

    url, ok := args["url"].(string)
    if !ok {
        return fmt.Errorf("url must be a string")
    }

    stream, _ := args["stream"].(bool)

    var body io.Reader
    if bodyData, ok := args["body"]; ok {
        bodyBytes, err := json.Marshal(bodyData)
        if err != nil {
            return fmt.Errorf("failed to marshal body: %v", err)
        }
        body = bytes.NewBuffer(bodyBytes)
    }

    // Create request
    req, err := http.NewRequest(method, url, body)
    if err != nil {
        return fmt.Errorf("failed to create request: %v", err)
    }

    // Add headers
    if headers, ok := args["headers"].(map[string]interface{}); ok {
        for key, value := range headers {
            if strValue, ok := value.(string); ok {
                req.Header.Set(key, strValue)
            }
        }
    }

    // Execute request
    client := &http.Client{}
    resp, err := client.Do(req)
    if err != nil {
        return fmt.Errorf("request failed: %v", err)
    }
    defer resp.Body.Close()

    // Handle streaming response if requested
    if stream {
        return c.handleStreamingResponse(resp, writer)
    }

    // Handle regular response
    return c.handleRegularResponse(resp, writer)
}

func (c *FetchCommand) handleStreamingResponse(resp *http.Response, writer ResponseWriter) error {
    // Send initial response with headers
    err := writer.Stream("start", map[string]interface{}{
        "statuscode": resp.StatusCode,
        "headers":    resp.Header,
    })
    if err != nil {
        return err
    }

    // Create a buffer for reading chunks
    buffer := make([]byte, 1024)
    for {
        n, err := resp.Body.Read(buffer)
        if n > 0 {
            chunk := buffer[:n]
            // Try to parse as JSON if possible
            var jsonData interface{}
            if err := json.Unmarshal(chunk, &jsonData); err == nil {
                writer.Stream("data", jsonData)
            } else {
                writer.Stream("data", string(chunk))
            }
        }
        if err == io.EOF {
            break
        }
        if err != nil {
            writer.Stream("error", map[string]string{"error": err.Error()})
            break
        }
    }

    return writer.End()
}

func (c *FetchCommand) handleRegularResponse(resp *http.Response, writer ResponseWriter) error {
    // Read response body
    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        return fmt.Errorf("failed to read response body: %v", err)
    }

    // Convert response headers
    headers := make(map[string][]string)
    for k, v := range resp.Header {
        headers[k] = v
    }

    // Parse JSON if content-type is application/json
    var responseBody interface{} = string(respBody)
    if ct := resp.Header.Get("Content-Type"); ct == "application/json" {
        var jsonBody interface{}
        if err := json.Unmarshal(respBody, &jsonBody); err != nil {
            return fmt.Errorf("failed to parse JSON response: %v", err)
        }
        responseBody = jsonBody
    }

    return writer.WriteJSON(resp.StatusCode, headers, responseBody)
}

func init() {
    RegisterCommand("fetch", &FetchCommand{})
}
