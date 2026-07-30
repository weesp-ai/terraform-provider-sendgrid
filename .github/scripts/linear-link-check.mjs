// Posts the `linear/linked` commit status that gates merging on a pull request
// being linked to a Linear issue. A copy of the canonical implementation in
// weesp-ai/.github-private (.github/actions/linear-link/), which every other
// weesp-ai repository calls instead of vendoring. GitHub does not let a public
// repository use an action or reusable workflow from a private or internal one,
// and this repository is public, so it keeps its own copy. Keep the two in step.
//
// Two sources of truth, in order:
//
//   1. Linear itself. Linear creates an Attachment on the issue whenever a PR is
//      linked, no matter how the link was made — branch name, PR title, a magic
//      word in the description, or someone attaching the PR by hand from the
//      Linear UI. Querying attachments back therefore catches the Linear-side
//      link, which nothing in the GitHub payload can see. Needs LINEAR_API_KEY.
//   2. The PR text. Scans title, description and branch name for a `DEV-<n>`
//      identifier — the same three places Linear's own integration looks (it
//      deliberately ignores comments and commit messages, so this does too).
//
// Source 2 alone is a complete check for every link made from the GitHub side,
// so a missing or expired LINEAR_API_KEY degrades to text matching rather than
// blocking every PR in the repo. When the key IS set, an identifier found in
// the text is confirmed against Linear before it counts, so a typo'd or deleted
// ticket fails instead of quietly passing.
//
// The result is written as a *commit status* rather than a check run because a
// status can be re-posted for the same context by any later workflow run. That
// is what lets the scheduled sweep flip a PR back to failing when its Linear
// link is removed from the Linear side, where no GitHub event ever fires.
//
// Inputs, all via the environment:
//   GITHUB_TOKEN       repo token with `statuses: write` + `pull-requests: read`
//   GITHUB_REPOSITORY  owner/repo
//   GITHUB_API_URL     API base (set by Actions; defaults to public GitHub)
//   PR_NUMBER          evaluate just this PR; unset means sweep all open PRs
//   RUN_URL            this workflow run, used as the status link on failure
//   LINEAR_API_KEY     optional; enables the authoritative Linear lookup
//   LINEAR_TEAM_KEYS   optional, comma-separated issue prefixes (default: DEV)

import {pathToFileURL} from 'node:url';

const STATUS_CONTEXT = 'linear/linked';
const WAIVER_LABEL = 'skip-linear';

// Bots that open PRs mechanically and can't be expected to file a ticket.
// Anything a human opens needs a ticket or the waiver label.
const EXEMPT_AUTHORS = ['dependabot[bot]', 'renovate[bot]', 'github-actions[bot]'];

const LINEAR_API = 'https://api.linear.app/graphql';
const apiBase = process.env.GITHUB_API_URL || 'https://api.github.com';
const [owner, repo] = (process.env.GITHUB_REPOSITORY || '').split('/');
const token = process.env.GITHUB_TOKEN;
const linearKey = process.env.LINEAR_API_KEY || '';
const teamKeys = (process.env.LINEAR_TEAM_KEYS || 'DEV')
    .split(',').map(k => k.trim().toUpperCase()).filter(Boolean);

// `DEV-362`, case-insensitively so a `dev-362-…` branch name counts. The team
// keys are anchored on word boundaries and never matched generically, because a
// generic `[A-Z]+-\d+` would read `CVE-2026-16389` or `UTF-8` as a ticket.
const IDENTIFIER_RE = new RegExp(`\\b(${teamKeys.join('|')})-(\\d+)\\b`, 'gi');

async function gh(path, init = {}) {
    const res = await fetch(path.startsWith('http') ? path : `${apiBase}${path}`, {
        ...init,
        headers: {
            accept: 'application/vnd.github+json',
            authorization: `Bearer ${token}`,
            'x-github-api-version': '2022-11-28',
            ...(init.body ? {'content-type': 'application/json'} : {}),
            ...(init.headers || {}),
        },
    });
    if (!res.ok) throw new Error(`GitHub ${init.method || 'GET'} ${path} -> ${res.status} ${await res.text()}`);
    return res.status === 204 ? null : res.json();
}

// Linear personal API keys are sent raw; OAuth tokens need the Bearer scheme.
async function linear(query, variables) {
    const res = await fetch(LINEAR_API, {
        method: 'POST',
        headers: {
            'content-type': 'application/json',
            authorization: linearKey.startsWith('lin_oauth_') ? `Bearer ${linearKey}` : linearKey,
        },
        body: JSON.stringify({query, variables}),
    });
    const json = await res.json().catch(() => null);
    if (!res.ok || !json || json.errors) {
        throw new Error(`Linear API ${res.status}: ${JSON.stringify(json?.errors ?? await res.text())}`);
    }
    return json.data;
}

const ISSUES_FOR_PR = `
query IssuesForPullRequest($url: String!) {
  issues(filter: {attachments: {url: {eq: $url}}}) { nodes { identifier url } }
}`;

const ISSUE_BY_NUMBER = `
query IssueByNumber($team: String!, $number: Float!) {
  issues(filter: {team: {key: {eq: $team}}, number: {eq: $number}}) { nodes { identifier url } }
}`;

