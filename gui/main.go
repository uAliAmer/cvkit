// Command cvkit-gui is a cross-platform desktop editor for cvkit CV JSON files.
// It edits the same cv_data.json the CLI uses and shares the internal/cv
// package for loading, rendering, validation, and serialization.
package main

import (
	"fmt"
	"image/color"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/uAliAmer/cvkit/internal/cv"
)

// lightTheme forces the light variant regardless of the OS setting.
type lightTheme struct{}

func (lightTheme) Color(n fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	return theme.DefaultTheme().Color(n, theme.VariantLight)
}
func (lightTheme) Font(s fyne.TextStyle) fyne.Resource     { return theme.DefaultTheme().Font(s) }
func (lightTheme) Icon(n fyne.ThemeIconName) fyne.Resource { return theme.DefaultTheme().Icon(n) }
func (lightTheme) Size(n fyne.ThemeSizeName) float32       { return theme.DefaultTheme().Size(n) }

type editor struct {
	app    fyne.App
	win    fyne.Window
	doc    *cv.CV
	keys   []string // skill keys, ordered
	vals   []string // skill values, parallel to keys
	path   string   // current file path ("" = unsaved)
	dirty  bool
	status *widget.Label
	valBar *widget.Label
	holder *fyne.Container // swapped on open/new
	tabs   *container.AppTabs
}

func main() {
	a := app.New()
	a.Settings().SetTheme(lightTheme{})
	w := a.NewWindow("CVKit")

	e := &editor{
		app: a, win: w, doc: emptyCV(),
		status: widget.NewLabel("New CV"),
		valBar: widget.NewLabel(""),
		holder: container.NewStack(),
	}
	e.refresh()
	e.updateTitle()

	top := container.NewVBox(e.toolbar(), e.valBar, widget.NewSeparator())
	w.SetContent(container.NewBorder(top, e.status, nil, nil, e.holder))
	w.Resize(fyne.NewSize(880, 760))
	w.ShowAndRun()
}

func emptyCV() *cv.CV { return &cv.CV{Skills: map[string]string{}} }

// ---- toolbar & actions -----------------------------------------------------

func (e *editor) toolbar() fyne.CanvasObject {
	b := func(label string, icon fyne.Resource, fn func()) *widget.Button {
		btn := widget.NewButtonWithIcon(label, icon, fn)
		return btn
	}
	save := b("Save", theme.DocumentSaveIcon(), e.actSave)
	save.Importance = widget.HighImportance
	return container.NewHBox(
		b("New", theme.DocumentCreateIcon(), e.actNew),
		b("Open", theme.FolderOpenIcon(), e.actOpen),
		save,
		b("Save As", theme.ContentCopyIcon(), e.actSaveAs),
		widget.NewSeparator(),
		b("Build .tex", theme.DocumentIcon(), e.actBuild),
		b("Export PDF", theme.DownloadIcon(), e.actPDF),
	)
}

func (e *editor) actNew() {
	e.confirmDiscard(func() {
		e.doc = emptyCV()
		e.keys, e.vals, e.path, e.dirty = nil, nil, "", false
		e.refresh()
		e.updateTitle()
		e.setStatus("New CV")
	})
}

func (e *editor) actOpen() {
	e.confirmDiscard(func() {
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
			e.path, e.dirty = path, false
			e.refresh()
			e.updateTitle()
			e.setStatus("Opened " + path)
		}, e.win)
	})
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
		e.updateTitle()
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
	e.dirty = false
	e.updateTitle()
	e.setStatus("Saved " + path)
}

func (e *editor) actBuild() {
	e.commitSkills()
	out := deriveTex(e.path)
	if out == "" {
		e.askSaveFirst("Build")
		return
	}
	if err := os.WriteFile(out, []byte(e.doc.RenderLaTeX()), 0o644); err != nil {
		e.fail(err)
		return
	}
	dialog.ShowInformation("Build", "Wrote "+out, e.win)
	e.setStatus("Wrote " + out)
}

