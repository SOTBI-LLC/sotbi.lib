package directoryv1_test

import (
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"

	directoryv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/directory/v1"
)

func TestDirectoryUserWireShape(t *testing.T) {
	t.Parallel()

	descriptor := directoryv1.File_api_directory_v1_service_proto.Messages().ByName("DirectoryUser")
	if descriptor == nil {
		t.Fatal("DirectoryUser descriptor is missing")
	}

	for _, expected := range []struct {
		name   protoreflect.Name
		number protoreflect.FieldNumber
		kind   protoreflect.Kind
	}{
		{"user_id", 1, protoreflect.Uint64Kind},
		{"position_id", 2, protoreflect.Uint64Kind},
		{"manager_user_id", 3, protoreflect.Uint64Kind},
		{"active", 4, protoreflect.BoolKind},
	} {
		field := descriptor.Fields().ByName(expected.name)
		if field == nil {
			t.Fatalf("field %s is missing", expected.name)
		}

		if field.Number() != expected.number || field.Kind() != expected.kind {
			t.Errorf(
				"field %s = number %d kind %s, want number %d kind %s",
				expected.name,
				field.Number(),
				field.Kind(),
				expected.number,
				expected.kind,
			)
		}
	}

	if !descriptor.Fields().ByName("manager_user_id").HasOptionalKeyword() {
		t.Error("manager_user_id must preserve optional presence")
	}

	for _, forbidden := range []protoreflect.Name{"name", "email", "staff_id", "position_name", "field_mask"} {
		if descriptor.Fields().ByName(forbidden) != nil {
			t.Errorf("forbidden field %s is present", forbidden)
		}
	}
}

func TestMaximumDirectoryResponsesRoundTripUnderSendLimit(t *testing.T) {
	t.Parallel()

	snapshotAt := timestamppb.New(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC))
	users := make([]*directoryv1.DirectoryUser, 5000)
	userIDs := make([]uint64, 5000)

	for index := range users {
		userID := uint64(index + 1)
		users[index] = &directoryv1.DirectoryUser{UserId: userID, PositionId: userID, Active: true}
		userIDs[index] = userID
	}

	responses := []proto.Message{
		&directoryv1.GetRosterSnapshotResponse{
			SnapshotAt: snapshotAt,
			Version:    directoryVersion,
			Users:      users,
		},
		&directoryv1.ListDirectReportsResponse{
			SnapshotAt: snapshotAt,
			Version:    directoryVersion,
			UserIds:    userIDs,
		},
	}

	for _, response := range responses {
		encoded, err := proto.Marshal(response)
		if err != nil {
			t.Fatalf("marshal %T: %v", response, err)
		}

		if len(encoded) > 2<<20 {
			t.Fatalf("%T encoded size = %d, exceeds 2 MiB", response, len(encoded))
		}

		decoded := response.ProtoReflect().New().Interface()
		if err := proto.Unmarshal(encoded, decoded); err != nil {
			t.Fatalf("unmarshal %T: %v", response, err)
		}

		if !proto.Equal(response, decoded) {
			t.Fatalf("round trip changed %T", response)
		}
	}
}
