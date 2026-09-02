package directoryv1_test

import (
	"strings"
	"testing"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	directoryv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/directory/v1"
)

const directoryVersion = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

func TestGeneratedDirectoryValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		validate func() error
		wantErr  bool
	}{
		{
			name: "positive user and position IDs are accepted",
			validate: func() error {
				return (&directoryv1.DirectoryUser{UserId: 1, PositionId: 2}).ValidateAll()
			},
		},
		{
			name: "zero user ID is rejected",
			validate: func() error {
				return (&directoryv1.GetUserRequest{}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "zero present manager ID is rejected",
			validate: func() error {
				return (&directoryv1.DirectoryUser{
					UserId:        1,
					PositionId:    2,
					ManagerUserId: proto.Uint64(0),
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "malformed version is rejected",
			validate: func() error {
				return (&directoryv1.GetRosterSnapshotResponse{
					SnapshotAt: timestamppb.New(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)),
					Version:    "not-a-version",
				}).ValidateAll()
			},
			wantErr: true,
		},
		{
			name: "direct reports above limit are rejected",
			validate: func() error {
				userIDs := make([]uint64, 5001)
				for index := range userIDs {
					userIDs[index] = uint64(index + 1)
				}

				return (&directoryv1.ListDirectReportsResponse{
					SnapshotAt: timestamppb.New(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC)),
					Version:    directoryVersion,
					UserIds:    userIDs,
				}).ValidateAll()
			},
			wantErr: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
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

func TestValidateMessageDirectoryResponses(t *testing.T) {
	t.Parallel()

	snapshotAt := timestamppb.New(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC))

	tests := []struct {
		name    string
		message proto.Message
		wantErr string
	}{
		{
			name: "sorted active roster is valid",
			message: &directoryv1.GetRosterSnapshotResponse{
				SnapshotAt: snapshotAt,
				Version:    directoryVersion,
				Users: []*directoryv1.DirectoryUser{
					{UserId: 1, PositionId: 10, Active: true},
					{UserId: 2, PositionId: 20, ManagerUserId: proto.Uint64(1), Active: true},
				},
			},
		},
		{
			name: "empty roster is valid",
			message: &directoryv1.GetRosterSnapshotResponse{
				SnapshotAt: snapshotAt,
				Version:    directoryVersion,
			},
		},
		{
			name: "roster rejects inactive user",
			message: &directoryv1.GetRosterSnapshotResponse{
				SnapshotAt: snapshotAt,
				Version:    directoryVersion,
				Users:      []*directoryv1.DirectoryUser{{UserId: 1, PositionId: 10}},
			},
			wantErr: "must contain only active users",
		},
		{
			name: "roster rejects out of order IDs",
			message: &directoryv1.GetRosterSnapshotResponse{
				SnapshotAt: snapshotAt,
				Version:    directoryVersion,
				Users: []*directoryv1.DirectoryUser{
					{UserId: 2, PositionId: 20, Active: true},
					{UserId: 1, PositionId: 10, Active: true},
				},
			},
			wantErr: "users must be ordered by user_id",
		},
		{
			name: "direct reports must be initialized and sorted",
			message: &directoryv1.ListDirectReportsResponse{
				SnapshotAt: snapshotAt,
				Version:    directoryVersion,
				UserIds:    []uint64{2, 1},
			},
			wantErr: "user_ids must be ordered",
		},
		{
			name:    "required user response is enforced",
			message: &directoryv1.GetUserResponse{},
			wantErr: "value is required",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := directoryv1.ValidateMessage(test.message)
			if test.wantErr == "" && err != nil {
				t.Fatalf("validation failed: %v", err)
			}

			if test.wantErr != "" && (err == nil || !strings.Contains(err.Error(), test.wantErr)) {
				t.Fatalf("validation error = %v, want substring %q", err, test.wantErr)
			}
		})
	}
}
