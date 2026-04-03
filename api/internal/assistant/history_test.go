package assistant

import (
	"testing"
	"time"
)

func TestBuildConversationMessages_Empty(t *testing.T) {
	msgs := buildConversationMessages(nil, "hello", 20)
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != "user" || msgs[0].Content != "hello" {
		t.Errorf("unexpected message: %+v", msgs[0])
	}
}

func TestBuildConversationMessages_WithHistory(t *testing.T) {
	history := []ChatMessage{
		{Role: "user", Content: "first", CreatedAt: time.Now()},
		{Role: "assistant", Content: "reply", CreatedAt: time.Now()},
	}
	msgs := buildConversationMessages(history, "second", 20)
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	if msgs[0].Content != "first" {
		t.Errorf("expected first history message, got %v", msgs[0].Content)
	}
	if msgs[2].Content != "second" {
		t.Errorf("expected current message last, got %v", msgs[2].Content)
	}
}

func TestBuildConversationMessages_ExcludesDuplicateCurrentMessage(t *testing.T) {
	history := []ChatMessage{
		{Role: "user", Content: "hello", CreatedAt: time.Now()},
	}
	msgs := buildConversationMessages(history, "hello", 20)
	// Should not have "hello" twice
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message (deduped), got %d", len(msgs))
	}
}

func TestBuildConversationMessages_CapsAtMax(t *testing.T) {
	var history []ChatMessage
	for i := 0; i < 30; i++ {
		role := "user"
		if i%2 == 1 {
			role = "assistant"
		}
		history = append(history, ChatMessage{Role: role, Content: "msg", CreatedAt: time.Now()})
	}
	msgs := buildConversationMessages(history, "current", 5)
	// 5 from history + 1 current = at most 6, but after dedup/alternation may be less
	if len(msgs) > 6 {
		t.Errorf("expected at most 6 messages, got %d", len(msgs))
	}
	// Last message should be current
	if msgs[len(msgs)-1].Content != "current" {
		t.Errorf("expected last message to be current")
	}
}

func TestBuildConversationMessages_EnforcesAlternation(t *testing.T) {
	history := []ChatMessage{
		{Role: "user", Content: "a", CreatedAt: time.Now()},
		{Role: "user", Content: "b", CreatedAt: time.Now()}, // duplicate role
		{Role: "assistant", Content: "c", CreatedAt: time.Now()},
	}
	msgs := buildConversationMessages(history, "d", 20)
	// Should skip "b" (consecutive user), result: user:a, assistant:c, user:d
	if len(msgs) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(msgs))
	}
	for i := 0; i < len(msgs)-1; i++ {
		if msgs[i].Role == msgs[i+1].Role {
			t.Errorf("roles not alternating at index %d: %s, %s", i, msgs[i].Role, msgs[i+1].Role)
		}
	}
}

func TestBuildConversationMessages_StartsWithUser(t *testing.T) {
	history := []ChatMessage{
		{Role: "assistant", Content: "orphan", CreatedAt: time.Now()},
		{Role: "user", Content: "a", CreatedAt: time.Now()},
		{Role: "assistant", Content: "b", CreatedAt: time.Now()},
	}
	msgs := buildConversationMessages(history, "c", 20)
	if msgs[0].Role != "user" {
		t.Errorf("expected first message role=user, got %s", msgs[0].Role)
	}
}
