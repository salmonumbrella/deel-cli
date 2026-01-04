package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersonUnmarshalJSON_NameComputation(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		expectedName string
	}{
		{
			name:         "both first and last name present",
			json:         `{"first_name": "John", "last_name": "Doe"}`,
			expectedName: "John Doe",
		},
		{
			name:         "only first name present",
			json:         `{"first_name": "John", "last_name": ""}`,
			expectedName: "John",
		},
		{
			name:         "only last name present",
			json:         `{"first_name": "", "last_name": "Doe"}`,
			expectedName: "Doe",
		},
		{
			name:         "both names empty",
			json:         `{"first_name": "", "last_name": ""}`,
			expectedName: "",
		},
		{
			name:         "first name only - last name missing from JSON",
			json:         `{"first_name": "Jane"}`,
			expectedName: "Jane",
		},
		{
			name:         "last name only - first name missing from JSON",
			json:         `{"last_name": "Smith"}`,
			expectedName: "Smith",
		},
		{
			name:         "neither name in JSON",
			json:         `{"id": "123"}`,
			expectedName: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var person Person
			err := json.Unmarshal([]byte(tt.json), &person)
			require.NoError(t, err)
			assert.Equal(t, tt.expectedName, person.Name)
		})
	}
}
