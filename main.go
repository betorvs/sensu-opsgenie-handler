package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/opsgenie/opsgenie-go-sdk-v2/heartbeat"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu-community/sensu-plugin-sdk/templates"

	"github.com/sensu/sensu-go/types"
)

const (
	notFound = "NOT FOUND"
	source   = "sensuGo"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	AuthToken             string
	APIRegion             string
	Team                  string
	EscalationTeam        string
	ScheduleTeam          string
	VisibilityTeams       string
	Priority              string
	SensuDashboard        string
	AliasTemplate         string
	MessageTemplate       string
	MessageLimit          int
	DescriptionTemplate   string
	DescriptionLimit      int
	IncludeEventInNote    bool
	WithAnnotations       bool
	WithLabels            bool
	FullDetails           bool
	HooksDetails          bool
	TitlePrettify         bool
	TagsTemplates         []string
	RemediationEvents     bool
	RemediationEventAlias string
	HeartbeatEvents       bool
	HeartbeatMap          string
}

var (
	plugin = Config{
		PluginConfig: sensu.PluginConfig{
			Name:     "sensu-opsgenie-handler",
			Short:    "The Sensu Go OpsGenie handler for incident management",
			Keyspace: "sensu.io/plugins/sensu-opsgenie-handler/config",
		},
	}

	options = []*sensu.PluginConfigOption{
		{
			Path:      "region",
			Env:       "OPSGENIE_REGION",
			Argument:  "region",
			Shorthand: "r",
			Default:   "us",
			Usage:     "The OpsGenie API Region (us or eu), use default from OPSGENIE_REGION env var",
			Value:     &plugin.APIRegion,
		},
		{
			Path:      "auth",
			Env:       "OPSGENIE_AUTHTOKEN",
			Argument:  "auth",
			Shorthand: "a",
			Default:   "",
			Secret:    true,
			Usage:     "The OpsGenie API authentication token, use default from OPSGENIE_AUTHTOKEN env var",
			Value:     &plugin.AuthToken,
		},
		{
			Path:      "team",
			Env:       "OPSGENIE_TEAM",
			Argument:  "team",
			Shorthand: "t",
			Default:   "",
			Usage:     "The OpsGenie Team, use default from OPSGENIE_TEAM env var: sre,ops (splitted by commas)",
			Value:     &plugin.Team,
		},
		{
			Path:      "escalation-team",
			Env:       "OPSGENIE_ESCALATION_TEAM",
			Argument:  "escalation-team",
			Shorthand: "",
			Default:   "",
			Usage:     "The OpsGenie Escalation Responders Team, use default from OPSGENIE_ESCALATION_TEAM env var: sre,ops (splitted by commas)",
			Value:     &plugin.EscalationTeam,
		},
		{
			Path:      "schedule-team",
			Env:       "OPSGENIE_SCHEDULE_TEAM",
			Argument:  "schedule-team",
			Shorthand: "",
			Default:   "",
			Usage:     "The OpsGenie Schedule Responders Team, use default from OPSGENIE_SCHEDULE_TEAM env var: sre,ops (splitted by commas)",
			Value:     &plugin.ScheduleTeam,
		},
		{
			Path:      "visibility-teams",
			Env:       "OPSGENIE_VISIBILITY_TEAMS",
			Argument:  "visibility-teams",
			Shorthand: "",
			Default:   "",
			Usage:     "The OpsGenie Visibility Responders Team, use default from OPSGENIE_VISIBILITY_TEAMS env var: sre,ops (splitted by commas)",
			Value:     &plugin.VisibilityTeams,
		},
		{
			Path:      "priority",
			Env:       "OPSGENIE_PRIORITY",
			Argument:  "priority",
			Shorthand: "p",
			Default:   "P3",
			Usage:     "The OpsGenie Alert Priority, use default from OPSGENIE_PRIORITY env var",
			Value:     &plugin.Priority,
		},
		{
			Path:      "sensuDashboard",
			Env:       "OPSGENIE_SENSU_DASHBOARD",
			Argument:  "sensuDashboard",
			Shorthand: "s",
			Default:   "disabled",
			Usage:     "The OpsGenie Handler will use it to create a source Sensu Dashboard URL. Use OPSGENIE_SENSU_DASHBOARD. Example: http://sensu-dashboard.example.local/c/~/n",
			Value:     &plugin.SensuDashboard,
		},
		{
			Path:      "aliasTemplate",
			Env:       "OPSGENIE_ALIAS_TEMPLATE",
			Argument:  "aliasTemplate",
			Shorthand: "A",
			Default:   "{{.Entity.Name}}/{{.Check.Name}}",
			Usage:     "The template for the alias to be sent",
			Value:     &plugin.AliasTemplate,
		},
		{
			Path:      "messageTemplate",
			Env:       "OPSGENIE_MESSAGE_TEMPLATE",
			Argument:  "messageTemplate",
			Shorthand: "m",
			Default:   "{{.Entity.Name}}/{{.Check.Name}}",
			Usage:     "The template for the message to be sent",
			Value:     &plugin.MessageTemplate,
		},
		{
			Path:      "messageLimit",
			Env:       "OPSGENIE_MESSAGE_LIMIT",
			Argument:  "messageLimit",
			Shorthand: "l",
			Default:   130,
			Usage:     "The maximum length of the message field",
			Value:     &plugin.MessageLimit,
		},
		{
			Path:      "descriptionTemplate",
			Env:       "OPSGENIE_DESCRIPTION_TEMPLATE",
			Argument:  "descriptionTemplate",
			Shorthand: "d",
			Default:   "{{.Check.Output}}",
			Usage:     "The template for the description to be sent",
			Value:     &plugin.DescriptionTemplate,
		},
		{
			Path:      "descriptionLimit",
			Env:       "OPSGENIE_DESCRIPTION_LIMIT",
			Argument:  "descriptionLimit",
			Shorthand: "L",
			Default:   15000,
			Usage:     "The maximum length of the description field",
			Value:     &plugin.DescriptionLimit,
		},
		{
			Path:      "titlePrettify",
			Env:       "",
			Argument:  "titlePrettify",
			Shorthand: "T",
			Default:   false,
			Usage:     "Remove all -, /, \\ and apply strings.Title in message title",
			Value:     &plugin.TitlePrettify,
		},
		{
			Path:      "includeEventInNote",
			Env:       "",
			Argument:  "includeEventInNote",
			Shorthand: "i",
			Default:   false,
			Usage:     "Include the event JSON in the payload sent to OpsGenie",
			Value:     &plugin.IncludeEventInNote,
		},
		{
			Path:      "withAnnotations",
			Env:       "",
			Argument:  "withAnnotations",
			Shorthand: "w",
			Default:   false,
			Usage:     "Include the event.metadata.Annotations in details to send to OpsGenie",
			Value:     &plugin.WithAnnotations,
		},
		{
			Path:      "withLabels",
			Env:       "",
			Argument:  "withLabels",
			Shorthand: "W",
			Default:   false,
			Usage:     "Include the event.metadata.Labels in details to send to OpsGenie",
			Value:     &plugin.WithLabels,
		},
		{
			Path:      "fullDetails",
			Env:       "",
			Argument:  "fullDetails",
			Shorthand: "F",
			Default:   false,
			Usage:     "Include the more details to send to OpsGenie like proxy_entity_name, occurrences and agent details arch and os",
			Value:     &plugin.FullDetails,
		},
		{
			Path:      "addHooksToDetails",
			Env:       "",
			Argument:  "addHooksToDetails",
			Shorthand: "",
			Default:   false,
			Usage:     "Include the checks.hooks in details to send to OpsGenie",
			Value:     &plugin.HooksDetails,
		},
		{
			Path:      "tagTemplate",
			Env:       "",
			Argument:  "tagTemplate",
			Shorthand: "",
			Default:   []string{"{{.Entity.Name}}", "{{.Check.Name}}", "{{.Entity.Namespace}}", "{{.Entity.EntityClass}}"},
			Usage:     "The template to assign for the incident in OpsGenie",
			Value:     &plugin.TagsTemplates,
		},
		{
			Path:      "remediation-events",
			Env:       "",
			Argument:  "remediation-events",
			Shorthand: "",
			Default:   false,
			Usage:     "Enable Remediation Events to send check.output to opsgenie using alert alias from remediation-event-alias configuration",
			Value:     &plugin.RemediationEvents,
		},
		{
			Path:      "remediation-event-alias",
			Env:       "",
			Argument:  "remediation-event-alias",
			Shorthand: "",
			Default:   "",
			Usage:     "Replace opsgenie alias with this value and add only output as node in opsgenie. Should be used with auto remediation checks",
			Value:     &plugin.RemediationEventAlias,
		},
		{
			Path:      "heartbeat",
			Env:       "",
			Argument:  "heartbeat",
			Shorthand: "",
			Default:   false,
			Usage:     "Enable Heartbeat Events",
			Value:     &plugin.HeartbeatEvents,
		},
		{
			Path:      "hearbeat-map",
			Env:       "",
			Argument:  "hearbeat-map",
			Shorthand: "",
			Default:   "",
			Usage:     "Map of entity/check to heartbeat name. E. entity/check=heartbeat_name,entity1/check1=heartbeat",
			Value:     &plugin.HeartbeatMap,
		},
	}
)

