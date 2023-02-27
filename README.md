# Youtube Livestream Scheduler

This is a utility that helps schedule a number of Youtube Live Broadcasts without dealing with overhead of running mutliple stream schedules at the same time and managing multiple keys. In short, it makes it possible to do last minute scheduling for Streams without having to manually create the broadcast schedule.

The tool can handle any number of Stream schedules and uses the popular [`cron`](https://www.ibm.com/docs/en/db2oc?topic=task-unix-cron-format) format for scheduling the various streams specified in the config.

## License

This project is licensed under [Apache2.0](/LICENSE).

## Contributing

If you'd like to contribute some code or have a suggestion/issue, please check the [CONTRIBUTING](/CONTRIBUTING.md) guidelines.

## Usage

1. Create a new project in the Google Developers Console:
  a. Go to the Google Developers Console.
  b. Create a new project by clicking on the "Create Project" button in the top right corner.
  c. Enter a name for your project and click on the "Create" button.

2. Enable the YouTube Data API v3:
  a. In the Google Developers Console, select your project.
  b. Click on the "Enable APIs and Services" button.
  c. Search for "YouTube Data API v3" and click on it.
  d. Click on the "Enable" button.

3. Create OAuth2.0 credentials:
  a. In the Google Developers Console, select your project.
  b. Click on the "Create Credentials" button and select "OAuth client ID".
  c. Select "Desktop App" as the application type.
  d. Enter a name for your OAuth client ID and click on the "Create" button.

4. Click on the "Download" button to download the client secret JSON file.
> Note: Make sure to update the `redirect_uris` field in the downloaded file to contain `urn:ietf:wg:oauth:2.0:oob`

Now you should be ready to start configuring `Streams` for use in the application.

## Configuration of Streams

Create a file somewhere to configure Streams (you can call it whatever you like). The file **MUST** be in YAML format, however.

Your file should follow be configured as shown:

```yaml
# this parameter is used by Google Oauth to configure the OIDC authorization flow
oauthConfigPath: "/path/to/oauth_config.json"
streams:
  - name: "Friendly name for logs and behind-the-scenes stuff"
    titlePrefix: "Title of Stream Broadcast"
    description: "example description for stream broadcast"
    schedule: "0 0 * * 0"
    delaySeconds: 1800 # this is the delay from when the stream is scheduled to be created and when it is set to start accepting data
    privacyLevel: "public" # can be one of 'private', 'public', 'unlisted'
```

> If you'd like you can copy the [example configuration]() and simply edit it.

```bash
# assumes you are currently in the git repository
cat $(pwd)/streams.config.example.yaml > ~/.yls.yaml
```

## Running the App

You will need a computer to run this on that can remain on 24/7 as this is a daemon process and is primarily meant to be run in the background.

The app will manage schedules for you once configured and will respond to configuration changes in real-time. This means if you need to make a change, you needn't restart the daemon.

### Usage

```bash
Usage:
  yls help [command] [flags]

Flags:
  -h, --help   help for help

Global Flags:
      --debug                  specifies whether Debug-level logs should be shown. This can be very noisy (be warned)
      --dry-run                specifies whether YLS should be run in dry-run mode. This means YLS will make no changes, but will help evaluate changes that would be done
      --oauth-config string    (required) the path to an associated OAuth configuration file (JSON) that is downloaded from Google for generation of the authorization token
  -c, --stream-config string   the path to the file which specifies configuration for youtube stream schedules
```

```bash
Usage:
  yls start [flags]

Flags:
  -h, --help             help for start
      --secrets string   (required) The path to a JSON file containing OAuth2.0 Access and Refresh Tokens

Global Flags:
      --debug                  specifies whether Debug-level logs should be shown. This can be very noisy (be warned)
      --dry-run                specifies whether YLS should be run in dry-run mode. This means YLS will make no changes, but will help evaluate changes that would be done
      --oauth-config string    (required) the path to an associated OAuth configuration file (JSON) that is downloaded from Google for generation of the authorization token
  -c, --stream-config string   the path to the file which specifies configuration for youtube stream schedules
```

> Note: the required flags (obviously) must be specified in each launch

## Maintainer / Author
- Ben Sykes (ben.sykes@statcan.gc.ca)
