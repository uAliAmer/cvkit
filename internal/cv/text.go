package cv

import (
	"strings"
)

// contactLine joins the present contact fields with " | ", expanding handles to
// full URLs for the Markdown variant when linkify is true.
func (c *CV) contactLine(linkify bool) string {
	var parts []string
	if c.Email != "" {
		parts = append(parts, c.Email)
	}
	if c.Phone != "" {
		parts = append(parts, c.Phone)
	}
	if c.Location != "" {
		parts = append(parts, c.Location)
	}
	if c.LinkedIn != "" {
		url, label := expandHandle(c.LinkedIn, "https://www.linkedin.com/in/")
		if linkify {
			parts = append(parts, "[in/"+label+"]("+url+")")
		} else {
			parts = append(parts, url)
		}
	}
	if c.GitHub != "" {
		url, label := expandHandle(c.GitHub, "https://github.com/")
		if linkify {
			parts = append(parts, "[@"+label+"]("+url+")")
		} else {
			parts = append(parts, url)
		}
	}
	return strings.Join(parts, " | ")
}

// projectURL returns the full URL form of a project link, or "" if none.
func projectURL(link string) string {
	link = strings.TrimSpace(link)
	if link == "" {
		return ""
	}
	if !strings.Contains(link, "://") {
		return "https://" + link
	}
	return link
}

// RenderMarkdown renders the CV as Markdown.
func (c *CV) RenderMarkdown() string {
	var b strings.Builder
	name := c.Name
	if name == "" {
		name = "Your Name"
	}
	b.WriteString("# " + name + "\n\n")
	if cl := c.contactLine(true); cl != "" {
		b.WriteString(cl + "\n")
	}
	if c.Summary != "" {
		b.WriteString("\n## Summary\n\n" + c.Summary + "\n")
	}
	if len(c.Experience) > 0 {
		b.WriteString("\n## Experience\n")
		for _, e := range c.Experience {
			b.WriteString("\n### " + e.Company + " — " + e.Title + "\n")
			b.WriteString(joinMeta(e.Dates, e.Location) + "\n")
			for _, bl := range linesOf(e.Bullets) {
				b.WriteString("- " + bl + "\n")
			}
		}
	}
	if len(c.Projects) > 0 {
		b.WriteString("\n## Projects\n")
		for _, p := range c.Projects {
			title := p.Name
			if url := projectURL(p.Link); url != "" {
				title = "[" + p.Name + "](" + url + ")"
			}
			b.WriteString("\n### " + title + "\n")
			if meta := joinMeta(p.Tech, p.Dates); meta != "" {
				b.WriteString(meta + "\n")
			}
			for _, bl := range linesOf(p.Bullets) {
				b.WriteString("- " + bl + "\n")
			}
		}
	}
	if len(c.Education) > 0 {
		b.WriteString("\n## Education\n")
		for _, e := range c.Education {
			b.WriteString("\n### " + e.School + " — " + e.Degree + "\n")
			b.WriteString(joinMeta(e.Dates, e.Location) + "\n")
			if e.Notes != "" {
				b.WriteString("- " + e.Notes + "\n")
			}
		}
	}
	if rows := c.skillRows(); len(rows) > 0 {
		b.WriteString("\n## Skills\n\n")
		for _, r := range rows {
			b.WriteString("- **" + r[0] + ":** " + r[1] + "\n")
		}
	}
	return b.String()
}

// RenderText renders the CV as plain text.
func (c *CV) RenderText() string {
	var b strings.Builder
	name := c.Name
	if name == "" {
		name = "Your Name"
	}
	b.WriteString(name + "\n")
	if cl := c.contactLine(false); cl != "" {
		b.WriteString(cl + "\n")
	}
	section := func(title string) {
		b.WriteString("\n" + strings.ToUpper(title) + "\n" + strings.Repeat("=", len(title)) + "\n")
	}
	if c.Summary != "" {
		section("Summary")
		b.WriteString(c.Summary + "\n")
	}
	if len(c.Experience) > 0 {
		section("Experience")
		for _, e := range c.Experience {
			b.WriteString("\n" + e.Company + " — " + e.Title + "\n")
			b.WriteString(joinMeta(e.Dates, e.Location) + "\n")
			for _, bl := range linesOf(e.Bullets) {
				b.WriteString("  * " + bl + "\n")
			}
		}
	}
	if len(c.Projects) > 0 {
		section("Projects")
		for _, p := range c.Projects {
			b.WriteString("\n" + p.Name + "\n")
			if url := projectURL(p.Link); url != "" {
				b.WriteString(url + "\n")
			}
			if meta := joinMeta(p.Tech, p.Dates); meta != "" {
				b.WriteString(meta + "\n")
			}
			for _, bl := range linesOf(p.Bullets) {
				b.WriteString("  * " + bl + "\n")
			}
		}
	}
	if len(c.Education) > 0 {
		section("Education")
		for _, e := range c.Education {
			b.WriteString("\n" + e.School + " — " + e.Degree + "\n")
			b.WriteString(joinMeta(e.Dates, e.Location) + "\n")
			if e.Notes != "" {
				b.WriteString("  * " + e.Notes + "\n")
			}
		}
	}
	if rows := c.skillRows(); len(rows) > 0 {
		section("Skills")
		for _, r := range rows {
			b.WriteString(r[0] + ": " + r[1] + "\n")
		}
	}
	return b.String()
}

// joinMeta joins two optional metadata strings with a comma, skipping empties.
func joinMeta(a, b string) string {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	switch {
	case a != "" && b != "":
		return a + ", " + b
	case a != "":
		return a
	default:
		return b
	}
}

// skillRows returns ordered (label, value) pairs, applying the same known-key
// labeling and source ordering as the LaTeX renderer.
func (c *CV) skillRows() [][2]string {
	if len(c.Skills) == 0 {
		return nil
	}
	order := c.skillsOrder
	if len(order) == 0 {
		for k := range c.Skills {
			order = append(order, k)
		}
	}
	var rows [][2]string
	for _, key := range order {
		val := strings.TrimSpace(c.Skills[key])
		if val == "" {
			continue
		}
		label, ok := skillLabels[key]
		if !ok {
			label = key
		}
		rows = append(rows, [2]string{label, val})
	}
	return rows
}