func main() {
	handler := sensu.NewGoHandler(&plugin.PluginConfig, options, checkArgs, executeHandler)
	handler.Execute()
}

func checkArgs(_ *types.Event) error {
	if len(plugin.AuthToken) == 0 {
		return fmt.Errorf("authentication token is empty")
	}
	// if len(plugin.Team) == 0 {
	// 	return fmt.Errorf("team is empty")
	// }
	if plugin.HeartbeatEvents && plugin.RemediationEvents {
		return fmt.Errorf("Cannot enable both options: --heartbeat and --remediation-events ")
	}
	return nil
}

// eventPriority func read priority in the event and return alerts.PX
// check.Annotations override Entity.Annotations
func eventPriority() alert.Priority {
	switch plugin.Priority {
	case "P5":
		return alert.P5

	case "P4":
		return alert.P4

	case "P3":
		return alert.P3

	case "P2":
		return alert.P2

	case "P1":
		return alert.P1

	default:
		return alert.P3
	}
}

func parseActions(event *types.Event) (output []string) {
	if event.Check.Annotations != nil && event.Check.Annotations["opsgenie_actions"] != "" {
		output = strings.Split(event.Check.Annotations["opsgenie_actions"], ",")
		return output
	}
	return output
}

// parseEventKeyTags func returns string, string, and []string with event data
// fist string contains custom templte string to use in message
// second string contains Entity.Name/Check.Name to use in alias
// []string contains Entity.Name Check.Name Entity.Namespace, event.Entity.EntityClass to use as tags in Opsgenie
func parseEventKeyTags(event *types.Event) (title string, alias string, tags []string) {
	alias, err1 := templates.EvalTemplate("title", plugin.AliasTemplate, event)
	if err1 != nil {
		return "", "", []string{}
	}

	// alias = fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
	title, err := templates.EvalTemplate("title", plugin.MessageTemplate, event)
	if err != nil {
		return "", "", []string{}
	}
	// tags = append(tags, event.Entity.Name, event.Check.Name, event.Entity.Namespace, event.Entity.EntityClass)
	for _, v := range plugin.TagsTemplates {
		tag, err := templates.EvalTemplate("tags", v, event)
		if err != nil {
			return "", "", []string{}
		}
		tags = append(tags, tag)
	}
	if plugin.TitlePrettify {
		newTitle := titlePrettify(title)
		return trim(newTitle, plugin.MessageLimit), alias, tags
	}
	return trim(title, plugin.MessageLimit), alias, tags
}

