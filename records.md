# Records

A requirement is a record. A record has a heading with its ID, a block of fields, and prose.
Text that does not have this shape is commentary, and it does not bind anything.

````markdown
### L2_PER_004: Acknowledgement is durable before deletion

```yaml
parent: L1_LNK_002
```

The system SHALL write an upload acknowledgement to durable storage before deleting the
acknowledged record.

Rationale: If the system deletes a record before its acknowledgement is durable, a power loss
can lose the record. The upload existed to move exactly that data, and nothing reports the loss.
The order of the write and the delete is the full mechanism. No retry can recover a record that
the system deleted.

Scenario: Power is removed between acknowledgement and deletion
- WHEN power is removed after an acknowledgement is received and before the record is deleted
- THEN the record is present on restart and is not uploaded a second time
````

| Piece | Form | Contents |
|---|---|---|
| Heading | `### <ID>: <title>` | The ID, and a title for people |
| Fields | A `yaml` block of `key: value` lines | `parent`, and as an option, `status` |
| Obligation | The first paragraph after the fields block | The one `SHALL` |
| Rationale | From the line that starts with `Rationale:` to the first `Scenario:` line. It can have more than one paragraph | Why the record exists, and where each number came from |
| Scenarios | One or more. Each is a `Scenario: <name>` line, then `- WHEN`, `- THEN` and `- AND` bullets | What the tests assert |

The pieces come in the order of the table. A record runs from its heading to the next heading of
any level, or to the end of the file. Put commentary under a heading that is not a record.

A record has no other markup. Keywords are in capitals. Labels are plain words at the start of
a line. Do not put prose in the fields block.

The records describe one system, and the system has features. A feature is one area of the
system, with its own file of records. For example, the `LNK` feature is the link that moves data,
with transport and session resumption.

## Fields

A record has two fields, `parent` and `status`, and no others.

### The parent field

Each record must have a `parent`. The parent is the requirement that the record exists to
serve. The level of a record sets which parents it can have:

| Level | Permitted parent |
|---|---|
| `L0` | `self`. L0 is the root. Give the reason for the record in its rationale |
| `L1` | An `L0` record, or an `L1` record in the same feature |
| `L2` | An `L0` or `L1` record |

A parent can be in a different feature. Do not write a record that only repeats its parent with
the name of a feature added. Skip a level instead.

Use the orphan test to find the correct parent. Imagine that you delete the parent. If you must
then also delete this record, the record is a child. If the record is still necessary, it is not
a child. It is a sibling with a dependency. Give it the same parent as the record it depends on,
and name that record in its text or its rationale. If a record has no true parent, the project
does not need it. Delete the record. Do not invent a parent.

### The status field

The `status` field is optional. A record without `status` is active. The only other value is
`tbr`, which means "to be resolved". A `tbr` record is a real requirement, but a number in it is
not chosen yet. Write a placeholder, for example `TBR_1`, where the number goes. In the
rationale, write who will choose the number and by which date.

To resolve a `tbr` record, follow these steps:

1. Put the number in place of the placeholder.
2. Remove `status: tbr`.
3. In the rationale, write where the number came from.

## IDs and levels

| Level | Question that it answers | ID | File |
|---|---|---|---|
| L0 System | What is the system for, and what is always true of it? | `L0_nnn` | `sdd/system.md` |
| L1 Feature | What does a feature do for its users? | `L1_<FEATURE>_nnn` | `sdd/features/<name>.md` |
| L2 Detail | How does a feature behave, exactly? | `L2_<FEATURE>_nnn` | `sdd/features/<name>.md` |

Write each requirement at the highest level that does not name an implementation. Use these
questions in this order:

1. If you can write it without the name of a feature, it is L0. A constraint on all features is
   L0, for example "no feature reads the wall clock".
2. If it names a feature but no internals of the feature, it is L1. A promise that one feature
   makes to another feature is an L1 record of the feature that makes the promise.
3. If it names an endpoint, a state, a field, or a number that has meaning only inside the
   feature, it is L2.

A feature can have no L1 records. Then its L2 records name an L0 record as their parent.

`<FEATURE>` is the code of the feature, in two to four capital letters. The file of the feature
declares it. To give a new record its number, find the highest number for its prefix in the
records and in the ledger. Add one. The command `sdd next L2 LNK` does this for you.

An ID has only capital letters, digits and the `_` character. A double click selects the whole ID, and
the ID is a legal part of a function name in most languages.

After an ID lands on the main branch, the ID is spent. Before that, nothing outside your branch
cites the ID, so you can renumber it. After that, do not use the ID again. Tests, rationales and
commits cite IDs. If you use an ID again, each old citation points at the wrong requirement, and
nothing tells you.

