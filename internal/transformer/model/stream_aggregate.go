package model

// MergeStreamingResponseAggregate folds a single streaming chunk into an
// aggregate completion-shaped response for logging, usage, and post-stream
// protocol transforms.
func MergeStreamingResponseAggregate(result, chunk *InternalLLMResponse) {
	if result == nil || chunk == nil {
		return
	}

	if chunk.ID != "" {
		result.ID = chunk.ID
	}
	if chunk.Created != 0 {
		result.Created = chunk.Created
	}
	if chunk.Model != "" {
		result.Model = chunk.Model
	}
	if chunk.SystemFingerprint != "" {
		result.SystemFingerprint = chunk.SystemFingerprint
	}
	if chunk.ServiceTier != "" {
		result.ServiceTier = chunk.ServiceTier
	}
	if chunk.Usage != nil {
		result.Usage = chunk.Usage
	}
	if chunk.Error != nil {
		result.Error = chunk.Error
	}

	for _, choice := range chunk.Choices {
		idx := ensureStreamingChoiceIndex(result, choice.Index)
		existingChoice := &result.Choices[idx]
		if existingChoice.Message == nil {
			existingChoice.Message = &Message{}
		}

		if choice.Message != nil {
			mergeStreamingMessage(existingChoice.Message, choice.Message)
		}
		if choice.Delta != nil && choice.Delta != choice.Message {
			mergeStreamingMessage(existingChoice.Message, choice.Delta)
		}

		if choice.FinishReason != nil {
			existingChoice.FinishReason = choice.FinishReason
		}
		mergeStreamingLogprobs(&existingChoice.Logprobs, choice.Logprobs)
	}
}

func mergeStreamingLogprobs(dst **LogprobsContent, src *LogprobsContent) {
	if src == nil {
		return
	}

	if *dst == nil {
		*dst = &LogprobsContent{}
	}

	for _, token := range src.Content {
		merged := TokenLogprob{
			Token:   token.Token,
			Logprob: token.Logprob,
		}
		if len(token.Bytes) > 0 {
			merged.Bytes = append([]int(nil), token.Bytes...)
		}
		if len(token.TopLogprobs) > 0 {
			merged.TopLogprobs = make([]TopLogprob, 0, len(token.TopLogprobs))
			for _, top := range token.TopLogprobs {
				copied := TopLogprob{
					Token:   top.Token,
					Logprob: top.Logprob,
				}
				if len(top.Bytes) > 0 {
					copied.Bytes = append([]int(nil), top.Bytes...)
				}
				merged.TopLogprobs = append(merged.TopLogprobs, copied)
			}
		}

		(*dst).Content = append((*dst).Content, merged)
	}
}

func ensureStreamingChoiceIndex(result *InternalLLMResponse, choiceIndex int) int {
	for idx := range result.Choices {
		if result.Choices[idx].Index == choiceIndex {
			return idx
		}
	}
	result.Choices = append(result.Choices, Choice{Index: choiceIndex, Message: &Message{}})
	return len(result.Choices) - 1
}

func mergeStreamingMessage(dst, src *Message) {
	if dst == nil || src == nil {
		return
	}

	if src.Role != "" {
		dst.Role = src.Role
	}
	if src.Name != nil && *src.Name != "" {
		dst.Name = src.Name
	}
	if src.ToolCallID != nil && *src.ToolCallID != "" {
		dst.ToolCallID = src.ToolCallID
	}
	if src.ToolCallName != nil && *src.ToolCallName != "" {
		dst.ToolCallName = src.ToolCallName
	}
	if src.ToolCallIsError != nil {
		dst.ToolCallIsError = src.ToolCallIsError
	}
	if src.Audio != nil {
		dst.Audio = src.Audio
	}

	if src.Content.Content != nil {
		appendStreamingTextContent(&dst.Content, *src.Content.Content)
	}
	if len(src.Content.MultipleContent) > 0 {
		appendStreamingContentParts(&dst.Content, src.Content.MultipleContent)
	}
	if len(src.Images) > 0 {
		appendStreamingContentParts(&dst.Content, src.Images)
	}

	if reasoning := src.GetReasoningContent(); reasoning != "" {
		if dst.ReasoningContent == nil {
			dst.ReasoningContent = new(string)
		}
		*dst.ReasoningContent += reasoning
	}
	if src.ReasoningSignature != nil && *src.ReasoningSignature != "" {
		if dst.ReasoningSignature == nil {
			dst.ReasoningSignature = new(string)
		}
		*dst.ReasoningSignature += *src.ReasoningSignature
	}

	for _, toolCall := range src.ToolCalls {
		dst.ToolCalls = mergeStreamingToolCall(dst.ToolCalls, toolCall)
	}

	if src.Refusal != "" {
		dst.Refusal += src.Refusal
	}
}

func appendStreamingTextContent(dst *MessageContent, text string) {
	if dst == nil || text == "" {
		return
	}

	if len(dst.MultipleContent) > 0 {
		last := len(dst.MultipleContent) - 1
		if dst.MultipleContent[last].Type == "text" {
			if dst.MultipleContent[last].Text == nil {
				dst.MultipleContent[last].Text = new(string)
			}
			*dst.MultipleContent[last].Text += text
			return
		}

		textCopy := text
		dst.MultipleContent = append(dst.MultipleContent, MessageContentPart{Type: "text", Text: &textCopy})
		return
	}

	if dst.Content == nil {
		dst.Content = new(string)
	}
	*dst.Content += text
}

func appendStreamingContentParts(dst *MessageContent, parts []MessageContentPart) {
	if dst == nil || len(parts) == 0 {
		return
	}

	if len(dst.MultipleContent) == 0 && !containsNonTextStreamingContent(parts) {
		for _, part := range parts {
			if part.Type == "text" && part.Text != nil {
				appendStreamingTextContent(dst, *part.Text)
			}
		}
		return
	}

	promoteStreamingTextContent(dst)
	dst.MultipleContent = append(dst.MultipleContent, parts...)
}

func promoteStreamingTextContent(dst *MessageContent) {
	if dst == nil || dst.Content == nil || *dst.Content == "" {
		return
	}

	text := *dst.Content
	dst.MultipleContent = append(dst.MultipleContent, MessageContentPart{Type: "text", Text: &text})
	dst.Content = nil
}

func containsNonTextStreamingContent(parts []MessageContentPart) bool {
	for _, part := range parts {
		if part.Type != "text" {
			return true
		}
	}

	return false
}

func mergeStreamingToolCall(toolCalls []ToolCall, delta ToolCall) []ToolCall {
	for i, tc := range toolCalls {
		if tc.Index == delta.Index {
			if delta.ID != "" {
				toolCalls[i].ID = delta.ID
			}
			if delta.Type != "" {
				toolCalls[i].Type = delta.Type
			}
			if delta.Function.Name != "" {
				toolCalls[i].Function.Name += delta.Function.Name
			}
			if delta.Function.Arguments != "" {
				toolCalls[i].Function.Arguments += delta.Function.Arguments
			}
			if delta.CacheControl != nil {
				toolCalls[i].CacheControl = delta.CacheControl
			}
			return toolCalls
		}
	}

	return append(toolCalls, delta)
}
