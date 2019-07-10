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

func eventKey(event *types.Event) string {
	return fmt.Sprintf("%s/%s", event.Entity.Name, event.Check.Name)
}

func eventPriority(event *types.Event) alertsv2.Priority {
	if event.Entity.Annotations != nil {
		m := make(map[string]string)
		m = event.Entity.Annotations
		switch m["opsgenie_priority"] {
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

	return manageIncident(event)
}

func manageIncident(event *types.Event) error {

	cli := new(ogcli.OpsGenieClient)
	cli.SetAPIKey(authToken)

	alertCli, _ := cli.AlertV2()

	teams := []alerts.TeamRecipient{
		&alerts.Team{Name: team},
	}

	request := alerts.CreateAlertRequest{
		Message:     eventKey(event),
		Alias:       event.Check.Name,
		Description: event.Check.Output,
		Teams:       teams,
		Entity:      event.Entity.Name,
		Source:      "sensu",
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
