package mailer

import (
	"strings"
	"testing"
)

func TestBuildRegistrationCodeMessage(t *testing.T) {
	message := string(buildRegistrationCodeMessage(
		"noreply@example.com",
		"AutomaticTools",
		"user@example.com",
		"123456",
		10,
	))

	for _, expected := range []string{
		"From: AutomaticTools <noreply@example.com>",
		"To: user@example.com",
		"123456",
		"10 分钟内有效",
	} {
		if !strings.Contains(message, expected) {
			t.Fatalf("message does not contain %q", expected)
		}
	}
}
