package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

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
	authToken string
	team      string
	stdin     *os.File
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

	return cmd
}

// eventKey func return Entity.Name/Check.Name to use in message and alias
func eventKey(event *types.Event) string {
	return fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
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

	// check if event has a alert
	hasAlert, _ := getAlert(event)
	// fmt.Printf("Has Alert: %s \n", hasAlert)
	if hasAlert == notFound && event.Check.Status != 0 {
		return createIncident(event)
	}

	if hasAlert != notFound && event.Check.Status == 0 {
		return closeAlert(event, hasAlert)
	}
	if event.Check.Status != 0 {
		return addNote(event, hasAlert)
	}
	return nil
}

// createIncident func create an alert in OpsGenie
func createIncident(event *types.Event) error {

	cli := new(ogcli.OpsGenieClient)
	cli.SetAPIKey(authToken)

	alertCli, _ := cli.AlertV2()

	teams := []alerts.TeamRecipient{
		&alerts.Team{Name: team},
	}

	request := alerts.CreateAlertRequest{
		Message:     eventKey(event),
		Alias:       eventKey(event),
		Description: event.Check.Output,
		Teams:       teams,
		Entity:      event.Entity.Name,
		Source:      source,
		Priority:    eventPriority(event),
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
func getAlert(event *types.Event) (string, error) {

	cli := new(ogcli.OpsGenieClient)
	cli.SetAPIKey(authToken)

	alertCli, _ := cli.AlertV2()

	response, err := alertCli.Get(alerts.GetAlertRequest{
		Identifier: &alerts.Identifier{
			Alias: eventKey(event),
		},
	})

	if err != nil {
		return notFound, nil
	}
	alert := response.Alert
	fmt.Printf("ID: %s, Message: %s ,Count: %d \n", alert.ID, alert.Message, alert.Count)
	return alert.ID, nil
}

// closeAlert func close an alert if status == 0
func closeAlert(event *types.Event, alertid string) error {
	cli := new(ogcli.OpsGenieClient)
	cli.SetAPIKey(authToken)

	alertCli, _ := cli.AlertV2()

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
func addNote(event *types.Event, alertid string) error {
	cli := new(ogcli.OpsGenieClient)
	cli.SetAPIKey(authToken)

	alertCli, _ := cli.AlertV2()

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
