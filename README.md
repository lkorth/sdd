# sdd

sdd driven development, in the tradition of GNU.

sdd is a form of [spec-driven development](https://www.ibm.com/think/topics/spec-driven-development)
inspired by [OpenSpec](https://github.com/Fission-AI/openspec).
In spec-driven development, you write the requirements before the code. In sdd, each
requirement is a small typed record, not a paragraph. Each record has an ID and a parent.
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

The commands are `show`, `ancestors`, `children`, `list`, `search`, `features`, `next` and
`delete`. Run `sdd --help` for the flags. Add `--json` to any command to get JSON.

The tool is a static binary with no runtime. The plugin ships it for Linux and macOS, on amd64
and arm64, at `skills/sdd/bin/sdd`. To use it yourself, put that directory on your `PATH`, or
run `go install github.com/lkorth/sdd/cmd/sdd@latest`.
