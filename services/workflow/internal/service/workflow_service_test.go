package service

import "testing"

func TestFirstApprovalStep(t *testing.T) {
	step, err := firstApprovalStep(`[{"name":"Prepare","type":"AUTOMATED_ACTION"},{"name":"Manager review","type":"APPROVAL","role":"ROLE_MANAGER"}]`)
	if err != nil {
		t.Fatalf("expected a valid approval step: %v", err)
	}
	if step.Name != "Manager review" || step.Role != "ROLE_MANAGER" {
		t.Fatalf("unexpected approval step: %#v", step)
	}
}

func TestFirstApprovalStepRejectsInvalidConfiguration(t *testing.T) {
	cases := []string{
		`not-json`,
		`[{"name":"Prepare","type":"AUTOMATED_ACTION"}]`,
		`[{"name":"Manager review","type":"APPROVAL"}]`,
	}
	for _, raw := range cases {
		if _, err := firstApprovalStep(raw); err == nil {
			t.Fatalf("expected invalid configuration to fail: %s", raw)
		}
	}
}
