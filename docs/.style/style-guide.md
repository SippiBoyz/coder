# Coder documentation style guide

This is the canonical prose style guide for the Coder documentation.
It tells you how to write the words. For decisions about what belongs
in the docs and what doesn't, see
[`content-guidelines.md`](./content-guidelines.md).

Each rule below is a policy decision the Coder docs team has made.
Where a Vale rule already enforces the policy, the rule name is
listed in a parenthetical so you can reproduce the warning locally.
Where the rule is documentation-only (not yet wired into Vale), the
parenthetical says so. The doctrine for adding Vale rules lives in
[`README.md`](./README.md).

## How to use this guide

- **Contributors**: read the section that matches what you are
  writing. Each rule includes a brief rationale and **Do** / **Don't**
  examples.
- **Reviewers**: cite the section in a review comment. Reviews are
  easier when the guidance is in one place.
- **AI agents**: read this page in full before editing anything under
  `docs/`. The Coder Agents and Claude Code guides
  ([`AGENTS.md`](../../AGENTS.md),
  [`.claude/docs/DOCS_STYLE_GUIDE.md`](../../.claude/docs/DOCS_STYLE_GUIDE.md))
  link here.

## Quick reference

| Area                                                              | Key rules                                                                                                                                      |
|-------------------------------------------------------------------|------------------------------------------------------------------------------------------------------------------------------------------------|
| [Voice and tone](#voice-and-tone)                                 | Second person; no first-person singular; "we" only for Coder; active voice; present tense                                                      |
| [Word choice](#word-choice)                                       | Canonical brand and product names; inclusive language; avoid weasel words; product prose prefers `stop` over `kill`, `turn off` over `disable` |
| [Capitalization and punctuation](#capitalization-and-punctuation) | Sentence-case headings; no gerund leads; no em-dashes; Oxford comma; US-style quotation                                                        |
| [Formatting](#formatting)                                         | Bold for UI; code font for identifiers; explicit language fence on every code block; descriptive link text                                     |
| [Numbers, units, and dates](#numbers-units-and-dates)             | Digits everywhere; non-breaking space between number and unit; `Month Day, Year` dates; 12-hour time with AM/PM                                |

## Voice and tone

### Address the reader directly

Use the second person ("you") in prose that gives the reader an
instruction or describes what the reader sees, types, or gets back.
Second person is direct, scales across audiences, and avoids the
ambiguity of "the user" (which user?) or generic constructions.

**Do**:

> You can connect to a workspace over SSH after you have installed
> the Coder CLI.

**Don't**:

> Users can connect to workspaces over SSH after the user has
> installed the Coder CLI.

*Documentation-only. No Vale rule.*

### Avoid first-person singular

First-person singular pronouns (`I`, `my`, `me`, `mine`, `I'm`,
`I've`) imply a single author speaking to a single reader, which is
the wrong register for product documentation. Rewrite in the second
person or in a neutral voice.

**Do**:

> You can configure the workspace timeout in the template settings.

**Don't**:

> I usually set the workspace timeout in the template settings.

*Enforced by `Coder.FirstPersonSingular`.*

### Reserve first-person plural for Coder

`We`, `us`, and `our` refer to Coder Technologies, the company.
Don't use them to mean "you and the docs together" or "you and the
author together". That construction tends to obscure who is taking
the action.

**Do**:

> We ship a new agent binary in each release.
>
> You can install the agent with the workspace template.

**Don't**:

> We can install the agent by running this command on our workspace.

*Enforced by `Coder.FirstPersonPlural`.*

### Active voice by default

Active voice puts the actor first and reads faster. Passive voice
is acceptable when the actor is genuinely unknown or irrelevant
(`The token is rotated every 24 hours.`), but the default is active.

**Do**:

> Coder rotates the agent token every 24 hours.

**Don't**:

> The agent token is rotated every 24 hours by Coder.

*Documentation-only. No Vale rule. Imprecise rules like
`Google.Passive` and `write-good.Passive` fire on every passive
construction including the legitimate ones, so they are excluded by
the DOCS-425 doctrine.*

### Present tense by default

Describe how the product works in the present tense. Future tense
("will") implies an event that hasn't happened yet at read time;
reserve it for genuine future events, like scheduled rollouts or
deprecations with a known date.

**Do**:

> The provisioner reads the template files and creates the
> workspace.

**Don't**:

> The provisioner will read the template files and will create the
> workspace.

*Documentation-only. No Vale rule.*

### Inclusive pronouns

Use the singular `they` when the subject's gender is unknown or
irrelevant. Avoid `he or she`, `(s)he`, and similar constructions.

**Do**:

> When a user opens a workspace, they connect to the agent over a
> Tailscale tunnel.

**Don't**:

> When a user opens a workspace, he or she connects to the agent
> over a Tailscale tunnel.

*Enforced by `Google.Gender` and `Google.GenderBias`.*

## Word choice

### Brand names

Use the canonical casing for third-party brand and product names.
The Coder docs team keeps a substitution list. The table below
reflects current canonical forms.

| Do         | Don't                                 |
|------------|---------------------------------------|
| HashiCorp  | Hashicorp, HASHICORP                  |
| GitHub     | Github, GITHUB                        |
| OpenTofu   | Opentofu, OpenTOFU                    |
| Kubernetes | kubernetes (in prose), K8s (in prose) |
| Terraform  | terraform (in prose)                  |
| JetBrains  | Jetbrains, jetbrains                  |
| VS Code    | VSCode, VSC, VS code                  |

Lowercase forms remain correct in code blocks, URLs, package names,
and Terraform provider sources, where the canonical form is
lowercase by convention.

*Enforced by `Coder.BrandNames`.*

### Coder product and feature names

Coder, the company and the product, is always capitalized.
Feature names are capitalized as proper nouns when referring to the
named feature. The underlying generic concept stays lowercase.

| Do              | Don't                                               |
|-----------------|-----------------------------------------------------|
| Coder           | coder (when referring to the product)               |
| AI Bridge       | AI bridge, AIBridge                                 |
| Workspace Proxy | workspace proxy (when referring to the feature)     |
| workspace       | Workspace (when referring to the generic concept)   |
| template        | Template (when referring to the generic concept)    |
| agent           | Agent (when referring to the generic concept)       |
| provisioner     | Provisioner (when referring to the generic concept) |

*Enforced by `Coder.ProductTerms` (planned).*

### Dev Container terminology

A development container that follows the
[Dev Container specification](https://containers.dev/) is a
**Dev Container** (capitalized) when referring to the named concept,
and **dev container** mid-sentence. `envbuilder` is the
implementation tool Coder uses to build dev containers. It is not
itself the concept.

**Do**:

> The template uses a Dev Container defined in a `devcontainer.json`
> file.
>
> Coder builds dev containers with `envbuilder`.

**Don't**:

> The template uses a devcontainer defined in a `devcontainer.json`
> file.
>
> Coder builds DevContainers with envbuilder.

*Enforced by `Coder.DevContainer` (planned).*

### Set up versus setup, Quickstart

`Set up` is the verb. `Setup` is the noun. `Quickstart` is one
word, always.

**Do**:

> Follow the Quickstart to set up your first workspace.
>
> The setup takes about 10 minutes.

**Don't**:

> Follow the Quick Start to setup your first workspace.
>
> The set-up takes about 10 minutes.

*Enforced by `Coder.SetupSetUp` (planned).*

### Learn more, not Next steps

End-of-page navigation that points the reader at related material
uses the heading **Learn more**, not **Next steps**. "Next steps"
implies the reader must follow a specific sequence; "Learn more"
frames the section as optional related reading, which matches the
Diátaxis distinction between a tutorial (sequenced) and a how-to or
reference (independent).

**Do**:

```markdown
## Learn more

- [Configure SSH access](./ssh.md)
- [Set workspace autostart](./autostart.md)
```

**Don't**:

```markdown
## Next steps

- [Configure SSH access](./ssh.md)
- [Set workspace autostart](./autostart.md)
```

*Enforced by `Coder.LearnMore` (planned).*

### Tutorial, not walkthrough

`Tutorial` is the standard term in technical documentation and
matches the Diátaxis category. `Walkthrough` is colloquial.

**Do**:

> This tutorial shows you how to deploy Coder on AWS.

**Don't**:

> This walkthrough shows you how to deploy Coder on AWS.

*Enforced by `Coder.Tutorial` (planned).*

### Inclusive-language substitutions

Use the industry-standard inclusive substitutions for terms that
have transitioned across the broader developer-tooling ecosystem.

| Do                      | Don't                                           |
|-------------------------|-------------------------------------------------|
| allowlist               | whitelist                                       |
| blocklist, denylist     | blacklist                                       |
| primary, main           | master (for the primary branch or controller)   |
| primary, hub, reference | master (general usage)                          |
| replica, secondary      | slave                                           |
| placeholder, dummy      | sanity check (alone), dummy (in pejorative use) |

*Enforced by `Coder.InclusiveLanguage` (planned), with additional
coverage from the curated `alex.*` lexicon.*

### Avoid weasel words

Words that minimize the difficulty of an action ("simply", "just",
"easy", "easily", "obviously", "of course", "clearly") assume the
reader's experience matches the author's. If something is "obvious"
to the author and not to the reader, the reader is left feeling
confused or condescended to. Cut the weasel word or restructure the
sentence.

**Do**:

> Run `coder login` to authenticate.

**Don't**:

> Simply run `coder login` to authenticate. It's easy!

*Enforced by `Coder.WeaselWords` (planned).*

### Stop, not kill; turn off, not disable

In product-facing prose, prefer `stop` over `kill` and `turn off`
over `disable`. The plain-language forms read better for a
non-technical audience and don't carry violent or ableist
connotations. Technical states (a `disabled` feature flag, a
`killed` process in a log file) keep their original term because
they reference a real state name.

**Do**:

> To stop a workspace, click **Stop** in the workspace dashboard.
>
> You can turn off auto-update in the template settings.

**Don't**:

> To kill a workspace, click **Kill** in the workspace dashboard.
>
> You can disable auto-update in the template settings.

*Enforced by `Coder.PlainLanguage` (planned), with the technical-state
exception scoped in the rule.*

## Capitalization and punctuation

### Sentence-case headings

Capitalize the first word of a heading or page title, plus any
proper nouns. Everything else is lowercase. This rule covers H1
through H6 and matches the way the heading reads aloud.

**Do**:

```markdown
# Configure your workspace
## Set up SSH access
### Connect through JetBrains Toolbox
```

**Don't**:

```markdown
# Configure Your Workspace
## Set Up SSH Access
### Connect Through JetBrains Toolbox
```

*Enforced by `Google.Headings` (scope adjusted to skip CLI flag
fragments and acronyms).*

### No gerund-leading headings

Don't start a heading with a present participle or gerund (an
`-ing` word acting as a noun or verb form). The imperative form
reads better for task headings. The noun form reads better for
concept headings. Reserve gerund-leading headings for the rare case
where neither alternative reads cleanly.

**Do**:

```markdown
## Install Coder
## Installation
## Configure your workspace
## Configuration reference
```

**Don't**:

```markdown
## Installing Coder
## Configuring your workspace
```

*Enforced by `Coder.GerundHeading`.*

### No trailing punctuation in headings

Headings are labels, not sentences. Drop terminal periods, question
marks, and exclamation points.

**Do**:

```markdown
## What is a workspace
## Quick reference
```

**Don't**:

```markdown
## What is a workspace?
## Quick reference!
```

*Enforced by `Google.HeadingPunctuation`.*

### No em-dashes or en-dashes

Em-dashes (&mdash;, U+2014), en-dashes (&ndash;, U+2013), and the
ASCII `--` fallback are banned in prose. Use a comma, semicolon, or
period, or restructure the sentence.

**Do**:

> The provisioner reads the template, runs `terraform plan`, and
> creates the workspace.

**Don't**:

> The provisioner reads the template&mdash;runs `terraform plan`&mdash;and
> creates the workspace.
>
> The provisioner reads the template -- runs `terraform plan` --
> and creates the workspace.

*Enforced by `scripts/check_emdash.sh` (existing CI script) and
`Coder.EmDash` (planned).*

### Oxford comma

Use a comma before the conjunction in a list of three or more
items.

**Do**:

> The provisioner builds, configures, and starts the workspace.

**Don't**:

> The provisioner builds, configures and starts the workspace.

*Enforced by `Google.OxfordComma`.*

### US-style quotation

Place commas and periods inside closing quotation marks. Semicolons
and colons stay outside. This is the United States convention and
matches the dominant style of the surrounding tech-docs ecosystem.

**Do**:

> The error message reads, "workspace not found."

**Don't**:

> The error message reads, "workspace not found".

*Enforced by `Google.Quotes`.*

### Semicolons sparingly

Prefer two sentences. A semicolon joins two complete thoughts when
they are tightly related and a period would lose the connection,
but in technical prose two sentences almost always read more
clearly.

**Do**:

> The provisioner uses Terraform. It reads the template files and
> creates the workspace.

**Don't**:

> The provisioner uses Terraform; it reads the template files and
> creates the workspace.

*Documentation-only. No Vale rule.*

### Exclamation points rare in prose

Exclamation points in body prose read as marketing copy or shouted
emphasis. Reserve them for code blocks, direct quotes from error
messages, and rare moments where genuine emphasis serves the
reader.

**Do**:

> Coder is ready to use.

**Don't**:

> Coder is ready to use!

*Enforced by `Google.Exclamation`.*

### Numeric ranges

Spell out the joiner in prose. Use `5 to 10` or `between 5 and 10`,
not `5-10`. In code blocks, terse reference material, and tables
where space matters, the hyphenated form is acceptable.

**Do**:

> The agent retries 5 to 10 times before giving up.

**Don't**:

> The agent retries 5-10 times before giving up.

*Enforced by `Google.Ranges`.*

## Formatting

### Bold for UI elements

Use bold for the literal text of UI elements the reader interacts
with: buttons, menu items, page titles, field labels, tab names.
Bold tells the reader "this is the thing you click or read".

**Do**:

> Navigate to **Templates** > **Settings** and select the
> **Schedule** tab.

**Don't**:

> Navigate to "Templates" > "Settings" and select the Schedule tab.
>
> Navigate to *Templates* > *Settings* and select the *Schedule*
> tab.

*Documentation-only. No Vale rule.*

### Italics for emphasis only

Reserve italics for genuine emphasis where bold would be too loud.
Don't use italics for UI elements, identifiers, or product names.

**Do**:

> Restarting the workspace deletes ephemeral state. Save your work
> *before* you click **Restart**.

**Don't**:

> Navigate to *Templates* > *Settings*.

*Documentation-only. No Vale rule.*

### Code font

Use backticks (inline code font) for: user input, command names,
flag names, filenames, file paths, environment variables, HTTP
verbs and status codes, configuration keys, code identifiers
(function names, struct names, package names), and placeholder
variables.

**Do**:

> Run `coder login --token <token>` to authenticate. Set
> `CODER_URL` in your environment first.
>
> The server returns `404 Not Found` when the workspace does not
> exist.

**Don't**:

> Run "coder login --token \<token\>" to authenticate. Set CODER_URL
> in your environment first.
>
> The server returns 404 when the workspace does not exist.

*Documentation-only. No Vale rule.*

### Code blocks with language fences

Every fenced code block declares a language. Use the most specific
language tag available: `sh` for shell, `tf` for Terraform, `yaml`
for YAML, `go` for Go, `json` for JSON, `text` for plain text with
no syntax to highlight.

**Do**:

````markdown
```sh
coder login --token <token>
```
````

**Don't**:

````markdown
```
coder login --token <token>
```
````

*Enforced by `markdownlint` rule `MD040`.*

### Callouts

Use the GitHub callout syntax for asides. Use them sparingly;
prose should carry the message.

| Callout          | Use for                                                                |
|------------------|------------------------------------------------------------------------|
| `> [!NOTE]`      | Additional context the reader benefits from but doesn't need to act on |
| `> [!TIP]`       | Optional optimization or shortcut                                      |
| `> [!WARNING]`   | Action that may cause data loss, downtime, or security exposure        |
| `> [!IMPORTANT]` | Required action the reader will miss otherwise                         |
| `> [!CAUTION]`   | Reserved for severe consequences; use rarely                           |

*Documentation-only. No Vale rule.*

### Tabs for parallel content

Use tabs when the reader picks one path that applies to their
situation: installation methods on different operating systems,
platform-specific commands, API client SDKs in different languages.
Don't use tabs to hide information the reader needs regardless of
choice.

*Documentation-only. No Vale rule.*

### Lists

Unordered lists are for items that have no required order. Ordered
lists are for sequential steps the reader follows in order. Steps
in an ordered list start with an imperative verb.

**Do**:

```markdown
1. Run `coder login` to authenticate.
2. Create the workspace template.
3. Build the workspace from the template.
```

**Don't**:

```markdown
1. The user runs `coder login` to authenticate.
2. Creating the workspace template comes next.
3. Then the workspace gets built from the template.
```

*Documentation-only. No Vale rule.*

### Tables

Use tables to compare options, list parameters, or show
permissions. Keep tables simple: avoid nested formatting and avoid
tables that would read better as prose.

*Documentation-only. No Vale rule.*

### Links

Link text describes what the reader gets at the destination.
Generic phrases like "click here" and "this link" tell the reader
nothing if they scan the link out of context. Screen readers
announce link text out of context too.

**Do**:

> See [the Coder CLI reference](../reference/cli/index.md) for the
> full command list.

**Don't**:

> See the Coder CLI reference [here](../reference/cli/index.md).
>
> [Click here](../reference/cli/index.md) for the full command list.

*Enforced by `Coder.LinkText` (planned).*

### Images

Every image declares descriptive alt text. The alt text describes
what the image shows or what purpose it serves; it is not a
caption. Captions go below the image in a `<small>` tag.

```markdown
![Template Insights dashboard showing weekly active users and
connection latency](../images/admin/templates/template-insights.png)

<small>The Template Insights dashboard. Active users in the left
panel; connection latency in the right panel.</small>
```

Screenshot policy lives in
[`content-guidelines.md`](./content-guidelines.md): include a
screenshot only when the topic would be confusing without one, and
capture the minimum surface area.

*Enforced by `markdownlint` for the alt-text requirement.*

## Numbers, units, and dates

### Digits everywhere

Use digits for all numbers in prose, including small whole numbers.
The traditional Chicago-style rule of "spell out one through nine"
optimizes for print journalism; digits are more accessible for the
international and non-native-English audience that reads Coder
docs, scan faster in technical prose, and stay legible through
machine translation.

If a sentence would start with a digit, restructure the sentence so
a word comes first. Don't spell out the number to avoid the leading
digit; that reintroduces the rule the digits-everywhere policy is
meant to remove.

**Do**:

> The agent retries 3 times before giving up.
>
> Workspaces auto-stop after 8 hours of inactivity.
>
> The workspace has 5 connected users.

**Don't**:

> The agent retries three times before giving up.
>
> Workspaces auto-stop after eight hours of inactivity.
>
> 5 users connected to the workspace. (sentence starts with a
> digit; restructure to put a word first)

*Enforced by `Coder.DigitsEverywhere` (planned, ships at `warning`
severity because the rule is preference, not hard policy).*

### Non-breaking space between number and unit

Insert a non-breaking space (`&nbsp;` in HTML, `U+00A0` in raw
Markdown) between a number and its unit so the pair never breaks
across a line. The visible result is the same as a regular space,
but the line breaker treats the number and unit as one token.

**Do**:

> The default timeout is 30&nbsp;seconds. Connection latency under
> 150&nbsp;ms shows green.

**Don't**:

> The default timeout is 30 seconds. Connection latency under
> 150ms shows green.

In code blocks, configuration values, and CLI output, the original
format is preserved (`30s`, `150ms`); the non-breaking-space rule
applies to prose only.

*Enforced by `Google.Units` (planned).*

### Date format

Write dates as `Month Day, Year` with a full month name and a
comma between day and year. The format is unambiguous across
locales, which the all-numeric forms (`07/31/2026` vs `31/07/2026`)
are not.

**Do**:

> Coder released version 2.20 on July 31, 2026.

**Don't**:

> Coder released version 2.20 on 07/31/2026.
>
> Coder released version 2.20 on 31 July 2026.
>
> Coder released version 2.20 on 2026-07-31.

In code blocks, configuration values, log lines, and API responses,
keep whatever format the source uses; ISO 8601 (`2026-07-31`) is
correct in those contexts.

*Enforced by `Google.DateFormat` (planned).*

### Time format

Write times in 12-hour format with a space and uppercase AM or PM.

**Do**:

> The maintenance window starts at 9 AM and ends at 5 PM.

**Don't**:

> The maintenance window starts at 9am and ends at 5pm.
>
> The maintenance window starts at 09:00 and ends at 17:00.

In code blocks and timestamps from logs or APIs, keep the source
format; the 12-hour rule is for prose only.

*Enforced by `Google.AMPM` (planned).*

### Ordinals

Spell out ordinals `first` through `ninth`. Use digits with a
suffix for `10th` and up. This is the one place the digits-
everywhere rule yields, because ordinals spelled out read more
naturally in prose at low counts.

**Do**:

> The first time you run `coder login`, the CLI prompts you for an
> access URL.
>
> The 10th workspace in the list is the oldest.

**Don't**:

> The 1st time you run `coder login`, the CLI prompts you for an
> access URL.
>
> The tenth workspace in the list is the oldest.

*Enforced by `Google.Ordinal` (planned).*

## Editor setup

A future revision of this guide will cover Vale editor integration
for VS Code, Cursor, JetBrains, and Neovim so contributors get
inline feedback before commit instead of CI failure after push.

## Vale enforcement

The repo-root `.vale.ini` loads only the Coder rule package by
default. Third-party rules from Google, alex, and write-good are
not enabled until a per-rule PR brings each back in.

Each enabled rule lands via a dedicated PR that:

1. Cleans the corpus to zero baseline findings.
2. Adds the rule line in `.vale.ini` at the rule author's chosen
   severity.
3. Adds the corresponding section to this style guide.

Severity is a deliberate per-rule choice from the three-tier
ladder:

- `error` blocks merge in CI. Use for hard policy where any
  violation is wrong.
- `warning` surfaces an annotation without failing CI. Use for
  strong guidance with legitimate human-judgment exceptions.
- `suggestion` surfaces a `notice` annotation. Use for soft
  guidance where the right fix is contextual.

The full doctrine, including the false-positive policy, lives in
[`README.md`](./README.md). Run `make lint/prose` to reproduce the
baseline locally.

## Relationship to `docs/about/contributing/documentation.md`

A public-facing prose summary lives today at
[`docs/about/contributing/documentation.md`](../about/contributing/documentation.md).
A follow-up PR will redirect that page to this guide; until then,
follow the public summary for anything the sections above do not
cover. New prose rules land here; the public page is frozen
pending the redirect.

## Third-party references

When this guide does not cover something, consult:

| Type of guidance    | Reference                                                                               |
|---------------------|-----------------------------------------------------------------------------------------|
| Spelling            | [Merriam-Webster](https://www.merriam-webster.com/)                                     |
| Style, nontechnical | [The Chicago Manual of Style](https://www.chicagomanualofstyle.org/home.html)           |
| Style, technical    | [Microsoft Writing Style Guide](https://learn.microsoft.com/en-us/style-guide/welcome/) |
