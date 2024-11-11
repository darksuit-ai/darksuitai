package _stream

import (
	_"bytes"
	"context"
	"strings"
)

func _streamDifferentiator(ctx context.Context, writer *StreamWriter, llmStreamData *LLMResult) ([]byte, bool) {
	const (
		bufferSize     int    = 4
		toolCallKey    string = `<tool_call>`
		toolCallKeyAlt string = `thought`
		StreamStartKey string = `<answer>`
	)

	var (
		packedWords      [bufferSize]string
		packedWordsIndex int
		completeWords    strings.Builder
		buffer           strings.Builder
		foundToolCall    bool
		startStream      bool
		actionReadiness  bool = true
	)

	for syllable := range llmStreamData.LLMResponse {
		println(syllable)
		select {
		case <-ctx.Done():
			return nil, false
		default:
			if startStream {
				completeWords.Reset()
				actionReadiness = false
				processedSyllable := toRawStringLiteral(syllable)
				writer.Write([]byte(processedSyllable))
				continue
			}

			completeWords.WriteString(syllable)

			if !foundToolCall {
				trimmedSyllable := strings.TrimLeft(syllable, " ")
				if trimmedSyllable != "" {
					if packedWordsIndex < bufferSize {
						packedWords[packedWordsIndex] = trimmedSyllable
						packedWordsIndex++
					}

					if packedWordsIndex == bufferSize {
						buffer.Reset()
						for _, word := range packedWords {
							buffer.WriteString(word)
						}

						if strings.Contains(buffer.String(), toolCallKey) || 
						   strings.Contains(buffer.String(), toolCallKeyAlt) {
							foundToolCall = true
						} else {
							startStream = true
							writer.Write([]byte(toRawStringLiteral(buffer.String()+" ")))
						}
						packedWordsIndex = 0
					}
				}
			}
		}
	}

	if !actionReadiness {
		return nil, false
	}
	return []byte(completeWords.String()), true
}



func toRawStringLiteral(s string) string {
	replacer := strings.NewReplacer(
		// `\`, `\\`,
		"\n", `\n`,
		// "\r", `\r`,
		// "\t", `\t`,
		// `"`, `\"`,
	)
	return replacer.Replace(s)
}

// func (sw *StreamWriter) processStream(input []byte) []byte {
//     // If we haven't seen the opening tag yet
//     if !sw.SeenOpenTag {
//         if idx := bytes.Index(input, []byte(`<answer>`)); idx != -1 {
//             sw.SeenOpenTag = true
//             // Return everything after the opening tag
//             return append(bytes.TrimSpace(bytes.TrimPrefix(input[idx:], []byte(`<answer>`))), ' ')
//         }
//         return []byte(``)
//     }

//     return input
// }
