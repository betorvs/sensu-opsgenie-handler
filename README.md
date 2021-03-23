# Sensu Go OpsGenie Handler
![Go Test](https://github.com/betorvs/sensu-opsgenie-handler/workflows/Go%20Test/badge.svg)
[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/betorvs/sensu-opsgenie-handler)

The Sensu Go OpsGenie Handler is a [Sensu Event Handler][3] which manages
[OpsGenie][2] incidents, for alerting operators. With this handler,
[Sensu][1] can trigger OpsGenie incidents.

This handler was inspired by [pagerduty plugin][6].

After version 1.0.0 we changed opsgenie [sdk][7] to [sdk-v2][8].

## Installation

Download the latest version of the sensu-opsgenie-handler from [releases][4],
or create an executable script from this source.

From the local path of the sensu-opsgenie-handler repository:
```
go build -o /usr/local/bin/sensu-opsgenie-handler main.go
```

## Configuration

Example Sensu Go handler definition:

```yml
type: Handler
api_version: core/v2
metadata:
  name: opsgenie
  namespace: default
spec:
  type: pipe
  command: sensu-opsgenie-handler
  env_vars:
  - OPSGENIE_TEAM=TEAM_NAME
  - OPSGENIE_REGION=us
  timeout: 10
  runtime_assets:
  - betorvs/sensu-opsgenie-handler
  filters:
  - is_incident
  secrets:
  - name: OPSGENIE_AUTHTOKEN
    secret: opgsgenie_authtoken
```

**Security Note:** Care should be taken to not expose the auth token for this handler by specifying it
on the command line or by directly setting the environment variable in the handler definition.  It is
suggested to make use of [secrets management][9] to surface it as an environment variable.  The
handler definition above references it as a secret.  Below is an example secrets definition that make
use of the built-in [env secrets provider][10].

```yml
---
type: Secret
api_version: secrets/v1
metadata:
  name: opsgenie_authtoken
spec:
  provider: env
  id: OPSGENIE_AUTHTOKEN
```

Example Sensu Go check definition:

```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: dummy-app-healthz
  namespace: default
  annotations:
    sensu.io/plugins/sensu-opsgenie-handler/config/priority: P2
spec:
  command: check-http -u http://localhost:8080/healthz
  subscriptions:
  - dummy
  handlers:
  - opsgenie
  interval: 60
  publish: true
```

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
      --addHooksToDetails            Include the checks.hooks in details to send to OpsGenie
  -a, --auth string                  The OpsGenie API authentication token, use default from OPSGENIE_AUTHTOKEN env var
  -L, --descriptionLimit int         The maximum length of the description field (default 15000)
  -d, --descriptionTemplate string   The template for the description to be sent (default "{{.Check.Output}}")
      --escalation-team string       The OpsGenie Escalation Responders Team, use default from OPSGENIE_ESCALATION_TEAM env var
  -F, --fullDetails                  Include the more details to send to OpsGenie like proxy_entity_name, occurrences and agent details arch and os
  -h, --help                         help for sensu-opsgenie-handler
  -i, --includeEventInNote           Include the event JSON in the payload sent to OpsGenie
  -l, --messageLimit int             The maximum length of the message field (default 130)
  -m, --messageTemplate string       The template for the message to be sent (default "{{.Entity.Name}}/{{.Check.Name}}")
  -p, --priority string              The OpsGenie Alert Priority, use default from OPSGENIE_PRIORITY env var (default "P3")
  -r, --region string                The OpsGenie API Region (us or eu), use default from OPSGENIE_REGION env var (default "us")
      --schedule-team string         The OpsGenie Schedule Responders Team, use default from OPSGENIE_SCHEDULE_TEAM env var
  -s, --sensuDashboard string        The OpsGenie Handler will use it to create a source Sensu Dashboard URL. Use OPSGENIE_SENSU_DASHBOARD. Example: http://sensu-dashboard.example.local/c/~/n (default "disabled")
  -t, --team string                  The OpsGenie Team, use default from OPSGENIE_TEAM env var
  -T, --titlePrettify                Remove all -, /, \ and apply strings.Title in message title
  -w, --withAnnotations              Include the event.metadata.Annotations in details to send to OpsGenie
  -W, --withLabels                   Include the event.metadata.Labels in details to send to OpsGenie

Use "sensu-opsgenie-handler [command] --help" for more information about a command.

```

**Note:** Make sure to set the `OPSGENIE_AUTHTOKEN` environment variable for sensitive credentials in production to prevent leaking into system process table. Please remember command arguments can be viewed by unprivileged users using commands such as `ps` or `top`. The `--auth` argument is provided as an override primarily for testing purposes. 

To configure OpsGenie Sensu Integration follow these first part in [OpsGenie Docs][5].

### To use Opsgenie Priority from Entity or Check

Please add this annotations inside sensu-agent:
```sh
# /etc/sensu/agent.yml example
annotations:
  sensu.io/plugins/sensu-opsgenie-handler/config/priority: "P1"
```

Or inside check:
```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: interval_check
  namespace: default
  annotations:
    sensu.io/plugins/sensu-opsgenie-handler/config/priority: P2
    documentation: https://docs.sensu.io/sensu-go/latest
spec:
  command: check-cpu.sh -w 75 -c 90
  subscriptions:
  - system
  handlers:
  - opsgenie
  interval: 60
  publish: true
```

### Argument Annotations

All arguments for this handler are tunable on a per entity or check basis based on annotations.  The
annotations keyspace for this handler is `sensu.io/plugins/sensu-opsgenie-handler/config`. It allows you to replace all flags, if it is a string type, like: `auth`, `priority`, `team`, `region`.

#### Examples

To change the team argument for a particular check, for that checks's metadata add the following:

```yml
type: CheckConfig
api_version: core/v2
metadata:
  annotations:
    sensu.io/plugins/sensu-opsgenie-handler/config/team: DevOps
[...]
```


### Asset creation

The easiest way to get this handler added to your Sensu environment, is to add it as an asset from Bonsai:

```sh
sensuctl asset add betorvs/sensu-opsgenie-handler --rename sensu-opsgenie-handler
```

See `sensuctl asset --help` for details on how to specify version.


## Contributing

See https://github.com/sensu/sensu-go/blob/master/CONTRIBUTING.md

[1]: https://github.com/sensu/sensu-go
[2]: https://www.opsgenie.com/ 
[3]: https://docs.sensu.io/sensu-go/5.0/reference/handlers/#how-do-sensu-handlers-work
[4]: https://github.com/betorvs/sensu-opsgenie-handler/releases
[5]: https://docs.opsgenie.com/docs/sensu-integration#section-add-sensu-integration-in-opsgenie
[6]: https://github.com/sensu/sensu-pagerduty-handler
[7]: https://github.com/opsgenie/opsgenie-go-sdk
[8]: https://github.com/opsgenie/opsgenie-go-sdk-v2
[9]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/
[10]: https://docs.sensu.io/sensu-go/latest/guides/secrets-management/#use-env-for-secrets-management
