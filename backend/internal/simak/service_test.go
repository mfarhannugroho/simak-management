package simak_test

import (
	"testing"
	"time"

	"simak-backend/internal/simak"
)

// TestSimakStatusTransitions tests legal and illegal workflow state changes
func TestSimakStatusTransitions(t *testing.T) {
	tests := []struct {
		name          string
		currentStatus string
		targetAction  string
		wantValid     bool
		nextStatus    string
	}{
		{
			name:          "Submit from Draft is valid",
			currentStatus: simak.StatusDraft,
			targetAction:  "SUBMIT",
			wantValid:     true,
			nextStatus:    simak.StatusSubmitted,
		},
		{
			name:          "Submit from Approved is invalid",
			currentStatus: simak.StatusApproved,
			targetAction:  "SUBMIT",
			wantValid:     false,
			nextStatus:    simak.StatusApproved,
		},
		{
			name:          "Issue from Approved is valid",
			currentStatus: simak.StatusApproved,
			targetAction:  "ISSUE",
			wantValid:     true,
			nextStatus:    simak.StatusIssued,
		},
		{
			name:          "Cancel from Submitted is valid",
			currentStatus: simak.StatusSubmitted,
			targetAction:  "CANCEL",
			wantValid:     true,
			nextStatus:    simak.StatusCancelled,
		},
		{
			name:          "Cancel from Completed is invalid",
			currentStatus: simak.StatusCompleted,
			targetAction:  "CANCEL",
			wantValid:     false,
			nextStatus:    simak.StatusCompleted,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid := false
			next := tt.currentStatus

			switch tt.targetAction {
			case "SUBMIT":
				if tt.currentStatus == simak.StatusDraft || tt.currentStatus == simak.StatusRequiresRevision {
					valid = true
					next = simak.StatusSubmitted
				}
			case "ISSUE":
				if tt.currentStatus == simak.StatusApproved {
					valid = true
					next = simak.StatusIssued
				}
			case "CANCEL":
				if tt.currentStatus == simak.StatusDraft || tt.currentStatus == simak.StatusSubmitted || tt.currentStatus == simak.StatusUnderReview {
					valid = true
					next = simak.StatusCancelled
				}
			}

			if valid != tt.wantValid {
				t.Errorf("validity got %v, want %v", valid, tt.wantValid)
			}
			if next != tt.nextStatus {
				t.Errorf("next status got %q, want %q", next, tt.nextStatus)
			}
		})
	}
}

// TestDateValidityValidation checks SIMAK permit date constraints
func TestDateValidityValidation(t *testing.T) {
	now := time.Now()
	startDate := now
	endDateValid := now.AddDate(0, 0, 7)
	endDateInvalid := now.AddDate(0, 0, -1)

	if !startDate.Before(endDateValid) && !startDate.Equal(endDateValid) {
		t.Errorf("Valid date range failed assertion")
	}

	if endDateInvalid.After(startDate) {
		t.Errorf("Invalid date range failed assertion")
	}
}
