package cv

import "testing"

func TestLint(t *testing.T) {
	c := &CV{
		Experience: []Experience{{
			Company: "X",
			// weak opener + no number; second bullet is clean
			Bullets: "Helped with various tasks\nShipped 12 features that cut latency 30%",
		}},
	}
	f := c.Lint()
	var weak, nonum bool
	for _, fi := range f {
		if fi.Where == "experience (X) bullet 1" {
			if contains(fi.Message, "weak opener") {
				weak = true
			}
			if contains(fi.Message, "no number") {
				nonum = true
			}
		}
		if fi.Where == "experience (X) bullet 2" {
			t.Errorf("clean bullet 2 should have no findings, got %q", fi.Message)
		}
	}
	if !weak || !nonum {
		t.Errorf("expected weak-opener and no-number findings on bullet 1; weak=%v nonum=%v", weak, nonum)
	}
}

func TestDiff(t *testing.T) {
	a := &CV{
		Skills:   map[string]string{"technical": "Go", "tools": "Git"},
		Projects: []Project{{Name: "P", Bullets: "old line"}},
	}
	b := &CV{
		Skills:   map[string]string{"technical": "Go", "extra": "Rust"},
		Projects: []Project{{Name: "P", Bullets: "new line"}},
	}
	d := Diff(a, b)
	var addExtra, dropTools, addLine, dropLine bool
	for _, l := range d {
		switch {
		case l == `+ skill "extra"`:
			addExtra = true
		case l == `- skill "tools"`:
			dropTools = true
		case contains(l, "+ [project") && contains(l, "new line"):
			addLine = true
		case contains(l, "- [project") && contains(l, "old line"):
			dropLine = true
		}
	}
	if !addExtra || !dropTools || !addLine || !dropLine {
		t.Errorf("diff missing expected changes: %v", d)
	}
}

func TestTailor(t *testing.T) {
	c := &CV{
		Skills:   map[string]string{"technical": "Go, FastAPI, PostgreSQL"},
		Projects: []Project{{Name: "API", Bullets: "Built a FastAPI service", Tech: "Go"}},
	}
	r := c.Tailor("We want FastAPI and Kubernetes experience.", 10)
	var matchedFastAPI, gapKubernetes bool
	for _, m := range r.Matched {
		if m.Term == "fastapi" {
			matchedFastAPI = true
		}
	}
	for _, g := range r.Gaps {
		if g.Term == "kubernetes" {
			gapKubernetes = true
		}
	}
	if !matchedFastAPI {
		t.Error("expected fastapi in matched")
	}
	if !gapKubernetes {
		t.Error("expected kubernetes in gaps")
	}
	if len(r.Surface) == 0 || r.Surface[0].Hits == 0 {
		t.Error("expected the API project to register keyword hits")
	}
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
