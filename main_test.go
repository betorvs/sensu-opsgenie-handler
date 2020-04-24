package main

import (
	"encoding/json"
	"testing"

	"github.com/opsgenie/opsgenie-go-sdk-v2/alert"
	"github.com/opsgenie/opsgenie-go-sdk-v2/client"
	v2 "github.com/sensu/sensu-go/api/core/v2"
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

func TestParseEventKeyTags(t *testing.T) {
	event := types.FixtureEvent("foo", "bar")
	_, err := json.Marshal(event)
	assert.NoError(t, err)
	title, tags := parseEventKeyTags(event)
	assert.Contains(t, title, "foo")
	assert.Contains(t, tags, "foo")
}

func TestParseAnnotations(t *testing.T) {
	event := v2.Event{
		Entity: &v2.Entity{
			ObjectMeta: v2.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Annotations: map[string]string{
					"documentation": "no-docs",
				},
			},
		},
		Check: &v2.Check{
			ObjectMeta: v2.ObjectMeta{
				Name: "test-check",
				Annotations: map[string]string{
					"playbook": "no-playbook",
				},
			},
			Output: "test output",
		},
	}
	// annotations := "documentation,playbook"
	description, _ := parseAnnotations(&event)
	assert.Contains(t, description, "documentation")
	assert.Contains(t, description, "playbook")

}

func TestEventPriority(t *testing.T) {
	event := v2.Event{
		Entity: &v2.Entity{
			ObjectMeta: v2.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
		Check: &v2.Check{
			ObjectMeta: v2.ObjectMeta{
				Name: "test-check",
				Annotations: map[string]string{
					"opsgenie_priority": "P1",
				},
			},
			Output: "test output",
		},
	}
	priority := eventPriority(&event)
	expectedValue := alert.P1
	assert.Contains(t, priority, expectedValue)
}

func TestStringInSlice(t *testing.T) {
	testSlice := []string{"foo", "bar", "test"}
	testString := "test"
	testResult := stringInSlice(testString, testSlice)
	assert.True(t, testResult)
}

func TestParseActions(t *testing.T) {
	event1 := v2.Event{
		Entity: &v2.Entity{
			ObjectMeta: v2.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
		Check: &v2.Check{
			ObjectMeta: v2.ObjectMeta{
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

	event2 := v2.Event{
		Entity: &v2.Entity{
			ObjectMeta: v2.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
		Check: &v2.Check{
			ObjectMeta: v2.ObjectMeta{
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

func TestParseAuthTokenAndTeams(t *testing.T) {
	event1 := v2.Event{
		Entity: &v2.Entity{
			ObjectMeta: v2.ObjectMeta{
				Name:      "test",
				Namespace: "default",
			},
		},
		Check: &v2.Check{
			ObjectMeta: v2.ObjectMeta{
				Name: "test-check",
				Annotations: map[string]string{
					"opsgenie_priority":  "P1",
					"opsgenie_authtoken": "aaaa-wwww-sssss-33333-zzzzz",
					"opsgenie_team":      "ops",
				},
			},
			Output: "test output",
		},
	}
	testAuth1 := parseOpsgenieAuthToken(&event1)
	expectedValueAuth1 := "aaaa-wwww-sssss-33333-zzzzz"
	assert.Contains(t, testAuth1, expectedValueAuth1)
	testTeams1 := parseOpsgenieTeams(&event1)
	expectedValueTeams1 := []alert.Responder{
		{Type: alert.EscalationResponder, Name: "ops"},
		{Type: alert.ScheduleResponder, Name: "ops"},
	}
	assert.Equal(t, testTeams1, expectedValueTeams1)

}

func TestSwitchOpsgenieRegion(t *testing.T) {
	expectedValueUS := client.API_URL
	expectedValueEU := client.API_URL_EU

	testUS := switchOpsgenieRegion()

	assert.Equal(t, testUS, expectedValueUS)

	apiRegion = "eu"

	testEU := switchOpsgenieRegion()

	assert.Equal(t, testEU, expectedValueEU)

	apiRegion = "EU"

	testEU2 := switchOpsgenieRegion()

	assert.Equal(t, testEU2, expectedValueEU)
}
