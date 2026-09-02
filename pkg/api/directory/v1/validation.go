package directoryv1

import (
	"fmt"

	"google.golang.org/protobuf/proto"
)

type allValidator interface {
	ValidateAll() error
}

// ValidateMessage composes generated PGV validation with Directory response
// invariants that cannot be expressed in the protobuf schema.
func ValidateMessage(message proto.Message) error {
	if message == nil || !message.ProtoReflect().IsValid() {
		return fmt.Errorf("validate directory protobuf message: message is required")
	}

	validator, ok := message.(allValidator)
	if ok {
		if err := validator.ValidateAll(); err != nil {
			return fmt.Errorf("validate directory protobuf message: %w", err)
		}
	}

	switch value := message.(type) {
	case *GetRosterSnapshotResponse:
		if err := validateRosterSnapshot(value); err != nil {
			return fmt.Errorf("validate directory protobuf message: %w", err)
		}
	case *ListDirectReportsResponse:
		if err := validateDirectReports(value); err != nil {
			return fmt.Errorf("validate directory protobuf message: %w", err)
		}
	}

	return nil
}

func validateRosterSnapshot(response *GetRosterSnapshotResponse) error {
	var previousUserID uint64

	for index, user := range response.Users {
		if !user.GetActive() {
			return fmt.Errorf("users must contain only active users")
		}

		if index > 0 && user.GetUserId() <= previousUserID {
			return fmt.Errorf("users must be ordered by user_id")
		}

		previousUserID = user.GetUserId()
	}

	return nil
}

func validateDirectReports(response *ListDirectReportsResponse) error {
	var previousUserID uint64
	for index, userID := range response.UserIds {
		if index > 0 && userID <= previousUserID {
			return fmt.Errorf("user_ids must be ordered")
		}

		previousUserID = userID
	}

	return nil
}
