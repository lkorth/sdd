# Finding the evidence

Read SKILL.md first. This file has the commands and the rules for Steps 1 and 6.

## What each source proves

| Source | It can show | It cannot show |
|---|---|---|
| Code | What the system does now | That anybody wanted it |
| Tests | What somebody chose to assert | Why. Also, that the test still runs |
| Commits and pull requests | Why a change happened, and what the team rejected | That the reason is still true |
| Issues and tickets | The first requirement, and the acceptance criteria | That the code met them |
| Design documents and diagrams | The plan, the constraints and the options | That the team built the plan |
| Comments and READMEs | What somebody claimed | That the claim is current |

Do not accept a requirement from one source alone. Use the code and the tests for what is true now.
Use the other sources for why.

## What you can reach

Find these facts before you start, and write them in your notes:

```bash
git rev-parse --is-shallow-repository       # true means that the history is cut off
git log --oneline | wc -l                   # the size of the history
git log --reverse --format='%h %ad %s' --date=short | head -5
git log --format=%s | grep -cE '\(#[0-9]+\)$|^Merge pull request #[0-9]+'    # commits that name a pull request
git remote -v                               # the host of the pull requests
gh auth status                              # the GitHub tool and its login
```

Then list each tool in this session that can read a document or a tracker. This includes MCP
connectors for Confluence, Jira, Linear, Notion, Google Drive, Slack and Figma, and a tool that
fetches web pages. Some sessions load tools on demand. If yours does, search for these tools by name.

If the history is shallow, ask the user to run `git fetch --unshallow`. It changes only the local
clone, but it can download a large amount of data.

If `gh` is missing, use the tool for your host, for example `glab`, or its API. If you have no tool
for pull requests, ask the user for access. The user can run `! gh auth login`. If the user declines,
use the pull request numbers in the commit subjects and the commit text only.

## Git history

Skip generated files, vendored code and bulk formatting commits.

```bash
# The areas that change most. Many changes point to a unique feature.
git log --format= --name-only --no-merges | sort | uniq -c | sort -rn | head -30

# The life of a file across renames.
git log --follow --format='%h %ad %s' --date=short -- <file>

# The life of one function. If git cannot find the function, use a line range: -L <start>,<end>:<file>
git log -L :<function>:<file> --format='%h %ad %s' --date=short -s

# When a number or a string appeared and changed.
git log -S '<value>' --format='%h %ad %s' --date=short -- <path>

# Who last changed each line, looking past whitespace and moved code.
git blame -w -C -C -C -L <start>,<end> -- <file>

# The fixes and the reverts in an area.
git log -i -E --grep='fix|bug|revert|regression|incident' --format='%h %ad %s' --date=short -- <path>

# The full text and the files of one commit.
git show --stat --format=%B <hash>
```

If `blame` stops at a refactor or a move, run it again on the parent of that commit:
`git blame <hash>^ -- <old-file>`.

Cite only a commit in the current history, because a pull request branch can disappear after a
squash merge. To make sure, run `git merge-base --is-ancestor <hash> HEAD`.

## Pull requests and issues

These commands are for GitHub. Use the same fields on another host. They only read.

```bash
# The pull requests that contain a commit.
gh api repos/{owner}/{repo}/commits/<hash>/pulls --jq '.[].number'

# The merged pull requests about a topic.
gh pr list --state merged --search '<term>' --json number,title --limit 50

# The whole conversation.
gh pr view <n> --json title,body,comments,reviews,commits,closingIssuesReferences

# The review comments on lines of code, which pr view does not include.
gh api --paginate repos/{owner}/{repo}/pulls/<n>/comments --jq '.[] | {path, body}'

# An issue.
gh issue view <n> --json title,body,comments,labels
```

Look for the problem statement, the acceptance criteria and the rejected options. Also look for the
edge cases that a reviewer raised, the incident that caused the change, and the follow-up work.

For a unique behavior, read each pull request that changed the behavior. For a standard feature,
read the first pull request and the fixes. Skip dependency updates and changes by bots.

## Linked content

A commit or a pull request often holds only a link to the real reasoning. Read that content.

Collect the links from each of these places:

1. Commit text: `git log --format=%B -- <path> | grep -oE 'https?://[^][ )>"]+' | sort | uniq -c | sort -rn`
2. Code comments: `git grep -nE 'https?://' -- <path>`
3. The bodies and the comments of pull requests and issues.
4. The documents that you already read.
5. Files in the repository: `git ls-files | grep -iE '\.(md|mdx|rst|adoc|txt|pdf|drawio|puml|plantuml|mmd|dot|svg|png|jpe?g)$'`

Also search the same text for references with no link. Examples are a ticket key such as `PAY-123`,
`ADR-7`, `RFC 12`, or "see the retries design doc". Search for that key or title with the tool for
the system that holds it.

Read each item as this table shows:

| Content | How to read it | What to take from it |
|---|---|---|
| Markdown, ADRs and RFCs in the repository | Read the file. Run `git log --follow` on it to see its age | Decisions, rejected options, constraints |
| Confluence, Notion, Google Docs, SharePoint | The MCP connector or API. Also read the page comments and the page date | Goals, non-goals, numbers, the final decision |
| Jira, Linear and other trackers | The MCP connector or tool | Acceptance criteria, comments, the final state |
| Mermaid, PlantUML, Graphviz and draw.io files | Read the source as text | Parts, flows, states, order |
| Images, PDFs and PR attachments | Open them with a tool that reads images or PDFs | The same, and the expected screens |
| Figma and Miro | The MCP connector, or an exported image | Flows and states |
| Chat threads and incident reviews | The MCP connector | The decision and the failure that caused it |

Save each download in its own new, empty directory outside the repository. Do not run a file that
you download.

If a link inside a document concerns the feature, follow it. If it does not, stop there.

If you cannot read a link, do not try to get past a login or use access that the user did not give.
Note the link and the features that it affects. Ask the user to connect a tool or to paste the
content. Do not guess the content from the title of the link.

## Weighing a document

1. Compare its date with the commits. A document from before the code shows the plan. A document from
   after the code can show what the team learned.
2. Read its status. Draft, rejected and superseded documents show options and not decisions. A
   rejected option can still explain a choice.
3. If the code differs from a document, read the later pull requests. If a later change explains the
   difference, the later decision wins. If not, the behavior is a conflict.
4. If a document promises something that the code does not do, the class is not built.
5. Read a diagram as a claim about the design. Make sure that the code has the same parts and order.

## Reading the signals

| Signal | Meaning |
|---|---|
| A revert | Somebody tried it and rejected it. Find out why. This is strong evidence |
| A fix commit that adds a test | The test states a requirement that the team first missed |
| A changed number | Read the reason. The latest value and its reason go in the rationale |
| A refactor commit | It carries no intent. Look behind it to the earlier commit |
| A vague squash message | Little evidence in the commit. Go to the pull request |
| A long review thread | A real decision. Read to the end, because a later comment can change the decision |
| A workaround comment | A constraint from outside the system. Find out whether the constraint still applies |
| A skipped test | A requirement that somebody gave up on, or a flaky test. Find out which |