// Every place Linear's GitHub integration looks for an identifier, and nowhere
// else — matching comments or commit messages would report a link that Linear
// never actually made.
function identifiersInText(pr) {
    const haystack = [pr.title || '', pr.body || '', pr.head?.ref || ''].join('\n');
    const found = new Map();
    for (const [, key, number] of haystack.matchAll(IDENTIFIER_RE)) {
        found.set(`${key.toUpperCase()}-${number}`, {key: key.toUpperCase(), number: Number(number)});
    }
    return [...found.values()];
}

async function issuesAttachedInLinear(pr) {
    const data = await linear(ISSUES_FOR_PR, {url: pr.html_url});
    return data.issues.nodes;
}

async function confirmIdentifiers(candidates) {
    const confirmed = [];
    for (const {key, number} of candidates) {
        const data = await linear(ISSUE_BY_NUMBER, {team: key, number});
        confirmed.push(...data.issues.nodes);
    }
    return confirmed;
}

// -> {state, description, target_url}
async function evaluate(pr) {
    const labels = (pr.labels || []).map(l => l.name);
    if (labels.includes(WAIVER_LABEL)) {
        return {state: 'success', description: `Waived by the "${WAIVER_LABEL}" label`};
    }
    if (EXEMPT_AUTHORS.includes(pr.user?.login)) {
        return {state: 'success', description: `Automated PR by ${pr.user.login} — no ticket required`};
    }

    const candidates = identifiersInText(pr);

    if (!linearKey) {
        return candidates.length
            ? {state: 'success', description: `Mentions ${describe(candidates.map(c => `${c.key}-${c.number}`))} (unverified — no Linear API key)`}
            : {state: 'failure', description: 'No Linear issue referenced in the title, description or branch name'};
    }

    let issues;
    try {
        // The Linear-side link first: it is the only one that survives a PR body
        // with no identifier in it.
        issues = await issuesAttachedInLinear(pr);
        if (!issues.length) issues = await confirmIdentifiers(candidates);
    } catch (err) {
        console.error(`  Linear lookup failed: ${err.message}`);
        // Don't call a PR unlinked because Linear was unreachable. A referenced
        // identifier is still positive evidence; anything else stays pending,
        // which blocks the merge without accusing the author of anything.
        return candidates.length
            ? {state: 'success', description: `Mentions ${describe(candidates.map(c => `${c.key}-${c.number}`))} (Linear unreachable — unverified)`}
            : {state: 'pending', description: 'Could not reach Linear to verify the link'};
    }

    if (!issues.length) {
        return {
            state: 'failure',
            description: candidates.length
                ? `${describe(candidates.map(c => `${c.key}-${c.number}`))} does not exist in Linear`
                : 'No Linear issue linked — add "Fixes DEV-<n>" to the description',
        };
    }
    return {
        state: 'success',
        description: `Linked to ${describe(issues.map(i => i.identifier))}`,
        target_url: issues[0].url,
    };
}

function describe(names) {
    return names.length > 2 ? `${names.slice(0, 2).join(', ')} +${names.length - 2} more` : names.join(', ');
}

// GitHub keeps every status ever posted for a context, so the half-hourly sweep
// would otherwise bury the PR timeline under identical entries. Only write when
// something actually changed.
async function postStatus(pr, result) {
    const target_url = result.target_url || process.env.RUN_URL;
    const existing = (await gh(`/repos/${owner}/${repo}/commits/${pr.head.sha}/statuses?per_page=100`))
        .find(s => s.context === STATUS_CONTEXT);
    if (existing && existing.state === result.state && existing.description === result.description) {
        console.log(`  #${pr.number} unchanged: ${result.state} — ${result.description}`);
        return;
    }
    await gh(`/repos/${owner}/${repo}/statuses/${pr.head.sha}`, {
        method: 'POST',
        body: JSON.stringify({
            state: result.state,
            context: STATUS_CONTEXT,
            description: result.description.slice(0, 140),
            target_url,
        }),
    });
    console.log(`  #${pr.number} -> ${result.state}: ${result.description}`);
}

async function openPullRequests() {
    const prs = [];
    for (let page = 1; page <= 10; page++) {
        const batch = await gh(`/repos/${owner}/${repo}/pulls?state=open&per_page=100&page=${page}`);
        prs.push(...batch);
        if (batch.length < 100) break;
    }
    return prs;
}

async function main() {
    const single = process.env.PR_NUMBER;
    const prs = single
        ? [await gh(`/repos/${owner}/${repo}/pulls/${single}`)]
        : await openPullRequests();

    console.log(`Evaluating ${prs.length} pull request(s) in ${owner}/${repo}` +
        `${linearKey ? '' : ' (no LINEAR_API_KEY — text matching only)'}`);

    let failed = 0;
    for (const pr of prs) {
        try {
            await postStatus(pr, await evaluate(pr));
        } catch (err) {
            // One unreachable PR must not abandon the rest of a sweep.
            failed++;
            console.error(`  #${pr.number} errored: ${err.message}`);
        }
    }
    if (failed) process.exitCode = 1;
}

// Guarded so the unit tests can import the decision logic without the script
// reaching for a GitHub token.
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) await main();

export {identifiersInText, evaluate, describe, STATUS_CONTEXT};
