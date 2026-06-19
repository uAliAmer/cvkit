package cv

import (
	"regexp"
	"sort"
	"strings"
)

// TailorReport ranks how well a CV matches a job description, all by keyword
// overlap — deterministic, no external services.
type TailorReport struct {
	Matched []TermCount // JD keywords present in the CV, by JD frequency
	Gaps    []TermCount // JD keywords absent from the CV, by JD frequency
	Surface []EntryHit  // CV entries ranked by how many JD keywords they hit
}

// TermCount is a keyword and how often it appears in the job description.
type TermCount struct {
	Term  string
	Count int
}

// EntryHit is a CV entry (project or job) and its JD-keyword hit count.
type EntryHit struct {
	Name string
	Hits int
}

var wordRe = regexp.MustCompile(`[a-zA-Z][a-zA-Z0-9+#.]*`)

// stopwords are common terms excluded from keyword matching.
var stopwords = map[string]bool{
	"the": true, "and": true, "for": true, "with": true, "you": true, "our": true,
	"are": true, "will": true, "have": true, "this": true, "that": true, "from": true,
	"your": true, "all": true, "but": true, "not": true, "can": true, "use": true,
	"who": true, "their": true, "they": true, "what": true, "such": true, "may": true,
	"per": true, "via": true, "into": true, "out": true, "any": true, "etc": true,
	"work": true, "team": true, "role": true, "job": true, "ability": true, "years": true,
	"experience": true, "including": true, "strong": true, "good": true, "well": true,
	"new": true, "using": true, "across": true, "within": true, "must": true, "should": true,
}

// Tailor compares a job description against the CV and returns a ranked report.
func (c *CV) Tailor(jd string, top int) TailorReport {
	jdFreq := tokenizeCount(jd)

	cvTerms := c.termSet()

	var matched, gaps []TermCount
	for term, n := range jdFreq {
		if cvTerms[term] {
			matched = append(matched, TermCount{term, n})
		} else {
			gaps = append(gaps, TermCount{term, n})
		}
	}
	sortTermCounts(matched)
	sortTermCounts(gaps)

	// Rank CV entries by distinct JD keywords they mention.
	jdSet := map[string]bool{}
	for t := range jdFreq {
		jdSet[t] = true
	}
	var surface []EntryHit
	for _, e := range c.Experience {
		surface = append(surface, EntryHit{"experience: " + e.Company, hits(jdSet, e.Bullets+" "+e.Title)})
	}
	for _, p := range c.Projects {
		surface = append(surface, EntryHit{"project: " + p.Name, hits(jdSet, p.Bullets+" "+p.Tech+" "+p.Name)})
	}
	sort.SliceStable(surface, func(i, j int) bool { return surface[i].Hits > surface[j].Hits })

	return TailorReport{
		Matched: clip(matched, top),
		Gaps:    clip(gaps, top),
		Surface: surface,
	}
}

// termSet collects every distinct keyword present anywhere in the CV.
func (c *CV) termSet() map[string]bool {
	set := map[string]bool{}
	addAll := func(s string) {
		for _, w := range tokenize(s) {
			set[w] = true
		}
	}
	addAll(c.Summary)
	for _, e := range c.Experience {
		addAll(e.Bullets)
		addAll(e.Title)
	}
	for _, p := range c.Projects {
		addAll(p.Bullets)
		addAll(p.Tech)
		addAll(p.Name)
	}
	for _, v := range c.Skills {
		addAll(v)
	}
	return set
}

func hits(jdSet map[string]bool, text string) int {
	seen := map[string]bool{}
	n := 0
	for _, w := range tokenize(text) {
		if jdSet[w] && !seen[w] {
			seen[w] = true
			n++
		}
	}
	return n
}

func tokenize(s string) []string {
	var out []string
	for _, m := range wordRe.FindAllString(strings.ToLower(s), -1) {
		m = strings.TrimRight(m, ".") // drop sentence punctuation, keep c++/c#
		if len(m) < 3 || stopwords[m] {
			continue
		}
		out = append(out, m)
	}
	return out
}

func tokenizeCount(s string) map[string]int {
	m := map[string]int{}
	for _, w := range tokenize(s) {
		m[w]++
	}
	return m
}

func sortTermCounts(t []TermCount) {
	sort.SliceStable(t, func(i, j int) bool {
		if t[i].Count != t[j].Count {
			return t[i].Count > t[j].Count
		}
		return t[i].Term < t[j].Term
	})
}

func clip(t []TermCount, n int) []TermCount {
	if n > 0 && len(t) > n {
		return t[:n]
	}
	return t
}
