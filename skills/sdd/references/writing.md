# Writing sdd records

Read SKILL.md first for the grammar and the fields.

## Choosing the level

Write each requirement at the highest level that does not name an implementation:

1. If the record names no feature, it is L0. A constraint on all features is L0.
2. If it names a feature but no internals of the feature, it is L1. A promise from one feature
   to another feature is an L1 record of the feature that makes the promise.
3. If it names an endpoint, a state, a field, or a number that has meaning only inside the
   feature, it is L2.

## Choosing the parent

| Level | Permitted parent |
|---|---|
| L0 | `self`. Give the reason for the record in its rationale |
| L1 | An L0 record, or an L1 record in the same feature |
| L2 | An L0 or L1 record |

A parent can be in a different feature. A linter can make sure that the parent exists and that
its level is permitted. It cannot make sure that the parent is the correct one. Use this test:

1. Imagine that you delete the parent.
2. If you must then also delete this record, the parent is correct.
3. If this record is still necessary, it is not a child. It is a sibling with a dependency.
   Give it the same parent as the record that it depends on. Name that record in its text or
   its rationale.
4. If no record passes the test, the project does not need this record. Delete it. Do not
   invent a parent.

Do not write a record that only repeats its parent with the name of a feature added. Skip a level
instead.

## Giving the ID

Run `sdd next <LEVEL> <FEATURE>`, for example `sdd next L2 LNK`. It finds the highest number for
the prefix in the records and in `sdd/retired.jsonl`, and adds one.
Before the ID lands on the main branch, you can renumber it. After that, the ID is spent. Do not
use it again. Tests, rationales and commits cite IDs. A reused ID makes each old citation point
at the wrong requirement.

If a spent ID leaves the records, retire it. An ID leaves the records in two
cases. You delete its record, or you give the record a new ID, for example in a feature rename.

A feature file declares its `<FEATURE>` code in the `feature:` key of its frontmatter. Put L1
records under `## L1` and L2 records under `## L2`.

## Numbers that are not chosen

If nobody chose a number yet, mark the record `status: tbr`. Write a placeholder such as `TBR_1`
where the number goes. In the rationale, write who will choose the number and by which date.

To resolve the record, follow these steps:

1. Put the number in place of the placeholder.
2. Remove `status: tbr`.
3. In the rationale, write where the number came from.

## Prose

Write records in simple English. sdd depends on the Simple English skill for the general
rules. These rules are special to records:

1. Write one obligation and one `SHALL` in each record. Two obligations are two records.
2. Use `SHALL` as the only modal verb that binds. In a record, do not use `must`, `should`,
   `may`, `might`, `could` or `would`, in capitals or in lowercase. Use `will` only for a fact
   and `can` only for a capability.
3. Name the actor and use the active voice: "the store rejects", not "is rejected". For an
   event such as "power is removed", you can leave out the actor.
4. Put the condition first: "If the link is down, the system SHALL ...".
5. Write sentences of 25 words or fewer, in simple tenses. If the `SHALL` sentence is long,
   define its terms in short sentences before it, in the same paragraph. Do not move a
   condition into the rationale, because the rationale does not bind. If precision and length
   conflict, keep the precision.
6. Use one word for one meaning in the whole specification. Define each term at its first use.
   Define a term that many records use once, in an L0 record.
7. Do not join clauses with semicolons or dashes. Write two sentences.
8. Write numbers, not adjectives such as `fast`, `reliable`, `robust`, `safe`, `sufficient` or
   `as needed`. Each of them stands for a number that nobody chose. Choose it, or mark the
   record `tbr`.
9. Write what the system does, not how it does it.
10. Make each scenario assert only what the `SHALL` states. If a `THEN` asserts more, a
    requirement is missing. Write that requirement.

The Simple English linter reads a `Scenario:` line and its `WHEN` line as one sentence, and
reports a condition at the end. Ignore that hit.

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

The `id` key is necessary. The keys `reason`, `date` (`YYYY-MM-DD`) and `pr` (an integer) are
optional. If you give a record a new ID, change each test and record that cites the old ID, in
the same pull request. Git keeps the text of the old record. To find it, run `git log -S 'L2_PER_003'`.

A live record can name a retired ID in its rationale, for example "replaces L2_PER_003". It
cannot name a retired ID as its parent, in its `SHALL`, or in a scenario.
