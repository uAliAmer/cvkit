# cvgen

[![CI](https://github.com/uAliAmer/cvgen/actions/workflows/ci.yml/badge.svg)](https://github.com/uAliAmer/cvgen/actions/workflows/ci.yml)

A single-binary Go CLI that turns one JSON file into a CV. One source of truth,
many outputs — a LaTeX résumé today, with role-specific variants built from the
same data.

It's a Go rewrite of a Python build script, and it does what the Python version
couldn't: build every variant in parallel, compile straight to PDF, watch for
changes, and validate the data before you ship.

## Install

```bash
go install github.com/uAliAmer/cvgen@latest
```

Or grab a prebuilt binary from the [releases page](https://github.com/uAliAmer/cvgen/releases)
(Linux, macOS, Windows — amd64 and arm64).

## Quick start

```bash
cvgen build                 # cv_data.json -> cv.tex
cvgen pdf                   # cv_data.json -> cv.pdf  (needs XeLaTeX)
cvgen validate              # sanity-check the JSON before building
```

## Commands

| Command | What it does |
|---|---|
| `cvgen build [in] [out]` | Render a CV JSON to a LaTeX `.tex`. |
| `cvgen build --all [dir]` | Build every `cv_data*.json` variant in `dir`, in parallel. |
| `cvgen pdf [in]` | Build then compile to PDF with XeLaTeX. `--keep-tex` to retain the `.tex`. |
| `cvgen sync [in] [out]` | Copy a validated JSON to the portfolio data path. |
| `cvgen validate [in]` | Check for missing fields and malformed entries. `--links` also HTTP-checks every project link. |
| `cvgen watch [in]` | Rebuild the `.tex` whenever the JSON changes. |

Defaults: input `cv_data.json`; output is derived from the input
(`cv_data.json` → `cv.tex`, `cv_data_qa.json` → `cv_qa.tex`).

## Variants

Role-specific résumés are just differently-named data files; the same generator
handles them all:

```bash
cvgen build cv_data_qa.json          # -> cv_qa.tex
cvgen build cv_data_sysadmin.json    # -> cv_sysadmin.tex
cvgen build --all                    # build them all at once
```

## Data format

`cv_data.json` holds your name, contact handles, summary, and arrays for
`experience`, `projects`, `education`, plus a `skills` object. Known skill keys
(`technical`, `business`, `tools`, `languages`) get nice labels; any other key
is used verbatim as its own label, which is how variants add custom skill
groups. Bullets are newline-separated strings.

## Compiling the LaTeX

The output uses sans fonts (tgheros / FiraMono), so compile with **XeLaTeX**,
not pdfLaTeX. `cvgen pdf` does this for you when a TeX distribution is
installed.

## Development

```bash
go test ./...     # unit tests
go vet ./...
go build -o cvgen .
```

Templates are embedded with `go:embed`, so the binary is fully self-contained —
no runtime files needed.

## License

MIT
