package motivationv1_test

import (
	"strings"
	"testing"

	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
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
			name: "create special accepts positive position",
			validate: func() error {
				return (&motivationv1.CreateCriterionRequest{
					IdempotencyKey: idempotencyKey,
					PositionId:     9007199254740993,
					Name:           "Cash register quality",
					MaxScore:       60,
					ValidFrom:      validDate,
				}).ValidateAll()
			},
		},
		{
			name: "create special without position is rejected",
			validate: func() error {
				return (&motivationv1.CreateCriterionRequest{
					IdempotencyKey: idempotencyKey,
					Name:           "Cash register quality",
					MaxScore:       60,
					ValidFrom:      validDate,
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "create special with zero position is rejected",
			validate: func() error {
				return (&motivationv1.CreateCriterionRequest{
					IdempotencyKey: idempotencyKey,
					PositionId:     0,
					Name:           "Cash register quality",
					MaxScore:       60,
					ValidFrom:      validDate,
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "create special with negative position is rejected",
			validate: func() error {
				return (&motivationv1.CreateCriterionRequest{
					IdempotencyKey: idempotencyKey,
					PositionId:     -5,
					Name:           "Cash register quality",
					MaxScore:       60,
					ValidFrom:      validDate,
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "special list with optional filters is accepted",
			validate: func() error {
				return (&motivationv1.ListCriteriaRequest{
					PositionId: proto.Int64(200),
					ActiveOn:   &motivationv1.Date{Year: 2026, Month: 9, Day: 1},
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
			name: "removed structural path max_score",
			request: func() *motivationv1.UpdateBaseCriteriaRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"max_score"}

				return request
			},
			wantErrText: "unsupported path",
		},
		{
			name: "removed structural path valid_from",
			request: func() *motivationv1.UpdateBaseCriteriaRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"valid_from"}

				return request
			},
			wantErrText: "unsupported path",
		},
		{
			name: "non-nullable masked value is required",
			request: func() *motivationv1.UpdateBaseCriteriaRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"description"}
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

func TestValidateUpdateCriterionRequest(t *testing.T) {
	t.Parallel()

	name := "Updated special name"
	description := "Updated special description"
	validRequest := func() *motivationv1.UpdateCriterionRequest {
		return &motivationv1.UpdateCriterionRequest{
			IdempotencyKey: idempotencyKey,
			Id:             resourceID,
			UpdateMask:     &fieldmaskpb.FieldMask{Paths: []string{"description"}},
			Description:    &description,
		}
	}

	tests := []struct {
		name        string
		request     func() *motivationv1.UpdateCriterionRequest
		wantErrText string
	}{
		{
			name: "allowed path",
			request: func() *motivationv1.UpdateCriterionRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"name"}
				request.Description = nil
				request.Name = &name

				return request
			},
		},
		{
			name: "valid_to can be cleared by presence in mask",
			request: func() *motivationv1.UpdateCriterionRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"valid_to"}
				request.Description = nil

				return request
			},
		},
		{
			name: "structural path position_id is rejected",
			request: func() *motivationv1.UpdateCriterionRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"position_id"}

				return request
			},
			wantErrText: "unsupported path",
		},
		{
			name: "empty mask",
			request: func() *motivationv1.UpdateCriterionRequest {
				request := validRequest()
				request.UpdateMask.Paths = nil

				return request
			},
			wantErrText: "must not be empty",
		},
		{
			name: "masked name without value is rejected",
			request: func() *motivationv1.UpdateCriterionRequest {
				request := validRequest()
				request.UpdateMask.Paths = []string{"name"}
				request.Description = nil

				return request
			},
			wantErrText: "value for update_mask path",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := motivationv1.ValidateUpdateCriterionRequest(test.request())
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

func TestCriterionScopeValidation(t *testing.T) {
	t.Parallel()

	validDate := &motivationv1.Date{Year: 2026, Month: 8, Day: 1}

	baseSnapshot := func() *motivationv1.PeriodCriterion {
		return &motivationv1.PeriodCriterion{
			Id:       resourceID,
			Name:     "Base quality",
			MaxScore: 40,
			Type:     motivationv1.CriterionType_CRITERION_TYPE_BASE,
			Source: &motivationv1.PeriodCriterion_BaseCriteriaId{
				BaseCriteriaId: "11111111-1111-4111-8111-111111111111",
			},
		}
	}
	specialSnapshot := func() *motivationv1.PeriodCriterion {
		return &motivationv1.PeriodCriterion{
			Id:       resourceID,
			Name:     "Special quality",
			MaxScore: 60,
			Type:     motivationv1.CriterionType_CRITERION_TYPE_SPECIAL,
			Source: &motivationv1.PeriodCriterion_CriterionId{
				CriterionId: "22222222-2222-4222-8222-222222222222",
			},
			PositionId: proto.Int64(200),
		}
	}

	tests := []struct {
		name        string
		message     proto.Message
		wantErrText string
	}{
		{name: "valid base snapshot", message: baseSnapshot()},
		{
			name: "valid special snapshot with position",
			message: &motivationv1.Period{
				Summary: &motivationv1.PeriodSummary{
					Id:           resourceID,
					Year:         2026,
					Month:        8,
					StartsAt:     validDate,
					EndsAt:       validDate,
					Status:       motivationv1.PeriodStatus_PERIOD_STATUS_OPEN,
					EffectiveCap: &motivationv1.Coefficient{},
				},
				Criteria: []*motivationv1.PeriodCriterion{baseSnapshot(), specialSnapshot()},
			},
		},
		{
			name: "base snapshot with special source is contradictory",
			message: func() proto.Message {
				value := baseSnapshot()
				value.Source = &motivationv1.PeriodCriterion_CriterionId{
					CriterionId: "22222222-2222-4222-8222-222222222222",
				}

				return value
			}(),
			wantErrText: "base criterion must reference base_criteria_id",
		},
		{
			name: "special snapshot with base source is contradictory",
			message: func() proto.Message {
				value := specialSnapshot()
				value.Source = &motivationv1.PeriodCriterion_BaseCriteriaId{
					BaseCriteriaId: "11111111-1111-4111-8111-111111111111",
				}

				return value
			}(),
			wantErrText: "special criterion must reference criterion_id",
		},
		{
			name: "special snapshot without position is rejected",
			message: func() proto.Message {
				value := specialSnapshot()
				value.PositionId = nil

				return value
			}(),
			wantErrText: "positive position_id",
		},
		{
			name: "base snapshot with position is rejected",
			message: func() proto.Message {
				value := baseSnapshot()
				value.PositionId = proto.Int64(200)

				return value
			}(),
			wantErrText: "must not carry a position_id",
		},
		{
			name: "unspecified snapshot type is rejected",
			message: &motivationv1.PeriodCriterion{
				Id:       resourceID,
				Name:     "Broken",
				MaxScore: 10,
				Source: &motivationv1.PeriodCriterion_BaseCriteriaId{
					BaseCriteriaId: "11111111-1111-4111-8111-111111111111",
				},
			},
			wantErrText: "CRITERION_TYPE_UNSPECIFIED",
		},
		{
			name: "valid base sheet line",
			message: &motivationv1.SheetCriterion{
				Id:                resourceID,
				PeriodCriterionId: "33333333-3333-4333-8333-333333333333",
				Name:              "Base quality",
				MaxScore:          40,
				Type:              motivationv1.CriterionType_CRITERION_TYPE_BASE,
				Source: &motivationv1.SheetCriterion_BaseCriteriaId{
					BaseCriteriaId: "11111111-1111-4111-8111-111111111111",
				},
			},
		},
		{
			name: "valid special sheet line",
			message: &motivationv1.SheetCriterion{
				Id:                resourceID,
				PeriodCriterionId: "33333333-3333-4333-8333-333333333333",
				Name:              "Special quality",
				MaxScore:          60,
				Type:              motivationv1.CriterionType_CRITERION_TYPE_SPECIAL,
				Source: &motivationv1.SheetCriterion_CriterionId{
					CriterionId: "22222222-2222-4222-8222-222222222222",
				},
				PositionId: proto.Int64(200),
			},
		},
		{
			name: "special sheet line without position is rejected",
			message: &motivationv1.SheetCriterion{
				Id:                resourceID,
				PeriodCriterionId: "33333333-3333-4333-8333-333333333333",
				Name:              "Special quality",
				MaxScore:          60,
				Type:              motivationv1.CriterionType_CRITERION_TYPE_SPECIAL,
				Source: &motivationv1.SheetCriterion_CriterionId{
					CriterionId: "22222222-2222-4222-8222-222222222222",
				},
			},
			wantErrText: "positive position_id",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := motivationv1.ValidateMessage(test.message)
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

func TestValidateSavePerformanceSheetRequest(t *testing.T) {
	t.Parallel()

	criterionA := "11111111-1111-4111-8111-111111111111"
	criterionB := "22222222-2222-4222-8222-222222222222"

	validScoreChange := func() *motivationv1.SaveSheetScoreChange {
		return &motivationv1.SaveSheetScoreChange{
			SheetCriterionId: criterionA,
			ScoreInput: &motivationv1.SaveSheetScoreChange_Score{
				Score: 5,
			},
		}
	}
	validAdjustmentChange := func() *motivationv1.SaveSheetAdjustmentChange {
		return &motivationv1.SaveSheetAdjustmentChange{
			AdjustmentInput: &motivationv1.SaveSheetAdjustmentChange_Value{
				Value: -3,
			},
			Comment: "remove carrot",
		}
	}
	validRequest := func() *motivationv1.SavePerformanceSheetRequest {
		return &motivationv1.SavePerformanceSheetRequest{
			IdempotencyKey: idempotencyKey,
			SheetId:        resourceID,
			Scores:         []*motivationv1.SaveSheetScoreChange{validScoreChange()},
			Adjustment:     validAdjustmentChange(),
		}
	}

	tests := []struct {
		name        string
		request     func() *motivationv1.SavePerformanceSheetRequest
		wantErrText string
	}{
		{
			name: "scores and adjustment are accepted",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				return &motivationv1.SavePerformanceSheetRequest{
					IdempotencyKey: idempotencyKey,
					SheetId:        resourceID,
					Scores: []*motivationv1.SaveSheetScoreChange{
						validScoreChange(),
						{
							SheetCriterionId: criterionB,
							ScoreInput:       &motivationv1.SaveSheetScoreChange_Score{Score: 0},
							Comment:          "keep zero",
						},
					},
				}
			},
		},
		{
			name: "adjustment only is accepted",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Scores = nil

				return request
			},
		},
		{
			name: "explicit zero score keeps presence",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Scores = []*motivationv1.SaveSheetScoreChange{
					{
						SheetCriterionId: criterionA,
						ScoreInput:       &motivationv1.SaveSheetScoreChange_Score{Score: 0},
					},
				}

				return request
			},
		},
		{
			name: "score entry without comment is accepted at the contract layer",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Scores = []*motivationv1.SaveSheetScoreChange{validScoreChange()}
				request.Scores[0].Comment = ""

				return request
			},
		},
		{
			name: "empty command is rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Scores = nil
				request.Adjustment = nil

				return request
			},
			wantErrText: "at least one",
		},
		{
			name: "duplicate criterion entries are rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Scores = []*motivationv1.SaveSheetScoreChange{
					validScoreChange(),
					{
						SheetCriterionId: criterionA,
						ScoreInput:       &motivationv1.SaveSheetScoreChange_Score{Score: 0},
					},
				}

				return request
			},
			wantErrText: "duplicate",
		},
		{
			name: "score entry without value is rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Scores = []*motivationv1.SaveSheetScoreChange{
					{
						SheetCriterionId: criterionA,
						Comment:          "no value",
					},
				}

				return request
			},
			wantErrText: "ScoreInput",
		},
		{
			name: "adjustment without value is rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Adjustment = &motivationv1.SaveSheetAdjustmentChange{
					Comment: "no value",
				}

				return request
			},
			wantErrText: "AdjustmentInput",
		},
		{
			name: "adjustment without comment is rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Adjustment = &motivationv1.SaveSheetAdjustmentChange{
					AdjustmentInput: &motivationv1.SaveSheetAdjustmentChange_Value{Value: 1},
				}

				return request
			},
			wantErrText: "Comment",
		},
		{
			name: "blank adjustment comment is rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Adjustment.Comment = "   "

				return request
			},
			wantErrText: "Comment",
		},
		{
			name: "non-UUID criterion id is rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.Scores = []*motivationv1.SaveSheetScoreChange{
					{
						SheetCriterionId: "not-a-uuid",
						ScoreInput:       &motivationv1.SaveSheetScoreChange_Score{Score: 1},
					},
				}

				return request
			},
			wantErrText: "valid UUID",
		},
		{
			name: "non-UUID sheet id is rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.SheetId = "not-a-uuid"

				return request
			},
			wantErrText: "valid UUID",
		},
		{
			name: "non-UUID idempotency key is rejected",
			request: func() *motivationv1.SavePerformanceSheetRequest {
				request := validRequest()
				request.IdempotencyKey = "not-a-uuid"

				return request
			},
			wantErrText: "valid UUID",
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			err := motivationv1.ValidateSavePerformanceSheetRequest(test.request())
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
