# Sensu Go OpsGenie Handler
TravisCI: [![TravisCI Build Status](https://travis-ci.org/betorvs/sensu-opsgenie-handler.svg?branch=master)](https://travis-ci.org/betorvs/sensu-opsgenie-handler)

The Sensu Go OpsGenie Handler is a [Sensu Event Handler][3] which manages
[OpsGenie][2] incidents, for alerting operators. With this handler,
[Sensu][1] can trigger OpsGenie incidents.

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
          "OPSGENIE_TEAM=TEAM_NAME"
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
  -a, --auth string   The OpsGenie V2 API authentication token, use default from OPSGENIE_AUTHTOKEN env var
  -h, --help          help for sensu-opsgenie-handler
  -t, --team string   The OpsGenie V2 API Team, use default from OPSGENIE_TEAM env var

```

**Note:** Make sure to set the `OPSGENIE_AUTHTOKEN` environment variable for sensitive credentials in production to prevent leaking into system process table. Please remember command arguments can be viewed by unprivileged users using commands such as `ps` or `top`. The `--auth` argument is provided as an override primarily for testing purposes. 


## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://www.opsgenie.com/ 
[3]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/betorvs/sensu-opsgenie-handler/releases

