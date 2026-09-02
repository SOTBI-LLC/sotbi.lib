package motivationv1_test

import (
	"strings"
	"testing"

	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"

	motivationv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/motivation/v1"
)

const (
	idempotencyKey = "12345678-1234-4234-8234-123456789abc"
	resourceID     = "abcdefab-cdef-4abc-8def-abcdefabcdef"
)

func TestGeneratedRequestValidation(t *testing.T) {
	t.Parallel()

	validDate := &motivationv1.Date{Year: 2026, Month: 8, Day: 1}

	tests := []struct {
		name     string
		validate func() error
		wantErr  bool
	}{
		{
			name: "create criteria accepts boundary values",
			validate: func() error {
				return (&motivationv1.CreateBaseCriteriaRequest{
					IdempotencyKey: idempotencyKey,
					Name:           "Delivery quality",
					MaxScore:       1,
					ValidFrom:      validDate,
				}).ValidateAll()
			},
		},
		{
			name: "uppercase UUID is not canonical",
			validate: func() error {
				return (&motivationv1.GetBaseCriteriaRequest{
					Id: "ABCDEFAB-CDEF-4ABC-8DEF-ABCDEFABCDEF",
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "zero max score is rejected",
			validate: func() error {
				return (&motivationv1.CreateBaseCriteriaRequest{
					IdempotencyKey: idempotencyKey,
					Name:           "Delivery quality",
					ValidFrom:      validDate,
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "blank required name is rejected",
			validate: func() error {
				return (&motivationv1.CreateBaseCriteriaRequest{
					IdempotencyKey: idempotencyKey,
					Name:           "   ",
					MaxScore:       10,
					ValidFrom:      validDate,
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "invalid month is rejected",
			validate: func() error {
				return (&motivationv1.OpenPeriodRequest{
					IdempotencyKey: idempotencyKey,
					Year:           2026,
					Month:          13,
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "invalid day is rejected",
			validate: func() error {
				return (&motivationv1.CreateBaseCriteriaRequest{
					IdempotencyKey: idempotencyKey,
					Name:           "Delivery quality",
					MaxScore:       10,
					ValidFrom:      &motivationv1.Date{Year: 2026, Month: 8, Day: 32},
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "zero coefficient cap is rejected",
			validate: func() error {
				return (&motivationv1.SetCoefficientCapRequest{
					IdempotencyKey: idempotencyKey,
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "omitted score is rejected",
			validate: func() error {
				return (&motivationv1.SetCriterionScoreRequest{
					IdempotencyKey:   idempotencyKey,
					SheetId:          resourceID,
					SheetCriterionId: resourceID,
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "explicit zero score is accepted",
			validate: func() error {
				return (&motivationv1.SetCriterionScoreRequest{
					IdempotencyKey:   idempotencyKey,
					SheetId:          resourceID,
					SheetCriterionId: resourceID,
					ScoreInput: &motivationv1.SetCriterionScoreRequest_Score{
						Score: 0,
					},
				}).ValidateAll()
			},
		},
		{
			name: "omitted adjustment is rejected",
			validate: func() error {
				return (&motivationv1.SetAdjustmentRequest{
					IdempotencyKey: idempotencyKey,
					SheetId:        resourceID,
					Comment:        "confirmed",
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "explicit zero adjustment is accepted",
			validate: func() error {
				return (&motivationv1.SetAdjustmentRequest{
					IdempotencyKey: idempotencyKey,
					SheetId:        resourceID,
					AdjustmentInput: &motivationv1.SetAdjustmentRequest_Value{
						Value: 0,
					},
					Comment: "confirmed",
				}).ValidateAll()
			},
		},
		{
			name: "present empty user filter is accepted",
			validate: func() error {
				return (&motivationv1.ListPerformanceSheetsRequest{
					PeriodId:     resourceID,
					UserIdFilter: &motivationv1.UserIDFilter{},
				}).ValidateAll()
			},
		},
		{
			name: "user filter above limit is rejected",
			validate: func() error {
				userIDs := make([]int64, 5001)
				for i := range userIDs {
					userIDs[i] = int64(i + 1)
				}

				return (&motivationv1.ListPerformanceSheetsRequest{
					PeriodId:     resourceID,
					UserIdFilter: &motivationv1.UserIDFilter{UserIds: userIDs},
				}).ValidateAll()
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := test.validate()
			if test.wantErr && err == nil {
				t.Fatal("validation succeeded, want error")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("validation failed: %v", err)
			}
		})
	}
}

func TestValidateMessageCalendarDate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		date    *motivationv1.Date
		wantErr bool
	}{
		{
			name: "leap day",
			date: &motivationv1.Date{Year: 2024, Month: 2, Day: 29},
		},
		{
			name:    "day after february",
			date:    &motivationv1.Date{Year: 2026, Month: 2, Day: 29},
			wantErr: true,
		},
		{
			name:    "day after april",
			date:    &motivationv1.Date{Year: 2026, Month: 4, Day: 31},
			wantErr: true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			request := &motivationv1.CreateBaseCriteriaRequest{
				IdempotencyKey: idempotencyKey,
				Name:           "Delivery quality",
				MaxScore:       10,
				ValidFrom:      test.date,
			}

			err := motivationv1.ValidateMessage(request)
			if test.wantErr && err == nil {
				t.Fatal("validation succeeded, want error")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("validation failed: %v", err)
			}
		})
	}
}

func TestValidateClosePeriodOperationTerminalOutcome(t *testing.T) {
	t.Parallel()

	validPeriod := func() *motivationv1.Period {
		return &motivationv1.Period{
			Summary: &motivationv1.PeriodSummary{
				Id:           resourceID,
				Year:         2026,
				Month:        8,
				StartsAt:     &motivationv1.Date{Year: 2026, Month: 8, Day: 1},
				EndsAt:       &motivationv1.Date{Year: 2026, Month: 8, Day: 31},
				Status:       motivationv1.PeriodStatus_PERIOD_STATUS_CLOSED,
				EffectiveCap: &motivationv1.Coefficient{},
			},
		}
	}

	setResult := func(operation *motivationv1.ClosePeriodOperation) {
		operation.TerminalOutcome = &motivationv1.ClosePeriodOperation_Result{
			Result: validPeriod(),
		}
	}
	setError := func(operation *motivationv1.ClosePeriodOperation) {
		operation.TerminalOutcome = &motivationv1.ClosePeriodOperation_Error{
			Error: &statuspb.Status{
				Code:    int32(codes.FailedPrecondition),
				Message: "period cannot be closed",
			},
		}
	}
	setNilResult := func(operation *motivationv1.ClosePeriodOperation) {
		operation.TerminalOutcome = &motivationv1.ClosePeriodOperation_Result{}
	}
	setNilError := func(operation *motivationv1.ClosePeriodOperation) {
		operation.TerminalOutcome = &motivationv1.ClosePeriodOperation_Error{}
	}

	tests := []struct {
		name       string
		state      motivationv1.ClosePeriodOperationState
		setOutcome func(*motivationv1.ClosePeriodOperation)
		wantErr    bool
	}{
		{
			name:  "queued",
			state: motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_QUEUED,
		},
		{
			name:  "running",
			state: motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_RUNNING,
		},
		{
			name:       "succeeded with result",
			state:      motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_SUCCEEDED,
			setOutcome: setResult,
		},
		{
			name:       "failed with error",
			state:      motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_FAILED,
			setOutcome: setError,
		},
		{
			name:    "unspecified state",
			state:   motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_UNSPECIFIED,
			wantErr: true,
		},
		{
			name:    "succeeded without result",
			state:   motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_SUCCEEDED,
			wantErr: true,
		},
		{
			name:       "succeeded with nil result",
			state:      motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_SUCCEEDED,
			setOutcome: setNilResult,
			wantErr:    true,
		},
		{
			name:       "succeeded with error",
			state:      motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_SUCCEEDED,
			setOutcome: setError,
			wantErr:    true,
		},
		{
			name:    "failed without error",
			state:   motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_FAILED,
			wantErr: true,
		},
		{
			name:       "failed with nil error",
			state:      motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_FAILED,
			setOutcome: setNilError,
			wantErr:    true,
		},
		{
			name:       "failed with result",
			state:      motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_FAILED,
			setOutcome: setResult,
			wantErr:    true,
		},
		{
			name:       "queued with result",
			state:      motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_QUEUED,
			setOutcome: setResult,
			wantErr:    true,
		},
		{
			name:       "running with error",
			state:      motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_RUNNING,
			setOutcome: setError,
			wantErr:    true,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			operation := &motivationv1.ClosePeriodOperation{
				OperationId:     resourceID,
				PeriodId:        resourceID,
				RequesterUserId: 123,
				State:           test.state,
				CreatedAt:       timestamppb.Now(),
			}
			if test.setOutcome != nil {
				test.setOutcome(operation)
			}

			err := motivationv1.ValidateMessage(operation)
			if test.wantErr && err == nil {
				t.Fatal("validation succeeded, want error")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("validation failed: %v", err)
			}
		})
	}
}

func TestValidateUpdateBaseCriteriaRequest(t *testing.T) {
	t.Parallel()

	name := "Updated name"
	validRequest := func() *motivationv1.UpdateBaseCriteriaRequest {
		return &motivationv1.UpdateBaseCriteriaRequest{
			IdempotencyKey: idempotencyKey,
			Id:             resourceID,
			UpdateMask:     &fieldmaskpb.FieldMask{Paths: []string{"name"}},
			Name:           &name,
		}
	}

	tests := []struct {
		name        string
		request     func() *motivationv1.UpdateBaseCriteriaRequest
		wantErrText string
	}{
		{
			name:    "allowed path",
			request: validRequest,
		},
		{
			name: "valid_to can be cleared by presence in mask",
			request: func() *motivationv1.UpdateBaseCriteriaRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"valid_to"}
				request.Name = nil

				return request
			},
		},
		{
			name: "unsupported path",
			request: func() *motivationv1.UpdateBaseCriteriaRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"updated_by"}

				return request
			},
			wantErrText: "unsupported path",
		},
		{
			name: "non-nullable masked value is required",
			request: func() *motivationv1.UpdateBaseCriteriaRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"max_score"}
				request.Name = nil

				return request
			},
			wantErrText: "value for update_mask path",
		},
		{
			name: "empty mask",
			request: func() *motivationv1.UpdateBaseCriteriaRequest {
				request := validRequest()
				request.UpdateMask.Paths = nil

				return request
			},
			wantErrText: "must not be empty",
		},
		{
			name: "PGV error is retained",
			request: func() *motivationv1.UpdateBaseCriteriaRequest {
				request := validRequest()
				request.Id = "not-a-uuid"

				return request
			},
			wantErrText: "valid UUID",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := motivationv1.ValidateUpdateBaseCriteriaRequest(test.request())
			if test.wantErrText == "" && err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			if test.wantErrText != "" &&
				(err == nil || !strings.Contains(err.Error(), test.wantErrText)) {
				t.Fatalf("error = %v, want text %q", err, test.wantErrText)
			}
		})
	}
}
