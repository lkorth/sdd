# sdd

sdd driven development, in the tradition of GNU.

sdd is a form of [spec-driven development](https://www.ibm.com/think/topics/spec-driven-development)
inspired by [OpenSpec](https://github.com/Fission-AI/openspec).
In spec-driven development, you write the requirements before the code. In sdd, each
requirement is a small typed record, not a paragraph. Each record has an ID and a parent.
For a system that already exists, an agent can [derive the records](deriving.md) from the code.
A parent is the requirement that a record exists to serve. With IDs and parents, a script
can answer questions that prose cannot answer:

1. Does a record point at a requirement that does not exist, or that you deleted?
2. Do two requirements use the same ID?
3. If you delete a requirement, which requirements depend on it?

## The core

| Piece | Summary | More |
|---|---|---|
| Record | `### ID: title`, a `yaml` block with `parent`, one `SHALL`, a `Rationale:` and a `Scenario:`, written in simple English | [records](records.md) |
| Levels | `L0` system, `L1` feature, `L2` detail. The ID prefix gives the level | [records](records.md#ids-and-levels) |

The repository holds only the current state: records, the ledger of retired IDs, code and
tests. Ephemeral artifacts like plans and trade-offs go in the pull request and in git history.

## What sdd is not

1. sdd is not a requirements freeze. It gives traceability without change boards or sign-off
   before work starts.
2. sdd is not a code generator. A specification states the intent. The code is the authority
   on behavior ([Naur](foundations.md#naur-the-limits-of-written-documentation)).

## Technical Writing

Write all text in sdd in simple English. This rule covers each record in the specs and each
doc in this directory. Simple English is a controlled form of English, in the spirit of
ASD-STE100 Simplified Technical English. Each sentence has one meaning for each reader, and a
reader outside the field understands it on one read. A requirement needs the same property,
because two readers who read it two ways build two different systems.

The rules, the one difference between a record and a doc, and the plugin that checks the text
are detaied in [records](records.md#writing-in-simple-english).

## Use with an agent

The `sdd` skill teaches an agent to read and write records. The agent loads it to work on code
that a record covers, and to add, change or retire a record. sdd depends on the
[Simple English](https://github.com/AminBlg/SimpleEnglish) skill for the general writing rules.

The `derive` skill teaches an agent to write the records of a system that already exists. The agent
reads the code, the tests, the git history, the pull requests, and the documents and diagrams that
they link to. Then it writes the records and reports where the evidence stops. To use it, run
`/sdd:derive`. For the method, read [deriving](deriving.md).

To install both plugins in Claude Code, run these commands in the directory that holds your
clone of this repository:

```bash
claude plugin marketplace add AminBlg/SimpleEnglish && claude plugin install simple-english@simple-english
claude plugin marketplace add ./sdd && claude plugin install sdd@sdd
```

## The sdd tool

The `sdd` tool answers questions about the records, so that you and your agent do not read
whole files. Each answer gives the file and the line range of each record:

```console
$ sdd show L2_PER_004
L2_PER_004: Acknowledgement is durable before deletion
sdd/features/persistence.md:11-25  parent L1_LNK_001
The system SHALL write an upload acknowledgement to durable storage before deleting the
acknowledged record.

$ sdd ancestors L2_PER_004 --short
L1_LNK_001  sdd/features/link.md:11-25  An interrupted upload completes without duplicate delivery
L0_001  sdd/system.md:5-18  Recorded data reaches the archive
```

To see each command, run `sdd --help`. To see the flags of one command, run
`sdd <command> -h`. Add `--json` to any command to get JSON.

### See the records in Obsidian

The `canvas` command prints an [Obsidian Canvas](https://jsoncanvas.org/) of the parent tree.
A canvas is a JSON file that Obsidian shows as a board of cards. Each record is a card below its
parent, with an arrow to the parent. Each card shows the record from its file, so you can read a
record on the board, or open its file with a double click. A group of cards at the left shows
each feature file. Each feature has one color. A red card is a parent that is missing or retired.

```bash
sdd canvas --out sdd/records.canvas
```

The paths in the canvas start at the vault, the directory that you open in Obsidian. The default
vault is the directory that holds `sdd/`. If you open a different directory, add `--vault <dir>`
the first time. After that, `--out` keeps the vault of the file.

Each card reads its record from the file, so a change to the text of a record shows at once. A
new record, a deleted record, a new parent or a new title needs a new canvas. If the canvas changed,
`--out` writes the file. When you move a card, Obsidian saves the canvas. The next `--out` puts the
card back, because the tool owns the layout.

Keep the canvas at `sdd/records.canvas`. Then three things refresh it:

1. `sdd delete` refreshes it after it deletes a record.
2. The `sdd` plugin has a Claude Code hook that refreshes it after each edit that Claude makes.
   The hook does nothing in a project with no canvas.
3. A git hook can refresh it before each commit. Put this script in `.git/hooks/pre-commit` and
   make it executable. If git tracks the canvas, the script also stages it:

```sh
#!/bin/sh
canvas=sdd/records.canvas
[ -f "$canvas" ] || exit 0
sdd canvas --out "$canvas" >/dev/null || exit 1
if git ls-files --error-unmatch "$canvas" >/dev/null 2>&1; then
  git add "$canvas"
fi
```

If you commit the canvas, add `sdd canvas --check sdd/records.canvas` to CI. If the canvas is not
current, it exits 1. If you do not commit the canvas, add it to `.gitignore`.

The tool is a static binary with no runtime. The plugin ships it for Linux and macOS, on amd64
and arm64, at `skills/sdd/bin/sdd`. To use it yourself, put that directory on your `PATH`, or
run `go install github.com/lkorth/sdd/cmd/sdd@latest`.
