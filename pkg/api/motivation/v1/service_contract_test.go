package motivationv1_test

import (
	"testing"

	statuspb "google.golang.org/genproto/googleapis/rpc/status"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"

	motivationv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/motivation/v1"
)

type validateAller interface {
	ValidateAll() error
}

func TestPerformanceEvaluationServiceDescriptor(t *testing.T) {
	t.Parallel()

	service := motivationv1.File_api_motivation_v1_service_proto.Services().
		ByName("PerformanceEvaluationService")
	if service == nil {
		t.Fatal("PerformanceEvaluationService descriptor is missing")
	}

	type methodContract struct {
		name            protoreflect.Name
		input           protoreflect.FullName
		output          protoreflect.FullName
		serverStreaming bool
	}

	want := []methodContract{
		{
			"CreateBaseCriteria",
			"motivation.v1.CreateBaseCriteriaRequest",
			"motivation.v1.CreateBaseCriteriaResponse",
			false,
		},
		{
			"UpdateBaseCriteria",
			"motivation.v1.UpdateBaseCriteriaRequest",
			"motivation.v1.UpdateBaseCriteriaResponse",
			false,
		},
		{
			"GetBaseCriteria",
			"motivation.v1.GetBaseCriteriaRequest",
			"motivation.v1.GetBaseCriteriaResponse",
			false,
		},
		{
			"ListBaseCriteria",
			"motivation.v1.ListBaseCriteriaRequest",
			"motivation.v1.ListBaseCriteriaResponse",
			false,
		},
		{
			"GetCoefficientCap",
			"motivation.v1.GetCoefficientCapRequest",
			"motivation.v1.GetCoefficientCapResponse",
			false,
		},
		{
			"SetCoefficientCap",
			"motivation.v1.SetCoefficientCapRequest",
			"motivation.v1.SetCoefficientCapResponse",
			false,
		},
		{
			"OpenPeriod",
			"motivation.v1.OpenPeriodRequest",
			"motivation.v1.OpenPeriodResponse",
			false,
		},
		{
			"ReopenPeriod",
			"motivation.v1.ReopenPeriodRequest",
			"motivation.v1.ReopenPeriodResponse",
			false,
		},
		{"GetPeriod", "motivation.v1.GetPeriodRequest", "motivation.v1.GetPeriodResponse", false},
		{
			"ListPeriods",
			"motivation.v1.ListPeriodsRequest",
			"motivation.v1.ListPeriodsResponse",
			false,
		},
		{
			"IncludeUser",
			"motivation.v1.IncludeUserRequest",
			"motivation.v1.IncludeUserResponse",
			false,
		},
		{
			"SetCriterionScore",
			"motivation.v1.SetCriterionScoreRequest",
			"motivation.v1.SetCriterionScoreResponse",
			false,
		},
		{
			"SetAdjustment",
			"motivation.v1.SetAdjustmentRequest",
			"motivation.v1.SetAdjustmentResponse",
			false,
		},
		{
			"GetPerformanceSheet",
			"motivation.v1.GetPerformanceSheetRequest",
			"motivation.v1.GetPerformanceSheetResponse",
			false,
		},
		{
			"ListPerformanceSheets",
			"motivation.v1.ListPerformanceSheetsRequest",
			"motivation.v1.ListPerformanceSheetsResponse",
			false,
		},
		{
			"StreamPerformanceSheets",
			"motivation.v1.StreamPerformanceSheetsRequest",
			"motivation.v1.StreamPerformanceSheetsResponse",
			true,
		},
		{
			"ListPayrollResults",
			"motivation.v1.ListPayrollResultsRequest",
			"motivation.v1.ListPayrollResultsResponse",
			false,
		},
		{
			"StreamPayrollResults",
			"motivation.v1.StreamPayrollResultsRequest",
			"motivation.v1.StreamPayrollResultsResponse",
			true,
		},
		{
			"StartClosePeriod",
			"motivation.v1.StartClosePeriodRequest",
			"motivation.v1.StartClosePeriodResponse",
			false,
		},
		{
			"GetClosePeriodOperation",
			"motivation.v1.GetClosePeriodOperationRequest",
			"motivation.v1.GetClosePeriodOperationResponse",
			false,
		},
		{
			"WatchClosePeriodOperation",
			"motivation.v1.WatchClosePeriodOperationRequest",
			"motivation.v1.WatchClosePeriodOperationResponse",
			true,
		},
	}

	if service.Methods().Len() != len(want) {
		t.Fatalf("method count = %d, want %d", service.Methods().Len(), len(want))
	}

	for _, expected := range want {
		expected := expected
		t.Run(string(expected.name), func(t *testing.T) {
			t.Parallel()

			method := service.Methods().ByName(expected.name)
			if method == nil {
				t.Fatalf("method %s is missing", expected.name)
			}

			if method.Input().FullName() != expected.input {
				t.Errorf("input = %s, want %s", method.Input().FullName(), expected.input)
			}

			if method.Output().FullName() != expected.output {
				t.Errorf("output = %s, want %s", method.Output().FullName(), expected.output)
			}

			if method.IsStreamingClient() {
				t.Error("client streaming must be disabled")
			}

			if method.IsStreamingServer() != expected.serverStreaming {
				t.Errorf(
					"server streaming = %t, want %t",
					method.IsStreamingServer(),
					expected.serverStreaming,
				)
			}
		})
	}

	for _, forbidden := range []protoreflect.Name{"Ping", "ClosePeriod", "CancelClosePeriod", "CancelOperation", "DeleteBaseCriteria"} {
		if method := service.Methods().ByName(forbidden); method != nil {
			t.Errorf("forbidden method %s is present", forbidden)
		}
	}
}

