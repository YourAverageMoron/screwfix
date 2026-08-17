# Git Agent Guide

This guide instructs agents how to work with Git in this repository: branching, commit messages, and the conventions that keep history readable and reviewable. Supporting detail lives in:
- [docs/git/BRANCHING.md](docs/git/BRANCHING.md) (branch naming conventions and prefixes)
- [docs/git/COMMIT_MESSAGES.md](docs/git/COMMIT_MESSAGES.md) (conventional commits rules)

## Operating Rules
- Never commit directly to `main` or long-lived release branches; create a branch and open a PR.
- Branch off an up-to-date base; rebase onto the target branch before opening a PR.
- One concern per branch. Split mixed work across branches.
- Name branches per [BRANCHING.md](docs/git/BRANCHING.md): `<type>/<ticket-><short-description>`, kebab-case, lowercase alphanumeric only.
- Write commit messages per [COMMIT_MESSAGES.md](docs/git/COMMIT_MESSAGES.md): Conventional Commits format with a clear subject and body when the "why" is not obvious.
- Keep branches short-lived; delete the remote branch after merge.

## Workflow
1. Update the base branch (`git fetch` then `git rebase origin/main` for features).
2. Create a descriptively named branch using the appropriate type prefix.
3. Make focused commits with conventional commit messages.
4. Rebase to resolve conflicts and keep history linear where possible.
5. Push and open a PR against the target branch.
6. After merge, delete the local and remote branches.

## References
- Branch naming guide: <https://dev.to/shnjd/git-good-best-practices-for-branch-naming-and-commit-messages-oj4>
- Conventional Commits: <https://www.conventionalcommits.org/>
