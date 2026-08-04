# terraform-provider-sendgrid — Project Mandate

> A Terraform provider for a focused slice of [SendGrid](https://sendgrid.com/): Authenticated Domains, their DNS validation, and Inbound Parse webhook rules. It is **not** a general-purpose SendGrid provider. Published on the [Terraform Registry](https://registry.terraform.io/providers/weesp-ai/sendgrid) and consumed by `weesp-ai/mono` and `weesp-ai/infrastructure`.

This file is the foundational mandate for AI coding agents working in this repo. It takes precedence over general defaults. Read it before writing code or running commands.

## Tone

Act as a friendly, professional teammate. Concise, opinionated when warranted, honest about uncertainty. No filler.

## What this repo is

- A Go Terraform provider built on **terraform-plugin-framework**, with the provider implementation under `internal/provider/` and generated reference docs under `docs/`.
- Three resources: `sendgrid_authenticated_domain`, `sendgrid_domain_validation` (polls until DNS validation passes), `sendgrid_inbound_parse_rule`. See `README.md` for the attribute reference.
- The API key comes **only** from the provider's `api_key` attribute, fed through Terraform's variable mechanism. The provider deliberately does **not** read keys from arbitrary environment variables — keep that property.
- **This is a published artifact.** Tagging a release publishes to the public Terraform Registry via `.github/workflows/release.yml`, and consumers pin `~> 0.2`. A breaking schema change is breaking for real users, not just for this repo.

## Issue tracking

**This repository is public, and it tracks work in its own GitHub Issues.** That is the exception to the org's Linear-only rule, and the reason is the readers: consumers of the provider can open an issue here, follow it, and see why a release changed — none of which they can do with an internal Linear ticket. Work is *also* tracked internally in **Linear** (workspace `weesp-ai`, team **Engineering**, identifiers `DEV-<n>`), and the Linear ticket links out to the GitHub Issue and the PR so the internal board stays complete.

So, for a change here:

1. **A GitHub Issue in this repository** states the case for the change, in public. It is the artifact an outside reader can find.
2. **The PR references that issue** — `Fixes #12` (or `Contributes to #12` when the issue takes more than one PR) as a paragraph of its own, last in `## Summary`.
3. **The Linear ticket cross-links both**, and carries whatever internal context doesn't belong in public.

The `linear/linked` status check is still required on `main`. Attaching the PR to its Linear ticket from inside Linear satisfies it — the check reads Linear's own attachments, not just the PR text — as does a bare `DEV-<n>` in the branch name. Neither obliges the public PR body to carry an identifier its readers cannot resolve.

**No issue for the work you're about to push?** Stop and ask before opening the PR — offer to file one and say what you'd put in it. Never invent one silently. When you file the Linear side too, it sets three things at creation: the **assignee** is the developer driving the session, never the bot; an initial **priority** derived from severity (how bad the impact is) and urgency (how soon it has to be dealt with) — `Urgent` for something actively breaking production or exploitable now, `High` for a real defect or exposure that should land this cycle, `Medium` for ordinary planned work, `Low` for cleanup and retroactive write-ups; and an **estimate** on the team's t-shirt scale (`XS`–`XL`). Say which priority and estimate you picked and why when you offer the ticket — they are a starting point for the developer to correct, not a call to make silently.

Opening a PR with **no** tracking at all needs the developer's explicit permission, asked for and given in that conversation; permission for one PR never carries to the next. Once given, add the `skip-linear` label so the waiver is recorded on the PR itself.

### What the issue says, and what the PR says

An issue makes the **case for a change**; a PR gives the **account of the change**. Say each thing once, in the artifact that owns it — the shared `CONTRIBUTING.md` §The ticket and the PR each have one job carries the full rules.

- **An issue title states a state; a PR title states a transition.** "A validated domain re-plans as tainted on every apply" is the issue; "Domain validation stops proposing a replacement for an already-valid domain" is its PR.
- **An issue may carry the diagnosis; it must not carry the design.** Naming the attribute that drifts is diagnosis. "Mark the attribute `Computed` and add a state upgrader" is design, and belongs in the PR that does it.
- **The PR doesn't re-argue the motivation** — but because this repository is public and its issues are the only tracker outside readers can follow, a PR here keeps enough of the case for the change to stand on its own. That is the one place this repo departs from the org-wide split.
- **When the implementation contradicts the issue, say so in the PR** and amend the issue. Never diverge quietly.

### Linking tickets in chat

In chat replies, render every Linear ticket identifier as a markdown link — `DEV-123` becomes `[DEV-123](https://linear.app/weesp-ai/issue/DEV-123)`. Chat only: in commit messages, PR titles and descriptions, branch names, and docs, keep the bare identifier (`Fixes DEV-362`) — that is the form Linear's status automation and the `linear/linked` check read.

## When stuck, read

| If you need…                                       | Read                                                        |
|----------------------------------------------------|-------------------------------------------------------------|
| Resource reference, dev overrides, usage           | `README.md`, `docs/`, `examples/`                            |
| Org-wide conventions (commit format & types, PR rules, bot workflow, GPG) | shared `CONTRIBUTING.md` in the internal `weesp-ai/docs` |

**This repository is public.** Unlike its siblings, its `CLAUDE.md` is world-readable and indexed, so it deliberately carries no internal hostnames, service-account identities, bucket names, or IAP token recipes. The org docs hub and the identity setup for cloud sessions are described in the internal `weesp-ai/docs` — read them from there, and don't copy those details back into this file.

**Reading docs — prefer the hub over cloning.** Docs in *this* repo you already have: read them from the working tree. For the org-wide `CONTRIBUTING.md` and the other shared docs, prefer the internal rendered docs hub over attaching a repo — a cloud session reads a page there with one authenticated HTTP GET, where `add_repo` clones a whole repository and mints credentials. The hub's address and its token recipe live in the internal `weesp-ai/docs`, deliberately not repeated here; `add_repo weesp-ai/docs` is the fallback, for when you need the raw Markdown source, write access, or don't have that recipe to hand.

## Commands

Run from the repo root — the `Makefile` wraps all of these:

```bash
make build     # go build -o terraform-provider-sendgrid .
make test      # go test ./...            (unit tests, no network)
make fmt       # go fmt ./...
make vet       # go vet ./...
make tidy      # go mod tidy
make docs      # regenerate docs/ with tfplugindocs
make install   # build into the local plugin dir for dev overrides
```

Acceptance tests hit the **live SendGrid API** and mutate real resources:

```bash
make testacc TF_VAR_sendgrid_api_key=SG.xxx    # TF_ACC=1, 30m timeout
```

CI (`.github/workflows/test.yml`) runs `go mod tidy` as a must-be-a-no-op, `gofmt -l`, `go vet`, `go test -race -count=1` and `go build` — run those locally before pushing. It does **not** run the acceptance tests.

## Conventions

- **Go** → follow the shared `CONTRIBUTING.md` §Go-specific. `.golangci.yml` is the lint contract.
- **Schema changes** → after touching a resource schema, regenerate `docs/` with `make docs` and commit it; the registry serves those pages.
- **Files** → end with a newline.
- **Commits** → the format and canonical type list are in the shared `CONTRIBUTING.md` §Commits.

## Acting as the bot identity

In Claude Code on the web sessions, the platform GitHub connector is OAuth-bound to the **Bracket Bot** user (`getbracket-bot`), so `mcp__github__*` tools authenticate as the bot for both reads and writes. Use `mcp__github__*` for **all** GitHub interactions so the human stays free to review.

- **Open PRs as ready-for-review, not drafts** — pass `draft: false` (or omit it) to `mcp__github__create_pull_request`.
- **Self-assign every PR you open** — immediately call `mcp__github__issue_write` with `method: "update"`, the new PR number, and `assignees: ["<bot login>"]`. Send `assignees` only — a `labels` array replaces the PR's entire label set.
- If `mcp__github__get_me` reports a login other than the bot, the connector wasn't switched — flag it and stop making writes.

### Acting as the GCP identity

Nothing in this repository touches GCP — the provider talks to SendGrid, and its tests are hermetic. Cloud sessions do carry a near-read-only GCP identity for the org, described in the internal `weesp-ai/docs`; every mutation still requires a human-driven PR.

### Replying to PR review comments

When you push a change that addresses a review comment, reply to that thread with **`Done.`** (just that) via `mcp__github__add_reply_to_pull_request_comment`. If you won't address it (incorrect, out of scope, handled elsewhere), reply with a one-sentence explanation instead — never `Done.` on something you didn't do.

Top-level review comments have no thread. When a bare acknowledgment is all a reply would say, add a 👍 reaction to the comment itself. Otherwise post one new PR comment via `mcp__github__add_issue_comment` that quotes the point being addressed, links the original comment's permalink, and responds. Never react-acknowledge feedback you didn't act on.

## Guardrails — ask before doing

Always pause and ask the user before any of the following, even if a similar action was approved earlier. Authorization for one invocation does not extend to others.

- **Git side-effects**: `git push --force` / `-f`, `git reset --hard`, `git rebase`, `git commit --amend`, branch deletion. Plain `git push` of new commits to a feature branch is routine.
- **GitHub side-effects**: merging, closing, or reopening PRs/issues; posting *new* PR or issue comments. Routine responses to review feedback per §Replying to PR review comments and self-assigning a PR you just opened are exempt.
- **Releasing**: pushing a tag publishes to the public Terraform Registry and cannot be unpublished. Always ask.
- **Acceptance tests**: `make testacc` creates and destroys real SendGrid Authenticated Domains and Inbound Parse rules against a live account. Always ask, and confirm which account.
- **Destructive shell**: `rm -rf`, deleting files outside the working tree, killing processes you didn't start.
- **External communication**: Slack/email or any outbound traffic from project credentials.

When uncertain whether an action is reversible or has side-effects, ask first.

## What NOT to do

- Don't read the API key from an ambient environment variable — `api_key` on the provider block is the only supported source.
- Don't rename or retype an existing schema attribute without a deliberate version bump; consumers pin `~> 0.2` and a state-incompatible change breaks their plans.
- Don't hand-edit `docs/` — it is generated by `make docs`.
- Don't grow this into a general-purpose SendGrid provider by reflex; new resources are a scoping decision, not a refactor.
- Don't `add_repo` a repo just to read its documentation — read the rendered page from the internal docs hub instead. See §When stuck, read.
- Don't open a PR that isn't linked to a GitHub Issue in this repository and tracked in Linear, and don't apply `skip-linear` without the developer's explicit say-so. See §Issue tracking.
- Don't write the fix into the issue or the issue's argument into the PR, and don't title an issue as though the fix already landed. See §What the issue says, and what the PR says.
- Don't put internal-only context in a public issue or PR — hostnames, service accounts, customer names, internal ticket detail. That belongs in the Linear ticket.
- Don't prefix PR titles with `type(scope):` — that format is commit-subject-only.
- Don't open PRs as draft.
- Don't add documentation or comments to code you didn't change.
- Don't reflow or rewrap a comment you're only partially editing — change just what changed and leave the surrounding line breaks intact.

## Always do

- After any Go change, run `make fmt`, `make vet` and `make test`; run `make tidy` if imports changed.
- Regenerate and commit `docs/` whenever a resource schema changes.
- Match the surrounding file's comment wrap width when writing a new comment, not a narrower default like 72.
- Write succinct PR titles that name **the change this branch makes** — the transition, where the issue title states the state — without a `type`/`scope` prefix. For one slice of a multi-PR issue, name the slice.
- Structure every PR description as `## Summary` (what this branch changes, enough of the case for the change that a public reader can follow it, then the `Fixes #<n>` reference as its own paragraph) and `## Technical details` (the how, and above all **why this way** — the obvious alternative you rejected and what defeats it), plus an optional `## Validation` when you have evidence you actually ran. Keep them compact; put long-form material in a Markdown file in this repository and link it by path rather than inlining it. Don't link a PR description here at the internal docs hub — this repository is public, and outside readers would hit an auth wall.
- Keep the PR title and description in sync with what's on the branch — re-read the cumulative diff after pushing new commits and update them if scope drifted.
- Files end with a trailing newline.