func TestClosePeriodOperationRoundTrip(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		operation *motivationv1.ClosePeriodOperation
		assert    func(*testing.T, *motivationv1.ClosePeriodOperation)
	}{
		{
			name: "running progress has no terminal outcome",
			operation: &motivationv1.ClosePeriodOperation{
				OperationId:     "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa",
				PeriodId:        "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb",
				RequesterUserId: 42,
				State:           motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_RUNNING,
				ProcessedSheets: 347,
				TotalSheets:     5000,
				CreatedAt:       timestamppb.Now(),
				StartedAt:       timestamppb.Now(),
			},
			assert: func(t *testing.T, got *motivationv1.ClosePeriodOperation) {
				t.Helper()

				if got.GetProcessedSheets() != 347 || got.GetTotalSheets() != 5000 {
					t.Errorf(
						"progress = %d/%d, want 347/5000",
						got.GetProcessedSheets(),
						got.GetTotalSheets(),
					)
				}

				if got.GetTerminalOutcome() != nil {
					t.Errorf("running operation has terminal outcome %T", got.GetTerminalOutcome())
				}
			},
		},
		{
			name: "succeeded operation carries Period result",
			operation: &motivationv1.ClosePeriodOperation{
				OperationId: "cccccccc-cccc-4ccc-8ccc-cccccccccccc",
				PeriodId:    "dddddddd-dddd-4ddd-8ddd-dddddddddddd",
				State:       motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_SUCCEEDED,
				CreatedAt:   timestamppb.Now(),
				FinishedAt:  timestamppb.Now(),
				TerminalOutcome: &motivationv1.ClosePeriodOperation_Result{
					Result: &motivationv1.Period{},
				},
			},
			assert: func(t *testing.T, got *motivationv1.ClosePeriodOperation) {
				t.Helper()

				if got.GetResult() == nil || got.GetError() != nil {
					t.Errorf("succeeded outcome = %T, want Period result", got.GetTerminalOutcome())
				}
			},
		},
		{
			name: "failed operation carries standard Status",
			operation: &motivationv1.ClosePeriodOperation{
				OperationId: "eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee",
				PeriodId:    "ffffffff-ffff-4fff-8fff-ffffffffffff",
				State:       motivationv1.ClosePeriodOperationState_CLOSE_PERIOD_OPERATION_STATE_FAILED,
				CreatedAt:   timestamppb.Now(),
				FinishedAt:  timestamppb.Now(),
				TerminalOutcome: &motivationv1.ClosePeriodOperation_Error{
					Error: &statuspb.Status{Code: 9, Message: "period cannot be closed"},
				},
			},
			assert: func(t *testing.T, got *motivationv1.ClosePeriodOperation) {
				t.Helper()

				if got.GetError() == nil || got.GetResult() != nil {
					t.Errorf("failed outcome = %T, want Status error", got.GetTerminalOutcome())
				}
			},
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			encoded, err := proto.Marshal(test.operation)
			if err != nil {
				t.Fatalf("marshal operation: %v", err)
			}

			var decoded motivationv1.ClosePeriodOperation
			if err := proto.Unmarshal(encoded, &decoded); err != nil {
				t.Fatalf("unmarshal operation: %v", err)
			}

			if !proto.Equal(test.operation, &decoded) {
				t.Fatalf(
					"round trip changed operation\n got: %v\nwant: %v",
					&decoded,
					test.operation,
				)
			}

			test.assert(t, &decoded)
		})
	}
}

