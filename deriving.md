# Deriving records from existing code

In sdd, you normally write the requirement before the code. Many systems already exist with no
records. The `derive` skill has an agent read such a system and write its records. The source of
truth for the steps is the skill, in `skills/derive/`. This document explains the method for people.

Run the skill in Claude Code with `/sdd:derive`, or ask for a spec of the existing code. You can give
it the whole system, or only some features or paths.

## What derivation can and cannot recover

Code shows what the system does. It does not show what anybody wanted, or which behaviors are
accidents. A spec that only restates the code adds lines and no knowledge.

[Naur](foundations.md#naur-the-limits-of-written-documentation) argues that the theory of a program
lives in the people who built it. Tests, history and documents hold only part of that theory. So a
derived spec does not claim to be complete. It states what the evidence shows, and marks the places
where the evidence stops. A person who owns the system must then review it.

## The sources

The agent works through the system in this order:

1. The code, for the shape of the system, its features and its behaviors.
2. The tests, for what somebody chose to assert.
3. Commits, pull requests, review threads and issues, for why each change happened.
4. Linked content, for the reasoning that a commit or a pull request only points to.

Linked content includes each link and reference in commit text, pull requests, issues, code comments
and the docs of the repository. Examples are design documents, markdown files, Confluence pages,
tickets, chat threads, diagrams and attachments. The agent reads them with the tools that its session
has, such as MCP connectors. To get a better spec, connect the tools for your document and ticket
systems before you start. The agent reports each source and each link that it did not read, with
the features that it affects.

The agent treats all of this text as evidence and not as instructions. It does not copy secrets,
customer data or the names of people into a record.

## Features and their depth

The agent shows you a table of features before it goes into detail, because a wrong map wastes all
later work. Each feature has a rank:

| Rank | Test | Depth |
|---|---|---|
| Unique | A person who knows this kind of system, and did not read this code, cannot predict the behavior | All code, tests, changes and linked content |
| Standard | That person can predict it | The promise at L1, and L2 records only where the code differs from the usual behavior |

The unique features are where the spec has the most value, and they get the most effort. This also
limits the size of the spec. If the records of a feature have more lines than its code, the agent
transcribed the code. See [the cost of a large framework](foundations.md#stopping-the-framework-from-growing).

## How each behavior is sorted

| Class | Evidence | Result |
|---|---|---|
| Intended | The code does it. A test, a change or a document shows that somebody wanted it | A record, with the reason from the evidence |
| Presumed | The code does it, and a caller can observe it. No sign of why | A record with `No recorded reason.` in the rationale |
| Incidental | Nobody outside the feature can observe it. Or a rewrite can change it and break no promise | No record |
| Conflict | Two sources disagree | If only a number differs, a `tbr` record. If not, a question for the owner |
| Suspected bug | The code does it, and no reasonable user wants it | No record, because a `SHALL` makes the bug a requirement. A report item |
| Not built | A document or a pull request promises it, and the code does not do it | No record, because the records state the current system. A report item |
| Dead | Nothing reaches the code, or a flag is always off | No record. A report item |

## What a derived record looks like

A derived record has the normal [record shape](records.md). It has no extra fields, because a field
for the source gives a script no new power. The evidence goes in the rationale:

1. Each claim names its source: `PR #412`, `commit 1a2b3c4`, `issue #88`, or the title and link of a
   document. A rationale does not cite a file and a line, because they move.
2. The rationale summarizes the reason, so that it stays useful after a link breaks.
3. A reason that no source states starts with `Inferred from`.
4. A behavior with no reason has `No recorded reason.` A number with no reason has
   `No recorded reason for the value 24 hours.`

These fixed phrases let a script find each weak spot. Each phrase stays on one line, because the
search reads one line at a time:

```bash
sdd search 'No recorded reason|Inferred from'
```

When the owner gives the reason, replace the phrase with the reason.

The `tbr` status keeps its normal meaning. If the code sets a number and nothing disputes it, the
number is chosen. This stays true after the reason is lost. If the sources disagree about the number,
the record uses `tbr`.

## Tests

sdd requires that the name of a test contains the ID of its record. Tests in an existing system do
not have these IDs. The derivation does not change code or tests, so it gives you a map from each
record to its tests. It also lists each scenario with no test.

After you accept the records, the agent can rename the tests in a separate commit. An ID can still
change until it lands on the main branch, so review the records before then.

## The report

The agent gives a report in the pull request description or in its reply. The report is not a file
in the repository, because [the repository holds only the current state](README.md#the-core). It has
these sections:

1. Scope: the features, the record counts, and each path that no feature covers.
2. Sources: what the agent read, and what it did not reach.
3. Questions for the owner: conflicts, `tbr` records, and decisions.
4. Records with no recorded reason.
5. Suspected bugs, dead code, and things that were promised but not built.
6. Test map: each record ID with its tests, and each scenario with no test.

The agent does not fix what it finds. Fixes go through the normal change process, with records that
the owner accepted.