// parseDescription func returns string with custom template string to use in description
func parseDescription(event *types.Event) (description string) {
	description, err := templates.EvalTemplate("description", plugin.DescriptionTemplate, event)
	if err != nil {
		return ""
	}
	// allow newlines to get expanded
	description = strings.Replace(description, `\n`, "\n", -1)
	return trim(description, plugin.DescriptionLimit)
}

// parseDetails func returns a map of string string with check information for the details field
func parseDetails(event *types.Event) map[string]string {
	details := make(map[string]string)
	details["subscriptions"] = fmt.Sprintf("%v", event.Check.Subscriptions)
	details["status"] = fmt.Sprintf("%d", event.Check.Status)
	details["interval"] = fmt.Sprintf("%d", event.Check.Interval)
	// only if true
	if plugin.FullDetails {
		details["output"] = event.Check.Output
		details["command"] = event.Check.Command
		details["proxy_entity_name"] = event.Check.ProxyEntityName
		details["state"] = event.Check.State
		details["ttl"] = fmt.Sprintf("%d", event.Check.Ttl)
		details["occurrences"] = fmt.Sprintf("%d", event.Check.Occurrences)
		details["occurrences_watermark"] = fmt.Sprintf("%d", event.Check.OccurrencesWatermark)
		details["handlers"] = fmt.Sprintf("%v", event.Check.Handlers)

		if event.Entity.EntityClass == "agent" {
			details["arch"] = event.Entity.System.GetArch()
			details["os"] = event.Entity.System.GetOS()
			details["platform"] = event.Entity.System.GetPlatform()
			details["platform_family"] = event.Entity.System.GetPlatformFamily()
			details["platform_version"] = event.Entity.System.GetPlatformVersion()
		}
	}
	if plugin.HooksDetails {
		if len(event.Check.Hooks) != 0 {
			for k, v := range event.Check.Hooks {
				detailNameLabel := fmt.Sprintf("hooks_%v_%s_label", k, v.Name)
				detailNameCommand := fmt.Sprintf("hooks_%v_%s_command", k, v.Name)
				detailNameOutput := fmt.Sprintf("hooks_%v_%s_output", k, v.Name)
				if v.Labels != nil {
					for key, value := range v.Labels {
						label := fmt.Sprintf("%s_%s", detailNameLabel, key)
						details[label] = value
					}
				}
				if v.Command != "" {
					details[detailNameCommand] = v.Command
				}
				if v.Output != "" {
					details[detailNameOutput] = v.Output
				}
			}
		}
	}
	// only if true
	if plugin.WithAnnotations {
		if event.Check.Annotations != nil {
			for key, value := range event.Check.Annotations {
				if !strings.Contains(key, "sensu.io/plugins/sensu-opsgenie-handler/config") {
					checkKey := fmt.Sprintf("%s_annotation_%s", "check", key)
					details[checkKey] = value
				}
			}
		}
		if event.Entity.Annotations != nil {
			for key, value := range event.Entity.Annotations {
				if !strings.Contains(key, "sensu.io/plugins/sensu-opsgenie-handler/config") {
					entityKey := fmt.Sprintf("%s_annotation_%s", "entity", key)
					details[entityKey] = value
				}
			}
		}
	}
	// only if true
	if plugin.WithLabels {
		if event.Check.Labels != nil {
			for key, value := range event.Check.Labels {
				checkKey := fmt.Sprintf("%s_label_%s", "check", key)
				details[checkKey] = value
			}
		}
		if event.Entity.Labels != nil {
			for key, value := range event.Entity.Labels {
				entityKey := fmt.Sprintf("%s_label_%s", "entity", key)
				details[entityKey] = value
			}
		}
	}

	if plugin.SensuDashboard != "disabled" {
		details["sensuDashboard"] = fmt.Sprintf("source: %s \n", sensuDashboard(event.Entity.Namespace, event.Entity.Name, event.Check.Name))
	}

	return details
}

