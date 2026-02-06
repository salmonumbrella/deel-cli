package api

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRedactedString_Value(t *testing.T) {
	r := NewRedactedString("secret-token")
	assert.Equal(t, "secret-token", r.Value())
}

func TestRedactedString_String(t *testing.T) {
	r := NewRedactedString("secret-token")
	assert.Equal(t, "[REDACTED]", r.String())
	assert.Equal(t, "[REDACTED]", fmt.Sprintf("%s", r))
	assert.Equal(t, "[REDACTED]", fmt.Sprintf("%v", r))
}

func TestRedactedString_GoString(t *testing.T) {
	r := NewRedactedString("secret-token")
	assert.Equal(t, "RedactedString{[REDACTED]}", fmt.Sprintf("%#v", r))
}

func TestRedactedString_MarshalJSON(t *testing.T) {
	r := NewRedactedString("secret-token")
	data, err := json.Marshal(r)
	assert.NoError(t, err)
	assert.Equal(t, `"[REDACTED]"`, string(data))
}

func TestRedactedString_InStruct(t *testing.T) {
	type container struct {
		Token RedactedString `json:"token"`
	}
	c := container{Token: NewRedactedString("my-secret")}
	data, err := json.Marshal(c)
	assert.NoError(t, err)
	assert.Contains(t, string(data), `"[REDACTED]"`)
	assert.NotContains(t, string(data), "my-secret")
}

func TestRedactedString_ZeroValue(t *testing.T) {
	var r RedactedString
	assert.Equal(t, "", r.Value())
	assert.Equal(t, "[REDACTED]", r.String())
}
