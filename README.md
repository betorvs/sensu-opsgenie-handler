# Sensu Go OpsGenie Handler
TravisCI: [![TravisCI Build Status](https://travis-ci.org/betorvs/sensu-opsgenie-handler.svg?branch=master)](https://travis-ci.org/betorvs/sensu-opsgenie-handler)
[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/betorvs/sensu-opsgenie-handler)

The Sensu Go OpsGenie Handler is a [Sensu Event Handler][3] which manages
[OpsGenie][2] incidents, for alerting operators. With this handler,
[Sensu][1] can trigger OpsGenie incidents.

This handler was inspired by [pagerduty plugin][6].

## Installation

Download the latest version of the sensu-opsgenie-handler from [releases][4],
or create an executable script from this source.

From the local path of the sensu-opsgenie-handler repository:
```
go build -o /usr/local/bin/sensu-opsgenie-handler main.go
```

## Configuration

Example Sensu Go handler definition:

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

Example Sensu Go check definition:

```json
{
    "api_version": "core/v2",
    "type": "CheckConfig",
    "metadata": {
        "namespace": "default",
        "name": "dummy-app-healthz"
    },
    "spec": {
        "command": "check-http -u http://localhost:8080/healthz",
        "subscriptions":[
            "dummy"
        ],
        "publish": true,
        "interval": 10,
        "handlers": [
            "opsgenie"
        ]
    }
}
```

## Usage Examples

Help:
```
Usage:
  sensu-opsgenie-handler [flags]

Flags:
  -a, --auth string             The OpsGenie V2 API authentication token, use default from OPSGENIE_AUTHTOKEN env var
  -h, --help                    help for sensu-opsgenie-handler
  -t, --team string             The OpsGenie V2 API Team, use default from OPSGENIE_TEAM env var
  -w, --withAnnotations string  To parse any annotation to send to OpsGenie inside message field. Use                                                OPSGENIE_ANNOTATIONS env var. Split them using comma (,).

```

**Note:** Make sure to set the `OPSGENIE_AUTHTOKEN` environment variable for sensitive credentials in production to prevent leaking into system process table. Please remember command arguments can be viewed by unprivileged users using commands such as `ps` or `top`. The `--auth` argument is provided as an override primarily for testing purposes. 

To configure OpsGenie Sensu Integration follow these first part in [OpsGenie Docs][5].

### To use Opsgenie Priority

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

### To parse any annotation

With this new feature you can include any annotation field in message to show inside OpsGenie alert. By default they will look for documentation and playbook. 


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


## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://www.opsgenie.com/ 
[3]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/betorvs/sensu-opsgenie-handler/releases
[5]: https://docs.opsgenie.com/docs/sensu-integration#section-add-sensu-integration-in-opsgenie
[6]: https://github.com/sensu/sensu-pagerduty-handler
