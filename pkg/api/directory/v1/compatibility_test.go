package directoryv1_test

import (
	"fmt"
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	directoryv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/directory/v1"
)

func TestDirectoryV1Compatibility(t *testing.T) {
	t.Parallel()

	file := directoryv1.File_api_directory_v1_service_proto
	for _, expected := range []struct {
		message string
		fields  []fieldShape
	}{
		{"DirectoryUser", []fieldShape{{"user_id", 1, protoreflect.Uint64Kind}, {"position_id", 2, protoreflect.Uint64Kind}, {"manager_user_id", 3, protoreflect.Uint64Kind}, {"active", 4, protoreflect.BoolKind}}},
		{"GetUserRequest", []fieldShape{{"user_id", 1, protoreflect.Uint64Kind}}},
		{"ListDirectReportsRequest", []fieldShape{{"manager_user_id", 1, protoreflect.Uint64Kind}}},
		{"GetRosterSnapshotResponse", []fieldShape{{"snapshot_at", 1, protoreflect.MessageKind}, {"version", 2, protoreflect.StringKind}, {"users", 3, protoreflect.MessageKind}}},
		{"ListDirectReportsResponse", []fieldShape{{"snapshot_at", 1, protoreflect.MessageKind}, {"version", 2, protoreflect.StringKind}, {"user_ids", 3, protoreflect.Uint64Kind}}},
	} {
		message := file.Messages().ByName(protoreflect.Name(expected.message))
		if err := checkFieldShapes(message, expected.fields); err != nil {
			t.Fatalf("%s compatibility: %v", expected.message, err)
		}
	}
}

func TestDirectoryV1CompatibilityRejectsBreakingField(t *testing.T) {
	t.Parallel()

	message := directoryv1.File_api_directory_v1_service_proto.Messages().ByName("DirectoryUser")

	err := checkFieldShapes(message, []fieldShape{{"user_id", 1, protoreflect.StringKind}})
	if err == nil {
		t.Fatal("wire-type change passed compatibility check")
	}
}

type fieldShape struct {
	name   protoreflect.Name
	number protoreflect.FieldNumber
	kind   protoreflect.Kind
}

func checkFieldShapes(message protoreflect.MessageDescriptor, expected []fieldShape) error {
	if message == nil {
		return fmt.Errorf("message is missing")
	}

	for _, shape := range expected {
		field := message.Fields().ByName(shape.name)
		if field == nil {
			return fmt.Errorf(
				"field %s is missing; reserve its name and number before removal",
				shape.name,
			)
		}

		if field.Number() != shape.number || field.Kind() != shape.kind {
			return fmt.Errorf(
				"field %s has number %d kind %s, want number %d kind %s",
				shape.name,
				field.Number(),
				field.Kind(),
				shape.number,
				shape.kind,
			)
		}
	}

	return nil
}
