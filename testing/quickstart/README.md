# Quickstart Test

The quickstart tests use [Playwright](https://playwright.dev/) to automate execution of the [Gort Quickstart](https://gort-guide-dev.readthedocs.io/en/latest/sections/quickstart.html) instructions to ensure they continue to perform as expected.

## Dependencies

In order to run these tests, you need to install Playwright along with the appropriate browser builds. From the root of this repo, run:

```
npm install && npx playwright install
```

You will also need [Go](https://go.dev/) and [Docker](https://www.docker.com/) in order to build and run Gort during the test.

## Debugging

You will need the following environment variables set:

* `SLACK_WORKSPACE` full URL for a Slack Workspace
* `SLACK_EMAIL` E-mail address to authenticate against the specified workspace
* `SLACK_PASSWORD` Password to authenticate to the specified Slack account

You can run the Slack test in [debug mode](https://playwright.dev/docs/1.17/debug) with the following command (run from the root of this repo):

```
PWDEBUG=console npx playwright test testing/quickstart/slack.spec.ts
```

## GitHub Actions

The tests are executed via a GitHub Action, triggered on pull requests and pushes. The Action configuration can be found at `.github/workflows/test.yaml`.