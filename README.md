# Gort

[![GoDoc](https://godoc.org/github.com/getgort/gort?status.svg)](https://godoc.org/github.com/getgort/gort)
[![Tests](https://github.com/getgort/gort/actions/workflows/test.yaml/badge.svg)](https://github.com/getgort/gort/actions/workflows/test.yaml)
[![Go Report Card](https://goreportcard.com/badge/github.com/getgort/gort)](https://goreportcard.com/report/github.com/getgort/gort)

**This project is a work in progress under active heavy development. It is not production (or even alpha) ready! Follow for updates!**

Gort is a chatbot framework designed from the ground up for chatops.

Gort brings the power of the command line to the place you collaborate with your team all the time -- your chat window. Its open-ended command bundle support allows developers to implement functionality in the language of their choice, while powerful access control means you can collaborate around even the most sensitive tasks with confidence. A focus on extensibility and adaptability means that you can respond quickly to the unexpected, without your team losing visibility.

## Rationale

Gort was initially conceived of as a Go re-implementation of Operable's [Cog Slack Bot](https://github.com/operable/cog), and while it remains heavily inspired by Cog, Gort has largely gone its own way.

Cog was originally designed as a distributed computation engine that was later re-branded as a chatops tool, and much of this original intent was reflected in its design, implementation, and feature set. As a result, many of Cog’s features, however innovative, went largely unused, and the codebase had become difficult to extend and maintain. These difficulties were compounded by its implementation language -- Elixir -- which has relatively few proficient developers.

The solution, which was discussed for many months on the [Cog Slack workspace](https://cogbot.slack.com), was to rewrite Cog from scratch in a more accessible language, such as [Go](http://golang.org), removing some of less-used functionality and reducing complexity in the process.

This gives us the opportunity to consider and possibly redefine what Cog was meant to be. To choose the features that make sense, and to discard those that don't. In this way, Gort can be described more as a “spiritual successor” to Cog than a faithful re-implementation: many things will change, others will cease to exist entirely.

## Features

The primary goal of this project is to re-implement the core features of Cog that made it stand out among other chatops tools. Specifically, to:

- define arbitrary command functionality in any programming language,
- package those commands into bundles that can be installed in Gort,
- allow users to trigger commands through Slack or another chat provider and be presented with the output,
- execute triggered commands anywhere a relay is installed using a tag-based targeting system,
- regulate the use of commands with a built-in authentication/authorization system,
- and record activity in an audit log.

This includes all of the [high-level features listed in the Cog documentation](https://web.archive.org/web/20191130061912/http://book.cog.bot/sections/introducing_cog.html#current-featuress).

<!-- ## Non-Goals

While some effort will be made to support existing functionality (such as Cog bundles), perfect compatibility is explicitly not guaranteed (however, a migration guide should be written eventually). -->

## Gort design

A WIP design doc, including rough milestones (but not dates) [can be seen here](https://docs.google.com/document/d/1u7LzEzPjT1L8_xkHL577cKeuQdCiCQAww8M0rx1QXEM/edit?usp=sharing). Feel free to add questions or comments.

## How to run the Gort controller

With Go installed, you can run (for testing) with: `go run . start`.

Note that you'll need [a proper API key in the config first](https://getgort.github.io/gort-guide/quickstart.html)!

## The Gort Client

The `gort` binary also serves as the controller administration CLI.

### Configuring client profiles

The `gort` client uses an INI-formatted configuration file, conventionally
located in the `profile` file in a `.gort` directory in your home directory.
This is where you can store connection credentials to allow `gort` to interact
with the Gort's Controller's REST API.

An example `.gort/profile` file might look like this:

```ini
[defaults]
profile = gort

[gort]
password = "seekrit#password"
url = https://gort.mycompany.com:4000
user = me

[preprod]
password = "anotherseekrit#password"
url = https://gort.preprod.mycompany.com:4000
user = me
```

Comments begin with a `#` character; if your password contains a `#`,
surround the entire password in quotes, as illustrated above.

You can store multiple "profiles" in this file, with a different name
for each (here, we have `gort` and `preprod`). Whichever one is noted
as the default (in the `defaults` section) will be used by
`gort`. However, you can pass the `--profile=$PROFILE` option to
`gort` to use a different set of credentials.

While you can add profiles to this file manually, you can also use the
`gort profile create` command to help.

### Getting help

The `gort` executable contains a number of commands and sub-commands.
Help is available for all of them by passing the `--help` option.
Start with `gort --help`, and go from there.

## Status of this project

Active heavy development. The date that various [milestones](TODO.md) have been achieved are listed below. The number and focus of present and future milestones are subject to change.

- Project created: 27 December 2018
- Milestone 1: 7 January 2019
- Milestone 2: 21 January 2019
- Milestone 3: 24 January 2019
- Milestone 4: 17 March 2019
- Milestone 5: 7 June 2021
- Milestone 6: 10 June 2021
- Milestone 7: 15 June 2021
- Milestone 8: _TBD_
- Milestone 9: _TBD_
- Release candidate 1: _TBD_
- Release!: _TBD_
