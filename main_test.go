package main

import (
	"encoding/json"
	"testing"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	"github.com/sensu/sensu-go/types"
	"github.com/stretchr/testify/assert"
)

func TestGetNote(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	eventJSON, err := json.Marshal(event)
	assert.NoError(t, err)
	note, err := getNote(event)
	assert.NoError(t, err)
	assert.Contains(t, note, "Event data update:\n\n")
	assert.Contains(t, note, string(eventJSON))
}

func TestEventPriority(t *testing.T) {
	plugin.Priority = "P1"
	priority := eventPriority()
	expectedValue := alert.P1
	assert.Contains(t, priority, expectedValue)
}

func TestParseActions(t *testing.T) {
	event1 := types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "test-check",
				Annotations: map[string]string{
					"opsgenie_priority": "P1",
					"opsgenie_actions":  "workaround",
				},
			},
			Output: "test output",
		},
	}
	test1 := parseActions(&event1)
	expectedValue1 := "workaround"
	assert.Contains(t, test1, expectedValue1)

	event2 := types.Event{
		Entity: &types.Entity{
			ObjectMeta: types.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
		Check: &types.Check{
			ObjectMeta: types.ObjectMeta{
				Name: "test-check",
				Annotations: map[string]string{
					"opsgenie_priority": "P1",
					"opsgenie_actions":  "workaround,bigrestart",
				},
			},
			Output: "test output",
		},
	}
	test2 := parseActions(&event2)
	expectedValue2 := "workaround"
	assert.Contains(t, test2, expectedValue2)
	expectedValue2a := "bigrestart"
	assert.Contains(t, test2, expectedValue2a)
}

func TestSwitchOpsgenieRegion(t *testing.T) {
	expectedValueUS := client.API_URL
	expectedValueEU := client.API_URL_EU

	testUS := switchOpsgenieRegion()

	assert.Equal(t, testUS, expectedValueUS)

	plugin.APIRegion = "eu"

	testEU := switchOpsgenieRegion()

	assert.Equal(t, testEU, expectedValueEU)

	plugin.APIRegion = "EU"

	testEU2 := switchOpsgenieRegion()

	assert.Equal(t, testEU2, expectedValueEU)
}

func TestParseDetails(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	event.Check.Output = "Check OK"
	event.Check.State = "passing"
	event.Check.Status = 0
	_, err := json.Marshal(event)
	assert.NoError(t, err)
	plugin.FullDetails = true
	det := parseDetails(event)
	assert.Equal(t, det["output"], "Check OK")
	assert.Equal(t, det["state"], "passing")
	assert.Equal(t, det["status"], "0")
}

func TestParseEventKeyTags(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	_, err := json.Marshal(event)
	assert.NoError(t, err)
	plugin.MessageTemplate = "{{.Entity.Name}}/{{.Check.Name}}"
	plugin.MessageLimit = 100
	plugin.TagsTemplates = []string{"{{.Entity.Name}}", "{{.Check.Name}}", "{{.Entity.Namespace}}", "{{.Entity.EntityClass}}"}
	title, alias, tags := parseEventKeyTags(event)
	assert.Contains(t, title, "foo")
	assert.Contains(t, alias, "foo")
	assert.Contains(t, tags, "foo")
}

func TestParseDescription(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	event.Check.Output = "Check OK"
	_, err := json.Marshal(event)
	assert.NoError(t, err)
	plugin.DescriptionTemplate = "{{.Check.Output}}"
	plugin.DescriptionLimit = 100
	description := parseDescription(event)
	assert.Equal(t, description, "Check OK")
}

func TestCheckArgs(t *testing.T) {
	assert := assert.New(t)
	event := types.FixtureEvent("entity1", "check1")
	assert.Error(checkArgs(event))
	plugin.AuthToken = "Testing"
	assert.NoError(checkArgs(event))
}

func TestTrim(t *testing.T) {
	testString := "This string is 33 characters long"
	assert.Equal(t, trim(testString, 40), testString)
	assert.Equal(t, trim(testString, 4), "This")
}

func TestTitlePrettify(t *testing.T) {
	test1 := "long-check-with-too-many-dashes"
	res1 := titlePrettify(test1)
	val1 := "Long Check With Too Many Dashes"
	assert.Equal(t, val1, res1)
	test2 := "long-check-with-too-many-dashes/and/slashes-and\\others"
	res2 := titlePrettify(test2)
	val2 := "Long Check With Too Many Dashes And Slashes And Others"
	assert.Equal(t, val2, res2)
}

func TestRespondersTeam(t *testing.T) {
	test1 := []alert.Responder{
		{Type: alert.TeamResponder, Name: "ops"},
	}
	plugin.Team = "ops"
	res1 := respondersTeam()
	assert.Equal(t, res1, test1)
}

func TestVisibilityTeams(t *testing.T) {
	test1 := []alert.Responder{
		{Type: alert.TeamResponder, Name: "ops"},
	}
	plugin.VisibilityTeams = "ops,"
	res1 := visibilityTeams()
	assert.Equal(t, res1, test1)
}
