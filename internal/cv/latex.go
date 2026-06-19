package cv

import (
	_ "embed"
	"strings"
)

//go:embed templates/preamble.tex
var preamble string

// RenderLaTeX renders the full .tex document for a CV.
func (c *CV) RenderLaTeX() string {
	var b strings.Builder
	b.WriteString(preamble)
	b.WriteString("\n")
	b.WriteString(c.heading())
	b.WriteString(sectionSummary(c.Summary))
	b.WriteString(c.sectionExperience())
	b.WriteString(c.sectionProjects())
	b.WriteString(c.sectionEducation())
	b.WriteString(c.sectionSkills())
	b.WriteString("\\end{document}\n")
	return b.String()
}

// expandHandle turns a bare handle into a full profile URL + display label.
// If the value already contains a scheme, it is used as-is.
func expandHandle(v, base string) (url, label string) {
	v = strings.TrimSpace(v)
	if strings.Contains(v, "://") {
		label = strings.TrimRight(v, "/")
		if i := strings.LastIndex(label, "/"); i >= 0 {
			label = label[i+1:]
		}
		return v, label
	}
	return base + strings.Trim(v, "/"), v
}

func (c *CV) heading() string {
	var parts []string
	if c.Phone != "" {
		parts = append(parts, `\faPhone* \texttt{`+esc(c.Phone)+`}`)
	}
	if c.Email != "" {
		parts = append(parts, `\faEnvelope \hspace{2pt} \texttt{`+esc(c.Email)+`}`)
	}
	if c.LinkedIn != "" {
		url, label := expandHandle(c.LinkedIn, "https://www.linkedin.com/in/")
		parts = append(parts, `\faLinkedin \hspace{2pt} \href{`+url+`}{\texttt{`+esc(label)+`}}`)
	}
	if c.GitHub != "" {
		url, label := expandHandle(c.GitHub, "https://github.com/")
		parts = append(parts, `\faGithub \hspace{2pt} \href{`+url+`}{\texttt{`+esc(label)+`}}`)
	}
	if c.Location != "" {
		parts = append(parts, `\faMapMarker* \hspace{2pt}\texttt{`+esc(c.Location)+`}`)
	}
	name := c.Name
	if name == "" {
		name = "Your Name"
	}
	sep := ` \hspace{1pt} $|$ \hspace{1pt} `
	return "\\begin{center}\n" +
		"    \\textbf{\\Huge " + esc(name) + "} \\\\ \\vspace{5pt}\n" +
		"    \\small " + strings.Join(parts, sep) + "\n" +
		"    \\\\ \\vspace{-3pt}\n\\end{center}\n\n"
}

func sectionSummary(s string) string {
	if s == "" {
		return ""
	}
	return "\\section{SUMMARY}\n\\small{" + esc(s) + "}\n\\vspace{4pt}\n\n"
}

func (c *CV) sectionExperience() string {
	if len(c.Experience) == 0 {
		return ""
	}
	var out []string
	out = append(out, "\\section{EXPERIENCE}", "  \\resumeSubHeadingListStart")
	for _, e := range c.Experience {
		out = append(out, "    \\resumeSubheading")
		out = append(out, "      {"+esc(e.Company)+"}{"+esc(e.Dates)+"}")
		out = append(out, "      {"+esc(e.Title)+"}{"+esc(e.Location)+"}")
		if b := items(linesOf(e.Bullets)); b != "" {
			out = append(out, strings.TrimRight(b, "\n"))
		}
	}
	out = append(out, "  \\resumeSubHeadingListEnd\n")
	return strings.Join(out, "\n") + "\n"
}

func (c *CV) sectionProjects() string {
	if len(c.Projects) == 0 {
		return ""
	}
	var out []string
	out = append(out, "\\section{PROJECTS}", "    \\resumeSubHeadingListStart")
	for _, p := range c.Projects {
		nameCell := "\\textbf{" + esc(p.Name) + "}"
		if link := strings.TrimSpace(p.Link); link != "" {
			url := link
			if !strings.Contains(link, "://") {
				url = "https://" + link
			}
			shown := url
			if i := strings.Index(shown, "://"); i >= 0 {
				shown = shown[i+3:]
			}
			shown = strings.TrimRight(strings.ReplaceAll(shown, "www.", ""), "/")
			nameCell += " \\hspace{4pt}{\\small\\ttfamily \\href{" + url +
				"}{\\textcolor{link-blue}{" + esc(shown) + "}}}"
		}
		out = append(out, "      \\item")
		out = append(out, "        \\begin{tabular*}{\\textwidth}[t]{l@{\\extracolsep{\\fill}}r}")
		out = append(out, "          "+nameCell+" & {\\color{dark-grey}\\small "+esc(p.Dates)+"}\\\\")
		if p.Tech != "" {
			out = append(out, "          {\\small\\emph{"+esc(p.Tech)+"}} & \\\\")
		}
		out = append(out, "        \\end{tabular*}\\vspace{-4pt}")
		if b := items(linesOf(p.Bullets)); b != "" {
			out = append(out, strings.TrimRight(b, "\n"))
		}
	}
	out = append(out, "    \\resumeSubHeadingListEnd\n")
	return strings.Join(out, "\n") + "\n"
}

func (c *CV) sectionEducation() string {
	if len(c.Education) == 0 {
		return ""
	}
	var out []string
	out = append(out, "\\section{EDUCATION}", "  \\resumeSubHeadingListStart")
	for _, e := range c.Education {
		out = append(out, "    \\resumeSubheading")
		out = append(out, "      {"+esc(e.School)+"}{"+esc(e.Dates)+"}")
		out = append(out, "      {"+esc(e.Degree)+"}{"+esc(e.Location)+"}")
		if e.Notes != "" {
			out = append(out, "        \\resumeItemListStart")
			out = append(out, "          \\resumeItem{"+esc(e.Notes)+"}")
			out = append(out, "        \\resumeItemListEnd")
		}
	}
	out = append(out, "  \\resumeSubHeadingListEnd\n")
	return strings.Join(out, "\n") + "\n"
}

// skillLabels maps known skill keys to display labels. Unknown keys are used
// verbatim as their own label.
var skillLabels = map[string]string{
	"technical": "Technical",
	"business":  "Business",
	"tools":     "Tools",
	"languages": "Languages",
}

func (c *CV) sectionSkills() string {
	skillRows := c.skillRows()
	if len(skillRows) == 0 {
		return ""
	}
	var rows []string
	for _, r := range skillRows {
		rows = append(rows, "     \\textbf{"+esc(r[0])+"} {: "+esc(r[1])+"}")
	}
	body := strings.Join(rows, " \\vspace{2pt} \\\\\n")
	return "\\section{SKILLS}\n" +
		" \\begin{itemize}[leftmargin=0in, label={}]\n" +
		"    \\small{\\item{\n" +
		body + "\n" +
		"    }}\n" +
		" \\end{itemize}\n\n"
}
