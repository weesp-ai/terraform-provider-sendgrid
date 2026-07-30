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

Work is tracked in **Linear** (workspace `weesp-ai`, team **Engineering**, identifiers `DEV-<n>`). The `weesp-ai/product` GitHub issue tracker is being migrated to Linear and then retired — **don't file new issues there**.

### Every PR needs a Linear ticket

**Don't open a PR that isn't linked to a Linear ticket.** The `linear/linked` status check is required on `main` and holds the merge until it is.

Link it from **a paragraph of its own, last in the `## Summary` section** — a standalone line after the summary paragraph, never a sentence tacked onto the end of it. When this is the only PR resolving the ticket, write `Fixes DEV-362` — `Fixes` (or `Closes` / `Resolves`) is what moves the ticket to Done when the PR merges. When the ticket is expected to take more than one PR, write `Contributes to DEV-362` instead: it links the PR without touching the ticket's status on merge, leaving the ticket to be closed by hand or by the PR that finishes it. A bare `DEV-<n>` or Linear URL in the title, description or branch name satisfies the check too, but drives no status transition. Attaching the PR from inside Linear also counts: the check reads Linear's own attachments, not just the PR text.

**No ticket for the work you're about to push?** Stop and ask before opening the PR — offer to file one and say what you'd put in it. Never invent a ticket silently. A ticket you file sets three things at creation: the **assignee** is the developer driving the session, never the bot; an initial **priority** derived from severity (how bad the impact is) and urgency (how soon it has to be dealt with) — `Urgent` for something actively breaking production or exploitable now, `High` for a real defect or exposure that should land this cycle, `Medium` for ordinary planned work, `Low` for cleanup and retroactive write-ups; and an **estimate** on the team's t-shirt scale (`XS`–`XL`). Say which priority and estimate you picked and why when you offer the ticket — they are a starting point for the developer to correct, not a call to make silently.

Opening a PR with **no** ticket at all needs the developer's explicit permission, asked for and given in that conversation; permission for one PR never carries to the next. Once given, add the `skip-linear` label so the waiver is recorded on the PR itself.

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
- Don't open a PR that isn't linked to a Linear ticket, and don't apply `skip-linear` without the developer's explicit say-so. See §Issue tracking.
- Don't prefix PR titles with `type(scope):` — that format is commit-subject-only.
- Don't open PRs as draft.
- Don't add documentation or comments to code you didn't change.
- Don't reflow or rewrap a comment you're only partially editing — change just what changed and leave the surrounding line breaks intact.

## Always do

- After any Go change, run `make fmt`, `make vet` and `make test`; run `make tidy` if imports changed.
- Regenerate and commit `docs/` whenever a resource schema changes.
- Match the surrounding file's comment wrap width when writing a new comment, not a narrower default like 72.
- Write succinct PR titles that describe the change, without a `type`/`scope` prefix.
- Structure every PR description with exactly two H2 sections: `## Summary` and `## Technical details`. Keep both compact; put long-form material in a Markdown file in this repository and link it by path rather than inlining it. Don't link a PR description here at the internal docs hub — this repository is public, and outside readers would hit an auth wall.
- Keep the PR title and description in sync with what's on the branch — re-read the cumulative diff after pushing new commits and update them if scope drifted.
- Files end with a trailing newline.
