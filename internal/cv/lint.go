package cv

import (
	"fmt"
	"regexp"
	"strings"
)

// LintFinding is one content-quality issue with where it was found.
type LintFinding struct {
	Where   string
	Message string
}

func (f LintFinding) String() string { return f.Where + ": " + f.Message }

// weakStarts are vague opening words that dilute a bullet's impact.
var weakStarts = map[string]bool{
	"helped": true, "worked": true, "assisted": true, "responsible": true,
	"various": true, "handled": true, "involved": true, "participated": true,
	"tasked": true, "duties": true, "supported": true,
}

var (
	digitRe   = regexp.MustCompile(`\d`)
	passiveRe = regexp.MustCompile(`\b(was|were|been|being|is|are)\b\s+\w+(ed|en)\b`)
)

// Lint runs content-quality heuristics over experience and project bullets and
// the summary. Findings are advisory, not correctness errors.
func (c *CV) Lint() []LintFinding {
	var f []LintFinding

	if s := strings.TrimSpace(c.Summary); s != "" {
		if n := len(strings.Fields(s)); n > 60 {
			f = append(f, LintFinding{"summary", fmt.Sprintf("%d words; consider trimming to under 60", n)})
		}
	}

	check := func(where, bullets string) {
		for i, b := range linesOf(bullets) {
			loc := fmt.Sprintf("%s bullet %d", where, i+1)
			first := strings.ToLower(strings.Trim(strings.Fields(b)[0], ".,:;"))
			if weakStarts[first] {
				f = append(f, LintFinding{loc, fmt.Sprintf("weak opener %q; lead with a strong action verb", first)})
			}
			if !digitRe.MatchString(b) {
				f = append(f, LintFinding{loc, "no number; quantify the impact if you can"})
			}
			if passiveRe.MatchString(strings.ToLower(b)) {
				f = append(f, LintFinding{loc, "reads as passive voice; prefer active"})
			}
			if len(b) > 220 {
				f = append(f, LintFinding{loc, fmt.Sprintf("%d chars; consider splitting", len(b))})
			}
		}
	}

	for _, e := range c.Experience {
		check("experience ("+e.Company+")", e.Bullets)
	}
	for _, p := range c.Projects {
		check("project ("+p.Name+")", p.Bullets)
	}
	return f
}
