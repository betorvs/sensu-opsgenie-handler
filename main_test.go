package main

import (
	"encoding/json"
	"testing"

	alerts "github.com/opsgenie/opsgenie-go-sdk/alertsv2"
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
	plugin.Annotations = "documentation,playbook"
	description := parseAnnotations(&event)
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
	expectedValue := alerts.P1
	assert.Contains(t, priority, expectedValue)
}

func TestStringInSlice(t *testing.T) {
	testSlice := []string{"foo", "bar", "test"}
	testString := "test"
	testResult := stringInSlice(testString, testSlice)
	assert.True(t, testResult)
}