func TestPresenceAndFixedPointRoundTrip(t *testing.T) {
	t.Parallel()

	original := &motivationv1.PerformanceSheet{
		Summary: &motivationv1.PerformanceSheetSummary{
			Id:          "11111111-1111-4111-8111-111111111111",
			PeriodId:    "22222222-2222-4222-8222-222222222222",
			UserId:      42,
			PositionId:  7,
			Coefficient: &motivationv1.Coefficient{TenThousandths: 30000},
		},
		Criteria: []*motivationv1.SheetCriterion{
			{
				Id:                "33333333-3333-4333-8333-333333333333",
				PeriodCriterionId: "44444444-4444-4444-8444-444444444444",
				Name:              "Unset score",
				MaxScore:          10,
			},
			{
				Id:                "55555555-5555-4555-8555-555555555555",
				PeriodCriterionId: "66666666-6666-4666-8666-666666666666",
				Name:              "Explicit zero score",
				MaxScore:          10,
				Score:             proto.Int32(0),
				History: []*motivationv1.ScoreChange{
					{
						Id:        "77777777-7777-4777-8777-777777777777",
						ChangedAt: timestamppb.Now(),
						From:      proto.Int32(0),
						To:        0,
					},
				},
			},
		},
		Adjustment: &motivationv1.Adjustment{Value: 0},
	}

	encoded, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal sheet: %v", err)
	}

	var decoded motivationv1.PerformanceSheet
	if err := proto.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal sheet: %v", err)
	}

	if !proto.Equal(original, &decoded) {
		t.Fatalf("round trip changed sheet\n got: %v\nwant: %v", &decoded, original)
	}

	if decoded.GetCriteria()[0].Score != nil {
		t.Error("unset score became present")
	}

	if decoded.GetCriteria()[1].Score == nil || decoded.GetCriteria()[1].GetScore() != 0 {
		t.Error("explicit zero score lost presence")
	}

	if decoded.GetCriteria()[1].GetHistory()[0].From == nil {
		t.Error("explicit zero history from lost presence")
	}

	if decoded.Adjustment == nil || decoded.GetAdjustment().GetValue() != 0 {
		t.Error("explicit zero adjustment lost presence")
	}

	if decoded.GetSummary().GetCoefficient().GetTenThousandths() != 30000 {
		t.Errorf(
			"coefficient = %d, want 30000",
			decoded.GetSummary().GetCoefficient().GetTenThousandths(),
		)
	}
}

func TestFiveThousandSheetSnapshotRoundTrip(t *testing.T) {
	t.Parallel()

	original := &motivationv1.ListPerformanceSheetsResponse{
		PerformanceSheets: make([]*motivationv1.PerformanceSheetSummary, 5000),
	}
	for i := range original.PerformanceSheets {
		original.PerformanceSheets[i] = &motivationv1.PerformanceSheetSummary{
			UserId:      int64(i + 1),
			PositionId:  int64(i + 10_000),
			Coefficient: &motivationv1.Coefficient{TenThousandths: uint32(i)},
		}
	}

	encoded, err := proto.Marshal(original)
	if err != nil {
		t.Fatalf("marshal snapshot: %v", err)
	}

	var decoded motivationv1.ListPerformanceSheetsResponse
	if err := proto.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}

	if len(decoded.GetPerformanceSheets()) != 5000 {
		t.Fatalf("sheet count = %d, want 5000", len(decoded.GetPerformanceSheets()))
	}

	if decoded.GetPerformanceSheets()[4999].GetUserId() != 5000 {
		t.Errorf("last UserID = %d, want 5000", decoded.GetPerformanceSheets()[4999].GetUserId())
	}
}

