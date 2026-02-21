package syncutil

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// NewJSONRequest creates an HTTP request with a JSON-encoded body.
func NewJSONRequest(ctx context.Context, method, url string, payload any) (*http.Request, error) {
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("marshaling request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
