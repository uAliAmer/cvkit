// Command cvkit-gui is a cross-platform desktop editor for cvkit CV JSON files.
// It edits the same cv_data.json the CLI uses and shares the internal/cv
// package for loading, rendering, validation, and serialization.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/uAliAmer/cvkit/internal/cv"
)

type editor struct {
	app    fyne.App
	win    fyne.Window
	doc    *cv.CV
	keys   []string // skill keys, ordered
	vals   []string // skill values, parallel to keys
	path   string   // current file path ("" = unsaved)
	status *widget.Label
	body   *fyne.Container // scrollable form body, rebuilt on structural change
}

func main() {
	a := app.New()
	w := a.NewWindow("CVKit")
	e := &editor{app: a, win: w, doc: emptyCV(), status: widget.NewLabel("New CV")}
	e.body = container.NewVBox()
	e.rebuild()

	w.SetContent(container.NewBorder(e.toolbar(), e.status, nil, nil, container.NewScroll(e.body)))
	w.Resize(fyne.NewSize(820, 720))
	w.ShowAndRun()
}

func emptyCV() *cv.CV {
	return &cv.CV{Skills: map[string]string{}}
}

// ---- toolbar & actions -----------------------------------------------------

func (e *editor) toolbar() fyne.CanvasObject {
	btn := widget.NewButton
	return container.NewHBox(
		btn("New", e.actNew),
		btn("Open", e.actOpen),
		btn("Save", e.actSave),
		btn("Save As", e.actSaveAs),
		widget.NewSeparator(),
		btn("Validate", e.actValidate),
		btn("Lint", e.actLint),
		btn("Build .tex", e.actBuild),
		btn("PDF", e.actPDF),
	)
}

func (e *editor) actNew() {
	e.doc = emptyCV()
	e.keys, e.vals, e.path = nil, nil, ""
	e.rebuild()
	e.setStatus("New CV")
}

func (e *editor) actOpen() {
	dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
		if err != nil || r == nil {
			return
		}
		defer r.Close()
		path := r.URI().Path()
		doc, err := cv.Load(path)
		if err != nil {
			e.fail(err)
			return
		}
		e.doc = doc
		e.loadSkills()
		e.path = path
		e.rebuild()
		e.setStatus("Opened " + path)
	}, e.win)
}

func (e *editor) actSave() {
	if e.path == "" {
		e.actSaveAs()
		return
	}
	e.writeTo(e.path)
}

func (e *editor) actSaveAs() {
	dialog.ShowFileSave(func(w fyne.URIWriteCloser, err error) {
		if err != nil || w == nil {
			return
		}
		path := w.URI().Path()
		w.Close()
		e.writeTo(path)
		e.path = path
	}, e.win)
}

func (e *editor) writeTo(path string) {
	e.commitSkills()
	out, err := e.doc.Marshal(e.keys)
	if err != nil {
		e.fail(err)
		return
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		e.fail(err)
		return
	}
	e.setStatus("Saved " + path)
}

func (e *editor) actValidate() {
	e.commitSkills()
	if p := e.doc.Validate(); len(p) > 0 {
		e.showList("Validation problems", p)
	} else {
		dialog.ShowInformation("Validate", "No problems found ✓", e.win)
	}
}

func (e *editor) actLint() {
	e.commitSkills()
	f := e.doc.Lint()
	if len(f) == 0 {
		dialog.ShowInformation("Lint", "No suggestions ✓", e.win)
		return
	}
	msgs := make([]string, len(f))
	for i, fi := range f {
		msgs[i] = fi.String()
	}
	e.showList("Lint suggestions", msgs)
}

func (e *editor) actBuild() {
	e.commitSkills()
	out := deriveTex(e.path)
	if out == "" {
		e.setStatus("Save the CV first, then Build")
		return
	}
	if err := os.WriteFile(out, []byte(e.doc.RenderLaTeX()), 0o644); err != nil {
		e.fail(err)
		return
	}
	e.setStatus("Wrote " + out)
}

