# Gort MVP To-Do

An incomplete list of tasks for Gort by milestone. Tasks listed within a milestone are mostly unordered.

This is ABSOLUTELY NOT an exhaustive list. Please feel free to add to it. If what you want to add doesn't have an open issue, please [create one of these as well](https://github.com/getgort/gort/issues).

## Milestone 4 (Command Bundles) -- _COMPLETE!_ ✅

- The **Gort Guide** : [started here](https://getgort.github.io/gort-guide/bundles.html) ✅
- Document how command bundles are _supposed_ to work. ✅
- Ensure command bundle behavior is consistent with documentation. ✅

## Milestone 5 (Focus on Security) -- _COMPLETE!_ ✅

- Encrypt Gort REST service communications. ✅
  - Allow a TLS certificate to be installed by Gort controller. ✅
    - MUST loudly warn on startup if one is not being used. ✅
  - Gortctl SHOULD warn when not using an encrypted connection. ✅
- Inject database password via an envvar so we don't have to add it as plain text in the config. ✅
  - Figure how we want to do it. ✅
  - Actually DO it. ✅

## Milestone 6 (Focus on Observability) -- _COMPLETE!_ ✅

- Functional health check endpoint. ✅
- Observability
  - Metrics endpoint ✅
  - Distributed tracing ✅ (only supports Jaeger for MVP)
- Command audit log ✅
  - All commands MUST be written to the database ✅
    - Including a unique identifier, Gort user ID and email, full command as typed, command+bundle (and version), originating adapter+channel, and status code ✅
  - Every command request MUST have a unique identifier ✅
    - All relevant log events MUST include this command identifier ✅

## Milestone 7 (Merge `gortctl` functionality into `gort` binary) -- _COMPLETE!_ ✅

- As it says on the tin: merge merge the [gortctl](https://github.com/getgort/gortctl) commands into the main Gort repo ✅
- Deprecate `gortctl` and archive [its repo](https://github.com/getgort/gortctl). ✅

## Milestone 8 (API-level authorization) -- _COMPLETE!_ ✅

- Document permissions model and how auth is expected to work; [execution rules](https://web.archive.org/web/20191130061912/http://book.cog.bot/sections/command_execution_rules.html). ✅
- Command (as entered by a user into chat) interpretation ✅
  - Command text tokenizer and parser ✅
  - Ability to specify command option behavior in bundles (equivalent to `rules/ParseOption`) ✅
- Rules
  - Rule tokenizer and parser ✅
- Roles
  - Database schema ✅
  - `role create|destroy` command ✅
  - `group grant|revoke` command ✅
- Bundle/command permission assignment ✅
- Runtime command authorization ✅

**_(Publicly release v0.8.0-alpha.0 at this point)_**

## Milestone 9 (TBD)

- TBD