If a spent ID leaves the records, [retire it](#retiring-an-id). An ID leaves the records in two
cases. You delete its record, or you give the record a new ID, for example in a feature rename.

### Citing a record in a test

Each scenario maps to a test. The name of that test contains the ID of the record, with no
change, for example `test_L2_PER_004_ack_is_durable_before_delete`. A test can cite more than
one record. Also name the ID in the commit that adds or changes the test.

## Files

```
sdd/
  system.md          L0 records
  features/<name>.md One file per feature: L1 records, then L2 records
  retired.jsonl      IDs that must not be used again
```

A feature file starts with frontmatter that gives the code of the feature. Then come its L1
records, and then its L2 records:

```markdown
---
feature: LNK
---

# Link

Transport and session resumption.

## L1

### L1_LNK_002: An interrupted upload completes without duplicate delivery
...

## L2

### L2_LNK_002: An interrupted upload resumes after the last acknowledgement
...
```

Use `###` headings only for records. A `###` heading that is not a record looks like a
requirement to a reader.

## Writing in simple English

Write all prose in a record in simple English. The rules come from ASD-STE100 Simplified
Technical English, the controlled language that aerospace uses for maintenance manuals. Its goal
is text with one meaning for each reader. That is also the goal of a requirement.

1. Write one obligation and one `SHALL` in each record. Two obligations are two records. Then
   you can test and retire each one by itself.
2. Use `SHALL` as the only modal verb that binds. Do not use `must`, `should`, `may`, `might`,
   `could` or `would` in a record, in capitals or in lowercase. Each of these words makes the
   reader decide how strong it is. Use `will` only to state a fact, and `can` only to state a
   capability.
3. Name the actor, and use the active voice: "the store rejects", not "is rejected". If the
   actor does not matter, as in an event such as "power is removed", you can leave it out.
4. Put the condition first: "If the link is down, the system SHALL ...". A scenario already does
   this, because `WHEN` comes before `THEN`.
5. Write sentences of 25 words or fewer, and use simple tenses. If the `SHALL` sentence becomes
   long, define its terms in short sentences before it, in the same paragraph. Those sentences
   bind, because they are in the obligation. Do not move a condition into the rationale,
   because the rationale does not bind.
6. Use one word for one meaning in the whole specification. Define each term at its first use.
   If many records use a term, define it once at L0, and then use it with no change.
7. Do not use semicolons or dashes to join clauses. Write two sentences.
8. Write numbers, not adjectives. Words such as `fast`, `reliable`, `robust`, `safe`,
   `sufficient` or `as needed` each stand for a number that nobody chose. Choose the number, or
   mark the record `tbr`.
9. Write what the system does, not how it does it. If the implementation can change and nobody
   sees a difference, the requirement does not include the implementation.
10. Make each scenario assert only what the `SHALL` states. If a `THEN` asserts more, a
    requirement is missing. Write that requirement.

These rules apply to records. The docs follow them too, with one difference: a doc uses "must"
to tell the reader what to do. In a record, `SHALL` binds the system. In a doc, "must" instructs
the person who writes the record.

### The Simple English plugin

Install [Simple English](https://github.com/AminBlg/SimpleEnglish) in the agent that writes your
specs. It teaches the agent the ASD-STE100 rules above. It checks each markdown file that
the agent writes, and gives feedback to the agent.

The Simple English linter reads the `Scenario:` line and the `WHEN` line below it as one
sentence. It then reports a condition at the end of a sentence. The record format puts the
condition first, so ignore that hit. It also cannot know that a precise `SHALL` sentence is
more important than a short one. If you cannot make a `SHALL` sentence short with writing rule
5, keep it precise.

## Retiring an ID

To delete a record, run `sdd delete <ID> --reason "<why>"`. It removes the record from its file
and adds the ledger line, with the date of today. If another record names the ID as its parent,
in its obligation or in a scenario, the command changes nothing and lists those records. Change
or delete them first. Then change or delete each test that cites the ID.

In the other case, you give a record a new ID. Retire each spent ID that leaves the records.
Add one line for the ID to `sdd/retired.jsonl`:

```json
{"id": "L2_PER_003", "reason": "Replaced by a shorter timeout"}
{"id": "L1_LNK_002", "reason": "Renamed to L1_NET_002"}
```

The `id` field is necessary. The `reason`, `date` and `pr` fields are optional
([schema](schemas/retired.schema.json)). If you give a record a new ID, change each test and
record that cites the old ID, in the same pull request. Git keeps the text of the old requirement. To find it,
run `git log -S 'L2_PER_003'`. The ledger keeps the one fact that must stay after the record is
gone: the ID is spent.

A live record can name a retired ID in its rationale, for example "replaces L2_PER_003". That is
history. A live record cannot name a retired ID as its parent, in its `SHALL`, or in a scenario.
