# Cog2

Cog2 is a Go reimplementation and significant simplification of the [Cog Slack Bot](https://github.com/operable/cog).

## Cog2 feature summary (ALL ARE WIP!)

### Retained features (under consideration)

All of the [high-level features listed here](https://book.cog.bot/sections/introducing_cog.html#current-features):

* Extensibility 
  * Build new bot commands in any language (Docker container command handlers)
  * Built-in templating
* Adaptability
  * Unix-style pipelines? (may or may not be included)
  * Output redirection?
* Security
  * Fine-grained command permissions: Users, Groups, and Roles
  * Audit logging
* Chat-Provider Agnostic (Slack, HipChat, others)

### Discarded features (under consideration)

Cog was originally designed as a distributed computation framework, and much of the functionality built for this is unused or seldom used. Features that are not currently being considered for support in Cog2 are as follows:
* Remote relays (but may be considered later)
* Unix-style pipelines? (may or may not be included)

## The road to 0.1.0

1. Slack things (0.0.1)
   1. Send a message to a channel (DONE)
   2. Establish a connection with channel (DONE)
   3. Make channel configurable (DONE)
1.  Container things (0.0.2)
   1. Run a container
   2. Run a container in response to a message
   3. Pass message into container at run
   4. Pass container output back into channel
1. Bundles (0.0.3)
1. Templates (0.0.4)
1. Users (0.1.0) - *POC feature complete* (NOT fully feature complete!)
