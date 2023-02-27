# Youtube Livestream Scheduler

This is a utility that helps schedule a number of Youtube Live Broadcasts without dealing with overhead of running mutliple stream schedules at the same time and managing multiple keys. In short, it makes it possible to do last minute scheduling for Streams without having to manually create the broadcast schedule.

The tool can handle any number of Stream schedules and uses the popular [`cron`](https://www.ibm.com/docs/en/db2oc?topic=task-unix-cron-format) format for scheduling the various streams specified in the config.

## License

This project is licensed under [Apache2.0](/LICENSE).

## Contributing

If you'd like to contribute some code or have a suggestion/issue, please check the [CONTRIBUTING](/CONTRIBUTING.md) guidelines.

## Usage

1. Create a new project in the Google Developers Console:
  - Go to the Google Developers Console.
  - Create a new project by clicking on the "Create Project" button in the top right corner.
  - Enter a name for your project and click on the "Create" button.

2. Enable the YouTube Data API v3:
  - In the Google Developers Console, select your project.
  - Click on the "Enable APIs and Services" button.
  - Search for "YouTube Data API v3" and click on it.
  - Click on the "Enable" button.

3. Create OAuth2.0 credentials:
  - In the Google Developers Console, select your project.
  - Click on the "Create Credentials" button and select "OAuth client ID".
  - Select "Desktop App" as the application type.
  - Enter a name for your OAuth client ID and click on the "Create" button.

4. Click on the "Download" button to download the client secret JSON file.
> Note: Make sure to update the `redirect_uris` field in the downloaded file to contain `urn:ietf:wg:oauth:2.0:oob`

Now you should be ready to start configuring `Streams` for use in the application.

## Configuration of Streams

Create a file somewhere to configure Streams (you can call it whatever you like). The file **MUST** be in YAML format, however.

Your file should follow be configured as shown:

```yaml
streams:
  - name: "Friendly name for logs and behind-the-scenes stuff"
    titlePrefix: "Title of Stream Broadcast"
    description: "example description for stream broadcast"
    schedule: "0 0 * * 0"
    delaySeconds: 1800 # this is the delay from when the stream is scheduled to be created and when it is set to start accepting data
    privacyLevel: "public" # can be one of 'private', 'public', 'unlisted'
```

> If you'd like you can copy the [example configuration](/streams.config.example.yaml) and simply edit it.

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
  yls start [flags]

Flags:
  -h, --help           help for start
  -i, --input string   the path to the file which specifies configuration for youtube stream schedules (default '$HOME/.yls.yaml')

Global Flags:
      --debug                  specifies whether Debug-level logs should be shown. This can be very noisy (be warned)
      --dry-run                specifies whether YLS should be run in dry-run mode. This means YLS will make no changes, but will help evaluate changes that would be done
      --oauth-config string    (required) the path to an associated OAuth configuration file (JSON) that is downloaded from Google for generation of the authorization token
      --secrets-cache string   A path to a file location that will be used to cache OAuth2.0 Access and Refresh Tokens (default "/Users/bensykes/.youtube_oauth2_credentials") 
```

### Extra Considerations

- The cache used for OAuth2.0 should be considered sensitive since it also contains refresh tokens in addition to access tokens. Access tokens are short-lived and would likely not be a huge threat, but refresh tokens tend to be longer-lived and can be exchanged for new access tokens
- To configure streaming software using a non-default stream-key, you might need to manually access the YouTube livestream portal. The API does not differenciate between "Live Now" and "Events". As a result, we cannot programmatically gain access to "Live Now" ingestion information such as the stream-key.

## Maintainer / Author
- Ben Sykes (ben.sykes@statcan.gc.ca)
