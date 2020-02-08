package main

import (
	"encoding/json"
	"testing"

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
