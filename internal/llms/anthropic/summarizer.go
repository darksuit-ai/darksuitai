package anthropic

import (
	"context"
	"strings"

	"github.com/darksuit-ai/darksuitai/internal/memory"

	"github.com/anthropics/anthropic-sdk-go"
)

// AnthropicSummarizer implements memory.Summarizer using the Anthropic SDK.
// It powers conversation compaction (Phase 4): folding older turns into a
// rolling, high-fidelity summary.
type AnthropicSummarizer struct {
	apiKey    string
	model     string
	maxTokens int64
}

// NewSummarizer builds an AnthropicSummarizer. If model is empty it defaults to
// claude-haiku-4-5 (summarization is a good fit for the fast, cheap tier).
func NewSummarizer(apiKey, model string) *AnthropicSummarizer {
	if model == "" {
		model = "claude-haiku-4-5"
	}
	return &AnthropicSummarizer{apiKey: apiKey, model: model, maxTokens: 1024}
}

const summarizerSystemPrompt = `You maintain a running summary of a conversation between a Human and an AI assistant.
You will be given the PRIOR SUMMARY (may be empty) and a batch of NEW TURNS.
Produce an updated summary that preserves all decision-relevant facts, names, numbers, decisions, open questions, and user preferences.
Be concise and high-fidelity. Output ONLY the updated summary text, with no preamble.`

// Summarize folds the given turns into an updated running summary.
func (s *AnthropicSummarizer) Summarize(ctx context.Context, priorSummary string, turns []memory.Turn) (string, error) {
	client, err := newSDKClient(s.apiKey)
	if err != nil {
		return "", err
	}

	var b strings.Builder
	b.WriteString("PRIOR SUMMARY:\n")
	if strings.TrimSpace(priorSummary) == "" {
		b.WriteString("(none)\n")
	} else {
		b.WriteString(priorSummary)
		b.WriteString("\n")
	}
	b.WriteString("\nNEW TURNS:\n")
	for _, t := range turns {
		if t.Human != "" {
			b.WriteString("Human: ")
			b.WriteString(t.Human)
			b.WriteString("\n")
		}
		if t.AI != "" {
			b.WriteString("AI: ")
			b.WriteString(t.AI)
			b.WriteString("\n")
		}
	}

	msg, err := client.Messages.New(ctx, anthropic.MessageNewParams{
		Model:     anthropic.Model(s.model),
		MaxTokens: s.maxTokens,
		System:    []anthropic.TextBlockParam{{Text: summarizerSystemPrompt}},
		Messages:  []anthropic.MessageParam{anthropic.NewUserMessage(anthropic.NewTextBlock(b.String()))},
	})
	if err != nil {
		return "", err
	}

	var out strings.Builder
	for _, block := range msg.Content {
		if block.Type == "text" {
			out.WriteString(block.Text)
		}
	}
	return strings.TrimSpace(out.String()), nil
}
