package cv

import (
	"bytes"
	"strings"

	"github.com/go-pdf/fpdf"
)

// RenderPDF renders the CV directly to a PDF, with no external dependencies
// (no LaTeX). It produces a clean, ATS-friendly single-column layout. For the
// fully typeset version, use RenderLaTeX + XeLaTeX instead.
func (c *CV) RenderPDF() ([]byte, error) {
	const (
		margin   = 15.0
		pageW    = 215.9 // US Letter
		contentW = pageW - 2*margin
		dateW    = 45.0
	)

	pdf := fpdf.New("P", "mm", "Letter", "")
	pdf.SetMargins(margin, margin, margin)
	pdf.SetAutoPageBreak(true, margin)
	pdf.AddPage()
	tr := pdf.UnicodeTranslatorFromDescriptor("") // UTF-8 -> cp1252

	dark := func() { pdf.SetTextColor(20, 20, 20) }
	gray := func() { pdf.SetTextColor(110, 110, 110) }
	blue := func() { pdf.SetTextColor(0, 102, 204) }
	t := func(s string) string { return tr(sanitizePDF(s)) }

	// Name
	dark()
	pdf.SetFont("Helvetica", "B", 22)
	name := c.Name
	if name == "" {
		name = "Your Name"
	}
	pdf.CellFormat(contentW, 10, t(name), "", 1, "C", false, 0, "")

	// Contact line: centered, with clickable shortened links.
	type cseg struct{ text, link string }
	var segs []cseg
	if c.Email != "" {
		segs = append(segs, cseg{c.Email, "mailto:" + c.Email})
	}
	if c.Phone != "" {
		segs = append(segs, cseg{c.Phone, ""})
	}
	if c.Location != "" {
		segs = append(segs, cseg{c.Location, ""})
	}
	if c.LinkedIn != "" {
		u, _ := expandHandle(c.LinkedIn, "https://www.linkedin.com/in/")
		segs = append(segs, cseg{shortenURL(u), u})
	}
	if c.GitHub != "" {
		u, _ := expandHandle(c.GitHub, "https://github.com/")
		segs = append(segs, cseg{shortenURL(u), u})
	}
	if len(segs) > 0 {
		pdf.SetFont("Helvetica", "", 9)
		const sep = "  |  "
		total := 0.0
		for i, s := range segs {
			if i > 0 {
				total += pdf.GetStringWidth(t(sep))
			}
			total += pdf.GetStringWidth(t(s.text))
		}
		pdf.SetX(margin + (contentW-total)/2)
		for i, s := range segs {
			if i > 0 {
				gray()
				pdf.CellFormat(pdf.GetStringWidth(t(sep)), 5, t(sep), "", 0, "L", false, 0, "")
			}
			if s.link != "" {
				blue()
			} else {
				gray()
			}
			pdf.CellFormat(pdf.GetStringWidth(t(s.text)), 5, t(s.text), "", 0, "L", false, 0, s.link)
		}
		pdf.Ln(5)
	}
	pdf.Ln(3)

	section := func(title string) {
		pdf.Ln(2)
		dark()
		pdf.SetFont("Helvetica", "B", 12)
		pdf.CellFormat(contentW, 6, t(strings.ToUpper(title)), "", 1, "L", false, 0, "")
		y := pdf.GetY()
		pdf.SetDrawColor(200, 200, 200)
		pdf.Line(margin, y, margin+contentW, y)
		pdf.Ln(1.5)
	}

	heading := func(left, right string) {
		dark()
		pdf.SetFont("Helvetica", "B", 11)
		pdf.CellFormat(contentW-dateW, 5.5, t(left), "", 0, "L", false, 0, "")
		gray()
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(dateW, 5.5, t(right), "", 1, "R", false, 0, "")
	}

	subheading := func(s string) {
		if s == "" {
			return
		}
		dark()
		pdf.SetFont("Helvetica", "I", 9.5)
		pdf.CellFormat(contentW, 5, t(s), "", 1, "L", false, 0, "")
	}

	bulletList := func(text string) {
		dark()
		pdf.SetFont("Helvetica", "", 10)
		for _, b := range linesOf(text) {
			x := pdf.GetX()
			pdf.CellFormat(4, 5, t("-"), "", 0, "L", false, 0, "")
			pdf.MultiCell(contentW-4, 5, t(b), "", "L", false)
			pdf.SetX(x)
		}
	}

	if s := strings.TrimSpace(c.Summary); s != "" {
		section("Summary")
		dark()
		pdf.SetFont("Helvetica", "", 10)
		pdf.MultiCell(contentW, 5, t(s), "", "L", false)
	}

	if len(c.Experience) > 0 {
		section("Experience")
		for _, e := range c.Experience {
			heading(e.Company, e.Dates)
			subheading(joinMeta(e.Title, e.Location))
			bulletList(e.Bullets)
			pdf.Ln(1.5)
		}
	}

	if len(c.Projects) > 0 {
		section("Projects")
		for _, p := range c.Projects {
			heading(p.Name, p.Dates)
			// meta line: tech (italic) then a clickable, shortened link.
			pdf.SetX(margin)
			wrote := false
			if p.Tech != "" {
				gray()
				pdf.SetFont("Helvetica", "I", 9.5)
				pdf.CellFormat(pdf.GetStringWidth(t(p.Tech)), 5, t(p.Tech), "", 0, "L", false, 0, "")
				wrote = true
			}
			if u := projectURL(p.Link); u != "" {
				if wrote {
					gray()
					pdf.SetFont("Helvetica", "I", 9.5)
					pdf.CellFormat(pdf.GetStringWidth(t("  -  ")), 5, t("  -  "), "", 0, "L", false, 0, "")
				}
				disp := shortenURL(u)
				blue()
				pdf.SetFont("Helvetica", "", 9.5)
				pdf.CellFormat(pdf.GetStringWidth(t(disp)), 5, t(disp), "", 0, "L", false, 0, u)
			}
			if wrote || projectURL(p.Link) != "" {
				pdf.Ln(5)
			}
			bulletList(p.Bullets)
			pdf.Ln(1.5)
		}
	}

	if len(c.Education) > 0 {
		section("Education")
		for _, e := range c.Education {
			heading(e.School, e.Dates)
			subheading(joinMeta(e.Degree, e.Location))
			bulletList(e.Notes)
			pdf.Ln(1.5)
		}
	}

	if rows := c.skillRows(); len(rows) > 0 {
		section("Skills")
		for _, r := range rows {
			dark()
			pdf.SetFont("Helvetica", "B", 10)
			lw := pdf.GetStringWidth(t(r[0]+": ")) + 1
			pdf.CellFormat(lw, 5, t(r[0]+":"), "", 0, "L", false, 0, "")
			pdf.SetFont("Helvetica", "", 10)
			pdf.MultiCell(contentW-lw, 5, t(r[1]), "", "L", false)
		}
	}

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// shortenURL strips the scheme, "www.", and trailing slash for display, while
// the full URL is still used as the clickable target.
func shortenURL(u string) string {
	if i := strings.Index(u, "://"); i >= 0 {
		u = u[i+3:]
	}
	u = strings.TrimPrefix(u, "www.")
	return strings.TrimRight(u, "/")
}

// sanitizePDF replaces characters the cp1252 core fonts can't render with
// ASCII-ish equivalents so the native PDF never shows "?" boxes.
func sanitizePDF(s string) string {
	r := strings.NewReplacer(
		"→", "->", "←", "<-", "↔", "<->",
		"—", "-", "–", "-",
		"“", "\"", "”", "\"", "‘", "'", "’", "'",
		"•", "-", "…", "...",
	)
	return r.Replace(s)
}
