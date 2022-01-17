import { test } from '@playwright/test';
import { parse, stringify } from 'yaml'
import { promises as fs } from 'fs';
import { Terminal } from 'playwright-terminal';

const doMakeImage = true;
const slackWorkspace = process.env.SLACK_WORKSPACE;
const slackEmail = process.env.SLACK_EMAIL;
const slackPassword = process.env.SLACK_PASSWORD;

test.use({
    contextOptions: {
        recordVideo: {
            dir: 'test-results/video/',
            size: { width: 640, height: 480 },
        },
    }
});

const terminalOptions = { logsEnabled: true, logDir: "test-results/logs"};

test.describe('quickstart', () => {
    test.afterEach(async () => {
        // Shut down any existing docker-compose runs
        var terminal = new Terminal(undefined, terminalOptions);
        // Output Gort (and dependency) logs to file
        await terminal.Execute("docker-compose", ["logs"]);
        // Stop Gort and dependencies
        await terminal.Execute("docker-compose", ["down"]);
    });

    // quickstart tests the flow described in the quickstart guide:
    //   https://guide.getgort.io/en/latest/sections/quickstart.html
    test('quickstart', async ({ page }) => {
        test.setTimeout(10 * 60 * 1000);

        // 2.2 Create Your Config File (https://guide.getgort.io/en/latest/sections/quickstart.html#create-your-configuration-file)
        await createConfigFile();
        
        // 2.3. Create Your Bot User (https://guide.getgort.io/en/latest/sections/quickstart.html#create-your-bot-user)
        // 2.3.1. Create a Slack Bot User
        await loginToSlack(
            page,
            slackWorkspace,
            slackEmail,
            slackPassword
        );
        await removeExistingSlackApps(page);
        const tokens = await configureSlackApp(page);
        await updateConfigFileForSlackApp(tokens.appToken,tokens.botToken);

        var terminal = new Terminal(page, terminalOptions);

        // 2.4. Build the Gort Image (Optional) (https://guide.getgort.io/en/latest/sections/quickstart.html#build-the-gort-image-optional)
        if (doMakeImage) {
            await terminal.Execute("make", ["image"]);
            // TODO: Verify that the image was built
        }

        // 2.5. Starting Containerized Gort (https://guide.getgort.io/en/latest/sections/quickstart.html#starting-containerized-gort)
        await startGort(terminal);

        // 2.6. Bootstrapping Gort (https://guide.getgort.io/en/latest/sections/quickstart.html#bootstrapping-gort)
        await bootstrapGort(terminal);

        // 2.7. Using Gort (https://guide.getgort.io/en/latest/sections/quickstart.html#using-gort)
        await page.goto(slackWorkspace); // Navigate to workspace

        // Create a channel (for a blank slate) and add Gort (alluded to in the instructions)
        const channel = "test-gort-" + Date.now();
        await createTestChannel(page, channel);
        await addGortToChannel(page);

        // Send a command
        await sendSlackMessage(page, channel, "!gort:version");

        // Expect Gort to respond
        await page.waitForSelector('text="Executing command: version"', {timeout: 2000});
        await page.waitForSelector('text=Gort ChatOps Engine v', {timeout: 5000});
        
        // TODO: Use the commands in the actual instructions (or update instructions)
        // await sendSlackMessage(page, channel, "!echo Hello, Gort!");
        // await sendSlackMessage(page, channel, "!echo:echo Hello, Gort!");
    });
})

async function createConfigFile() {
    const data = await fs.readFile("config.yml", "utf8");
    let config = parse(data);
    delete config.kubernetes;
    delete config.discord;
    await fs.writeFile("development.yml", stringify(config));
}

async function updateConfigFileForSlackApp(appToken: string, botToken: string) {
    const data = await fs.readFile("development.yml", "utf8");
    let config = parse(data);
    config.slack[0].app_token = appToken;
    config.slack[0].bot_token = botToken;
    await fs.writeFile("development.yml", stringify(config));
}

async function loginToSlack(page, workspaceURL, email, password) {
    // Navigate to the workspace
    await page.goto(workspaceURL);

    // Log in to the workspace
    await page.fill('[placeholder="name@work-email.com"]', email);
    await page.fill('[placeholder="Your password"]', password);
    await page.click('#signin_btn');

    // Use Slack in the browser rather than the app
    await page.click('text=use Slack in your browser');
}

async function removeExistingSlackApps(page) {
    await page.goto('https://api.slack.com/apps');
    while (await page.$('text=Gort', { timeout: 1000 }) !== null) {
        await page.click('text=Gort');
        await page.click('button:has-text("Delete App")');
        await page.click('text=Yes, I’m sure');
    }
}

