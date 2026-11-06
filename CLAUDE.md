# sdd

This repository defines sdd. It has two copies of the same rules. The docs are for people:
`README.md`, `records.md`, `deriving.md`, `foundations.md` and `schemas/`. The `sdd` skill in
`skills/sdd/` and the `derive` skill in `skills/derive/` are for agents.

## Keep the skill and the docs the same

If you change a rule in the docs or in a schema, change the skill in the same commit. If you
change the skill, change the docs in the same commit. A rule that is in only one of them is a
defect.

| Docs | Skill |
|---|---|
| `records.md`: shape, fields, IDs, test citations and files | `skills/sdd/SKILL.md` |
| `records.md`: levels, parents, `tbr`, prose rules, retirement | `skills/sdd/references/writing.md` |
| `schemas/retired.schema.json` | "Retiring an ID" in `skills/sdd/references/writing.md` |
| `sdd --help`, the usage text in `cmd/sdd/main.go` | "Finding records" in `skills/sdd/SKILL.md` |
| `deriving.md`: sources, feature ranks, classes, rationale phrases, tests, report | `skills/derive/SKILL.md` and `skills/derive/references/evidence.md` |
| `foundations.md` | No copy. It explains why sdd exists, and an agent does not need it to read or write a record |

After you change the skill, run `claude plugin validate .` and fix each error. Do not change the
plugin version before the first release.

## The sdd tool

The Go source of the `sdd` tool is in `cmd/sdd/` and `internal/record/`. The parser in
`internal/record/` follows the record grammar in `records.md`. If you change that grammar, change
the parser, its tests and `testdata/` in the same commit.

The plugin ships the binaries in `skills/sdd/bin/`. After you change the Go source, follow these
steps:

1. Run `make binaries`. It runs the tests, builds each binary, and runs the launcher with the
   binary for this machine. If a step fails, it stops, and `skills/sdd/bin/` does not change.
   Do not build a binary into `skills/sdd/bin/` in a different way.
2. Commit the binaries with the source change.

### The canvas

`sdd canvas` writes an Obsidian Canvas, in the [JSON Canvas](https://jsoncanvas.org/spec/1.0/)
format. The source is `cmd/sdd/canvas.go`. Before you change it, read the `json-canvas` skill in
[kepano/obsidian-skills](https://github.com/kepano/obsidian-skills). That skill gives the
conventions of Obsidian for IDs, colors, layout and file nodes.

The plugin hook in `hooks/` refreshes `sdd/records.canvas` after each edit that Claude makes. If
you change the name of that file or the flags of `--out`, make these changes in the same commit.
Change the hook, the git hook in `README.md`, and `refreshCanvas` in `cmd/sdd/commands.go`.

## Writing

Write all text in this repository in simple English. Use the Simple English skill.
