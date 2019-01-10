# Cog2

Cog2 is a re-imagining and re-implementation of the current version of Operable's Cog [Cog Slack Bot](https://github.com/operable/cog) (Cog version 1, or Cog1).

## Overview

Operable Cog is described in its literature as a chatbot that brings the power of the command line to your chat window, and provides a powerful set of features that allow developers to implement arbitrary functionality, regulated by a powerful authorization layer.

Cog1 was originally designed as a distributed computation engine that was later rebranded as a chatops tool, and much of the original intent is reflected in its design, implementation, and featureset. As a result, many of Cog1’s features, however innovative, go largely unused, and the codebase has become difficult to extend and maintain. These difficulties are compounded by its implementation language -- Elixir -- which has few proficient developers.

The solution, which has been discussed for many months on the [Cog Slack workspace](https://cogbot.slack.com), is to rewrite Cog from scratch in a more accessible language, such as [Go](http://golang.org), removing superfluous functionality and reducing complexity in the process.

This gives us the opportunity to consider and possibly redefine what Cog is meant to be. To choose the features that make sense, and to discard those that don’t. In this way, Cog2 can be described more as a “spiritual successor” to Cog1 than a faithful reimplementation: many things will change, others will cease to exist entirely.

## Goals
The primary goal of this project is reimplement in Go the core features of Cog that distinguish it from other chatops tools. Namely, the ability to: 

* define arbitrary command functionality in any programming language,
* package those commands into bundles that can be installed in Cog,
* allow a user to trigger commands through Cog and be presented with the output,
* regulate the use of commands with a built-in authentication/authorization system,
* and record activity in an audit log.

This includes all of the [high-level features listed in the Cog1 documentation](https://book.cog.bot/sections/introducing_cog.html#current-features).

## Non-Goals  
While some effort will be made to support existing functionality (such as Cog1 bundles), perfect compatibility is explicitly not guaranteed (however, a migration guide should be written eventually).

## Cog2 design

A WIP design doc, including rough milestones (but not dates) [can be seen here](https://docs.google.com/document/d/1u7LzEzPjT1L8_xkHL577cKeuQdCiCQAww8M0rx1QXEM/edit?usp=sharing). Feel free to add questions or comments.

## How to run

With Go installed, you can run (for testing) with: `go run . start`

Note that you'll need a proper API key in the config first!