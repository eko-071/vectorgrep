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

func NewClient(baseURL string, timeout time.Duration) *Client {
	return &Client{
		baseURL:    baseURL,
		httpClient: &http.Client{Timeout: timeout},
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

func (c *Client) decodeErrorResponse(resp *http.Response, service string) error {
	var errBody struct {
		Detail string `json:"detail"`
	}
	json.NewDecoder(resp.Body).Decode(&errBody)
	msg := errBody.Detail
	if msg == "" {
		msg = fmt.Sprintf("%s returned status %d", service, resp.StatusCode)
	}
	return &StatusError{StatusCode: resp.StatusCode, Message: msg}
}

func (c *Client) Embed(text string) ([]float32, error) {
	body, err := json.Marshal(embedRequest{Text: text})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	resp, err := c.httpClient.Post(c.baseURL+"/embed", "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, &StatusError{StatusCode: http.StatusServiceUnavailable, Message: "embedding service unreachable"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeErrorResponse(resp, "embedding service")
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
		return nil, c.decodeErrorResponse(resp, "ingestion service")
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

func (c *Client) Search(query string, topK int, scoreThreshold float64) (*searchResponse, error) {
	params := url.Values{}
	params.Set("q", query)
	params.Set("top_k", fmt.Sprintf("%d", topK))
	if scoreThreshold > 0 {
		params.Set("score_threshold", fmt.Sprintf("%.2f", scoreThreshold))
	}

	resp, err := c.httpClient.Get(c.baseURL + "/search?" + params.Encode())
	if err != nil {
		return nil, &StatusError{StatusCode: http.StatusServiceUnavailable, Message: "search service unreachable"}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.decodeErrorResponse(resp, "search service")
	}

	var result searchResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &result, nil
}
