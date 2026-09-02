package directoryv1

import (
	"time"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/durationpb"
)

// ErrorReason is a stable, bounded reason attached to an expected Directory
// RPC failure with google.rpc.ErrorInfo.
type ErrorReason string

const (
	ErrorReasonUserNotFound              ErrorReason = "USER_NOT_FOUND"
	ErrorReasonDataInconsistent          ErrorReason = "DIRECTORY_DATA_INCONSISTENT"
	ErrorReasonRosterLimitExceeded       ErrorReason = "ROSTER_LIMIT_EXCEEDED"
	ErrorReasonDirectReportLimitExceeded ErrorReason = "DIRECT_REPORT_LIMIT_EXCEEDED"
	ErrorReasonRateLimited               ErrorReason = "RATE_LIMITED"
	ErrorReasonAuthRoleMissing           ErrorReason = "AUTH_ROLE_MISSING"
)

// NewStatusError creates the public rich gRPC status for one expected
// Directory error. Retry guidance is deliberately limited to rate limiting.
func NewStatusError(reason ErrorReason) error {
	code := codeForErrorReason(reason)
	result := status.New(code, string(reason))

	withDetails, err := result.WithDetails(&errdetails.ErrorInfo{
		Reason: string(reason),
		Domain: "directory.v1",
	})
	if err != nil {
		return status.Error(codes.Internal, "create directory status error")
	}

	if reason == ErrorReasonRateLimited {
		withDetails, err = withDetails.WithDetails(&errdetails.RetryInfo{
			RetryDelay: durationpb.New(time.Second),
		})
		if err != nil {
			return status.Error(codes.Internal, "create directory retry status error")
		}
	}

	return withDetails.Err()
}

func codeForErrorReason(reason ErrorReason) codes.Code {
	switch reason {
	case ErrorReasonUserNotFound:
		return codes.NotFound
	case ErrorReasonDataInconsistent:
		return codes.FailedPrecondition
	case ErrorReasonRosterLimitExceeded,
		ErrorReasonDirectReportLimitExceeded,
		ErrorReasonRateLimited:
		return codes.ResourceExhausted
	case ErrorReasonAuthRoleMissing:
		return codes.PermissionDenied
	default:
		return codes.Internal
	}
}
