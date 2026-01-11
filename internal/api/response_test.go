package api

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testPerson is a simple struct for testing generic decode functions.
type testPerson struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func TestDecodeJSON_Success(t *testing.T) {
	raw := json.RawMessage(`{"id": "123", "name": "Alice"}`)

	result, err := decodeJSON[testPerson](raw)

	require.NoError(t, err)
	assert.Equal(t, "123", result.ID)
	assert.Equal(t, "Alice", result.Name)
}

func TestDecodeJSON_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`{invalid json}`)

	result, err := decodeJSON[testPerson](raw)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestDecodeJSON_EmptyObject(t *testing.T) {
	raw := json.RawMessage(`{}`)

	result, err := decodeJSON[testPerson](raw)

	require.NoError(t, err)
	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.Name)
}

func TestDecodeJSON_NullValue(t *testing.T) {
	raw := json.RawMessage(`null`)

	result, err := decodeJSON[testPerson](raw)

	require.NoError(t, err)
	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.Name)
}

func TestDecodeData_Success(t *testing.T) {
	raw := json.RawMessage(`{"data": {"id": "456", "name": "Bob"}}`)

	result, err := decodeData[testPerson](raw)

	require.NoError(t, err)
	assert.Equal(t, "456", result.ID)
	assert.Equal(t, "Bob", result.Name)
}

func TestDecodeData_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`{invalid json}`)

	result, err := decodeData[testPerson](raw)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestDecodeData_MissingDataField(t *testing.T) {
	raw := json.RawMessage(`{"id": "789", "name": "Charlie"}`)

	result, err := decodeData[testPerson](raw)

	require.NoError(t, err)
	// When data field is missing, we get zero values
	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.Name)
}

func TestDecodeData_NullData(t *testing.T) {
	raw := json.RawMessage(`{"data": null}`)

	result, err := decodeData[testPerson](raw)

	require.NoError(t, err)
	assert.Equal(t, "", result.ID)
	assert.Equal(t, "", result.Name)
}

func TestDecodeList_Success(t *testing.T) {
	raw := json.RawMessage(`{
		"data": [
			{"id": "1", "name": "Alice"},
			{"id": "2", "name": "Bob"}
		],
		"page": {
			"next": "cursor123",
			"total": 100
		}
	}`)

	result, err := decodeList[testPerson](raw)

	require.NoError(t, err)
	require.Len(t, result.Data, 2)
	assert.Equal(t, "1", result.Data[0].ID)
	assert.Equal(t, "Alice", result.Data[0].Name)
	assert.Equal(t, "2", result.Data[1].ID)
	assert.Equal(t, "Bob", result.Data[1].Name)
	assert.Equal(t, "cursor123", result.Page.Next)
	assert.Equal(t, 100, result.Page.Total)
}

func TestDecodeList_EmptyList(t *testing.T) {
	raw := json.RawMessage(`{"data": [], "page": {"next": "", "total": 0}}`)

	result, err := decodeList[testPerson](raw)

	require.NoError(t, err)
	assert.Empty(t, result.Data)
	assert.Equal(t, "", result.Page.Next)
	assert.Equal(t, 0, result.Page.Total)
}

func TestDecodeList_InvalidJSON(t *testing.T) {
	raw := json.RawMessage(`{invalid json}`)

	result, err := decodeList[testPerson](raw)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to parse response")
}

func TestDecodeList_NoPageField(t *testing.T) {
	raw := json.RawMessage(`{"data": [{"id": "1", "name": "Alice"}]}`)

	result, err := decodeList[testPerson](raw)

	require.NoError(t, err)
	require.Len(t, result.Data, 1)
	assert.Equal(t, "1", result.Data[0].ID)
	// Page defaults to zero values when missing
	assert.Equal(t, "", result.Page.Next)
	assert.Equal(t, 0, result.Page.Total)
}

func TestDecodeList_NullData(t *testing.T) {
	raw := json.RawMessage(`{"data": null, "page": {"next": "abc"}}`)

	result, err := decodeList[testPerson](raw)

	require.NoError(t, err)
	assert.Nil(t, result.Data)
	assert.Equal(t, "abc", result.Page.Next)
}

func TestWrapData_Success(t *testing.T) {
	person := testPerson{ID: "123", Name: "Alice"}

	result := wrapData(person)

	assert.Equal(t, person, result.Data)
}

func TestWrapData_EmptyStruct(t *testing.T) {
	person := testPerson{}

	result := wrapData(person)

	assert.Equal(t, testPerson{}, result.Data)
}

func TestWrapData_WithMap(t *testing.T) {
	data := map[string]string{"key": "value"}

	result := wrapData(data)

	assert.Equal(t, "value", result.Data["key"])
}

func TestWrapData_MarshalJSON(t *testing.T) {
	person := testPerson{ID: "123", Name: "Alice"}

	wrapped := wrapData(person)
	jsonBytes, err := json.Marshal(wrapped)

	require.NoError(t, err)
	assert.JSONEq(t, `{"data": {"id": "123", "name": "Alice"}}`, string(jsonBytes))
}
