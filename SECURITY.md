# Security Policy

## Reporting a vulnerability

Please report security issues privately, not through a public issue.

Use GitHub's **[Report a vulnerability](https://github.com/Lucky2356/system-hub/security/advisories/new)**
button (Security → Advisories), which keeps the report private until a fix is
ready.

Please include:

- what the issue is and where in the code,
- how to reproduce it,
- the platform (Linux or Windows) and the version (`system-hub --version`).

You can expect an acknowledgement within a few days. There is no bounty program;
this is a personal open-source project.

## Scope

System Hub runs local, read-only diagnostics and manages local services and
containers. It has no network server and opens no ports, so the attack surface
is the local machine.

Things worth reporting:

- a way to make the app run a command it should not — the safe-command list, the
  service/container/image name handling, or the file browser,
- a path that escapes the directories the file browser is meant to confine to,
- an argument-injection path into `systemctl`, `docker`, `journalctl`,
  `sc`/`wevtutil`, or any other tool the app shells out to.

Managing services on Windows requires administrator rights, and on Linux may
require polkit/sudo — that a non-elevated user cannot start a service is expected
behaviour, not a vulnerability.

## Supported versions

Fixes land on the latest release. Please upgrade before reporting.
