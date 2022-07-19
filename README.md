# Sensu Go OpsGenie Handler

Deprecated, please go to: https://github.com/sensu/sensu-opsgenie-handler

![Go Test](https://github.com/betorvs/sensu-opsgenie-handler/workflows/Go%20Test/badge.svg)
[![Sensu Bonsai Asset](https://img.shields.io/badge/Bonsai-Download%20Me-brightgreen.svg?colorB=89C967&logo=sensu)](https://bonsai.sensu.io/assets/betorvs/sensu-opsgenie-handler)

## Table of Contents
- [Overview](#overview)
- [Configuration](#configuration)
- [Usage examples](#usage-examples)
- [Others Configurations](#others-configurations)
  - [To use Opsgenie Priority from Entity or Check](#to-use-opsgenie-priority-from-entity-or-check)
  - [Argument Annotations](#argument-annotations)
  - [Asset registration](#asset-registration)
- [Installation from source](#installation-from-source)
- [Additional notes](#additional-notes)
  - [Option remediation handler](#option-remediation-handler)
  - [Option keepalived handler](#option-keepalived-handler)
- [Contributing](#contributing)

## Overview

The Sensu Go OpsGenie Handler is a [Sensu Event Handler][3] which manages
[OpsGenie][2] incidents, for alerting operators. With this handler,
[Sensu][1] can trigger OpsGenie incidents.

This handler was inspired by [pagerduty plugin][6].

After version 1.0.0 we changed opsgenie [sdk][7] to [sdk-v2][8].


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
      --addHooksToDetails                Include the checks.hooks in details to send to OpsGenie
  -A, --aliasTemplate string             The template for the alias to be sent (default "{{.Entity.Name}}/{{.Check.Name}}")
  -a, --auth string                      The OpsGenie API authentication token, use default from OPSGENIE_AUTHTOKEN env var
  -L, --descriptionLimit int             The maximum length of the description field (default 15000)
  -d, --descriptionTemplate string       The template for the description to be sent (default "{{.Check.Output}}")
      --escalation-team string           The OpsGenie Escalation Responders Team, use default from OPSGENIE_ESCALATION_TEAM env var: sre,ops (splitted by commas)
  -F, --fullDetails                      Include the more details to send to OpsGenie like proxy_entity_name, occurrences and agent details arch and os
      --hearbeat-map string              Map of entity/check to heartbeat name. E. entity/check=heartbeat_name,entity1/check1=heartbeat
      --heartbeat                        Enable Heartbeat Events
  -h, --help                             help for sensu-opsgenie-handler
  -i, --includeEventInNote               Include the event JSON in the payload sent to OpsGenie
  -l, --messageLimit int                 The maximum length of the message field (default 130)
  -m, --messageTemplate string           The template for the message to be sent (default "{{.Entity.Name}}/{{.Check.Name}}")
  -p, --priority string                  The OpsGenie Alert Priority, use default from OPSGENIE_PRIORITY env var (default "P3")
  -r, --region string                    The OpsGenie API Region (us or eu), use default from OPSGENIE_REGION env var (default "us")
      --remediation-event-alias string   Replace opsgenie alias with this value and add only output as node in opsgenie. Should be used with auto remediation checks
      --remediation-events               Enable Remediation Events to send check.output to opsgenie using alert alias from remediation-event-alias configuration
      --schedule-team string             The OpsGenie Schedule Responders Team, use default from OPSGENIE_SCHEDULE_TEAM env var: sre,ops (splitted by commas)
  -s, --sensuDashboard string            The OpsGenie Handler will use it to create a source Sensu Dashboard URL. Use OPSGENIE_SENSU_DASHBOARD. Example: http://sensu-dashboard.example.local/c/~/n (default "disabled")
      --tagTemplate strings              The template to assign for the incident in OpsGenie (default [{{.Entity.Name}},{{.Check.Name}},{{.Entity.Namespace}},{{.Entity.EntityClass}}])
  -t, --team string                      The OpsGenie Team, use default from OPSGENIE_TEAM env var: sre,ops (splitted by commas)
  -T, --titlePrettify                    Remove all -, /, \ and apply strings.Title in message title
      --visibility-teams string          The OpsGenie Visibility Responders Team, use default from OPSGENIE_VISIBILITY_TEAMS env var: sre,ops (splitted by commas)
  -w, --withAnnotations                  Include the event.metadata.Annotations in details to send to OpsGenie
  -W, --withLabels                       Include the event.metadata.Labels in details to send to OpsGenie

Use "sensu-opsgenie-handler [command] --help" for more information about a command.

```

**Note:** Make sure to set the `OPSGENIE_AUTHTOKEN` environment variable for sensitive credentials in production to prevent leaking into system process table. Please remember command arguments can be viewed by unprivileged users using commands such as `ps` or `top`. The `--auth` argument is provided as an override primarily for testing purposes. 

To configure OpsGenie Sensu Integration follow these first part in [OpsGenie Docs][5].

## Others Configurations

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


### Asset registration

The easiest way to get this handler added to your Sensu environment, is to add it as an asset from Bonsai:

```sh
sensuctl asset add betorvs/sensu-opsgenie-handler --rename sensu-opsgenie-handler
```

See `sensuctl asset --help` for details on how to specify version.

## Installation from source

Download the latest version of the sensu-opsgenie-handler from [releases][4],
or create an executable script from this source.

From the local path of the sensu-opsgenie-handler repository:
```
go build -o /usr/local/bin/sensu-opsgenie-handler main.go
```


## Additional notes

Both options presented here changes how this handler works, only use this for specific cases and remember to not apply any filter, because in both cases, they will send only in case of status `!= 0`.

### Option remediation handler

Using this option we enable opsgenie handler to add extra properties in a previous alert created with remediation actions (checks). 

Flags: `--remediation-events` and `--remediation-event-alias`. More info about [remediation][11].

Proposed Workflow:

1. Entity -> Check_Http -> opsgenie, remediation handlers
2. Remediation handler -> Run Check_Http_remediation to Entity 
3. Entity -> Check_Http_remediation -> opsgenie_remediation handler

In opsgenige original alert Check_Http we will find a note with Check_Http_remediation.Check.Output and a detail with name `remediation_CHECK-NAME_source` with sensu url if `--sensuDashboard` flag was configured.

In check definition :
```yml
---
type: CheckConfig
api_version: core/v2
metadata:
  name: Check_Http_remediation
  namespace: default
  annotations:
    sensu.io/plugins/sensu-opsgenie-handler/config/remediation-event-alias: "entity/Check_Http"
spec:
  command: collect_smart_evidences.sh
  subscriptions:
  - system
  handlers:
  - opsgenie_remediation
  interval: 60
  publish: false
```

And Handler:
```yml
type: Handler
api_version: core/v2
metadata:
  name: opsgenie_remediation
  namespace: default
spec:
  type: pipe
  command: sensu-opsgenie-handler --remediation-events -s https://sensu-dashboard.example.com/c/~/n
  env_vars:
  - OPSGENIE_REGION=us
  timeout: 10
  runtime_assets:
  - betorvs/sensu-opsgenie-handler
  filters: null
```

More ideas about remediation try this [plugin][12].

### Option keepalived handler

This option enable opsgenie plugin to send heatbeat pings instead creating new alerts. This options could fit in keepalive for important network assets or important integrations (like [alert manager plugin][14]). If this check fails, and this plugin receives an event with status `!= 0` it discard this event. Opsgenie will alert you using heartbeat configuration. 

Flags `--heartbeat` and `--hearbeat-map` can map a entity/check to a [heartbeat][13] in opsgenie. 

And Handler:
```yml
type: Handler
api_version: core/v2
metadata:
  name: opsgenie_heartbeat
  namespace: default
spec:
  type: pipe
  command: sensu-opsgenie-handler --heartbeat --hearbeat-map webserver01/check-nginx=heartbeat_webserver01_nginx
  env_vars:
  - OPSGENIE_REGION=us
  timeout: 10
  runtime_assets:
  - betorvs/sensu-opsgenie-handler
  filters: null
```

Using all with `--heartbeat-map`:
```
--hearbeat-map webserver01/check-nginx=heartbeat_webserver01_nginx,webserver01/all=heartbeat_webserver01_all,all/check-nginx=heartbeat_all_nginx
```
In order: should match entity/check; should match entity with any check; should match any entity with check-nginx.


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
[11]: https://github.com/sensu/sensu-remediation-handler
[12]: https://github.com/betorvs/sensu-dynamic-check-mutator
[13]: https://docs.opsgenie.com/docs/heartbeat-api
[14]: https://github.com/betorvs/sensu-alertmanager-events