async function configureSlackApp(page) {
    // Use this link to create a new Slack app: https://api.slack.com/apps?new_app=1.
    await page.goto('https://api.slack.com/apps?new_app=1');

    // Choose to create your app “From an app manifest”.
    await page.click('button:has-text("From an app manifestBETAUse a manifest file to add your app’s basic info, scopes")');

    // Select your workspace and click “Next”.
    await page.click('[aria-label="Select a team"]');
    await page.click('#team-picker_option_0 >> :nth-match(div:has-text("telliott Test"), 4)');
    await page.click('text=Next');

    // Copy the contents of the manifest file slackapp.yaml into the code box below “Enter app manifest below”, 
    // replacing the existing content. Click “Next”.
    await page.click('div.CodeMirror-lines');
    
    if (process.platform == "darwin") {
        await page.keyboard.press("Meta+A");
    } else {
        await page.keyboard.press("Control+A");
    }

    await page.keyboard.press("Delete");
    const appYaml = await fs.readFile("slackapp.yaml", "utf8");
    await page.fill('textarea', appYaml);
    await page.click('text=Next');

    // Review the summary and click “Create” to create your app.
    await page.waitForSelector('text=Bot Scopes (9)'); // Verify scopes are visible
    await page.click('text="Create"');

    // On the left-hand bar, under “Settings”, click “Basic Information”.
    await page.click('text=Basic Information');

    // Under “App-Level Tokens”, click “Generate Token and Scopes”.
    await page.click('text=Generate Token and Scopes');

    // Enter a name for your token, click “Add Scope” and select “connections:write”. Click “Generate”.
    await page.fill('[placeholder="This will be how you refer to your token"]', 'Gort Scope');
    await page.click('text=Add Scope');
    await page.click('text=connections:write');
    await page.click('div[role="dialog"] button:has-text("Generate")');

    // Copy the app token that starts with xapp- and paste it into the slack section of your development.yml config file as app_token. Click “Done”.
    await page.waitForSelector('text="Token"');
    await new Promise(r => setTimeout(r, 1000)); // TODO: Figure out what to wait for
    const appToken = await page.inputValue('#app_level_token_string');
    await page.click('text=Done');

    // On the left-hand bar, under “Settings”, click “Install App”.
    await page.click('text=Install App');
    await page.click('text=Install to Workspace');

    // You’ll get a screen that says something like “Gort is requesting permission to access the $NAME Slack workspace”; click “Allow”
    await page.click('text=Allow');

    // At the top of the screen, you should see “OAuth Tokens for Your Workspace” containing a “Bot User OAuth Token” that starts with xoxb-. Copy that value, and paste it into the slack section of your development.yml config file as bot_token.
    await page.waitForSelector('text=Bot User OAuth Token');
    const botToken = await page.inputValue('#bot_oauth_access_token');

    return {
    appToken: appToken,
    botToken: botToken,
    };
}

function checkGort() {
    var https = require('https');
    const options = {
        hostname: 'localhost',
        port: 4000,
        path: '/',
        method: 'GET',
        agent: new https.Agent({
            rejectUnauthorized: false
        })
    }

    return new Promise<boolean> ((resolve) => {
        const req = https.get(options,(res) => {
            return resolve(true);
        });

        req.on('error', (e) => {
            return resolve(false);
        });
    }); 
}


// startGort initializes Gort and waits until it is up and running
async function startGort(terminal) {
    await terminal.Execute("docker-compose", ["up", "-d"]);

    var running = false;
    while (!running) {
        running = await checkGort();
        if (!running) {
            await new Promise(f => setTimeout(f, 1000));
        }
    }
}

async function bootstrapGort(terminal) {
    // Note that -F is used here (not mentioned in Quickstart)
    // This makes it easier to re-run tests
    await terminal.Execute("go", ["run", ".", "bootstrap", "-F", "--allow-insecure", "https://localhost:4000"]);
}

// TODO: Add expected response
async function sendSlackMessage(page, channel, message) {
    // Post message and press enter
    await page.fill(`[aria-label="Message to ${channel}"]`, message);
    await page.press(`[aria-label="Message to ${channel}"]`, 'Enter');
}


async function createTestChannel(page, name) {
    await page.click('[data-qa=message_input]');
    if (process.platform == "darwin") {
        await page.press('[data-qa=message_input]', 'Meta+Shift+l');
    } else {
        await page.press('[data-qa=message_input]', 'Control+Shift+l');
    }

    await page.click('[data-qa=channel_browser_channel_create_btn]');

    await page.fill('[placeholder="e.g. plan-budget"]', name);
    await page.click('[aria-label="Create a channel"]')
    await page.click('[aria-label="Close"]');
    await page.click('text=Skip for Now');
}

async function addGortToChannel(page) {
    await page.click('[aria-label="View 1 member. Includes you."]');
    await page.click('button[role="tab"]:has-text("Integrations")');
    await page.click('text=Add an App');
    await page.click('[aria-label="Add app"]');
  }