func TestPresenceSensitiveFields(t *testing.T) {
	t.Parallel()

	fields := []struct {
		message protoreflect.MessageDescriptor
		name    protoreflect.Name
	}{
		{(&motivationv1.BaseCriteria{}).ProtoReflect().Descriptor(), "valid_to"},
		{(&motivationv1.SheetCriterion{}).ProtoReflect().Descriptor(), "score"},
		{(&motivationv1.PerformanceSheet{}).ProtoReflect().Descriptor(), "adjustment"},
		{(&motivationv1.ScoreChange{}).ProtoReflect().Descriptor(), "from"},
		{(&motivationv1.Period{}).ProtoReflect().Descriptor(), "stamped_cap"},
		{(&motivationv1.ClosePeriodOperation{}).ProtoReflect().Descriptor(), "started_at"},
		{(&motivationv1.ClosePeriodOperation{}).ProtoReflect().Descriptor(), "finished_at"},
	}

	for _, field := range fields {
		field := field
		t.Run(string(field.message.Name())+"/"+string(field.name), func(t *testing.T) {
			t.Parallel()

			descriptor := field.message.Fields().ByName(field.name)
			if descriptor == nil {
				t.Fatalf("field %s.%s is missing", field.message.FullName(), field.name)
			}

			if !descriptor.HasPresence() {
				t.Errorf(
					"field %s.%s does not preserve presence",
					field.message.FullName(),
					field.name,
				)
			}
		})
	}
}

func TestCanonicalWireShapes(t *testing.T) {
	t.Parallel()

	type fieldContract struct {
		message protoreflect.MessageDescriptor
		name    protoreflect.Name
		kind    protoreflect.Kind
	}

	fields := []fieldContract{
		{(&motivationv1.Date{}).ProtoReflect().Descriptor(), "year", protoreflect.Int32Kind},
		{(&motivationv1.Date{}).ProtoReflect().Descriptor(), "month", protoreflect.Int32Kind},
		{(&motivationv1.Date{}).ProtoReflect().Descriptor(), "day", protoreflect.Int32Kind},
		{
			(&motivationv1.Coefficient{}).ProtoReflect().Descriptor(),
			"ten_thousandths",
			protoreflect.Uint32Kind,
		},
		{(&motivationv1.BaseCriteria{}).ProtoReflect().Descriptor(), "id", protoreflect.StringKind},
		{
			(&motivationv1.PeriodCriterion{}).ProtoReflect().Descriptor(),
			"base_criteria_id",
			protoreflect.StringKind,
		},
		{
			(&motivationv1.PerformanceSheetSummary{}).ProtoReflect().Descriptor(),
			"id",
			protoreflect.StringKind,
		},
		{
			(&motivationv1.PerformanceSheetSummary{}).ProtoReflect().Descriptor(),
			"user_id",
			protoreflect.Int64Kind,
		},
		{
			(&motivationv1.PerformanceSheetSummary{}).ProtoReflect().Descriptor(),
			"position_id",
			protoreflect.Int64Kind,
		},
		{
			(&motivationv1.ScoreChange{}).ProtoReflect().Descriptor(),
			"changed_at",
			protoreflect.MessageKind,
		},
	}

	for _, expected := range fields {
		expected := expected
		t.Run(string(expected.message.Name())+"/"+string(expected.name), func(t *testing.T) {
			t.Parallel()

			field := expected.message.Fields().ByName(expected.name)
			if field == nil {
				t.Fatalf("field %s.%s is missing", expected.message.FullName(), expected.name)
			}

			if field.Kind() != expected.kind {
				t.Errorf("kind = %s, want %s", field.Kind(), expected.kind)
			}
		})
	}

	if got := (&motivationv1.ScoreChange{}).ProtoReflect().Descriptor().Fields().ByName("changed_at").Message().FullName(); got != "google.protobuf.Timestamp" {
		t.Errorf("ScoreChange.changed_at type = %s, want google.protobuf.Timestamp", got)
	}

	for _, enum := range []protoreflect.EnumDescriptor{
		motivationv1.PeriodStatus(0).Descriptor(),
		motivationv1.ClosePeriodOperationState(0).Descriptor(),
	} {
		if enum.Values().Get(0).Number() != 0 {
			t.Errorf("enum %s does not reserve zero for unspecified", enum.FullName())
		}
	}

	messages := motivationv1.File_api_motivation_v1_service_proto.Messages()
	for i := range messages.Len() {
		message := messages.Get(i)
		for j := range message.Fields().Len() {
			field := message.Fields().Get(j)
			if field.Kind() == protoreflect.FloatKind || field.Kind() == protoreflect.DoubleKind {
				t.Errorf(
					"floating-point field %s.%s is forbidden",
					message.FullName(),
					field.Name(),
				)
			}
		}
	}
}

