package cc

import (
	"encoding/json"
	"os"
)

// JSONResponse is a small JSON envelope for cc commands.
type JSONResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message,omitempty"`
	Error   string `json:"error,omitempty"`
	Data    any    `json:"data,omitempty"`
}

// EmitJSON writes a response to stdout.
func EmitJSON(r JSONResponse) error {
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	return enc.Encode(r)
}
