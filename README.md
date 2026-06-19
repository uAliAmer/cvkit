# cvkit

[![CI](https://github.com/uAliAmer/cvkit/actions/workflows/ci.yml/badge.svg)](https://github.com/uAliAmer/cvkit/actions/workflows/ci.yml)

A single-binary Go CLI that turns one JSON file into a CV. One source of truth,
many outputs — a LaTeX/PDF résumé, Markdown, and plain text, with role-specific
variants built from the same data.

It's a Go rewrite of a Python build script, and it does what the Python version
couldn't: build every variant in parallel, compile straight to PDF, watch for
changes, and validate the data before you ship.

## Install

```bash
go install github.com/uAliAmer/cvkit@latest
```

This installs to `$(go env GOPATH)/bin` (usually `~/go/bin`). If `cvkit` isn't
found afterward, that directory isn't on your `PATH`:

```bash
echo 'export PATH="$HOME/go/bin:$PATH"' >> ~/.bashrc && source ~/.bashrc
```

Or skip Go entirely — grab a prebuilt binary from the
[releases page](https://github.com/uAliAmer/cvkit/releases) (Linux, macOS,
Windows; amd64 and arm64) and drop it in a `PATH` directory:

```bash
# example: Linux amd64
tar -xzf cvkit_*_linux_amd64.tar.gz
sudo mv cvkit /usr/local/bin/
```

## Example

See [`examples/cv_data.json`](examples/cv_data.json) for a complete sample, the
generated [`examples/cv_example.tex`](examples/cv_example.tex), and the compiled
result: **[examples/cv_example.pdf](examples/cv_example.pdf)**.

```bash
cvkit build examples/cv_data.json examples/cv_example.tex   # render the .tex
cvkit pdf   examples/cv_data.json                           # compile to PDF (needs XeLaTeX)
```

## Quick start

```bash
cvkit build                 # cv_data.json -> cv.tex
cvkit pdf                   # cv_data.json -> cv.pdf  (needs XeLaTeX)
cvkit validate              # sanity-check the JSON before building
```

## Commands

| Command | What it does |
|---|---|
| `cvkit build [in] [out]` | Render a CV JSON to a LaTeX `.tex`. |
| `cvkit build --all [dir]` | Build every `cv_data*.json` variant in `dir`, in parallel. |
| `cvkit export [in] [out]` | Render to another format: `-f tex`, `-f md`, or `-f txt`. |
| `cvkit pdf [in]` | Build then compile to PDF with XeLaTeX. `--keep-tex` to retain the `.tex`. |
| `cvkit sync [in] [out]` | Validate then copy the JSON to the portfolio data path. `--force` to sync despite problems. |
| `cvkit validate [in]` | Check for missing fields and malformed entries. `--links` also HTTP-checks every project link. |
| `cvkit lint [in]` | Flag content-quality issues: weak verbs, missing metrics, passive voice, over-long bullets. `--strict` to fail. |
| `cvkit diff <base> <other>` | Show what differs between two variants (skills, projects, bullets). |
| `cvkit new <role>` | Scaffold `cv_data_<role>.json` from a base (`--from`). |
| `cvkit tailor [in] --jd <file>` | Match the CV against a job description; show matched keywords, gaps, and which entries to surface. |
| `cvkit watch [in]` | Rebuild the `.tex` whenever the JSON changes. |

Defaults: input `cv_data.json`; output is derived from the input
(`cv_data.json` → `cv.tex`, `cv_data_qa.json` → `cv_qa.tex`).

## Variants

Role-specific résumés are just differently-named data files; the same generator
handles them all:

```bash
cvkit build cv_data_qa.json          # -> cv_qa.tex
cvkit build cv_data_sysadmin.json    # -> cv_sysadmin.tex
cvkit build --all                    # build them all at once
```

## Desktop app

`cvkit-gui` is a cross-platform desktop editor (Linux/macOS/Windows) for the
same CV JSON, built with [Fyne](https://fyne.io). Open/edit/save a `cv_data.json`,
add or remove experience/projects/skills, and run Validate, Lint, Build `.tex`,
or PDF from the toolbar — all sharing the CLI's rendering and validation.

```bash
go install github.com/uAliAmer/cvkit/gui@latest   # installs the 'gui' binary
```

Building from source needs a C toolchain and OpenGL/X11 dev headers (Linux:
`libgl1-mesa-dev xorg-dev`); see the Fyne docs. Prebuilt desktop binaries are on
the releases page.

## Data format

`cv_data.json` holds your name, contact handles, summary, and arrays for
`experience`, `projects`, `education`, plus a `skills` object. Known skill keys
(`technical`, `business`, `tools`, `languages`) get nice labels; any other key
is used verbatim as its own label, which is how variants add custom skill
groups. Bullets are newline-separated strings.

## Compiling the LaTeX

The output uses sans fonts (tgheros / FiraMono), so compile with **XeLaTeX**,
not pdfLaTeX. `cvkit pdf` does this for you when a TeX distribution is
installed.

## Development

```bash
go test ./...     # unit tests
go vet ./...
go build -o cvkit .
```

Templates are embedded with `go:embed`, so the binary is fully self-contained —
no runtime files needed.

## License

MIT
