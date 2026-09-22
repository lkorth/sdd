# Foundations

## Backus: framework and changeable parts

John Backus, *"Can Programming Be Liberated from the von Neumann Style?"* ACM Turing Award Lecture, *CACM* 21(8), 1978.

### Framework compared with changeable parts

In §6 of his lecture, Backus divides a language into a fixed structure and parts that users add:

> "Let us distinguish two parts of a programming language. First, its **framework** which
> gives the overall rules of the system, and second, its **changeable parts**, whose
> existence is anticipated by the framework but whose particular behavior is not specified by
> it... Now suppose a language had a small framework which could accommodate a great variety
> of powerful features entirely as changeable parts."

The same division applies to a system of specifications:

| | Programming languages | Specification systems |
|---|---|---|
| Framework | Syntax, statement forms, typing rules | The record shape: fields, permitted values, level rules |
| Changeable parts | Functions, user data types, procedures | The requirements |

The goal is a small framework that lets the requirements say much. Each field in the record
shape must give the requirements a new power. For example, a script can check them, trace them,
or delete them without risk. If a field does not do this, it is only a cost. If a field has the
same value on each record, it is framework that carries no information.

### Stopping the framework from growing

Backus wrote that conventional systems often grow "fat and weak". The framework becomes more
complex, but the language can say no more than before.

This explains why heavy specification tools often do not get used. Some practitioners compared
the Spec Kit of GitHub with plain step-by-step prompting. They reported about 9 to 10 times the
authoring time. In one case, the tool made 2,577 lines of markdown for 689 lines of code. These
reports come from single people, not from controlled studies. But all of them point the same way.

A large framework around prose requirements that no tool can read causes this cost. Fields and
ceremony must let a script check, trace and delete requirements. If they do not, they cost more
than they give.

### Forms that combine requirements

Backus said that conventional languages have no forms that combine parts, and no algebraic
properties. Prose requirements have the same two problems:

1. You cannot build, derive or collect prose statements into larger statements.
2. A script cannot search prose for requirements without a parent, broken references, missing
   tests, or what a deletion affects.

The typed record is a form that combines requirements. Each record has an ID, a parent and typed
references. Together, the records make a directed graph. A script can use this graph to find
records without a parent, broken references, missing tests and duplicate IDs.

### One definition for each rule

Backus warned against rules that are spread across many places.

In many projects, the docs, the prompt templates and the guides for developers each repeat a
rule. Only one tool enforces it. The copies then drift apart.

Give each rule one definition that is the authority, and make each other copy point to it. In
sdd, `records.md` is that definition.

## Naur: the limits of written documentation

Peter Naur, *"Programming as Theory Building."* *Microprocessing and Microprogramming* 15 (1985).

Naur argues that documentation cannot hold all of a program:

> "The **death of a program** happens when the programmer team possessing its theory is
> dissolved. A dead program may continue to be used for execution... The actual state of death
> becomes visible when demands for modifications of the program cannot be **intelligently
> answered**."


> "Programming properly should be regarded as an activity by which the programmers form or
> achieve a certain kind of insight, a **theory**, of the matters at hand."

> "The **program text and its documentation has proved insufficient** as a carrier of some of
> the most important design ideas."

### Why each record has a rationale

For Naur, a theory is the knowledge that lets an engineer explain and justify each action.

This is the function of the rationale. The `SHALL` statement says what the system does. The
rationale says why the requirement exists, and why each number has its value. Without a
rationale, a team can build a requirement but cannot change it with understanding. That is the
test that Naur gives for a dead program.

### What this means for work with agents

Naur saw that new people learn the theory of a program when they work with the people who
already maintain it.

An agent keeps no shared theory from one session to the next. Each session starts with nothing.
This gives two rules:

1. Each record must have a rationale. Nothing carries the reasons from one session to the next,
   so the repository must hold them in writing.
2. A specification states intent, not the full theory of the program. A specification cannot
   foresee each detail. The specification is the authority on intent. The code is the authority
   on behavior.

## BNF: syntax in a formal notation, meaning in prose

*ALGOL 60 Report* (Peter Naur, ed., 1960).

Backus-Naur Form (BNF) put the syntax of a programming language in a short formal notation with
a closed set of symbols. It left the meaning of each construct to prose.

The typed record uses the same pattern. A closed grammar controls the IDs, the parents and the
statuses. The obligation stays in prose that people can read, in a `SHALL` statement.

BNF did not remove prose from the definition of meaning. Also, implementations found unusual
cases, and the first versions needed changes. In the same way, a script can check the structure
of the records, and the prose describes the behavior.