func (e *editor) actPDF() {
	bin, err := exec.LookPath("xelatex")
	if err != nil {
		dialog.ShowError(fmt.Errorf("xelatex not found in PATH.\nInstall a TeX distribution (TeX Live / MacTeX) to export PDF."), e.win)
		return
	}
	e.commitSkills()
	tex := deriveTex(e.path)
	if tex == "" {
		e.askSaveFirst("Export PDF")
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
	dialog.ShowInformation("Export PDF", "Wrote "+pdf, e.win)
	e.setStatus("Wrote " + pdf)
}

// ---- layout (tabs) ---------------------------------------------------------

func (e *editor) refresh() {
	idx := 0
	if e.tabs != nil {
		idx = e.tabs.SelectedIndex()
	}
	e.tabs = container.NewAppTabs(
		container.NewTabItemWithIcon("Basics", theme.AccountIcon(), vscroll(e.basicsTab())),
		container.NewTabItemWithIcon("Experience", theme.HistoryIcon(), vscroll(e.expSection())),
		container.NewTabItemWithIcon("Projects", theme.GridIcon(), vscroll(e.projSection())),
		container.NewTabItemWithIcon("Education", theme.DocumentIcon(), vscroll(e.eduSection())),
		container.NewTabItemWithIcon("Skills", theme.ListIcon(), vscroll(e.skillsSection())),
	)
	if idx >= 0 && idx < len(e.tabs.Items) {
		e.tabs.SelectIndex(idx)
	}
	e.holder.Objects = []fyne.CanvasObject{e.tabs}
	e.holder.Refresh()
	e.revalidate()
}

func vscroll(o fyne.CanvasObject) fyne.CanvasObject {
	s := container.NewVScroll(o)
	s.SetMinSize(fyne.NewSize(0, 560))
	return s
}

func (e *editor) basicsTab() fyne.CanvasObject {
	d := e.doc
	hdr := widget.NewForm(
		e.row("Full name", &d.Name, "e.g. John Doe"),
		e.row("Email", &d.Email, "you@example.com"),
		e.row("Phone", &d.Phone, "+1 555 010 2020"),
		e.row("Location", &d.Location, "City, Country"),
		e.row("LinkedIn", &d.LinkedIn, "username or full URL"),
		e.row("GitHub", &d.GitHub, "username or full URL"),
		e.row("Telegram", &d.Telegram, "username (optional)"),
	)
	return container.NewVBox(
		widget.NewCard("Who you are", "", hdr),
		widget.NewCard("Summary", "A short 2–3 line pitch", e.multiline(&d.Summary, "What you do and what you're great at...")),
		widget.NewCard("About", "One paragraph per line (used by the web portfolio)", e.aboutEntry()),
	)
}

func (e *editor) expSection() fyne.CanvasObject {
	d := e.doc
	return e.list("Work experience", "No jobs yet. Click “Add job”.", "Add job",
		func() int { return len(d.Experience) },
		func(i int) fyne.CanvasObject {
			x := &d.Experience[i]
			return widget.NewForm(
				e.row("Company", &x.Company, ""),
				e.row("Job title", &x.Title, ""),
				e.dateItem("When", &x.Dates),
				e.row("Location", &x.Location, ""),
				e.bullets("Highlights", &x.Bullets),
			)
		},
		func() { d.Experience = append(d.Experience, cv.Experience{}) },
		func(i int) { d.Experience = remove(d.Experience, i) },
	)
}

func (e *editor) projSection() fyne.CanvasObject {
	d := e.doc
	return e.list("Projects", "No projects yet. Click “Add project”.", "Add project",
		func() int { return len(d.Projects) },
		func(i int) fyne.CanvasObject {
			p := &d.Projects[i]
			return widget.NewForm(
				e.row("Name", &p.Name, ""),
				e.row("Link", &p.Link, "https://github.com/you/project"),
				e.dateItem("When", &p.Dates),
				e.row("Tech", &p.Tech, "Go, PostgreSQL, React"),
				e.bullets("Highlights", &p.Bullets),
			)
		},
		func() { d.Projects = append(d.Projects, cv.Project{}) },
		func(i int) { d.Projects = remove(d.Projects, i) },
	)
}

func (e *editor) eduSection() fyne.CanvasObject {
	d := e.doc
	return e.list("Education", "No schools yet. Click “Add school”.", "Add school",
		func() int { return len(d.Education) },
		func(i int) fyne.CanvasObject {
			ed := &d.Education[i]
			return widget.NewForm(
				e.row("School", &ed.School, ""),
				e.row("Degree", &ed.Degree, ""),
				e.dateItem("When", &ed.Dates),
				e.row("Location", &ed.Location, ""),
				e.bullets("Notes", &ed.Notes),
			)
		},
		func() { d.Education = append(d.Education, cv.Education{}) },
		func(i int) { d.Education = remove(d.Education, i) },
	)
}

func (e *editor) skillsSection() fyne.CanvasObject {
	return e.list("Skills", "Add groups like Technical, Tools, Languages.", "Add skill group",
		func() int { return len(e.keys) },
		func(i int) fyne.CanvasObject {
			return widget.NewForm(
				e.row("Group", &e.keys[i], "Technical"),
				e.row("Items", &e.vals[i], "Go, Docker, PostgreSQL"),
			)
		},
		func() { e.keys = append(e.keys, ""); e.vals = append(e.vals, "") },
		func(i int) { e.keys = remove(e.keys, i); e.vals = remove(e.vals, i) },
	)
}

// list renders a titled, add/remove-able section over a slice abstraction.
func (e *editor) list(title, empty, addLabel string, count func() int, item func(i int) fyne.CanvasObject, add func(), del func(i int)) fyne.CanvasObject {
	items := container.NewVBox()
	var render func()
	render = func() {
		items.Objects = nil
		if count() == 0 {
			items.Add(widget.NewLabel(empty))
		}
		for i := 0; i < count(); i++ {
			i := i
			rm := widget.NewButtonWithIcon("", theme.DeleteIcon(), func() {
				dialog.ShowConfirm("Remove", "Remove this entry?", func(ok bool) {
					if ok {
						del(i)
						e.touch()
						render()
					}
				}, e.win)
			})
			rm.Importance = widget.LowImportance
			head := container.NewBorder(nil, nil, nil, rm, widget.NewLabelWithStyle(fmt.Sprintf("#%d", i+1), fyne.TextAlignLeading, fyne.TextStyle{Bold: true}))
			items.Add(widget.NewCard("", "", container.NewVBox(head, item(i))))
		}
		items.Refresh()
	}
	render()
	addBtn := widget.NewButtonWithIcon(addLabel, theme.ContentAddIcon(), func() { add(); e.touch(); render() })
	addBtn.Importance = widget.MediumImportance
	return container.NewVBox(
		widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		items, addBtn,
	)
}

// ---- widget helpers --------------------------------------------------------

func (e *editor) row(label string, target *string, placeholder string) *widget.FormItem {
	en := widget.NewEntry()
	en.SetText(*target)
	if placeholder != "" {
		en.SetPlaceHolder(placeholder)
	}
	en.OnChanged = func(s string) { *target = s; e.touch() }
	return widget.NewFormItem(label, en)
}

func (e *editor) bullets(label string, target *string) *widget.FormItem {
	en := widget.NewMultiLineEntry()
	en.SetText(*target)
	en.SetPlaceHolder("One achievement per line. Start with a verb, add a number.")
	en.OnChanged = func(s string) { *target = s; e.touch() }
	en.SetMinRowsVisible(3)
	return widget.NewFormItem(label, en)
}

func (e *editor) multiline(target *string, placeholder string) fyne.CanvasObject {
	en := widget.NewMultiLineEntry()
	en.SetText(*target)
	en.SetPlaceHolder(placeholder)
	en.OnChanged = func(s string) { *target = s; e.touch() }
	en.SetMinRowsVisible(3)
	return en
}

func (e *editor) aboutEntry() fyne.CanvasObject {
	d := e.doc
	en := widget.NewMultiLineEntry()
	en.SetText(strings.Join(d.About, "\n"))
	en.SetMinRowsVisible(3)
	en.OnChanged = func(s string) {
		var ps []string
		for _, l := range strings.Split(s, "\n") {
			if t := strings.TrimSpace(l); t != "" {
				ps = append(ps, t)
			}
		}
		d.About = ps
		e.touch()
	}
	return en
}

// ---- date range picker -----------------------------------------------------

var months = []string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}

