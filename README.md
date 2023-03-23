# Youtube Livestream Scheduler

This is a utility that helps schedule a number of Youtube Live Broadcasts without dealing with overhead of running mutliple stream schedules at the same time and managing multiple keys. In short, it makes it possible to do last minute scheduling for Streams without having to manually create the broadcast schedule.

The tool can handle any number of Stream schedules and uses the popular [`cron`](https://www.ibm.com/docs/en/db2oc?topic=task-unix-cron-format) format for scheduling the various streams specified in the config.

## License

This project is licensed under [Apache2.0](/LICENSE).

## Contributing

If you'd like to contribute some code or have a suggestion/issue, please check the [CONTRIBUTING](/CONTRIBUTING.md) guidelines.

## Installation

### Binary

Download the binary from the [releases page](https://gitlab.sykesdev.ca/standalone-projects/yls/-/releases)

### Docker

You can find all related docker images in [Docker Hub](https://hub.docker.com/repository/docker/sykeben/yls/general)

## Usage

1. Create a new project in the [Google Developers Console](https://console.cloud.google.com/):
  - Go to the [Google Developers Console](https://console.cloud.google.com/).
  - Create a new project by clicking on the `Create Project` button in the top right corner.
  - Enter a name for your project and click on the `Create` button.

2. Enable the YouTube Data API v3:
  - In the [Google Developers Console](https://console.cloud.google.com/), select your project.
  - Click on the `Enable APIs and Services` button.
  - Search for `YouTube Data API v3` and click on it.
  - Click on the `Enable` button.

3. Create OAuth2.0 credentials:
  - In the [Google Developers Console](https://console.cloud.google.com/), select your project.
  - Click on the `Create Credentials` button and select `OAuth client ID`.
  - Select `Desktop App` as the application type.
  - Enter a name for your OAuth client ID and click on the `Create` button.

4. Click on the `Download` button to download the client secret JSON file.
> Note: Make sure to update the `redirect_uris` field in the downloaded file to contain `urn:ietf:wg:oauth:2.0:oob`

Now you should be ready to start configuring `Streams` for use in the application.

## Configuration of Streams

Create a file somewhere to configure Streams (you can call it whatever you like). The file **MUST** be in YAML format, however.

Your file should follow be configured as shown:

```yaml
streams:
  - name: "Friendly name for logs and behind-the-scenes stuff"
    title: "Title of Stream Broadcast"
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

### Publishers

As of `v0.2.x`, YLS now supports publishers. While more publishers can easily be extended through the `Publisher` interface, currently the following publishers are supported:

- Wordpress

## Configuring a Publisher

Publishers can be configured via config files and/or through commandline arguments and/or environment variables.

For the wordpress publisher an example configuration may include the following options:

| Name | env variable | Description | Default | Example |
| ---- | ------------ | ----------- | ------- | -------- |
|publish| `N/A` |Specifies whether livestreams as part of this schedule should be published using a publisher|`false`|`false`|
|wp-config| `YLS_WP_CONFIG` |the path to a file containing configuration (in YAML) for a wordpress publisher|`""`|`"/config/wp-pub.conf.yaml"`|
|wp-base-url| `YLS_WP_BASE_URL` |the base URL for a the wordpress v2 API|`""`|`"https://wordpress.example.ca/wp-json/wp/v2"`|
|wp-username| `YLS_WP_USERNAME` |the username for the user or service account to use for wordpress publishing|`""`|`"ben.sykes"`|
|wp-app-token| `YLS_WP_APP_TOKEN` |the wordpress App token to use to authenticate the identified wordpress user|`""`|`"DadH6 aiUW GIsY 62Yt"`|
|wp-page-id|`N/A`|(optional) a page ID for a wordpress page to publish changes to (if not specified, a page will be created)|`0`|`55555`|
|wp-page-template|`YLS_WP_PAGE_TEMPLATE`|a string that contains a gotemplate-compatible HTLM page template to use to construct wordpress page content|`""`|`"<h1>Live stream</h1><p>Available at: {{ .StreamURLShare }}</p>"`|

> As more publishers are added, options will be extended to configure specific traits of each of those implementations

It is worth noting that options above (for publishers) that are not required or optional are only not required **if** `publish` is `false`.

### Extra Considerations

- The cache used for OAuth2.0 should be considered sensitive since it also contains refresh tokens in addition to access tokens. Access tokens are short-lived and would likely not be a huge threat, but refresh tokens tend to be longer-lived and can be exchanged for new access tokens
- To configure streaming software using a non-default stream-key, you might need to manually access the YouTube livestream portal. The API does not differenciate between "Live Now" and "Events". As a result, we cannot programmatically gain access to "Live Now" ingestion information such as the stream-key.

## Maintainer / Author
- Ben Sykes (ben.sykes@statcan.gc.ca)
