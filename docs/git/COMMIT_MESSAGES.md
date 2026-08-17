# Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification.

## Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

- `scope` is optional
- `body` and `footer` are optional, omit when the "why" is obvious

## Types

| Type       | Use when                                           |
|------------|----------------------------------------------------|
| `feat`     | A new feature                                      |
| `fix`      | A bug fix                                          |
| `docs`     | Documentation only changes                         |
| `style`    | Formatting, whitespace, semicolons — no code logic |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `perf`     | A change that improves performance                 |
| `test`     | Adding or correcting tests                         |
| `chore`    | Build, tooling, dependencies — no production code  |
| `ci`       | CI pipeline changes                                |
| `build`    | Build system or external dependencies              |
| `revert`   | Reverting a previous commit                        |

## Rules

1. **Subject** — imperative mood: "add", "fix", "refactor", not "added" or "fixes".
2. **Subject length** — ≤ 50 characters. No trailing period.
3. **Subject casing** — lowercase, no code-style capitalisation.
4. **Scope** — short noun identifying the area of change (e.g. `auth`, `api`, `cli`).
5. **Body** — wrap at 72 characters. Explain **why**, not what — the diff shows what.
6. **Breaking change** — add `!` after type/scope and a `BREAKING CHANGE:` footer.
   ```
   feat(api)!: drop support for v1 endpoints

   BREAKING CHANGE: v1 endpoints removed; migrate to v2.
   ```
7. **Reference issues** — in the footer: `Closes #123`, `Refs #456`.
8. **One concern per commit** — split mixed concerns (feature + refactor + formatting) into separate commits.
9. **Atomic commits** — each commit must build and pass tests. No "WIP" or "fix typo" follow-ups.
10. **No commit to a broken main** — branch and PR instead.
11. Prefer a scope that matches the service or package being changed.

