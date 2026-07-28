# Security Policy

## Supported versions

Security fixes are actively provided for the **latest released version** of zlog only.

| Version              | Supported          |
| -------------------- | ------------------ |
| Latest release       | Yes                |
| Older tagged releases| No (best effort)   |

Please upgrade to the newest release when a security fix is published.

## Reporting a vulnerability

**Do not** open a public GitHub issue or pull request to disclose a security vulnerability.

Report vulnerabilities privately via GitHub by contacting
[@GokselKUCUKSAHIN](https://github.com/GokselKUCUKSAHIN).

Please include as much of the following as you can:

- A short summary of the impact
- Steps to reproduce (or a minimal proof of concept)
- Affected version(s) or commit
- Any suggested fix or patch, if you have one

## What to expect

- We aim to acknowledge valid reports promptly after they are received.
- If the issue is confirmed, we will work on a fix and publish a new release when appropriate.
- We prefer coordinated disclosure: please allow time for a fix before sharing details publicly.

## Scope notes

zlog is a lightweight wrapper around Go's standard `log/slog` package with **no third-party dependencies**.

Security-relevant reports for this project typically involve issues in zlog itself (for example unexpected behavior in configuration loading, output handling, or related helpers). Misuse such as logging secrets or credentials from application code is generally outside the scope of library vulnerabilities, though suggestions that help users avoid common pitfalls are welcome through normal contribution channels.

## Prefer private contact

When public disclosure could put users at risk, always use the private GitHub contact above rather than filing a public issue.
