package eventbus

import "testing"

func TestStandardEventTopicsAreUniqueAndCoverTicketProjection(t *testing.T) {
	want := map[string]bool{
		TopicTicketCreated:       false,
		TopicTicketAssigned:      false,
		TopicTicketStatusChanged: false,
	}
	seen := make(map[string]struct{})
	for _, topic := range standardEventTopics() {
		if topic == "" {
			t.Fatal("standard event topic must not be empty")
		}
		if _, duplicate := seen[topic]; duplicate {
			t.Fatalf("duplicate standard event topic %q", topic)
		}
		seen[topic] = struct{}{}
		if _, required := want[topic]; required {
			want[topic] = true
		}
	}
	for topic, found := range want {
		if !found {
			t.Errorf("ticket projection topic %q is not bound to the dead-letter queue", topic)
		}
	}
}
