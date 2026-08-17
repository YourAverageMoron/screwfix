# Branch Naming

Conventions for naming Git branches. Follow the [shnjd guide](https://dev.to/shnjd/git-good-best-practices-for-branch-naming-and-commit-messages-oj4).

Commit message rules are a separate concern — see [Commit Messages](./COMMIT_MESSAGES.md).

## Format

```
<type>/<ticket-><short-description>
```

- `type` — required prefix (see below).
- `ticket-` — optional issue/ticket identifier.
- `short-description` — kebab-case summary of the change.

## Rules

1. **Descriptive** — name conveys the branch's purpose at a glance (`feature/login`, not `stuff`).
2. **Kebab-case** — separate words with hyphens. `bugfix/fix-login-issue`, not `bugfix/fixLoginIssue` or `bugfix/fix_login_issue`.
3. **Lowercase alphanumeric only** — `a-z`, `0-9`, and hyphens. No spaces, underscores, or special characters.
4. **No redundant hyphens** — no consecutive or trailing hyphens (`feat/new--login-` is wrong).
5. **Short** — concise enough to read in `git branch` / PR lists; long enough to be clear.
6. **One concern** — a branch does one thing. Split mixed work across branches.
7. **Branch from up-to-date base** — rebase onto the target branch before opening a PR.

## Type Prefixes

| Prefix     | Use for                                              |
|------------|------------------------------------------------------|
| `feature/` | New functionality                                    |
| `bugfix/`  | Non-critical bug fixes, often tied to an issue       |
| `hotfix/`  | Critical production fixes (fast-tracked)             |
| `release/` | Release prep and final revisions (`release/v1.2.0`)  |
| `docs/`    | Documentation only                                   |
| `chore/`   | Tooling, deps, build, refactors with no behaviour change |
| `spike/` | Throwaway spikes and prototypes                  |

## Ticket Numbers

When using an issue tracker (Jira, GitHub Issues), include the ticket ID after the prefix for traceability:

```
feature/PROJ-123-footer-links
bugfix/gh-456-null-pointer-in-parser
```

- Keep the ID exactly as the tracker spells it (`PROJ-123`, `gh-456`).
- Only one ticket per branch. If work spans tickets, split branches.

## Examples

```
feature/oauth2-login
bugfix/PROJ-88-navbar-overflow-on-safari
hotfix/gh-12-db-connection-leak
release/v1.4.0
docs/update-go-package-versioning
chore/bump-golangci-lint
spike/canary-deploy
```

## Lifecycle

- Branch off the appropriate base (`main` for features/hotfixes, the release target for `release/`).
- Keep branches short-lived; merge or delete once the PR lands.
- Delete the remote branch after merge.
- Tag releases on `release/` branches per [Go Module Versioning](../go/PACKAGE_VERSIONING.md).

## Don'ts

- No author initials, dates, or personal names in the branch (`rj/feature/login`).
- No generic names (`wip`, `tmp`, `fix`, `test`).
- No branch per developer — one branch per concern.
- Don't commit directly to `main` or long-lived release branches; use a PR.

## References

- Guide: <https://dev.to/shnjd/git-good-best-practices-for-branch-naming-and-commit-messages-oj4>
- Commit messages: [Commit Messages](./COMMIT_MESSAGES.md)
- Release tagging: [Go Module Versioning](../go/PACKAGE_VERSIONING.md)
