# cvkit

[![CI](https://github.com/uAliAmer/cvkit/actions/workflows/ci.yml/badge.svg)](https://github.com/uAliAmer/cvkit/actions/workflows/ci.yml)

Build your CV from one JSON file — as a **PDF, LaTeX résumé, Markdown, or plain
text**. cvkit is both a desktop app and a command-line tool that share one data
source, so the GUI and CLI always produce the same result. Make role-specific
variants (QA, backend, PM…) from the same history without copy-pasting.

- **Desktop app** — a friendly editor with live validation and one-click PDF
  export. No LaTeX, no terminal needed.
- **CLI** — scriptable build/validate/lint/diff/tailor, parallel variant builds,
  and watch mode.

PDF export is rendered **natively in Go — no LaTeX install required** — so it
works out of the box on any machine.

---

## Desktop app (easiest)

`cvkit-gui` runs on Linux, macOS, and Windows. Download a prebuilt binary from
the [releases page](https://github.com/uAliAmer/cvkit/releases)
(`cvkit-gui_*_linux_x64`, `_macos_arm64`, `_windows_x64`), unpack, and run it.

What you get:

- **Tabbed editor** — Basics, Experience, Projects, Education, Skills.
- **Month/year date pickers** with a *Present* toggle — no date typing.
- **Live validation bar** that updates as you type (missing name, email, etc.).
- **Export PDF** in one click — native, clickable shortened links, no LaTeX.
- **Build .tex** for the fully typeset version, plus Save/Open/New with an
  unsaved-changes guard. Light theme.

Install from source (needs a C toolchain + OpenGL/X11 headers; on Linux
`libgl1-mesa-dev xorg-dev`):

```bash
go install github.com/uAliAmer/cvkit/gui@latest   # installs the 'gui' binary
```

---

## Command-line tool

```bash
go install github.com/uAliAmer/cvkit@latest
```

This installs to `$(go env GOPATH)/bin` (usually `~/go/bin`). If `cvkit` isn't
found afterward, add that directory to your `PATH`:

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

Or grab a prebuilt CLI binary from the
[releases page](https://github.com/uAliAmer/cvkit/releases) (Linux, macOS,
Windows; amd64 and arm64) and drop it in a `PATH` directory:

```bash
tar -xzf cvkit_*_linux_amd64.tar.gz && sudo mv cvkit /usr/local/bin/
```

### Quick start

```bash
cvkit validate                       # sanity-check the JSON
cvkit export -f pdf                  # cv_data.json -> cv.pdf  (native, no LaTeX)
cvkit build                          # cv_data.json -> cv.tex  (typeset source)
```

### Commands

| Command | What it does |
|---|---|
| `cvkit export [in] [out]` | Render to `-f pdf` (native, no LaTeX), `-f md`, `-f txt`, or `-f tex`. |
| `cvkit build [in] [out]` | Render a CV JSON to a LaTeX `.tex`. |
| `cvkit build --all [dir]` | Build every `cv_data*.json` variant in `dir`, in parallel. |
| `cvkit pdf [in]` | Typeset PDF via XeLaTeX (best-looking; needs a TeX install). `--keep-tex` keeps the `.tex`. |
| `cvkit validate [in]` | Check for missing fields and malformed entries. `--links` HTTP-checks every project link. |
| `cvkit lint [in]` | Flag weak verbs, missing metrics, passive voice, over-long bullets. `--strict` to fail. |
| `cvkit diff <base> <other>` | Show what differs between two variants (skills, projects, bullets). |
| `cvkit new <role>` | Scaffold `cv_data_<role>.json` from a base (`--from`). |
| `cvkit tailor [in] --jd <file>` | Match the CV against a job description; show matched keywords, gaps, and which entries to surface. |
| `cvkit sync [in] [out]` | Validate then copy the JSON to a portfolio data path. `--force` to override. |
| `cvkit watch [in]` | Rebuild the `.tex` whenever the JSON changes. |

Defaults: input `cv_data.json`; output is derived from the input
(`cv_data.json` → `cv.tex`, `cv_data_qa.json` → `cv_qa.tex`).

---

## Example

See [`examples/cv_data.json`](examples/cv_data.json) for a complete sample and
the rendered outputs:
[`cv_example.pdf`](examples/cv_example.pdf) ·
[`cv_example.tex`](examples/cv_example.tex) ·
[`cv_example.md`](examples/cv_example.md) ·
[`cv_example.txt`](examples/cv_example.txt).

```bash
cvkit export -f pdf examples/cv_data.json   # native PDF, no LaTeX
cvkit build examples/cv_data.json examples/cv_example.tex
```

## Variants

Role-specific résumés are just differently-named data files; the same generator
handles them all:

```bash
cvkit new qa                         # -> cv_data_qa.json (edit for the role)
cvkit build cv_data_qa.json          # -> cv_qa.tex
cvkit build --all                    # build every variant at once
cvkit diff cv_data.json cv_data_qa.json
```

## Data format

`cv_data.json` holds your name, contact handles, summary, and arrays for
`experience`, `projects`, `education`, plus a `skills` object. Known skill keys
(`technical`, `business`, `tools`, `languages`) get nice labels; any other key
is used verbatim as its own label, which is how variants add custom skill
groups. Bullets are newline-separated strings.

## PDF: two ways

- **Native (default, no deps):** `cvkit export -f pdf` and the GUI's *Export
  PDF* render a clean, ATS-friendly PDF directly in Go — clickable shortened
  links, works anywhere.
- **Typeset (LaTeX):** `cvkit build` emits a `.tex` using sans fonts
  (tgheros / FiraMono); compile with **XeLaTeX** (not pdfLaTeX). `cvkit pdf`
  does this automatically when a TeX distribution is installed.

## Development

```bash
go test ./...     # unit tests (incl. LaTeX golden parity)
go vet ./...
go build ./...    # CLI + GUI (GUI needs the OpenGL/X11 headers above)
```

Templates are embedded with `go:embed`, so binaries are self-contained — no
runtime files needed.

## License

MIT