func (e *editor) actPDF() {
	bin, err := exec.LookPath("xelatex")
	if err != nil {
		dialog.ShowError(fmt.Errorf("xelatex not found in PATH; install a TeX distribution to export PDF"), e.win)
		return
	}
	e.commitSkills()
	tex := deriveTex(e.path)
	if tex == "" {
		e.setStatus("Save the CV first, then PDF")
		return
	}
	if err := os.WriteFile(tex, []byte(e.doc.RenderLaTeX()), 0o644); err != nil {
		e.fail(err)
		return
	}
	dir := filepath.Dir(tex)
	cmd := exec.Command(bin, "-interaction=nonstopmode", "-halt-on-error", "-output-directory", dir, filepath.Base(tex))
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		dialog.ShowError(fmt.Errorf("xelatex failed:\n%s", tailBytes(out, 600)), e.win)
		return
	}
	pdf := strings.TrimSuffix(tex, ".tex") + ".pdf"
	e.setStatus("Wrote " + pdf)
}

// ---- form ------------------------------------------------------------------

func (e *editor) rebuild() {
	e.body.Objects = nil
	d := e.doc

	e.body.Add(widget.NewCard("Header", "", e.headerForm()))
	e.body.Add(widget.NewCard("Summary", "", multiline(&d.Summary)))
	e.body.Add(widget.NewCard("About (one paragraph per line)", "", aboutEntry(d)))
	e.body.Add(e.expSection())
	e.body.Add(e.projSection())
	e.body.Add(e.eduSection())
	e.body.Add(e.skillsSection())
	e.body.Refresh()
}

func (e *editor) headerForm() fyne.CanvasObject {
	d := e.doc
	return widget.NewForm(
		row("Name", &d.Name), row("Email", &d.Email), row("Phone", &d.Phone),
		row("Location", &d.Location), row("LinkedIn", &d.LinkedIn),
		row("GitHub", &d.GitHub), row("Telegram", &d.Telegram),
	)
}

func (e *editor) expSection() fyne.CanvasObject {
	d := e.doc
	return e.list("Experience",
		func() int { return len(d.Experience) },
		func(i int) fyne.CanvasObject {
			x := &d.Experience[i]
			return widget.NewForm(
				row("Company", &x.Company), row("Title", &x.Title),
				row("Dates", &x.Dates), row("Location", &x.Location),
				wide("Bullets", &x.Bullets),
			)
		},
		func() { d.Experience = append(d.Experience, cv.Experience{}) },
		func(i int) { d.Experience = remove(d.Experience, i) },
	)
}

func (e *editor) projSection() fyne.CanvasObject {
	d := e.doc
	return e.list("Projects",
		func() int { return len(d.Projects) },
		func(i int) fyne.CanvasObject {
			p := &d.Projects[i]
			return widget.NewForm(
				row("Name", &p.Name), row("Link", &p.Link),
				row("Dates", &p.Dates), row("Tech", &p.Tech),
				wide("Bullets", &p.Bullets),
			)
		},
		func() { d.Projects = append(d.Projects, cv.Project{}) },
		func(i int) { d.Projects = remove(d.Projects, i) },
	)
}

func (e *editor) eduSection() fyne.CanvasObject {
	d := e.doc
	return e.list("Education",
		func() int { return len(d.Education) },
		func(i int) fyne.CanvasObject {
			ed := &d.Education[i]
			return widget.NewForm(
				row("School", &ed.School), row("Degree", &ed.Degree),
				row("Dates", &ed.Dates), row("Location", &ed.Location),
				wide("Notes", &ed.Notes),
			)
		},
		func() { d.Education = append(d.Education, cv.Education{}) },
		func(i int) { d.Education = remove(d.Education, i) },
	)
}

