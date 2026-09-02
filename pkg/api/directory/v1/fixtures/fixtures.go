// Package fixtures provides versioned Directory v1 contract scenarios for
// provider and consumer acceptance suites.
package fixtures

import (
	"context"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	directoryv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/directory/v1"
)

const Version = "directory.v1"

const contentVersion = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"

// Scenario supplies one externally observable request/result boundary. A
// provider owns its transport setup; the fixture deliberately contains no
// credentials or service subject.
type Scenario struct {
	Name            string
	Request         proto.Message
	Response        proto.Message
	ExpectedError   directoryv1.ErrorReason
	ExpectedCode    codes.Code
	ExpectedContext error
	ValidResponse   bool
}

// Scenarios returns independent messages so callers can safely mutate them.
func Scenarios() []Scenario {
	snapshotAt := timestamppb.New(time.Date(2026, 9, 2, 0, 0, 0, 0, time.UTC))

	oversizedReports := make([]uint64, 5001)
	for index := range oversizedReports {
		oversizedReports[index] = uint64(index + 1)
	}

	return []Scenario{
		{
			Name:    "valid roster",
			Request: &directoryv1.GetRosterSnapshotRequest{},
			Response: &directoryv1.GetRosterSnapshotResponse{
				SnapshotAt: snapshotAt,
				Version:    contentVersion,
				Users: []*directoryv1.DirectoryUser{
					{UserId: 1, PositionId: 10, Active: true},
					{UserId: 2, PositionId: 20, ManagerUserId: proto.Uint64(1), Active: true},
				},
			},
			ValidResponse: true,
		},
		{
			Name:    "empty roster",
			Request: &directoryv1.GetRosterSnapshotRequest{},
			Response: &directoryv1.GetRosterSnapshotResponse{
				SnapshotAt: snapshotAt,
				Version:    contentVersion,
				Users:      []*directoryv1.DirectoryUser{},
			},
			ValidResponse: true,
		},
		{
			Name:    "inactive user",
			Request: &directoryv1.GetUserRequest{UserId: 3},
			Response: &directoryv1.GetUserResponse{
				User: &directoryv1.DirectoryUser{UserId: 3, PositionId: 30, Active: false},
			},
			ValidResponse: true,
		},
		{
			Name:         "malformed lookup",
			Request:      &directoryv1.GetUserRequest{},
			ExpectedCode: codes.InvalidArgument,
		},
		{
			Name: "oversized direct reports",
			Request: &directoryv1.ListDirectReportsRequest{
				ManagerUserId: 1,
			},
			Response: &directoryv1.ListDirectReportsResponse{
				SnapshotAt: snapshotAt,
				Version:    contentVersion,
				UserIds:    oversizedReports,
			},
			ExpectedError: directoryv1.ErrorReasonDirectReportLimitExceeded,
		},
		{
			Name:    "manager change",
			Request: &directoryv1.GetUserRequest{UserId: 2},
			Response: &directoryv1.GetUserResponse{
				User: &directoryv1.DirectoryUser{
					UserId:        2,
					PositionId:    20,
					ManagerUserId: proto.Uint64(4),
					Active:        true,
				},
			},
			ValidResponse: true,
		},
		{Name: "cancellation", ExpectedCode: codes.Canceled, ExpectedContext: context.Canceled},
		{
			Name:            "deadline",
			ExpectedCode:    codes.DeadlineExceeded,
			ExpectedContext: context.DeadlineExceeded,
		},
		{
			Name:          "authentication",
			ExpectedCode:  codes.PermissionDenied,
			ExpectedError: directoryv1.ErrorReasonAuthRoleMissing,
		},
		{
			Name:          "typed not found",
			ExpectedCode:  codes.NotFound,
			ExpectedError: directoryv1.ErrorReasonUserNotFound,
		},
	}
}
