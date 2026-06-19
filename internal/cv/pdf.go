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
	t := func(s string) string { return tr(sanitizePDF(s)) }

	// Name
	dark()
	pdf.SetFont("Helvetica", "B", 22)
	name := c.Name
	if name == "" {
		name = "Your Name"
	}
	pdf.CellFormat(contentW, 10, t(name), "", 1, "C", false, 0, "")

	// Contact line
	if cl := c.contactLine(false); cl != "" {
		gray()
		pdf.SetFont("Helvetica", "", 9)
		pdf.CellFormat(contentW, 5, t(cl), "", 1, "C", false, 0, "")
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
			meta := p.Tech
			if u := projectURL(p.Link); u != "" {
				meta = joinMeta(meta, u)
			}
			subheading(meta)
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
