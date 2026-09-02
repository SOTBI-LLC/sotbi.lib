package directoryv1_test

import (
	"testing"

	"google.golang.org/protobuf/reflect/protoreflect"

	directoryv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/directory/v1"
)

func TestDirectoryServiceDescriptor(t *testing.T) {
	t.Parallel()

	service := directoryv1.File_api_directory_v1_service_proto.Services().ByName("DirectoryService")
	if service == nil {
		t.Fatal("DirectoryService descriptor is missing")
	}

	type methodContract struct {
		name   protoreflect.Name
		input  protoreflect.FullName
		output protoreflect.FullName
	}

	want := []methodContract{
		{
			"GetRosterSnapshot",
			"directory.v1.GetRosterSnapshotRequest",
			"directory.v1.GetRosterSnapshotResponse",
		},
		{"GetUser", "directory.v1.GetUserRequest", "directory.v1.GetUserResponse"},
		{
			"ListDirectReports",
			"directory.v1.ListDirectReportsRequest",
			"directory.v1.ListDirectReportsResponse",
		},
	}

	if service.Methods().Len() != len(want) {
		t.Fatalf("method count = %d, want %d", service.Methods().Len(), len(want))
	}

	for _, expected := range want {
		t.Run(string(expected.name), func(t *testing.T) {
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

			if method.IsStreamingClient() || method.IsStreamingServer() {
				t.Error("Directory v1 methods must be unary")
			}
		})
	}

	for _, forbidden := range []protoreflect.Name{
		"CreateUser", "UpdateUser", "DeleteUser", "ListUsers", "StreamRosterSnapshot",
	} {
		if service.Methods().ByName(forbidden) != nil {
			t.Errorf("forbidden method %s is present", forbidden)
		}
	}
}
