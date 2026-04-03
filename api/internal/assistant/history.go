package assistant

import "github.com/dimon2255/agentic-ecommerce/api/pkg/anthropic"

// buildConversationMessages converts DB chat history to Anthropic RichMessages,
// caps at maxMessages, enforces alternating roles starting with user, and
// appends the current user message.
func buildConversationMessages(history []ChatMessage, currentMessage string, maxMessages int) []anthropic.RichMessage {
	// Exclude the current message if it was already saved to DB
	if len(history) > 0 {
		last := history[len(history)-1]
		if last.Role == "user" && last.Content == currentMessage {
			history = history[:len(history)-1]
		}
	}

	// Cap at maxMessages (keep most recent)
	if len(history) > maxMessages {
		history = history[len(history)-maxMessages:]
	}

	// Convert to RichMessages, enforcing alternation
	var messages []anthropic.RichMessage
	for _, m := range history {
		if len(messages) > 0 && messages[len(messages)-1].Role == m.Role {
			continue
		}
		messages = append(messages, anthropic.RichMessage{
			Role:    m.Role,
			Content: m.Content,
		})
	}

	// Ensure starts with user role
	if len(messages) > 0 && messages[0].Role != "user" {
		messages = messages[1:]
	}

	// Append current user message
	messages = append(messages, anthropic.RichMessage{
		Role:    "user",
		Content: currentMessage,
	})

	return messages
}
