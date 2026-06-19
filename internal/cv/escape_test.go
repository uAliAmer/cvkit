package cv

import "testing"

func TestEsc(t *testing.T) {
	cases := map[string]string{
		"":      "",
		"a & b": `a \& b`,
		"100%":  `100\%`,
		"a_b":   `a\_b`,
		"x → y": `x $\rightarrow$ y`,
		// NOTE: matches build_cv.py exactly — the {} from \textbackslash{} get
		// re-escaped by the later brace rules. Faithful port of a latent quirk.
		`\path`:        `\textbackslash\{\}path`,
		"{braces}":     `\{braces\}`,
		"C# $5 #1 ~ ^": `C\# \$5 \#1 \textasciitilde{} \textasciicircum{}`,
	}
	for in, want := range cases {
		if got := esc(in); got != want {
			t.Errorf("esc(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestLinesOf(t *testing.T) {
	got := linesOf("  one  \n\n  two\n   \nthree")
	want := []string{"one", "two", "three"}
	if len(got) != len(want) {
		t.Fatalf("linesOf len = %d, want %d (%q)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("line %d = %q, want %q", i, got[i], want[i])
		}
	}
}

func TestExpandHandle(t *testing.T) {
	url, label := expandHandle("uAliAmer", "https://github.com/")
	if url != "https://github.com/uAliAmer" || label != "uAliAmer" {
		t.Errorf("bare handle: got (%q,%q)", url, label)
	}
	url, label = expandHandle("https://github.com/uAliAmer/", "https://github.com/")
	if url != "https://github.com/uAliAmer/" || label != "uAliAmer" {
		t.Errorf("full url: got (%q,%q)", url, label)
	}
}