// sensuDashboard
func sensuDashboard(namespace, entity, check string) string {
	return fmt.Sprintf("%s/%s/events/%s/%s", plugin.SensuDashboard, namespace, entity, check)
}

// switchOpsgenieRegion func
func switchOpsgenieRegion() client.ApiUrl {
	var region client.ApiUrl
	apiRegionLowCase := strings.ToLower(plugin.APIRegion)
	switch apiRegionLowCase {
	case "eu":
		region = client.API_URL_EU
	case "us":
		region = client.API_URL
	default:
		region = client.API_URL
	}
	return region
}

func executeHandler(event *types.Event) error {
	alertClient, err := alert.NewClient(&client.Config{
		ApiKey:         plugin.AuthToken,
		OpsGenieAPIURL: switchOpsgenieRegion(),
	})
	if err != nil {
		return fmt.Errorf("failed to create opsgenie client: %s", err)
	}
	// always create an alert in opsgenie if status != 0
	if event.Check.Status != 0 && !plugin.RemediationEvents && !plugin.HeartbeatEvents {
		return createIncident(alertClient, event)
	}

	// if RemediationEvents true: change behaviour of opsgenie plugin
	if plugin.RemediationEvents && event.Check.Status == 0 {
		hasAlert, _ := getAlert(alertClient, plugin.RemediationEventAlias)
		details := make(map[string]string)
		if plugin.SensuDashboard != "disabled" {
			name := fmt.Sprintf("remediation_%s_source", event.Check.Name)
			details[name] = sensuDashboard(event.Entity.Namespace, event.Entity.Name, event.Check.Name)
		}
		notes := fmt.Sprintf("%s ", event.Check.Output)
		return updateAlert(alertClient, notes, hasAlert, details)
	}
	if plugin.RemediationEvents && event.Check.Status != 0 {
		fmt.Printf("not sending alert because --remediation-events is enabled %s/%s", event.Entity.Name, event.Check.Name)
		return nil
	}

	// if heartbeat true: match entity/check with heartbeat
	if plugin.HeartbeatEvents && event.Check.Status == 0 && plugin.HeartbeatMap != "" {
		return heartbeatEvent(event)
	}
	if plugin.HeartbeatEvents && event.Check.Status != 0 {
		fmt.Printf("not sending alert because --heartbeat is enabled %s/%s", event.Entity.Name, event.Check.Name)
		return nil
	}

	// check if event has a alert
	_, alias, _ := parseEventKeyTags(event)
	hasAlert, _ := getAlert(alertClient, alias)

	// close incident if status == 0
	if hasAlert != notFound && event.Check.Status == 0 {
		return closeAlert(alertClient, event, hasAlert)
	}

	return nil
}

