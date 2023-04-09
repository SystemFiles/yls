# Youtube Livestream Scheduler

This is a utility that helps schedule a number of Youtube Live Broadcasts without dealing with overhead of running mutliple stream schedules at the same time and managing multiple keys. In short, it makes it possible to do last minute scheduling for Streams without having to manually create the broadcast schedule.

The tool can handle any number of Stream schedules and uses the popular [`cron`](https://www.ibm.com/docs/en/db2oc?topic=task-unix-cron-format) format for scheduling the various streams specified in the config.

## License

This project is licensed under [Apache2.0](/LICENSE).

## Contributing

If you'd like to contribute some code or have a suggestion/issue, please check the [CONTRIBUTING](/CONTRIBUTING.md) guidelines.

## Prerequisites

There are some additional dependencies used by this project that are dynamically linked during compile-time. Dynamic linking is used to help reduce the binary size for faster download during installation and so that the application can get updates for linked dependencies from the OS directly. 

Below are the required dependencies by OS

Debian-based systems

* g++
* g++-9
* libstdc++-9-dev
* jq (single-line command install only)

> These can be installed from the build-essential bundle `apt install -y build-essential`

Fedora:

* gcc
* gcc-c++
* kernel-devel
* jq (single-line command install only)

> These can be install using dnf `dnf install gcc gcc-c++ kernel-devel`

## Installation

### Binary

Download the binary from the [releases page](https://gitlab.sykesdev.ca/standalone-projects/yls/-/releases)

```bash
ARCH="amd64"; VERSION="$(curl -s 'https://gitlab.sykesdev.ca/api/v4/projects/62/releases?sort=desc' | jq -rc '.[0].name' | sed -e 's/v//g')"; mkdir -pv $HOME/bin && curl -sSLo yls.tar.gz "https://gitlab.sykesdev.ca/standalone-projects/yls/-/releases/v${VERSION}/downloads/yls_${VERSION}_$(uname -s | tr '[A-Z]' '[a-z]')_${ARCH}.tar.gz" && tar -C $HOME/bin -zxf ./yls.tar.gz yls && rm yls.tar.gz
```

> Note: does not work on Windows

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
  - name: Example Stream
    title: Example Stream
    description: "Example Description ... .. ... .."
    schedule: "*/1 * * * *" # every 10 minutes
    delaySeconds: 1800 # delay stream start time for 30 minutes
    privacyLevel: "unlisted"
    # publisher:
    #   wordpress:
    #     host: example.ca
    #     port: 80
    #     tls: no
    #     username: admin
    #     appToken: "AAAA BBBB CCCC DDDD EEEE FFFF"
    #     content:
    #       meta:
    #         status: published
    #       content: |
    #         <h1>Hello</h1>
```

> If you'd like you can copy the [example configuration](/streams.config.example.yaml) and simply edit it.

```bash
# assumes you are currently in the git repository
cat $(pwd)/streams.config.example.yaml > ~/.yls.yaml
```

## Running the App

You will need a computer to run this on that can remain on 24/7 as this is a daemon process and is primarily meant to be run in the background.

> WARNING: right now I don't know how to configure headless access to the Youtube Data API V3 ([might be impossible](https://developers.google.com/youtube/v3/guides/moving_to_oauth))

### Publishers

As of `v0.2.x`, YLS now supports publishers. While more publishers can easily be extended through the `Publisher` interface, currently the following publishers are supported:

- Wordpress

#### Planned Publishers

I'd like to expand the built-in publishers at some point (just need to find the time) to include the following (and more?)

- Instagram
- Twitter
- Facebook
- Webhook
- Discord
- Slack

### Extra Considerations

- The cache used for OAuth2.0 should be considered sensitive since it also contains refresh tokens in addition to access tokens. Access tokens are short-lived and would likely not be a huge threat, but refresh tokens tend to be longer-lived and can be exchanged for new access tokens
- To configure streaming software using a non-default stream-key, you might need to manually access the YouTube livestream portal. The API does not differenciate between "Live Now" and "Events". As a result, we cannot programmatically gain access to "Live Now" ingestion information such as the stream-key.

## Maintainer / Author
- Ben Sykes (ben.sykes@statcan.gc.ca)
