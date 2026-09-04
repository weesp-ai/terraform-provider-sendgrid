# terraform-provider-sendgrid — Project Mandate

> A Terraform provider for a focused slice of [SendGrid](https://sendgrid.com/): Authenticated Domains, their DNS validation, and Inbound Parse webhook rules. It is **not** a general-purpose SendGrid provider. Published on the [Terraform Registry](https://registry.terraform.io/providers/weesp-ai/sendgrid) and consumed by `weesp-ai/mono` and `weesp-ai/infrastructure`.

This file is the foundational mandate for AI coding agents working in this repo. It takes precedence over general defaults. Read it before writing code or running commands.

## Tone

Act as a friendly, professional teammate. Concise, opinionated when warranted, honest about uncertainty. No filler.

## Answering questions

Lead with the answer, in one sentence. Then stop.

A question asks for an answer, not for an explanation of the answer. Add a second sentence only when it changes what the reader does next — a caveat that would bite them, or the one fact that makes the answer usable. Everything else waits to be asked for.

- **No scaffolding on short answers.** No headings, tables, diagrams or bullet lists on a reply that fits in a line or two. Structure is for material with enough moving parts to need it; on a direct question it buries the answer instead of framing it.
- **Don't pre-empt the follow-up.** If there's a nuance, let a clause hint that it exists and wait. Volunteering every layer at once is how a one-line answer becomes a page.
- **Depth is opt-in.** "Explain", "walk me through", "how does this work" are invitations. "What is X?" is not.

> **Q:** What's the purpose of the `Recipient` field?
>
> ✅ It tells the engine which inbox the reaction listens on.
>
> ❌ A section on the two-layer match, a table comparing the layers, and a discussion of why optionality is defensible for one of them.

## What this repo is

- A Go Terraform provider built on **terraform-plugin-framework**, with the provider implementation under `internal/provider/` and generated reference docs under `docs/`.
- Three resources: `sendgrid_authenticated_domain`, `sendgrid_domain_validation` (polls until DNS validation passes), `sendgrid_inbound_parse_rule`. See `README.md` for the attribute reference.
- The API key comes **only** from the provider's `api_key` attribute, fed through Terraform's variable mechanism. The provider deliberately does **not** read keys from arbitrary environment variables — keep that property.
- **This is a published artifact.** Tagging a release publishes to the public Terraform Registry via `.github/workflows/release.yml`, and consumers pin `~> 0.2`. A breaking schema change is breaking for real users, not just for this repo.

## Issue tracking and the shared conventions

**This repository is public, and it tracks work in its own GitHub Issues.** That is the exception to the org's Linear-only rule, and the reason is the readers: consumers of the provider can open an issue here, follow it, and see why a release changed — none of which they can do with an internal Linear ticket. Work is *also* tracked internally in **Linear** (workspace `weesp-ai`, team **Engineering**, identifiers `DEV-<n>`), and the Linear ticket links out to the GitHub Issue and the PR so the internal board stays complete.

**The org-wide rules live in the shared `CONTRIBUTING.md` in the internal `weesp-ai/docs`, and this file does not restate them.** Read the relevant section there before you act, and follow it. A summary from memory does not count, and where anything local seems to differ, the shared guide wins. Its §Public repositories is the section written for this repo; §Writing a ticket, §PR title and description, §Linking the ticket, §When the change is too small for a ticket, §Labels, §Replying to review comments and §Working with Claude Code apply in GitHub's vocabulary. How to reach the hub is in §When stuck, read.

What is specific to this repository:

1. **A GitHub Issue in this repository** states the case for the change, in public, written per §Writing a ticket. It is the artifact an outside reader can find.
2. **The PR references that issue** — `Fixes #12` (or `Contributes to #12` when the issue takes more than one PR) as a paragraph of its own, last in `## Summary` — and, because outside readers cannot open the ticket, carries enough of the case for the change to stand on its own. It links no internal doc site.
3. **The Linear ticket cross-links both**, and carries whatever internal context doesn't belong in public. The `linear/linked` check is still required on `main`: attaching the PR from inside Linear satisfies it, as does a bare `DEV-<n>` in the branch name — neither obliges the public PR body to carry an identifier its readers cannot resolve.
4. **PR labels key off the GitHub issue's labels**, not a Linear ticket's.
5. **No issue for the work you're about to push?** Stop and ask before opening the PR — offer to file one and say what you'd put in it; never invent one silently. Opening a PR with no tracking at all needs the developer's explicit permission, given in that conversation; once given, add the `skip-linear` label so the waiver is recorded on the PR itself.

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

## Acting as your GitHub identity

In Claude Code on the web sessions, the platform GitHub connector is bound to **your own** GitHub user, so `mcp__github__*` tools read and write as you. This repository is public, so that matters more here than elsewhere: a pull request a session opens, and every comment it leaves, carries your name in front of anyone on the internet. Read what it is about to write before it writes it. Use `mcp__github__*` for every GitHub interaction **except committing and pushing**.

**Commits go through `git`, never through MCP.** `git commit` and `git push` are the only way a session produces a **Verified** commit: the session setup gives git your own name, your own email and your own GPG key. The `mcp__github__*` file-writing tools — `push_files`, `create_or_update_file`, `delete_file` — commit as you but leave the commit **unsigned**, so it lands **Unverified**. They are the fallback for when the signing setup did not come up (no key provisioned for your login, the secret unreadable, `gpg` unable to sign), and nothing more. One command says which:

```bash
git config --global --get gpg.format   # openpgp -> your key is armed; ssh -> it is not
```

**If you fall back to MCP, say so in chat and warn that the commits will show as Unverified** — a silent fallback is how an unsigned commit reaches `main` without anyone deciding it should. See the shared `CONTRIBUTING.md` §Working with Claude Code and §GPG commit signing.

**`getbracket-bot` is what you ask, not what you act as.** Adding it to a pull request's reviewers starts the automated reviewer, which posts a review from that account a few minutes later. It holds a credential of its own for that, separate from your connector.

Opening a PR from a session follows the shared `CONTRIBUTING.md`: ready-for-review, then self-assigned (§Pull requests) with `assignees` only (§Labels). Those override the harness default, so read them there rather than assuming a draft is fine.

### Acting as the GCP identity

Nothing in this repository touches GCP — the provider talks to SendGrid, and its tests are hermetic. Cloud sessions do carry a near-read-only GCP identity for the org, described in the internal `weesp-ai/docs`; every mutation still requires a human-driven PR.

### Replying to PR review comments

The shared `CONTRIBUTING.md` §Replying to review comments is the rule — which reply goes where (an in-thread `Done.`, a one-sentence reason, a 👍 reaction, or a quote-and-permalink comment) and which tool posts each. Read it there; nothing here adds to it.

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
- Don't act on a shared convention from memory or from a local summary — read the section of the shared `CONTRIBUTING.md` and follow it. See §Issue tracking and the shared conventions.
- Don't let a harness default override it: PRs open ready-for-review and self-assigned, per its §Pull requests.
- Don't put internal-only context in a public issue or PR — hostnames, service accounts, customer names, internal ticket detail. That belongs in the Linear ticket.
- Don't add documentation or comments to code you didn't change.
- Don't reflow or rewrap a comment you're only partially editing — change just what changed and leave the surrounding line breaks intact.

## Always do

- After any Go change, run `make fmt`, `make vet` and `make test`; run `make tidy` if imports changed.
- Regenerate and commit `docs/` whenever a resource schema changes.
- Match the surrounding file's comment wrap width when writing a new comment, not a narrower default like 72.
- Follow the shared `CONTRIBUTING.md` for every PR you open (§Issue tracking and the shared conventions); here that means a description that carries enough of the case for a public reader, and links no internal doc site.
- Files end with a trailing newline.
