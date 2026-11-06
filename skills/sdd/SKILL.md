---
name: sdd
description: Read and write sdd requirement records, the typed requirements in sdd/system.md and sdd/features/*.md. Use it to find a record, to implement, test or review code against a record, or to cite a record ID. Also use it before you add, edit or retire a record.
---

# sdd records

A record is one requirement. It has an ID, a parent, and exactly one `SHALL`. Text that does
not have the record shape is commentary, and it does not bind anything.

## Finding records

To find or read a record, use the `sdd` tool at `bin/sdd` in the directory of this skill. Do not
read a whole feature file to find one record. The tool finds the `sdd/` directory from the
working directory. For another project, add `--root <dir>`.

| Command | Prints |
|---|---|
| `sdd show ID...` | The location, the fields and the obligation of each record |
| `sdd ancestors ID` | The obligation of each parent, nearest first, up to L0 |
| `sdd children ID` | One line for each child. Add `--all` for each descendant |
| `sdd list` | One line for each record. Filter with `--level`, `--feature` and `--status` |
| `sdd search PATTERN` | Each record line that matches a regular expression, ignoring case |
| `sdd features` | Each feature code, its file and its record counts |
| `sdd next LEVEL [CODE]` | The next free ID, for example `sdd next L2 LNK` |
| `sdd delete ID` | Deletes the record and retires its ID. Read the writing rules first |
| `sdd canvas` | An Obsidian Canvas of the parent tree, as JSON. `--out FILE` writes it if it changed. `--check FILE` exits 1 if it is stale |

`show`, `ancestors`, `children` and `list` also take `--short` for one line for each record, and
`--full` for the whole record. Add `--json` to any command to get JSON.

Each answer gives a location such as `sdd/features/link.md:29-43`. To read more, read only those
lines. If `show` names a retired ID, it prints the ledger line, for example "Renamed to
L1_NET_002". If an ID is missing or the search finds nothing, the exit code is 1.

If the project has `sdd/records.canvas`, `delete` and a hook of this plugin refresh it after each
change. Do not edit the canvas yourself. If the hook reports an error, fix the record file that
it names.

## Reading a record

1. The obligation paragraph binds. That is the `SHALL` sentence and the sentences before it in
   the same paragraph.
2. The rationale does not bind. It tells you why the record exists and where each
   number came from. Read it before you change code that the record covers.
3. Each `Scenario:` maps to a test. The name of that test contains the record ID with no change,
   for example `test_L2_PER_004_ack_is_durable_before_delete`. Name the ID in the commit too.
4. If a record has `status: tbr`, a number in it is not chosen yet. Do not choose a value for a
   placeholder such as `TBR_1` in code. Ask the person who owns the record.
5. A record states what the system does. The code is the authority on how it does it.

## Grammar

A record has these pieces, in this order:

````markdown
### L2_PER_004: Acknowledgement is durable before deletion

```yaml
parent: L1_LNK_002
```

The system SHALL write an upload acknowledgement to durable storage before deleting the
acknowledged record.

Rationale: If the system deletes a record before its acknowledgement is durable, a power loss
can lose the record. No retry can recover a record that the system deleted.

Scenario: Power is removed between acknowledgement and deletion
- WHEN power is removed after an acknowledgement is received and before the record is deleted
- THEN the record is present on restart and is not uploaded a second time
````

| Piece | Form | Binds | Contents |
|---|---|---|---|
| Heading | `### <ID>: <title>` | No | The ID, and a title for people |
| Fields | A `yaml` block of `key: value` lines | Yes | `parent`, and as an option, `status`. No prose |
| Obligation | The first paragraph after the fields block | Yes | Exactly one `SHALL`. Short sentences before it can define its terms |
| Rationale | From `Rationale:` to the first `Scenario:` line. It can have more than one paragraph | No | Why the record exists, and where each number came from |
| Scenarios | One or more. Each is a `Scenario: <name>` line, then `- WHEN`, `- THEN` and `- AND` lines | Yes | What the tests assert |

A record runs from its heading to the next heading of any level, or to the end of the file.
Only records use `###` headings. Put commentary under a heading that is not a record. A record
has no other markup.

## Fields

A record has two fields and no others.

| Field | Values | Meaning |
|---|---|---|
| `parent` | An ID, or `self` | Necessary. The requirement that this record exists to serve. Only an L0 record uses `self` |
| `status` | `tbr` | Optional. No `status` means that the record is active. `tbr` means "to be resolved": the requirement is real, but a number in it is not chosen yet |

## IDs

| Level | ID | File |
|---|---|---|
| L0 System | `L0_nnn` | `sdd/system.md` |
| L1 Feature | `L1_<FEATURE>_nnn` | `sdd/features/<name>.md` |
| L2 Detail | `L2_<FEATURE>_nnn` | `sdd/features/<name>.md` |

A feature is one area of the system, with its own file of records. `<FEATURE>` is the code of the
feature, in two to four capital letters. An ID has only capital letters, digits and the `_`
character. `sdd/retired.jsonl` lists the spent IDs that no record uses now.

## Writing a record

Before you add, edit or retire a record, read [writing.md](references/writing.md). It has the
rules for levels, parents, IDs, `tbr` numbers, prose and retirement.
