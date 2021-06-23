# The Gort Guide

## _This project is a work in progress under active heavy development. It is not production (or even alpha) ready! Follow [the Gort project](https://github.com/getgort/gort) for updates!_

Gort is a chatbot framework designed from the ground up for chatops.

Gort brings the power of the command line to the place you collaborate with your team all the time -- your chat window. Its open-ended command bundle support allows developers to implement functionality in the language of their choice, while powerful access control means you can collaborate around even the most sensitive tasks with confidence. A focus on extensibility and adaptability means that you can respond quickly to the unexpected, without your team losing visibility.

## Features

The primary goal of this project is to re-implement the core features of Cog that made it stand out among other chatops tools. Specifically, to:

* define arbitrary command functionality in any programming language,
* package those commands into bundles that can be installed in Gort,
* allow users to trigger commands through Slack or another chat provider and be presented with the output,
* execute triggered commands anywhere a relay is installed using a tag-based targeting system,
* regulate the use of commands with a built-in authentication/authorization system,
* and record activity in an audit log.

## User Documentation

1. [Quick Start](quickstart.md)
2. [Architecture](architecture.md)
3. [Getting Started](getting-started.md)
3. [Bootstrapping Gort](bootstrapping.md)
4. [Command Bundles](bundles.md)
5. [Permissions and Rules](rules.md)
6. [Command Execution Rules](execution-rules.md)
7. [Users and Groups](users+groups.md)
