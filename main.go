package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/opsgenie/opsgenie-go-sdk/alertsv2"
	alerts "github.com/opsgenie/opsgenie-go-sdk/alertsv2"
	ogcli "github.com/opsgenie/opsgenie-go-sdk/client"
	"github.com/sensu-community/sensu-plugin-sdk/sensu"
	"github.com/sensu/sensu-go/types"
)

const (
	notFound = "NOT FOUND"
	source   = "sensuGo"
)

// Config represents the handler plugin config.
type Config struct {
	sensu.PluginConfig
	APIURL          string
	AuthToken       string
	Team            string
	Annotations     string
	SensuDashboard  string
	MessageTemplate string
	MessageLimit	int
}

// used to handle getting text/template or html/template
type templater interface {
	Execute(wr io.Writer, data interface{}) error
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
			Path:      "url",
			Env:       "OPSGENIE_APIURL",
			Argument:  "url",
			Shorthand: "u",
			Default:   "https://api.opsgenie.com",
			Usage:     "The OpsGenie V2 API URL, use default from OPSGENIE_APIURL env var",
			Value:     &plugin.APIURL,
		},
		{
			Path:      "auth",
			Env:       "OPSGENIE_AUTHTOKEN",
			Argument:  "auth",
			Shorthand: "a",
			Default:   "",
			Usage:     "The OpsGenie V2 API authentication token, use default from OPSGENIE_AUTHTOKEN env var",
			Value:     &plugin.AuthToken,
		},
		{
			Path:      "team",
			Env:       "OPSGENIE_TEAM",
			Argument:  "team",
			Shorthand: "t",
			Default:   "",
			Usage:     "The OpsGenie V2 API Team, use default from OPSGENIE_TEAM env var",
			Value:     &plugin.Team,
		},
		{
			Path:      "withAnnotations",
			Env:       "OPSGENIE_ANNOTATIONS",
			Argument:  "withAnnotations",
			Shorthand: "w",
			Default:   "documentation,playbook",
			Usage:     "The OpsGenie Handler will parse check and entity annotations with these values. Use OPSGENIE_ANNOTATIONS env var with commas, like: documentation,playbook",
			Value:     &plugin.Annotations,
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
			Default:   100,
			Usage:     "The maximum length of the message field",
			Value:     &plugin.MessageLimit,
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
	if len(plugin.Team) == 0 {
		return fmt.Errorf("team is empty")
	}
	return nil
}

// parseEventKeyTags func return string and []string with event data
// string contains Entity.Name/Check.Name to use in message and alias
// []string contains Entity.Name Check.Name Entity.Namespace, event.Entity.EntityClass to use as tags in Opsgenie
func parseEventKeyTags(event *types.Event) (title string, tags []string) {
	title, err := resolveTemplate(plugin.MessageTemplate, event)
	if err != nil {
		return "", []string{}
	}
	tags = append(tags, event.Entity.Name, event.Check.Name, event.Entity.Namespace, event.Entity.EntityClass)
	return trim(title, plugin.MessageLimit), tags
}

// eventPriority func read priority in the event and return alerts.PX
// check.Annotations override Entity.Annotations
func eventPriority(event *types.Event) alertsv2.Priority {
	if event.Check.Annotations != nil {
		switch event.Check.Annotations["opsgenie_priority"] {
		case "P5":
			return alerts.P5

		case "P4":
			return alerts.P4

		case "P3":
			return alerts.P3

		case "P2":
			return alerts.P2

		case "P1":
			return alerts.P1

		default:
			return alerts.P3

		}
	} else if event.Entity.Annotations != nil {
		switch event.Entity.Annotations["opsgenie_priority"] {
		case "P5":
			return alerts.P5

		case "P4":
			return alerts.P4

		case "P3":
			return alerts.P3

		case "P2":
			return alerts.P2

		case "P1":
			return alerts.P1

		default:
			return alerts.P3

		}
	} else {
		return alerts.P3
	}

}

// stringInSlice checks if a slice contains a specific string
func stringInSlice(a string, list []string) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

