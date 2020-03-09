package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/opsgenie/opsgenie-go-sdk/alertsv2"
	alerts "github.com/opsgenie/opsgenie-go-sdk/alertsv2"
	ogcli "github.com/opsgenie/opsgenie-go-sdk/client"
	"github.com/sensu/sensu-go/types"
	"github.com/spf13/cobra"
)

const (
	notFound = "NOT FOUND"
	source   = "sensuGo"
)

var (
	authToken      string
	team           string
	apiURL         string
	annotations    string
	sensuDashboard string
	stdin          *os.File
)

func main() {
	rootCmd := configureRootCommand()
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func configureRootCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sensu-opsgenie-handler",
		Short: "The Sensu Go OpsGenie handler for incident management",
		RunE:  run,
	}

	/*
	   Security Sensitive flags
	     - default to using envvar value
	     - do not mark as required
	     - manually test for empty value
	*/
	cmd.Flags().StringVarP(&authToken,
		"auth",
		"a",
		os.Getenv("OPSGENIE_AUTHTOKEN"),
		"The OpsGenie V2 API authentication token, use default from OPSGENIE_AUTHTOKEN env var")

	cmd.Flags().StringVarP(&team,
		"team",
		"t",
		os.Getenv("OPSGENIE_TEAM"),
		"The OpsGenie V2 API Team, use default from OPSGENIE_TEAM env var")

	cmd.Flags().StringVarP(&apiURL,
		"url",
		"u",
		os.Getenv("OPSGENIE_APIURL"),
		"The OpsGenie V2 API URL, use default from OPSGENIE_APIURL env var")

	cmd.Flags().StringVarP(&annotations,
		"withAnnotations",
		"w",
		os.Getenv("OPSGENIE_ANNOTATIONS"),
		"The OpsGenie Handler will parse check and entity annotations with these values. Use OPSGENIE_ANNOTATIONS env var with commas, like: documentation,playbook")

	cmd.Flags().StringVar(&sensuDashboard,
		"sensuDashboard",
		os.Getenv("OPSGENIE_SENSU_DASHBOARD"),
		"The OpsGenie Handler will use it to create a source Sensu Dashboard URL. Use OPSGENIE_SENSU_DASHBOARD. Example: http://sensu-dashboard.example.local/c/~/n")

	return cmd
}

// // eventKey func return Entity.Name/Check.Name to use in message and alias
// func eventKey(event *types.Event) string {
// 	return fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
// }

// // eventTags func return Entity.Name Check.Name Entity.Namespace, event.Entity.EntityClass to use as tags in Opsgenie
// func eventTags(event *types.Event) (tags []string) {
// 	tags = append(tags, event.Entity.Name, event.Check.Name, event.Entity.Namespace, event.Entity.EntityClass)
// 	return tags
// }

// parseEventKeyTags func return string and []string with event data
// string contains Entity.Name/Check.Name to use in message and alias
// []string contains Entity.Name Check.Name Entity.Namespace, event.Entity.EntityClass to use as tags in Opsgenie
func parseEventKeyTags(event *types.Event) (title string, tags []string) {
	title = fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
	tags = append(tags, event.Entity.Name, event.Check.Name, event.Entity.Namespace, event.Entity.EntityClass)
	return title, tags
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
	if annotations == "" {
		annotations = "documentation,playbook"
	}
	tags := strings.Split(annotations, ",")
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
	if sensuDashboard != "disabled" {
		output += fmt.Sprintf("source: %s/%s/events/%s/%s \n", sensuDashboard, event.Entity.Namespace, event.Entity.Name, event.Check.Name)
	}
	output += fmt.Sprintf("check output: %s", event.Check.Output)
	return output
}

// run func do everything
func run(cmd *cobra.Command, args []string) error {
	if len(args) != 0 {
		_ = cmd.Help()
		return fmt.Errorf("invalid argument(s) received")
	}

	if authToken == "" {
		_ = cmd.Help()
		return fmt.Errorf("authentication token is empty")

	}

	if team == "" {
		_ = cmd.Help()
		return fmt.Errorf("team is empty")

	}

	if apiURL == "" {
		apiURL = "https://api.opsgenie.com"
	}

	if sensuDashboard == "" {
		sensuDashboard = "disabled"
	}

	if stdin == nil {
		stdin = os.Stdin
	}

	eventJSON, err := ioutil.ReadAll(stdin)
	if err != nil {
		return fmt.Errorf("failed to read stdin: %s", err)
	}

	event := &types.Event{}
	err = json.Unmarshal(eventJSON, event)
	if err != nil {
		return fmt.Errorf("failed to unmarshal stdin data: %s", err)
	}

	if err = event.Validate(); err != nil {
		return fmt.Errorf("failed to validate event: %s", err)
	}

	if !event.HasCheck() {
		return fmt.Errorf("event does not contain check")
	}

	// starting client instance
	cli := new(ogcli.OpsGenieClient)
	cli.SetAPIKey(authToken)
	cli.SetOpsGenieAPIUrl(strings.TrimSuffix(apiURL, "/"))
	alertCli, _ := cli.AlertV2()

	// check if event has a alert
	hasAlert, _ := getAlert(alertCli, event)
	// fmt.Printf("Has Alert: %s \n", hasAlert)
	if hasAlert == notFound && event.Check.Status != 0 {
		return createIncident(alertCli, event)
	}

	if hasAlert != notFound && event.Check.Status == 0 {
		return closeAlert(alertCli, event, hasAlert)
	}
	if event.Check.Status != 0 {
		return addNote(alertCli, event, hasAlert)
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
		&alerts.Team{Name: team},
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

// addNode func add a note inside an alert if status != 0
func addNote(alertCli *ogcli.OpsGenieAlertV2Client, event *types.Event, alertid string) error {

	note := fmt.Sprintf("Last Check Status: %s, Output: %s", event.Check.State, event.Check.Output)

	request := alerts.AddNoteRequest{
		Identifier: &alerts.Identifier{
			ID: alertid,
		},
		Source: source,
		Note:   note,
	}

	response, err := alertCli.AddNote(request)

	if err != nil {
		fmt.Println(err.Error())
	}
	fmt.Println("RequestID: " + response.RequestID)
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
