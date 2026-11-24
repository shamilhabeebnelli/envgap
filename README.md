# envgap

<img src="logo.svg" alt="envgap logo" width="64" align="right">

A small Go command that compares a `.env.example` file with an actual `.env` file and reports **key names and line numbers only**. Values are never printed, so the output is safe to paste into a ticket or CI log.

Free, MIT, standard library only, no network access, read-only.

```
$ envgap env.example env.actual
duplicate           actual   PORT line 3,4
extra               actual   SENTRY_DSN line 5
invalid-line        actual   - line 6
missing             actual   DEBUG line 4
placeholder         actual   API_KEY line 2
5 finding(s)
```

## Install and run

Needs Go 1.21 or newer.

```
go install github.com/shamilhabeebnelli/envgap@latest
envgap [flags] EXAMPLE_FILE ACTUAL_FILE
```

Flags: `-json` (machine output), `-ignore-extra` (skip keys only in the actual file), `-optional A,B` (keys allowed to be missing or empty).

Exit codes: `0` clean, `1` findings, `2` usage or read error.

## What it reports

| kind | meaning |
| --- | --- |
| `missing` | key is in the example, not in the actual file |
| `empty` | key is present but has an empty value |
| `extra` | key is in the actual file only |
| `placeholder` | actual value is identical to an example value that looks like a placeholder (`changeme`, `your_...`, `<...>`, and similar) |
| `duplicate` | key is defined more than once in one file (the last definition is the one compared) |
| `invalid-line` | line is not a comment, blank, or `KEY=VALUE` |
| `unclosed-quote` | a quote opens and does not close on the same line (multi-line values are not supported) |
| `space-around-equals` | spaces around `=`, which some loaders treat differently |
| `bom`, `crlf` | byte order mark or Windows line endings |

## Privacy

Findings carry a kind, a key name, and line numbers. Tests plant sentinel values in valid lines, invalid lines and unclosed quotes and check that none appear in text or JSON output. Key names still appear, so do not share output if key names are themselves sensitive. A file read error prints the path you passed, not any content.

## Limits

- Not a dotenv loader. No `${VAR}` interpolation, no multi-line values, no escape decoding. Different loaders (Docker Compose, dotenv libraries, shells) disagree on edge cases, so a clean result does not prove every loader reads the file the same way.
- The placeholder check is a word list. It misses placeholders it does not know and can flag real values that contain those words, but only when they equal the example value.
- It compares two files you name. It does not search a repository or check which keys your code actually reads.
- It does not check that values are valid or secret.

## Existing alternatives

Several tools already cover dotenv linting and comparing, for example [dotenv-linter](https://github.com/dotenv-linter/dotenv-linter) and [shahadulhaider/envdiff](https://github.com/shahadulhaider/envdiff). envgap is a deliberately small one: two files, no values in output, stable exit codes. See `RESEARCH.md`.

## Develop

```
go vet ./...
go test ./...
```

MIT licensed. See `LICENSE`.
