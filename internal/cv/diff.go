package cv

import (
	"fmt"
	"sort"
	"strings"
)

// Diff returns a human-readable list of differences from a (base) to b (other),
// grouped by section. Lines are prefixed with "+" (added in b), "-" (removed in
// b), or "~" (changed). Empty result means the two are equivalent.
func Diff(a, b *CV) []string {
	var out []string
	add := func(s string) { out = append(out, s) }

	if a.Summary != b.Summary {
		add("~ summary changed")
	}

	// Skills: compare by key.
	skillsA, skillsB := a.skillMap(), b.skillMap()
	for _, k := range sortedUnion(skillsA, skillsB) {
		va, oka := skillsA[k]
		vb, okb := skillsB[k]
		switch {
		case oka && !okb:
			add(fmt.Sprintf("- skill %q", k))
		case !oka && okb:
			add(fmt.Sprintf("+ skill %q", k))
		case va != vb:
			add(fmt.Sprintf("~ skill %q changed", k))
		}
	}

	diffBulletSet(&out, "experience", expByKey(a), expByKey(b))
	diffBulletSet(&out, "project", projByKey(a), projByKey(b))

	return out
}

// diffBulletSet diffs two maps of name->bullets, reporting added/removed
// entries and, for shared entries, added/removed bullet lines.
func diffBulletSet(out *[]string, kind string, a, b map[string]string) {
	for _, name := range sortedUnion(a, b) {
		ba, oka := a[name]
		bb, okb := b[name]
		switch {
		case oka && !okb:
			*out = append(*out, fmt.Sprintf("- %s %q", kind, name))
		case !oka && okb:
			*out = append(*out, fmt.Sprintf("+ %s %q", kind, name))
		default:
			setA := toSet(linesOf(ba))
			setB := toSet(linesOf(bb))
			for _, l := range linesOf(ba) {
				if !setB[l] {
					*out = append(*out, fmt.Sprintf("  - [%s %q] %s", kind, name, l))
				}
			}
			for _, l := range linesOf(bb) {
				if !setA[l] {
					*out = append(*out, fmt.Sprintf("  + [%s %q] %s", kind, name, l))
				}
			}
		}
	}
}

func (c *CV) skillMap() map[string]string {
	m := make(map[string]string, len(c.Skills))
	for k, v := range c.Skills {
		m[k] = strings.TrimSpace(v)
	}
	return m
}

func expByKey(c *CV) map[string]string {
	m := make(map[string]string, len(c.Experience))
	for _, e := range c.Experience {
		m[e.Company] = e.Bullets
	}
	return m
}

func projByKey(c *CV) map[string]string {
	m := make(map[string]string, len(c.Projects))
	for _, p := range c.Projects {
		m[p.Name] = p.Bullets
	}
	return m
}

func toSet(items []string) map[string]bool {
	s := make(map[string]bool, len(items))
	for _, it := range items {
		s[it] = true
	}
	return s
}

func sortedUnion(a, b map[string]string) []string {
	seen := map[string]bool{}
	var keys []string
	for k := range a {
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	for k := range b {
		if !seen[k] {
			seen[k] = true
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}