func TestCommandAndFilterRequestShapes(t *testing.T) {
	t.Parallel()

	file := motivationv1.File_api_motivation_v1_service_proto
	mutationRequests := []protoreflect.Name{
		"CreateBaseCriteriaRequest",
		"UpdateBaseCriteriaRequest",
		"SetCoefficientCapRequest",
		"OpenPeriodRequest",
		"ReopenPeriodRequest",
		"IncludeUserRequest",
		"SetCriterionScoreRequest",
		"SetAdjustmentRequest",
		"StartClosePeriodRequest",
	}

	for _, name := range mutationRequests {
		name := name
		t.Run(string(name), func(t *testing.T) {
			t.Parallel()

			message := file.Messages().ByName(name)
			if message == nil {
				t.Fatalf("message %s is missing", name)
			}

			key := message.Fields().ByName("idempotency_key")
			if key == nil || key.Kind() != protoreflect.StringKind {
				t.Errorf("%s must expose a string idempotency_key", name)
			}

			for _, forbidden := range []protoreflect.Name{"actor_user_id", "audit_timestamp", "changed_at"} {
				if message.Fields().ByName(forbidden) != nil {
					t.Errorf("%s lets the caller set %s", name, forbidden)
				}
			}
		})
	}

	updateMask := (&motivationv1.UpdateBaseCriteriaRequest{}).ProtoReflect().
		Descriptor().
		Fields().
		ByName("update_mask")
	if updateMask == nil || updateMask.Message().FullName() != "google.protobuf.FieldMask" {
		t.Fatal("UpdateBaseCriteriaRequest.update_mask must be google.protobuf.FieldMask")
	}

	for _, message := range []protoreflect.MessageDescriptor{
		(&motivationv1.ListPerformanceSheetsRequest{}).ProtoReflect().Descriptor(),
		(&motivationv1.StreamPerformanceSheetsRequest{}).ProtoReflect().Descriptor(),
	} {
		filter := message.Fields().ByName("user_id_filter")
		if filter == nil || !filter.HasPresence() {
			t.Errorf(
				"%s.user_id_filter must distinguish absent from present-empty",
				message.FullName(),
			)
		}
	}

	scoreRequest := (&motivationv1.SetCriterionScoreRequest{}).ProtoReflect().Descriptor()
	if scoreRequest.Fields().ByName("sheet_criterion_id") == nil {
		t.Error("SetCriterionScoreRequest.sheet_criterion_id is missing")
	}

	if scoreRequest.Fields().ByName("period_criterion_id") != nil {
		t.Error("SetCriterionScoreRequest must not accept period_criterion_id")
	}

	capRequest := (&motivationv1.GetCoefficientCapRequest{}).ProtoReflect().Descriptor()
	for _, forbidden := range []protoreflect.Name{"year", "month", "period_id"} {
		if capRequest.Fields().ByName(forbidden) != nil {
			t.Errorf("GetCoefficientCapRequest must not contain %s", forbidden)
		}
	}
}

func TestGeneratedValidationAPI(t *testing.T) {
	t.Parallel()

	validators := []validateAller{
		&motivationv1.CreateBaseCriteriaRequest{},
		&motivationv1.UpdateBaseCriteriaRequest{},
		&motivationv1.SetCoefficientCapRequest{},
		&motivationv1.OpenPeriodRequest{},
		&motivationv1.ReopenPeriodRequest{},
		&motivationv1.IncludeUserRequest{},
		&motivationv1.SetCriterionScoreRequest{},
		&motivationv1.SetAdjustmentRequest{},
		&motivationv1.StartClosePeriodRequest{},
	}

	for _, validator := range validators {
		if validator == nil {
			t.Fatal("generated validator is nil")
		}
	}
}
