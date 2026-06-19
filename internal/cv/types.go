// Package cv loads CV data and renders it to LaTeX and portfolio JSON.
package cv

import (
	"encoding/json"
	"fmt"
	"os"
)

// CV is the single source of truth, mirroring cv_data.json.
type CV struct {
	Name       string            `json:"name"`
	Phone      string            `json:"phone"`
	Email      string            `json:"email"`
	Location   string            `json:"location"`
	LinkedIn   string            `json:"linkedin"`
	GitHub     string            `json:"github"`
	Telegram   string            `json:"telegram"`
	Summary    string            `json:"summary"`
	About      []string          `json:"about"`
	Experience []Experience      `json:"experience"`
	Projects   []Project         `json:"projects"`
	Education  []Education       `json:"education"`
	Skills     map[string]string `json:"skills"`

	// skillsOrder preserves the key order from the source JSON, since Go maps
	// are unordered but the rendered SKILLS section must be stable.
	skillsOrder []string
}

// Experience is one job entry. Bullets is a newline-joined string.
type Experience struct {
	Company  string `json:"company"`
	Dates    string `json:"dates"`
	Title    string `json:"title"`
	Location string `json:"location"`
	Bullets  string `json:"bullets"`
}

// Project is one project entry. Bullets is a newline-joined string.
type Project struct {
	Name    string `json:"name"`
	Link    string `json:"link"`
	Dates   string `json:"dates"`
	Tech    string `json:"tech"`
	Bullets string `json:"bullets"`
}

// Education is one school entry.
type Education struct {
	School   string `json:"school"`
	Dates    string `json:"dates"`
	Degree   string `json:"degree"`
	Location string `json:"location"`
	Notes    string `json:"notes"`
}

// Load reads and parses a CV JSON file, capturing skills key order.
func Load(path string) (*CV, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var c CV
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, fmt.Errorf("parse %s: %w", path, err)
	}
	c.skillsOrder = skillKeyOrder(raw)
	return &c, nil
}

// skillKeyOrder extracts the "skills" object key order from raw JSON, since
// encoding/json discards map ordering.
func skillKeyOrder(raw []byte) []string {
	var top map[string]json.RawMessage
	if json.Unmarshal(raw, &top) != nil {
		return nil
	}
	skills, ok := top["skills"]
	if !ok {
		return nil
	}
	dec := json.NewDecoder(newByteReader(skills))
	// expect opening '{'
	if t, err := dec.Token(); err != nil || t != json.Delim('{') {
		return nil
	}
	var order []string
	for dec.More() {
		key, err := dec.Token()
		if err != nil {
			break
		}
		order = append(order, key.(string))
		// skip the value
		var skip json.RawMessage
		if dec.Decode(&skip) != nil {
			break
		}
	}
	return order
}
