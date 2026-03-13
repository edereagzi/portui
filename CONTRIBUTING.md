# Contributing

## Workflow

1. Pick an issue.
2. Create a branch from `main`.
3. Open a PR using the repository PR template.
4. Link the issue in the PR body using `Closes #<issue-number>`.

## Branch naming

- `feature/<issue-number>-<short-slug>`
- `bug/<issue-number>-<short-slug>`
- `task/<issue-number>-<short-slug>`

Examples:

- `feature/6-operator-filtering`
- `feature/7-kill-impact-summary`
- `task/10-windows-feasibility-spike`

## Pull request rules

- Keep PRs scoped to one issue.
- Run `go test ./...` and `go build ./...` before opening PR.
- Avoid unrelated file changes.
