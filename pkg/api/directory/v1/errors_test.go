package directoryv1_test

import (
	"testing"

	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	directoryv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/directory/v1"
)

func TestDirectoryStatusErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		reason        directoryv1.ErrorReason
		wantCode      codes.Code
		wantRetryInfo bool
	}{
		{"absent user", directoryv1.ErrorReasonUserNotFound, codes.NotFound, false},
		{
			"inconsistent data",
			directoryv1.ErrorReasonDataInconsistent,
			codes.FailedPrecondition,
			false,
		},
		{
			"roster limit",
			directoryv1.ErrorReasonRosterLimitExceeded,
			codes.ResourceExhausted,
			false,
		},
		{
			"direct report limit",
			directoryv1.ErrorReasonDirectReportLimitExceeded,
			codes.ResourceExhausted,
			false,
		},
		{"temporary rate limit", directoryv1.ErrorReasonRateLimited, codes.ResourceExhausted, true},
		{"missing role", directoryv1.ErrorReasonAuthRoleMissing, codes.PermissionDenied, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			rpcStatus := status.Convert(directoryv1.NewStatusError(test.reason))
			if rpcStatus.Code() != test.wantCode {
				t.Fatalf("status code = %s, want %s", rpcStatus.Code(), test.wantCode)
			}

			var errorInfo *errdetails.ErrorInfo

			hasRetryInfo := false

			for _, detail := range rpcStatus.Details() {
				switch value := detail.(type) {
				case *errdetails.ErrorInfo:
					errorInfo = value
				case *errdetails.RetryInfo:
					hasRetryInfo = true

					if value.RetryDelay == nil || value.RetryDelay.AsDuration() <= 0 {
						t.Fatal("RetryInfo must contain a bounded positive delay")
					}
				}
			}

			if errorInfo == nil || errorInfo.Reason != string(test.reason) {
				t.Fatalf("ErrorInfo reason = %v, want %s", errorInfo, test.reason)
			}

			if hasRetryInfo != test.wantRetryInfo {
				t.Fatalf("RetryInfo presence = %t, want %t", hasRetryInfo, test.wantRetryInfo)
			}
		})
	}
}