func (e *editor) skillsSection() fyne.CanvasObject {
	return e.list("Skills (key : value)",
		func() int { return len(e.keys) },
		func(i int) fyne.CanvasObject {
			return widget.NewForm(row("Key", &e.keys[i]), row("Value", &e.vals[i]))
		},
		func() { e.keys = append(e.keys, ""); e.vals = append(e.vals, "") },
		func(i int) { e.keys = remove(e.keys, i); e.vals = remove(e.vals, i) },
	)
}

// list renders a titled, add/remove-able section over a slice abstraction.
func (e *editor) list(title string, count func() int, item func(i int) fyne.CanvasObject, add func(), del func(i int)) fyne.CanvasObject {
	items := container.NewVBox()
	var render func()
	render = func() {
		items.Objects = nil
		for i := 0; i < count(); i++ {
			i := i
			rm := widget.NewButton("Remove", func() { del(i); render() })
			items.Add(container.NewBorder(nil, nil, nil, rm, item(i)))
			items.Add(widget.NewSeparator())
		}
		items.Refresh()
	}
	render()
	addBtn := widget.NewButton("+ Add", func() { add(); render() })
	return widget.NewCard(title, "", container.NewVBox(items, addBtn))
}

// ---- widget helpers --------------------------------------------------------

func row(label string, target *string) *widget.FormItem {
	en := widget.NewEntry()
	en.SetText(*target)
	en.OnChanged = func(s string) { *target = s }
	return widget.NewFormItem(label, en)
}

func wide(label string, target *string) *widget.FormItem {
	en := widget.NewMultiLineEntry()
	en.SetText(*target)
	en.OnChanged = func(s string) { *target = s }
	en.SetMinRowsVisible(3)
	return widget.NewFormItem(label, en)
}

func multiline(target *string) fyne.CanvasObject {
	en := widget.NewMultiLineEntry()
	en.SetText(*target)
	en.OnChanged = func(s string) { *target = s }
	en.SetMinRowsVisible(3)
	return en
}

func aboutEntry(d *cv.CV) fyne.CanvasObject {
	en := widget.NewMultiLineEntry()
	en.SetText(strings.Join(d.About, "\n"))
	en.OnChanged = func(s string) {
		var ps []string
		for _, l := range strings.Split(s, "\n") {
			if t := strings.TrimSpace(l); t != "" {
				ps = append(ps, t)
			}
		}
		d.About = ps
	}
	en.SetMinRowsVisible(3)
	return en
}

// ---- skills <-> map glue ---------------------------------------------------

func (e *editor) loadSkills() {
	e.keys, e.vals = nil, nil
	for _, k := range e.doc.SkillsOrder() {
		e.keys = append(e.keys, k)
		e.vals = append(e.vals, e.doc.Skills[k])
	}
}

func (e *editor) commitSkills() {
	m := make(map[string]string, len(e.keys))
	var order []string
	for i, k := range e.keys {
		k = strings.TrimSpace(k)
		if k == "" {
			continue
		}
		m[k] = e.vals[i]
		order = append(order, k)
	}
	e.doc.Skills = m
	e.doc.SetSkillsOrder(order)
}

// ---- misc ------------------------------------------------------------------

func (e *editor) setStatus(s string) { e.status.SetText(s) }
func (e *editor) fail(err error)     { dialog.ShowError(err, e.win) }

func (e *editor) showList(title string, lines []string) {
	dialog.ShowInformation(title, strings.Join(lines, "\n"), e.win)
}

// deriveTex maps a cv_data[_role].json path to its .tex sibling; "" if no path.
func deriveTex(path string) string {
	if path == "" {
		return ""
	}
	base := strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
	switch {
	case base == "cv_data":
		base = "cv"
	case strings.HasPrefix(base, "cv_data_"):
		base = "cv_" + strings.TrimPrefix(base, "cv_data_")
	}
	return filepath.Join(filepath.Dir(path), base+".tex")
}

func remove[T any](s []T, i int) []T {
	return append(s[:i], s[i+1:]...)
}

func tailBytes(b []byte, n int) string {
	if len(b) > n {
		b = b[len(b)-n:]
	}
	return string(b)
}