const noneOpt = "—"

func monthSelect() *widget.Select {
	opts := append([]string{noneOpt}, months...)
	s := widget.NewSelect(opts, nil)
	s.SetSelectedIndex(0)
	return s
}

func yearSelect() *widget.Select {
	now := time.Now().Year()
	opts := []string{noneOpt}
	for y := now + 5; y >= 1980; y-- {
		opts = append(opts, strconv.Itoa(y))
	}
	s := widget.NewSelect(opts, nil)
	s.SetSelectedIndex(0)
	return s
}

func (e *editor) dateItem(label string, target *string) *widget.FormItem {
	sM, sY := monthSelect(), yearSelect()
	eM, eY := monthSelect(), yearSelect()
	present := widget.NewCheck("Present", nil)

	if s, en, pres, ok := parseRange(*target); ok {
		setSel(sM, s.month)
		setSel(sY, s.year)
		if pres {
			present.SetChecked(true)
		} else {
			setSel(eM, en.month)
			setSel(eY, en.year)
		}
	}

	apply := func() {
		*target = formatRange(sM.Selected, sY.Selected, eM.Selected, eY.Selected, present.Checked)
		if present.Checked {
			eM.Disable()
			eY.Disable()
		} else {
			eM.Enable()
			eY.Enable()
		}
		e.touch()
	}
	for _, s := range []*widget.Select{sM, sY, eM, eY} {
		s.OnChanged = func(string) { apply() }
	}
	present.OnChanged = func(bool) { apply() }
	if present.Checked {
		eM.Disable()
		eY.Disable()
	}

	from := container.NewHBox(widget.NewLabel("From"), sM, sY)
	to := container.NewHBox(widget.NewLabel("To"), eM, eY, present)
	return widget.NewFormItem(label, container.NewVBox(from, to))
}