// handle with heartbeat option
func heartbeatEvent(event *types.Event) error {
	heartbeats, err := parseHeartbeatMap(plugin.HeartbeatMap)
	if err != nil {
		return err
	}
	// match entity/check names
	entity_check := fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
	if heartbeats[entity_check] != "" {
		fmt.Printf("Pinging heartbeat %s \n", heartbeats[entity_check])
		errPing := pingHeartbeat(heartbeats[entity_check])
		if errPing != nil {
			return errPing
		}
	}
	// use any check
	entity_all := fmt.Sprintf("%s/all", event.Entity.Name)
	if heartbeats[entity_all] != "" {
		// ping all alerts
		fmt.Printf("Pinging heartbeat %s with entity/all defined\n", heartbeats[entity_all])
		errPing := pingHeartbeat(heartbeats[entity_all])
		if errPing != nil {
			return errPing
		}
		return nil
	}
	// use any entity
	all_check := fmt.Sprintf("all/%s", event.Check.Name)
	if heartbeats[all_check] != "" {
		// ping all alerts
		fmt.Printf("Pinging heartbeat %s with all/check defined\n", heartbeats[all_check])
		errPing := pingHeartbeat(heartbeats[all_check])
		if errPing != nil {
			return errPing
		}
		return nil
	}
	if heartbeats["all"] != "" {
		// ping all alerts
		fmt.Printf("Pinging heartbeat %s with all/all defined\n", heartbeats["all"])
		errPing := pingHeartbeat(heartbeats["all"])
		if errPing != nil {
			return errPing
		}
		return nil
	}
	if len(heartbeats) != 0 {
		fmt.Println("Not pinging any heartbeat because entity/check defined do not match")
	}
	return nil
}

