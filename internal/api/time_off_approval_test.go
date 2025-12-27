package api

import (
	"context"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestApproveRejectTimeOff(t *testing.T) {
	tests := []struct {
		name           string
		params         ApproveRejectParams
		serverResponse any
		serverStatus   int
		wantErr        bool
		errContains    string
		wantStatus     string
	}{
		{
			name: "approve request success",
			params: ApproveRejectParams{
				RequestID: "req_123",
				Action:    "approve",
				Comment:   "Approved for vacation",
			},
			serverResponse: map[string]any{
				"data": map[string]any{
					"request_id": "req_123",
					"status":     "approved",
					"comment":    "Approved for vacation",
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantStatus:   "approved",
		},
		{
			name: "reject request success",
			params: ApproveRejectParams{
				RequestID: "req_456",
				Action:    "reject",
				Comment:   "Insufficient coverage",
			},
			serverResponse: map[string]any{
				"data": map[string]any{
					"request_id": "req_456",
					"status":     "rejected",
					"comment":    "Insufficient coverage",
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantStatus:   "rejected",
		},
		{
			name: "invalid action",
			params: ApproveRejectParams{
				RequestID: "req_789",
				Action:    "invalid",
			},
			serverStatus: http.StatusBadRequest,
			wantErr:      true,
		},
		{
			name: "request not found",
			params: ApproveRejectParams{
				RequestID: "req_notfound",
				Action:    "approve",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mockServer(t, http.MethodPost, "/rest/v2/time-off-requests/approve-reject", tt.serverStatus, tt.serverResponse)
			defer server.Close()

			client := testClient(server)
			result, err := client.ApproveRejectTimeOff(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				if tt.errContains != "" {
					assert.Contains(t, err.Error(), tt.errContains)
				}
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.params.RequestID, result.RequestID)
			assert.Equal(t, tt.wantStatus, result.Status)
		})
	}
}

func TestValidateTimeOffRequest(t *testing.T) {
	tests := []struct {
		name           string
		params         ValidateTimeOffParams
		serverResponse any
		serverStatus   int
		wantErr        bool
		wantValid      bool
	}{
		{
			name: "valid request",
			params: ValidateTimeOffParams{
				ProfileID: "profile_123",
				Type:      "vacation",
				StartDate: "2024-06-01",
				EndDate:   "2024-06-05",
			},
			serverResponse: map[string]any{
				"data": map[string]any{
					"valid":    true,
					"errors":   []string{},
					"warnings": []string{},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantValid:    true,
		},
		{
			name: "invalid with errors",
			params: ValidateTimeOffParams{
				ProfileID: "profile_123",
				Type:      "vacation",
				StartDate: "2024-06-01",
				EndDate:   "2024-06-30",
			},
			serverResponse: map[string]any{
				"data": map[string]any{
					"valid": false,
					"errors": []string{
						"Insufficient balance",
						"Exceeds maximum consecutive days",
					},
					"warnings": []string{},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantValid:    false,
		},
		{
			name: "valid with warnings",
			params: ValidateTimeOffParams{
				ProfileID: "profile_123",
				Type:      "vacation",
				StartDate: "2024-12-20",
				EndDate:   "2024-12-27",
			},
			serverResponse: map[string]any{
				"data": map[string]any{
					"valid":  true,
					"errors": []string{},
					"warnings": []string{
						"Overlaps with holiday period",
					},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantValid:    true,
		},
		{
			name: "profile not found",
			params: ValidateTimeOffParams{
				ProfileID: "invalid_profile",
				Type:      "vacation",
				StartDate: "2024-06-01",
				EndDate:   "2024-06-05",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mockServer(t, http.MethodPost, "/rest/v2/time-off-requests/validate", tt.serverStatus, tt.serverResponse)
			defer server.Close()

			client := testClient(server)
			result, err := client.ValidateTimeOffRequest(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.wantValid, result.Valid)
		})
	}
}

func TestGetWorkSchedule(t *testing.T) {
	tests := []struct {
		name           string
		profileID      string
		serverResponse any
		serverStatus   int
		wantErr        bool
	}{
		{
			name:      "get work schedule success",
			profileID: "profile_123",
			serverResponse: map[string]any{
				"data": map[string]any{
					"profile_id":    "profile_123",
					"work_days":     []string{"monday", "tuesday", "wednesday", "thursday", "friday"},
					"hours_per_day": 8.0,
					"start_time":    "09:00",
					"end_time":      "17:00",
					"timezone":      "America/New_York",
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name:         "profile not found",
			profileID:    "invalid_profile",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedPath := "/rest/v2/profiles/" + tt.profileID + "/work-schedule"
			server := mockServer(t, http.MethodGet, expectedPath, tt.serverStatus, tt.serverResponse)
			defer server.Close()

			client := testClient(server)
			result, err := client.GetWorkSchedule(context.Background(), tt.profileID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Equal(t, tt.profileID, result.ProfileID)
			assert.NotEmpty(t, result.WorkDays)
		})
	}
}

func TestGetEntitlements(t *testing.T) {
	tests := []struct {
		name           string
		profileID      string
		serverResponse any
		serverStatus   int
		wantErr        bool
		wantCount      int
	}{
		{
			name:      "get entitlements success",
			profileID: "profile_123",
			serverResponse: map[string]any{
				"data": []map[string]any{
					{
						"id":           "ent_1",
						"type":         "vacation",
						"total_days":   20.0,
						"used_days":    5.0,
						"pending_days": 2.0,
						"balance":      13.0,
						"year":         2024,
					},
					{
						"id":           "ent_2",
						"type":         "sick",
						"total_days":   10.0,
						"used_days":    1.0,
						"pending_days": 0.0,
						"balance":      9.0,
						"year":         2024,
					},
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    2,
		},
		{
			name:      "no entitlements",
			profileID: "profile_new",
			serverResponse: map[string]any{
				"data": []map[string]any{},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
			wantCount:    0,
		},
		{
			name:         "profile not found",
			profileID:    "invalid_profile",
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expectedPath := "/rest/v2/profiles/" + tt.profileID + "/entitlements"
			server := mockServer(t, http.MethodGet, expectedPath, tt.serverStatus, tt.serverResponse)
			defer server.Close()

			client := testClient(server)
			result, err := client.GetEntitlements(context.Background(), tt.profileID)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, result)
			assert.Len(t, result, tt.wantCount)
		})
	}
}

func TestSyncExternalTimeOff(t *testing.T) {
	tests := []struct {
		name           string
		params         SyncTimeOffParams
		serverResponse any
		serverStatus   int
		wantErr        bool
	}{
		{
			name: "sync external time off success",
			params: SyncTimeOffParams{
				ProfileID:  "profile_123",
				ExternalID: "ext_456",
				Type:       "vacation",
				StartDate:  "2024-07-01",
				EndDate:    "2024-07-05",
				Status:     "approved",
			},
			serverResponse: map[string]any{
				"data": map[string]any{
					"id":          "req_789",
					"external_id": "ext_456",
					"status":      "synced",
				},
			},
			serverStatus: http.StatusOK,
			wantErr:      false,
		},
		{
			name: "sync with conflict",
			params: SyncTimeOffParams{
				ProfileID:  "profile_123",
				ExternalID: "ext_conflict",
				Type:       "vacation",
				StartDate:  "2024-07-01",
				EndDate:    "2024-07-05",
				Status:     "approved",
			},
			serverStatus: http.StatusConflict,
			wantErr:      true,
		},
		{
			name: "invalid profile",
			params: SyncTimeOffParams{
				ProfileID:  "invalid",
				ExternalID: "ext_456",
				Type:       "vacation",
				StartDate:  "2024-07-01",
				EndDate:    "2024-07-05",
				Status:     "approved",
			},
			serverStatus: http.StatusNotFound,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := mockServerWithBody(t, http.MethodPost, "/rest/v2/time-off/sync", func(t *testing.T, body map[string]any) {
				assert.Equal(t, tt.params.ProfileID, body["profile_id"])
				assert.Equal(t, tt.params.ExternalID, body["external_id"])
			}, tt.serverStatus, tt.serverResponse)
			defer server.Close()

			client := testClient(server)
			err := client.SyncExternalTimeOff(context.Background(), tt.params)

			if tt.wantErr {
				require.Error(t, err)
				return
			}

			require.NoError(t, err)
		})
	}
}
