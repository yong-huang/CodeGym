package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// Client talks to a local Ollama instance (设计 §9).
type Client struct {
	Endpoint string
	Model    string
	Timeout  time.Duration
	HTTP     *http.Client
}

// New builds an Ollama client. Endpoint/model can be overridden by the
// OLLAMA_ENDPOINT / OLLAMA_MODEL environment variables.
func New(endpoint, model string) *Client {
	if endpoint == "" {
		endpoint = firstNonEmpty(os.Getenv("OLLAMA_ENDPOINT"), "http://localhost:11434")
	}
	if model == "" {
		model = firstNonEmpty(os.Getenv("OLLAMA_MODEL"), "qwen3.5:latest")
	}
	return &Client{Endpoint: endpoint, Model: model, Timeout: 300 * time.Second,
		HTTP: &http.Client{Timeout: 300 * time.Second}}
}

// Generate runs a non-streaming completion and returns the response text.
func (c *Client) Generate(ctx context.Context, prompt string) (string, error) {
	body, _ := json.Marshal(map[string]any{"model": c.Model, "prompt": prompt, "stream": false})
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint+"/api/generate", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return "", fmt.Errorf("Ollama 不可用: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("Ollama 返回 %d", resp.StatusCode)
	}
	var out struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Response, nil
}

// SetConfig updates the endpoint/model at runtime (page-configurable).
func (c *Client) SetConfig(endpoint, model string) {
	if endpoint != "" {
		c.Endpoint = endpoint
	}
	if model != "" {
		c.Model = model
	}
}

// Models lists the models available on the configured Ollama server
// (used both for the settings dropdown and for connectivity tests).
func (c *Client) Models() ([]string, error) {
	req, err := http.NewRequest(http.MethodGet, c.Endpoint+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.HTTP.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Ollama 不可用: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama 返回 %d", resp.StatusCode)
	}
	var out struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, err
	}
	var names []string
	for _, m := range out.Models {
		names = append(names, m.Name)
	}
	return names, nil
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}