// createIncident func create an alert in OpsGenie
func createIncident(alertClient *alert.Client, event *types.Event) error {
	var (
		note string
		err  error
	)

	if plugin.IncludeEventInNote {
		note, err = getNote(event)
		if err != nil {
			return err
		}
	}
	teams := respondersTeam()
	visibilityTeams := visibilityTeams()

	title, alias, tags := parseEventKeyTags(event)

	actions := parseActions(event)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	createResult, err := alertClient.Create(ctx, &alert.CreateAlertRequest{
		Message:     title,
		Alias:       alias,
		Description: parseDescription(event),
		Responders:  teams,
		VisibleTo:   visibilityTeams,
		Actions:     actions,
		Tags:        tags,
		Details:     parseDetails(event),
		Entity:      event.Entity.Name,
		Source:      source,
		Priority:    eventPriority(),
		Note:        note,
	})
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Create request ID: " + createResult.RequestId)
	}
	return nil
}

func respondersTeam() []alert.Responder {
	local := []alert.Responder{}
	if plugin.EscalationTeam != "" {
		teamsList := splitStringInSlice(plugin.EscalationTeam)
		if len(teamsList) != 0 {
			for _, v := range teamsList {
				if v != "" {
					team := []alert.Responder{
						{Type: alert.EscalationResponder, Name: v},
					}
					local = append(local, team...)
				}
			}
		}
		// escalation := []alert.Responder{
		// 	{Type: alert.EscalationResponder, Name: plugin.EscalationTeam},
		// }
		// local = append(local, escalation...)
	}
	if plugin.ScheduleTeam != "" {
		teamsList := splitStringInSlice(plugin.ScheduleTeam)
		if len(teamsList) != 0 {
			for _, v := range teamsList {
				if v != "" {
					team := []alert.Responder{
						{Type: alert.ScheduleResponder, Name: v},
					}
					local = append(local, team...)
				}
			}
		}
		// schedule := []alert.Responder{
		// 	{Type: alert.ScheduleResponder, Name: plugin.ScheduleTeam},
		// }
		// local = append(local, schedule...)
	}
	if plugin.Team != "" {
		teamsList := splitStringInSlice(plugin.Team)
		if len(teamsList) != 0 {
			for _, v := range teamsList {
				if v != "" {
					team := []alert.Responder{
						{Type: alert.TeamResponder, Name: v},
					}
					local = append(local, team...)
				}
			}
		}
	}
	return local
}

func visibilityTeams() []alert.Responder {
	local := []alert.Responder{}
	teamsList := splitStringInSlice(plugin.VisibilityTeams)
	if len(teamsList) != 0 {
		for _, v := range teamsList {
			if v != "" {
				team := []alert.Responder{
					{Type: alert.TeamResponder, Name: v},
				}
				local = append(local, team...)
			}
		}
	}
	return local
}

// getAlert func get a alert using an alias.
func getAlert(alertClient *alert.Client, title string) (string, error) {
	// _, title, _ := parseEventKeyTags(event)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	fmt.Printf("Checking for alert %s \n", title)
	getResult, err := alertClient.Get(ctx, &alert.GetAlertRequest{
		IdentifierType:  alert.ALIAS,
		IdentifierValue: title,
	})
	if err != nil {
		return notFound, nil
	}
	fmt.Printf("ID: %s, Message: %s, Count: %d \n", getResult.Id, getResult.Message, getResult.Count)
	return getResult.Id, nil
}

// closeAlert func close an alert if status == 0
func closeAlert(alertClient *alert.Client, event *types.Event, alertid string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	notes := fmt.Sprintf("Closed Automatically\n %s", event.Check.Output)
	closeResult, err := alertClient.Close(ctx, &alert.CloseAlertRequest{
		IdentifierType:  alert.ALERTID,
		IdentifierValue: alertid,
		Source:          source,
		Note:            notes,
	})
	if err != nil {
		fmt.Printf("[ERROR] Not Closed: %s \n", err)
	}
	fmt.Printf("RequestID %s to Close %s \n", alertid, closeResult.RequestId)

	return nil
}

