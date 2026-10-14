# sdd

This repository defines sdd. It has two copies of the same rules. The docs are for people:
`README.md`, `records.md`, `foundations.md` and `schemas/`. The `sdd` skill in `skills/sdd/` is
for agents.

## Keep the skill and the docs the same

If you change a rule in the docs or in a schema, change the skill in the same commit. If you
change the skill, change the docs in the same commit. A rule that is in only one of them is a
defect.

| Docs | Skill |
|---|---|
| `records.md`: shape, fields, IDs, test citations and files | `skills/sdd/SKILL.md` |
| `records.md`: levels, parents, `tbr`, prose rules, retirement | `skills/sdd/references/writing.md` |
| `schemas/retired.schema.json` | "Retiring an ID" in `skills/sdd/references/writing.md` |
| `README.md`: the sdd tool | "Finding records" in `skills/sdd/SKILL.md` |
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

## Writing

Write all text in this repository in simple English. Use the Simple English skill.
