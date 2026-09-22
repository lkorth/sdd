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
| `foundations.md` | No copy. It explains why sdd exists, and an agent does not need it to read or write a record |

After you change the skill, follow these steps:

1. Increase `version` in `.claude-plugin/plugin.json` and in `.claude-plugin/marketplace.json`.
   Use the same value in both files.
2. Run `claude plugin validate .` and fix each error.

## Writing

Write all text in this repository in simple English. Use the Simple English skill.
