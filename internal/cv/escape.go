package cv

import (
	"bytes"
	"strings"
)

// escPair is one (find, replace) substitution for LaTeX escaping.
type escPair struct{ from, to string }

// escTable mirrors build_cv.py's ESC list. Backslash must come first so that
// backslashes inserted by later replacements are not themselves re-escaped.
var escTable = []escPair{
	{"\\", `\textbackslash{}`},
	{"&", `\&`},
	{"%", `\%`},
	{"$", `\$`},
	{"#", `\#`},
	{"_", `\_`},
	{"{", `\{`},
	{"}", `\}`},
	{"~", `\textasciitilde{}`},
	{"^", `\textasciicircum{}`},
	// unicode that may lack a glyph in the CV font -> safe LaTeX
	{"→", `$\rightarrow$`},
	{"←", `$\leftarrow$`},
}

// esc escapes LaTeX special characters in plain user text. Empty in, empty out.
func esc(s string) string {
	if s == "" {
		return ""
	}
	for _, p := range escTable {
		s = strings.ReplaceAll(s, p.from, p.to)
	}
	return s
}

// linesOf splits a newline-joined bullets string into trimmed, non-empty lines.
func linesOf(text string) []string {
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if t := strings.TrimSpace(l); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// items renders a slice of bullet strings as a resumeItem list, or "" if empty.
func items(bullets []string) string {
	if len(bullets) == 0 {
		return ""
	}
	var b bytes.Buffer
	b.WriteString("      \\resumeItemListStart\n")
	for _, item := range bullets {
		b.WriteString("        \\resumeItem{" + esc(item) + "}\n")
	}
	b.WriteString("      \\resumeItemListEnd\n")
	return b.String()
}

// newByteReader is a tiny helper so types.go can stream a json.RawMessage.
func newByteReader(b []byte) *bytes.Reader { return bytes.NewReader(b) }