func setSel(s *widget.Select, v string) {
	if v == "" {
		return
	}
	for _, o := range s.Options {
		if o == v {
			s.SetSelected(v)
			return
		}
	}
}

type datePart struct{ month, year string }

var rangeSep = regexp.MustCompile(`\s*(?:--|–|-)\s*`)

func parseRange(s string) (start, end datePart, present, ok bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return
	}
	parts := rangeSep.Split(s, 2)
	start, ok = parseDate(parts[0])
	if len(parts) == 2 {
		if strings.EqualFold(strings.TrimSpace(parts[1]), "present") {
			present = true
		} else {
			end, _ = parseDate(parts[1])
		}
	}
	return
}

func parseDate(tok string) (datePart, bool) {
	f := strings.Fields(strings.TrimSpace(tok))
	var dp datePart
	for _, w := range f {
		if isYear(w) {
			dp.year = w
		} else if m := normMonth(w); m != "" {
			dp.month = m
		}
	}
	return dp, dp.year != "" || dp.month != ""
}

func isYear(s string) bool {
	if len(s) != 4 {
		return false
	}
	_, err := strconv.Atoi(s)
	return err == nil
}

func normMonth(s string) string {
	if len(s) < 3 {
		return ""
	}
	low := strings.ToLower(s[:3])
	p := strings.ToUpper(low[:1]) + low[1:] // "mar" -> "Mar"
	for _, m := range months {
		if m == p {
			return m
		}
	}
	return ""
}

func formatDate(m, y string) string {
	m, y = clean(m), clean(y)
	switch {
	case m != "" && y != "":
		return m + " " + y
	case y != "":
		return y
	default:
		return ""
	}
}

func formatRange(sM, sY, eM, eY string, present bool) string {
	start := formatDate(sM, sY)
	end := ""
	if present {
		end = "Present"
	} else {
		end = formatDate(eM, eY)
	}
	switch {
	case start != "" && end != "":
		return start + " -- " + end
	case start != "":
		return start
	default:
		return end
	}
}

func clean(s string) string {
	if s == noneOpt {
		return ""
	}
	return s
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

// ---- live validation, status, dialogs --------------------------------------

func (e *editor) touch() {
	e.dirty = true
	e.updateTitle()
	e.revalidate()
}

func (e *editor) revalidate() {
	probs := e.doc.Validate()
	if len(probs) == 0 {
		e.valBar.SetText("✓ Looks good — ready to build")
		return
	}
	if len(probs) > 3 {
		probs = append(probs[:3], fmt.Sprintf("+%d more", len(probs)-3))
	}
	e.valBar.SetText("⚠ " + strings.Join(probs, "   •   "))
}

func (e *editor) updateTitle() {
	name := "Untitled"
	if e.path != "" {
		name = filepath.Base(e.path)
	}
	if e.dirty {
		name += " *"
	}
	e.win.SetTitle("CVKit — " + name)
}

func (e *editor) setStatus(s string) { e.status.SetText(s) }
func (e *editor) fail(err error)     { dialog.ShowError(err, e.win) }

func (e *editor) askSaveFirst(action string) {
	dialog.ShowInformation(action, "Save the CV first, then "+action+".", e.win)
}

func (e *editor) confirmDiscard(then func()) {
	if !e.dirty {
		then()
		return
	}
	dialog.ShowConfirm("Unsaved changes", "Discard your unsaved changes?", func(ok bool) {
		if ok {
			then()
		}
	}, e.win)
}

// ---- misc ------------------------------------------------------------------

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