// getNote func creates a note with whole event in json format
func getNote(event *types.Event) (string, error) {
	eventJSON, err := json.Marshal(event)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("Event data update:\n\n%s", eventJSON), nil
}

// time func returns only the first n bytes of a string
func trim(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}

func titlePrettify(s string) string {
	var title string
	title = strings.ReplaceAll(s, "-", " ")
	title = strings.ReplaceAll(title, "\\", " ")
	title = strings.ReplaceAll(title, "/", " ")
	title = strings.Title(title)
	return title
}

// updateAlert func update alert with status == 0
func updateAlert(alertClient *alert.Client, notes string, alertid string, details map[string]string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if len(details) != 0 {
		// update with details and sensu source url
		updateAlert, err := alertClient.AddDetails(ctx, &alert.AddDetailsRequest{
			IdentifierType:  alert.ALERTID,
			IdentifierValue: alertid,
			Source:          source,
			Note:            notes,
			Details:         details,
		})
		if err != nil {
			fmt.Printf("Not updated: %s \n", err)
		}
		fmt.Printf("RequestID with details %s to update %s \n", alertid, updateAlert.RequestId)
	} else {
		// update without details and just add check.output to notes
		updateAlert, err := alertClient.AddNote(ctx, &alert.AddNoteRequest{
			IdentifierType:  alert.ALERTID,
			IdentifierValue: alertid,
			Source:          source,
			Note:            notes,
		})
		if err != nil {
			fmt.Printf("Not updated: %s \n", err)
		}
		fmt.Printf("RequestID %s to update %s \n", alertid, updateAlert.RequestId)
	}
	return nil
}

func pingHeartbeat(name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	heartbeatClient, err := heartbeat.NewClient(&client.Config{
		ApiKey:         plugin.AuthToken,
		OpsGenieAPIURL: switchOpsgenieRegion(),
	})
	if err != nil {
		return err
	}

	hearbeatResult, err := heartbeatClient.Ping(ctx, name)
	if err != nil {
		return err
	}
	fmt.Printf("Heartbeat %s was requested with %s and response time %v and message %s", name, hearbeatResult.RequestId, hearbeatResult.ResponseTime, hearbeatResult.Message)

	return nil
}

func parseHeartbeatMap(s string) (map[string]string, error) {
	// entity1/check1=heartbeat1,entity2/check2=heartbeat2
	heartbeats := make(map[string]string)
	temp := makeMap(s)
	for k, v := range temp {
		if strings.Contains(v, "/") {
			return heartbeats, fmt.Errorf("hearbeat wrong format: entity/check=heartbeat_name")
		}
		if k != "" && v != "" {
			key := k
			if !strings.Contains(k, "/") {
				key = "all"
			}
			heartbeats[key] = v
		}
	}
	return heartbeats, nil
}

func makeMap(s string) map[string]string {
	temp := make(map[string]string)
	if strings.Contains(s, ",") {
		splited := strings.Split(s, ",")
		for _, v := range splited {
			a, b := splitString(v, "=")
			if a != "" && b != "" {
				temp[a] = b
			}
		}
	} else {
		a, b := splitString(s, "=")
		if a != "" && b != "" {
			temp[a] = b
		}
	}
	return temp
}

func splitString(s, div string) (string, string) {
	if div != "" {
		splited := strings.Split(s, div)
		if len(splited) == 2 {
			return splited[0], splited[1]
		}
	}
	return "", ""
}

func splitStringInSlice(s string) (stringList []string) {
	if strings.Contains(s, ",") {
		if strings.Contains(s, ",") {
			tmpList := strings.Split(s, ",")
			if len(tmpList) != 0 {
				for _, v := range tmpList {
					if v != "" {
						stringList = append(stringList, v)
					}
				}
			}
		}
	} else {
		stringList = []string{s}
	}
	return stringList
}
