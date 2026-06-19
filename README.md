# cvgen

A Go CLI that turns one JSON data source into a CV across multiple formats —
LaTeX/PDF, portfolio JSON, Markdown, and plaintext.

> **Status:** early WIP. Scaffolding in progress.

## Goals

- Single source of truth (`cv_data.json`) → many outputs.
- Role-specific variants from the same generator.
- Parallel multi-variant builds, local PDF compilation, file watching, and
  content validation.
- Ships as a single cross-compiled binary (`go install` / Homebrew).

## Planned commands

```
cvgen build [input] [output]   # JSON -> LaTeX .tex
cvgen sync  [input] [output]   # JSON -> portfolio cv.json
cvgen build --all              # build every variant in parallel
cvgen pdf   [input]            # compile to PDF (XeLaTeX)
cvgen watch                    # auto-rebuild on save
cvgen validate [input]         # schema-check the JSON
```

## License

MIT
