package inspection_test

import (
	"testing"
)

// TestPhysicalVerificationItemStatus verifies business rules for physical matching
func TestPhysicalVerificationItemStatus(t *testing.T) {
	tests := []struct {
		name        string
		expectedQty int
		actualQty   int
		itemFound   bool
		wantStatus  string
	}{
		{
			name:        "Exact Match 1:1",
			expectedQty: 2,
			actualQty:   2,
			itemFound:   true,
			wantStatus:  "MATCH",
		},
		{
			name:        "Quantity Mismatch Under",
			expectedQty: 3,
			actualQty:   1,
			itemFound:   true,
			wantStatus:  "MISMATCH",
		},
		{
			name:        "Quantity Mismatch Over",
			expectedQty: 1,
			actualQty:   2,
			itemFound:   true,
			wantStatus:  "MISMATCH",
		},
		{
			name:        "Unregistered Item Added At Gate",
			expectedQty: 0,
			actualQty:   1,
			itemFound:   false,
			wantStatus:  "UNREGISTERED",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := "MATCH"
			if !tt.itemFound || tt.expectedQty == 0 {
				status = "UNREGISTERED"
			} else if tt.actualQty != tt.expectedQty {
				status = "MISMATCH"
			}

			if status != tt.wantStatus {
				t.Errorf("got %q, want %q", status, tt.wantStatus)
			}
		})
	}
}

// TestOverallInspectionStatus evaluates inspection level matching
func TestOverallInspectionStatus(t *testing.T) {
	tests := []struct {
		name       string
		itemStates []string
		wantStatus string
	}{
		{
			name:       "All items match",
			itemStates: []string{"MATCH", "MATCH", "MATCH"},
			wantStatus: "MATCH",
		},
		{
			name:       "One item mismatch triggers MISMATCH",
			itemStates: []string{"MATCH", "MISMATCH", "MATCH"},
			wantStatus: "MISMATCH",
		},
		{
			name:       "Unregistered item triggers MISMATCH",
			itemStates: []string{"MATCH", "UNREGISTERED"},
			wantStatus: "MISMATCH",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			overall := "MATCH"
			for _, s := range tt.itemStates {
				if s != "MATCH" {
					overall = "MISMATCH"
					break
				}
			}

			if overall != tt.wantStatus {
				t.Errorf("got %q, want %q", overall, tt.wantStatus)
			}
		})
	}
}
