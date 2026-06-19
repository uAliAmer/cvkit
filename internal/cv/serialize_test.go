package cv

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestMarshalPreservesSkillOrder(t *testing.T) {
	c, err := Load("testdata/sample.json")
	if err != nil {
		t.Fatal(err)
	}
	// sample.json order: technical, tools, "Custom Group"
	order := c.SkillsOrder()
	want := []string{"technical", "tools", "Custom Group"}
	if len(order) != len(want) {
		t.Fatalf("SkillsOrder = %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("SkillsOrder[%d] = %q, want %q", i, order[i], want[i])
		}
	}

	out, err := c.Marshal(nil)
	if err != nil {
		t.Fatal(err)
	}
	// Skill keys must appear in source order, not alphabetical.
	iTech := strings.Index(string(out), `"technical"`)
	iTools := strings.Index(string(out), `"tools"`)
	iCustom := strings.Index(string(out), `"Custom Group"`)
	if !(iTech < iTools && iTools < iCustom) {
		t.Errorf("skills not in source order: technical=%d tools=%d custom=%d", iTech, iTools, iCustom)
	}

	// Re-parsing the output must yield the same skill order (round-trip).
	c2, err := loadBytes(out)
	if err != nil {
		t.Fatal(err)
	}
	for i, k := range c2.SkillsOrder() {
		if k != want[i] {
			t.Errorf("round-trip order[%d] = %q, want %q", i, k, want[i])
		}
	}
}

// loadBytes parses CV JSON from memory (test helper mirroring Load).
func loadBytes(raw []byte) (*CV, error) {
	var c CV
	if err := json.Unmarshal(raw, &c); err != nil {
		return nil, err
	}
	c.skillsOrder = skillKeyOrder(raw)
	return &c, nil
}
