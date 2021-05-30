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

func TestSplitString(t *testing.T) {
	test1 := "key1=value1"
	res1 := "key1"
	res2 := "value1"
	val1, val2 := splitString(test1, "=")
	assert.Equal(t, res1, val1)
	assert.Equal(t, res2, val2)
	assert.NotEqual(t, res1, val2)
}

func TestMakeRewriteAnnotation(t *testing.T) {
	test1 := "key1=value1,key2=value2"
	res1 := map[string]string{"key1": "value1", "key2": "value2"}
	val1 := makeMap(test1)
	assert.Equal(t, res1, val1)
	test2 := "key1=value1,key2/subkey2=value2"
	res2 := map[string]string{"key1": "value1", "key2/subkey2": "value2"}
	val2 := makeMap(test2)
	assert.Equal(t, res2, val2)
	test3 := "key1=value1,key2=value2,"
	res3 := map[string]string{"key1": "value1", "key2": "value2"}
	val3 := makeMap(test3)
	assert.Equal(t, res3, val3)
}

func TestParseHeartbeatMap(t *testing.T) {
	test1 := "entity1/check1=heartbeat1,entity2/check2=heartbeat2"
	expected1 := map[string]string{"entity1/check1": "heartbeat1", "entity2/check2": "heartbeat2"}
	res1, err1 := parseHeartbeatMap(test1)
	assert.Equal(t, expected1, res1)
	assert.NoError(t, err1)

	test2 := "entity1/check1=heartbeat1,"
	expected2 := map[string]string{"entity1/check1": "heartbeat1"}
	res2, err2 := parseHeartbeatMap(test2)
	assert.Equal(t, expected2, res2)
	assert.NoError(t, err2)

	test3 := "heartbeat1=entity1/check1,"
	_, err3 := parseHeartbeatMap(test3)
	assert.Error(t, err3)

}

func TestSplitStringInSlice(t *testing.T) {
	s1 := "test"
	res1 := splitStringInSlice(s1)
	expected1 := []string{"test"}
	assert.Equal(t, expected1, res1)
	s2 := "test,"
	res2 := splitStringInSlice(s2)
	expected2 := []string{"test"}
	assert.Equal(t, expected2, res2)
	s3 := "test,test1"
	res3 := splitStringInSlice(s3)
	expected3 := []string{"test", "test1"}
	assert.Equal(t, expected3, res3)
	s4 := ","
	res4 := splitStringInSlice(s4)
	var expected4 []string
	expected5 := 0
	assert.Equal(t, expected4, res4)
	assert.Equal(t, expected5, len(res4))
}
