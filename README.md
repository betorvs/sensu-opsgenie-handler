# Sensu Go OpsGenie Handler
[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/nixwiz/sensu-opsgenie-handler)
![Go Test](https://github.com/nixwiz/sensu-opsgenie-handler/workflows/Go%20Test/badge.svg)
![goreleaser](https://github.com/nixwiz/sensu-opsgenie-handler/workflows/goreleaser/badge.svg)

# Sensu Go OpsGenie Handler

## Table of Contents
- [Overview](#overview)
- [Files](#files)
- [Usage examples](#usage-examples)
- [Configuration](#configuration)
  - [Asset registration](#asset-registration)
  - [Handler definition](#handler-definition)
  - [Argument Annotations](#argument-annotations)
  - [Proxy support](#proxy-support)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
- [Contributing](#contributing)

## Overview

The Sensu Go OpsGenie Handler is a [Sensu Event Handler][3] which manages
[OpsGenie][2] incidents, for alerting operators. With this handler,
[Sensu][1] can trigger OpsGenie incidents.

This handler was inspired by [pagerduty plugin][6].

## Files

N/A

## Usage Examples

Help:
```
The Sensu Go OpsGenie handler for incident management

Usage:
  sensu-opsgenie-handler [flags]
  sensu-opsgenie-handler [command]

Available Commands:
  help        Help about any command
  version     Print the version number of this plugin

Flags:
  -a, --auth string              The OpsGenie V2 API authentication token, use default from OPSGENIE_AUTHTOKEN env var
  -h, --help                     help for sensu-opsgenie-handler
  -l, --messageLimit int         The maximum length of the message field (default 100)
  -m, --messageTemplate string   The template for the message to be sent (default "{{.Entity.Name}}/{{.Check.Name}}")
  -s, --sensuDashboard string    The OpsGenie Handler will use it to create a source Sensu Dashboard URL. Use OPSGENIE_SENSU_DASHBOARD. Example: http://sensu-dashboard.example.local/c/~/n (default "disabled")
  -t, --team string              The OpsGenie V2 API Team, use default from OPSGENIE_TEAM env var
  -u, --url string               The OpsGenie V2 API URL, use default from OPSGENIE_APIURL env var (default "https://api.opsgenie.com")
  -w, --withAnnotations string   The OpsGenie Handler will parse check and entity annotations with these values. Use OPSGENIE_ANNOTATIONS env var with commas, like: documentation,playbook (default "documentation,playbook")

Use "sensu-opsgenie-handler [command] --help" for more information about a command.

```

To configure OpsGenie Sensu Integration follow these first part in [OpsGenie Docs][5].

**Note 1:** Make sure to set the `OPSGENIE_AUTHTOKEN` environment variable for sensitive credentials in production to prevent leaking into system process table. Please remember command arguments can be viewed by unprivileged users using commands such as `ps` or `top`. The `--auth` argument is provided as an override primarily for testing purposes. 

**Note 2:** With the message being a template, be aware that OpsGenie uses that as the key
for an event, if you include an item in the message template that may change (such as
.Check.Output) between event creation and resolution, it could cause a mismatch and the
event will not resolve.

#### To use Opsgenie Priority

Please add this annotations inside sensu-agent:
```sh
# /etc/sensu/agent.yml example
annotations:
  opsgenie_priority: "P1"
```

Or inside check:
```sh
{
  "type": "CheckConfig",
  "api_version": "core/v2",
  "metadata": {
    "name": "interval_check",
    "namespace": "default",
    "annotations": {
        "opsgenie_priority": "P2",
        "documentation": "https://docs.sensu.io/sensu-go/latest"
    }
  },
  "spec": {
    "command": "check-cpu.sh -w 75 -c 90",
    "subscriptions": ["system"],
    "handlers": ["opsgenie"],
    "interval": 60,
    "publish": true
  }
}
```

#### To parse any annotation

With this new feature you can include any annotation field in message to show inside OpsGenie alert. By default they will look for documentation and playbook. 

## Configuration
### Sensu Go
#### Asset registration

[Sensu Assets][7] are the best way to make use of this plugin. If you're not using an asset, please
consider doing so! If you're using sensuctl 5.13 with Sensu Backend 5.13 or later, you can use the
following command to add the asset:

```
sensuctl asset add nixwiz/sensu-opsgenie-handler
```

If you're using an earlier version of sensuctl, you can find the asset on the [Bonsai Asset Index][8].


#### Handler definition

```json
{
    "api_version": "core/v2",
    "type": "Handler",
    "metadata": {
        "namespace": "default",
        "name": "opsgenie"
    },
    "spec": {
        "type": "pipe",
        "command": "sensu-opsgenie-handler",
        "env_vars": [
          "OPSGENIE_AUTHTOKEN=SECRET",
          "OPSGENIE_TEAM=TEAM_NAME",
          "OPSGENIE_APIURL=https://api.eu.opsgenie.com"
        ],
        "timeout": 10,
        "filters": [
            "is_incident"
        ]
    }
}
```

### Argument Annotations

All arguments for this handler are tunable on a per entity or check basis based on annotations.  The
annotations keyspace for this handler is `sensu.io/plugins/sensu-opsgenie-handler/config`.

#### Examples

To change the team argument for a particular check, for that checks's metadata add the following:

```yml
type: CheckConfig
api_version: core/v2
metadata:
  annotations:
    sensu.io/plugins/sensu-opsgenie-handler/config/team: WebOps
[...]
```

### Proxy Support

This handler supports the use of the environment variables HTTP_PROXY,
HTTPS_PROXY, and NO_PROXY (or the lowercase versions thereof). HTTPS_PROXY takes
precedence over HTTP_PROXY for https requests.  The environment values may be
either a complete URL or a "host[:port]", in which case the "http" scheme is assumed.

### Asset creation

The easiest way to get this handler added to your Sensu environment, is to add it as an asset from Bonsai:

```sh
sensuctl asset add betorvs/sensu-opsgenie-handler --rename sensu-opsgenie-handler
```

See `sensuctl asset --help` for details on how to specify version.

Another option is to manually register the asset by providing a URL to the tar.gz file, and sha512 hash for that file: 

```sh
sensuctl asset create sensu-opsgenie-handler --url "https://assets.bonsai.sensu.io/fba8c41f2b5bc817f8fb201144627042a3e31ee3/sensu-opsgenie-handler_0.0.4_linux_amd64.tar.gz" --sha512 "5eda4b31371fae83860604dedbf8527d0d6919bfae8e4f5b33f71bd314f6d706ef80356b14f11d7d2f86923df722338a3d11b84fa1e35323959120b46b738487"
```

### Sensu Core

See [this plugin][9]

## Installation from source

Download the latest version of the sensu-opsgenie-handler from [releases][4],
or create an executable script from this source.

From the local path of the sensu-opsgenie-handler repository:
```
go build -o /usr/local/bin/sensu-opsgenie-handler main.go
```

## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://www.opsgenie.com/ 
[3]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/betorvs/sensu-opsgenie-handler/releases
[5]: https://docs.opsgenie.com/docs/sensu-integration#section-add-sensu-integration-in-opsgenie
[6]: https://github.com/sensu/sensu-pagerduty-handler
[7]: https://docs.sensu.io/sensu-go/latest/reference/assets/
[8]: https://bonsai.sensu.io/
[9]: https://github.com/sensu-plugins/sensu-plugins-opsgenie
