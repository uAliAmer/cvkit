package cv

import (
	"encoding/json"
	"fmt"
	"os"
)

// ReadValidated reads a JSON file and confirms it parses as valid JSON,
// returning the raw bytes. Used by sync, which copies the source verbatim
// (matching sync_portfolio_data.py) but refuses to propagate broken JSON.
func ReadValidated(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if !json.Valid(raw) {
		return nil, fmt.Errorf("%s is not valid JSON", path)
	}
	return raw, nil
}
