# Gort MVP To-Do

An incomplete list of tasks for Gort by milestone. Tasks listed within a milestone are mostly unordered.

This is ABSOLUTELY NOT an exhaustive list. Please feel free to add to it. If what you want to add doesn't have an open issue, please [create one of these as well](https://github.com/clockworksoul/Gort/issues).

## Milestone 4 (Command Bundles)

- The **Gort Guide** initial draft (started: https://docs.google.com/document/d/1GetZrlyBLpzYKpPvI7ffZlBgwt_bBuqodQhToaYnY_4/edit?usp=sharing)
- Document how command bundles are _supposed_ to work.
- Ensure command bundle behavior is consistent with documentation.

## Milestone 5 (Focus on Security)

- Encrypt Gort REST service communications
  - Allow a TLS certificate to be installed by Gort service
    - MUST loudly warn on startup if one is not being used
  - Gortctl SHOULD warn when not using an encrypted connection
- Inject database password so we don't have to add it as plain text in the config.
  - Figure how we want to do it.
  - Actually DO it.

## Milestone 6 (Focus on Observability)

- Observability
  - Metrics endpoint
  - Distributed tracing
- Command audit log
  - All commands MUST be written to the database
    - Including a unique identifier, Gort user ID and email, full command as typed, command+bundle (and version), originating adapter+channel, and status code
  - Every command request MUST have a unique identifier
    - All relevant log events MUST include this command identifier

## Milestone 7 (API-level authorization)

- Document permissions model and how auth is expected to work.
- Command invocations can be associated with a user.
- User and group permission assignment
- Bundle/command permission assignment
- Audit log for all API actions (user, timestamp, action taken, method (Slack, API, etc.))

***(Publically release v0.7.0-alpha.0 at this point)***

## Milestone 8 (Remote relays)

- Document relay architecture
  - Requirement: Allow a quick-start-friendly "simple mode" with a local relay
  - Requirement: Support relay tagging and selection at the bundle/command level
- Break the `relay` package into a standalone service
- Include support for Kafka (and others? NATS?)

## Milestone X (Necessary but not attached to a milestone)

_This space intentionally left blank._


