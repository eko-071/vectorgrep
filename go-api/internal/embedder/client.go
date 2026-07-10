package embedder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

type StatusError struct {
	StatusCode int
	Message    string
}

func (e *StatusError) Error() string {
	return e.Message
}

type embedRequest struct {
	Text string `json:"text"`
}

type embedResponse struct {
	Embedding []float32 `json:"embedding"`
}

func (c *Client) Embed(text string) ([]float32, error) {
	body, err := json.Marshal(embedRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("call embedding service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("embedding service returned status %d", resp.StatusCode)
	}

	var result embedResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result.Embedding, nil
}

type ingestRequest struct {
	Command string `json:"command"`
}

type IngestResult struct {
	Command       string `json:"command"`
	ChunksIndexed int    `json:"chunks_indexed"`
}

func (c *Client) Ingest(command string) (*IngestResult, error) {
	body, err := json.Marshal(ingestRequest{Command: command})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/ingest", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, &StatusError{StatusCode: http.StatusServiceUnavailable, Message: "embedding service unreachable"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Detail string `json:"detail"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		msg := errBody.Detail
		if msg == "" {
			msg = fmt.Sprintf("ingestion service returned status %d", resp.StatusCode)
		}
		return nil, &StatusError{StatusCode: resp.StatusCode, Message: msg}
	}

	var result IngestResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}

type SearchResult struct {
	ID       string         `json:"id"`
	Text     string         `json:"text"`
	Metadata map[string]any `json:"metadata"`
	Score    float64        `json:"score"`
}

type searchResponse struct {
	Query   string         `json:"query"`
	Results []SearchResult `json:"results"`
}

func (c *Client) Search(query string, topK int) (*searchResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("top_k", fmt.Sprintf("%d", topK))

	resp, err := c.httpClient.Get(c.baseURL + "/search?" + params.Encode())
	if err != nil {
		return nil, &StatusError{StatusCode: http.StatusServiceUnavailable, Message: "search service unreachable"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errBody struct {
			Detail string `json:"detail"`
		}
		json.NewDecoder(resp.Body).Decode(&errBody)
		msg := errBody.Detail
		if msg == "" {
			msg = fmt.Sprintf("search service returned status %d", resp.StatusCode)
		}
		return nil, &StatusError{StatusCode: resp.StatusCode, Message: msg}
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
