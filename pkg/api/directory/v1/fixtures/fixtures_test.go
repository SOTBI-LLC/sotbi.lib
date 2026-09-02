package fixtures_test

import (
	"testing"

	directoryv1 "github.com/SOTBI-LLC/sotbi.lib/pkg/api/directory/v1"
	"github.com/SOTBI-LLC/sotbi.lib/pkg/api/directory/v1/fixtures"
)

func TestScenariosExposeIndependentValidResponses(t *testing.T) {
	t.Parallel()

	scenarios := fixtures.Scenarios()
	if len(scenarios) < 10 {
		t.Fatalf("scenario count = %d, want at least 10", len(scenarios))
	}

	for _, scenario := range scenarios {
		if !scenario.ValidResponse {
			continue
		}

		if err := directoryv1.ValidateMessage(scenario.Response); err != nil {
			t.Fatalf("%s fixture validation: %v", scenario.Name, err)
		}
	}

	first := fixtures.Scenarios()
	second := fixtures.Scenarios()

	first[0].Response.(*directoryv1.GetRosterSnapshotResponse).Users[0].UserId = 99
	if got := second[0].Response.(*directoryv1.GetRosterSnapshotResponse).Users[0].GetUserId(); got != 1 {
		t.Fatalf("fixture shares mutable state: got user ID %d, want 1", got)
	}
}