// parseAnnotations func try to find a predeterminated keys
func parseAnnotations(event *types.Event) string {
	var output string
	tags := strings.Split(plugin.Annotations, ",")
	if event.Check.Annotations != nil {
		for key, value := range event.Check.Annotations {
			if stringInSlice(key, tags) {
				output += fmt.Sprintf("%s: %s \n", key, value)
			}
		}
	}
	if event.Entity.Annotations != nil {
		for key, value := range event.Entity.Annotations {
			if stringInSlice(key, tags) {
				output += fmt.Sprintf("%s: %s \n", key, value)
			}
		}
	}
	if plugin.SensuDashboard != "disabled" {
		output += fmt.Sprintf("source: %s/%s/events/%s/%s \n", plugin.SensuDashboard, event.Entity.Namespace, event.Entity.Name, event.Check.Name)
	}
	output += fmt.Sprintf("check output: %s", event.Check.Output)
	return output
}

func executeHandler(event *types.Event) error {
	// starting client instance
	cli := new(ogcli.OpsGenieClient)
	cli.SetAPIKey(plugin.AuthToken)
	cli.SetOpsGenieAPIUrl(strings.TrimSuffix(plugin.APIURL, "/"))
	alertCli, _ := cli.AlertV2()

	// check if event has a alert
	hasAlert, _ := getAlert(alertCli, event)
	if event.Check.Status != 0 {
		return createIncident(alertCli, event)
	}
	// close incident if status == 0
	if hasAlert != notFound && event.Check.Status == 0 {
		return closeAlert(alertCli, event, hasAlert)
	}

	return nil
}

// createIncident func create an alert in OpsGenie
func createIncident(alertCli *ogcli.OpsGenieAlertV2Client, event *types.Event) error {
	note, err := getNote(event)
	if err != nil {
		return err
	}

	teams := []alerts.TeamRecipient{
		&alerts.Team{Name: plugin.Team},
	}
	title, tags := parseEventKeyTags(event)

	request := alerts.CreateAlertRequest{
		Message:     title,
		Alias:       title,
		Description: parseAnnotations(event),
		Teams:       teams,
		Entity:      event.Entity.Name,
		Source:      source,
		Priority:    eventPriority(event),
		Note:        note,
		Tags:        tags,
	}

	response, err := alertCli.Create(request)
	if err != nil {
		fmt.Println(err.Error())
	} else {
		fmt.Println("Create request ID: " + response.RequestID)
	}
	return nil
}

// getAlert func get a alert using an alias.
func getAlert(alertCli *ogcli.OpsGenieAlertV2Client, event *types.Event) (string, error) {
	title, _ := parseEventKeyTags(event)
	response, err := alertCli.Get(alerts.GetAlertRequest{
		Identifier: &alerts.Identifier{
			Alias: title,
		},
	})

	if err != nil {
		return notFound, nil
	}
	alert := response.Alert
	fmt.Printf("ID: %s, Message: %s, Count: %d \n", alert.ID, alert.Message, alert.Count)
	return alert.ID, nil
}

// closeAlert func close an alert if status == 0
func closeAlert(alertCli *ogcli.OpsGenieAlertV2Client, event *types.Event, alertid string) error {

	identifier := alerts.Identifier{
		ID: alertid,
	}
	closeRequest := alerts.CloseRequest{
		Identifier: &identifier,
		Source:     source,
		Note:       "Closed Automatically",
	}

	response, err := alertCli.Close(closeRequest)

	if err != nil {
		fmt.Printf("[ERROR] Not Closed: %s", err)
	}
	fmt.Printf("RequestID %s to Close %s", alertid, response.RequestID)

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

// resolveTemplate func resolves events in go templates
func resolveTemplate(templateValue string, event *types.Event) (string, error) {
	var (
		resolved bytes.Buffer
		tmpl     templater
		err      error
	)

	tmpl, err = template.New("test").Parse(templateValue)

	if err != nil {
		return "", err
	}

	err = tmpl.Execute(&resolved, *event)
	if err != nil {
		return "", err
	}

	return resolved.String(), nil
}

// time func returns only the first n bytes of a string
func trim(s string, n int) string {
	if len(s) > n {
		return s[:n]
	}
	return s
}
