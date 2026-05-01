package openai

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/samber/lo"

	"github.com/xiaoli0412/octopus-xiaoli-repo/internal/transformer/model"
)

func TestConvertToLLMResponseFromResponsesPreservesMixedTextImageOrder(t *testing.T) {
	t.Parallel()

	status := "completed"
	text1 := "hello"
	text2 := "world"
	image := "ZmFrZQ=="
	format := "png"

	resp := convertToLLMResponseFromResponses(&ResponsesResponse{
		Object:    "response",
		ID:        "resp-mixed",
		Model:     "gpt-4.1",
		CreatedAt: 1,
		Status:    &status,
		Output: []ResponsesItem{
			{
				Type:    "message",
				Content: &ResponsesInput{Items: []ResponsesItem{{Type: "output_text", Text: &text1}}},
			},
			{
				Type:         "image_generation_call",
				Result:       &image,
				OutputFormat: &format,
			},
			{
				Type:    "message",
				Content: &ResponsesInput{Items: []ResponsesItem{{Type: "output_text", Text: &text2}}},
			},
		},
	})

	if resp == nil || len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatalf("response malformed: %#v", resp)
	}
	message := resp.Choices[0].Message
	if message.Content.Content != nil {
		t.Fatalf("content string = %#v, want multipart content", message.Content.Content)
	}
	if len(message.Content.MultipleContent) != 3 {
		t.Fatalf("len(MultipleContent) = %d, want 3", len(message.Content.MultipleContent))
	}
	if message.Content.MultipleContent[0].Type != "text" || message.Content.MultipleContent[0].Text == nil || *message.Content.MultipleContent[0].Text != "hello" {
		t.Fatalf("first content part = %#v, want leading text", message.Content.MultipleContent[0])
	}
	if message.Content.MultipleContent[1].Type != "image_url" || message.Content.MultipleContent[1].ImageURL == nil || message.Content.MultipleContent[1].ImageURL.URL != "data:image/png;base64,ZmFrZQ==" {
		t.Fatalf("second content part = %#v, want image", message.Content.MultipleContent[1])
	}
	if message.Content.MultipleContent[2].Type != "text" || message.Content.MultipleContent[2].Text == nil || *message.Content.MultipleContent[2].Text != "world" {
		t.Fatalf("third content part = %#v, want trailing text", message.Content.MultipleContent[2])
	}
}

func TestConvertToLLMResponseFromResponsesPreservesRefusal(t *testing.T) {
	t.Parallel()

	status := "completed"
	refusal := "cannot comply"

	resp := convertToLLMResponseFromResponses(&ResponsesResponse{
		Object:    "response",
		ID:        "resp-refusal",
		Model:     "gpt-4.1",
		CreatedAt: 2,
		Status:    &status,
		Output: []ResponsesItem{{
			Type:    "message",
			Role:    "assistant",
			Refusal: &refusal,
		}},
	})

	if resp == nil || len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatalf("response malformed: %#v", resp)
	}
	if got := resp.Choices[0].Message.Refusal; got != "cannot comply" {
		t.Fatalf("refusal = %q, want cannot comply", got)
	}
}

func TestConvertToLLMResponseFromResponsesAssignsSequentialToolCallIndexes(t *testing.T) {
	t.Parallel()

	status := "completed"

	resp := convertToLLMResponseFromResponses(&ResponsesResponse{
		Object:    "response",
		ID:        "resp-tools",
		Model:     "gpt-4.1",
		CreatedAt: 3,
		Status:    &status,
		Output: []ResponsesItem{
			{
				Type:      "function_call",
				CallID:    "call-weather",
				Name:      "weather_lookup",
				Arguments: "{\"city\":\"Tokyo\"}",
			},
			{
				Type:      "function_call",
				CallID:    "call-time",
				Name:      "time_lookup",
				Arguments: "{\"zone\":\"Asia/Tokyo\"}",
			},
		},
	})

	if resp == nil || len(resp.Choices) != 1 || resp.Choices[0].Message == nil {
		t.Fatalf("response malformed: %#v", resp)
	}
	toolCalls := resp.Choices[0].Message.ToolCalls
	if len(toolCalls) != 2 {
		t.Fatalf("tool call count = %d, want 2", len(toolCalls))
	}
	if toolCalls[0].Index != 0 || toolCalls[1].Index != 1 {
		t.Fatalf("tool call indexes = [%d %d], want [0 1]", toolCalls[0].Index, toolCalls[1].Index)
	}
}

func TestConvertToLLMResponseFromResponsesPreservesContentFilterIncompleteReason(t *testing.T) {
	t.Parallel()

	status := "incomplete"

	resp := convertToLLMResponseFromResponses(&ResponsesResponse{
		Object:            "response",
		ID:                "resp-content-filter",
		Model:             "gpt-4.1",
		CreatedAt:         4,
		Status:            &status,
		IncompleteDetails: &ResponsesIncompleteDetails{Reason: "content_filter"},
		Output: []ResponsesItem{{
			Type: "message",
			Role: "assistant",
			Content: &ResponsesInput{Items: []ResponsesItem{{
				Type: "output_text",
				Text: lo.ToPtr("filtered"),
			}}},
		}},
	})

	if resp == nil || len(resp.Choices) != 1 {
		t.Fatalf("response malformed: %#v", resp)
	}
	if resp.Choices[0].FinishReason == nil || *resp.Choices[0].FinishReason != "content_filter" {
		t.Fatalf("finish reason = %#v, want content_filter", resp.Choices[0].FinishReason)
	}
}

func TestResponseOutboundTransformStreamRestoresToolCallFromAddedPlusDoneWithoutDelta(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	responseID := "resp-tool-done"
	responseModel := "gpt-4.1"
	status := "in_progress"
	outputIndex := 0
	itemID := "item-call-1"
	arguments := "{\"city\":\"Tokyo\"}"

	events := []ResponsesStreamEvent{
		{
			Type: "response.created",
			Response: &ResponsesResponse{
				ID:     responseID,
				Model:  responseModel,
				Status: &status,
			},
		},
		{
			Type:        "response.output_item.added",
			OutputIndex: outputIndex,
			Item: &ResponsesItem{
				ID:     itemID,
				Type:   "function_call",
				CallID: "call-weather",
				Name:   "weather_lookup",
			},
		},
		{
			Type:        "response.function_call_arguments.done",
			OutputIndex: outputIndex,
			ItemID:      &itemID,
			CallID:      "call-weather",
			Arguments:   arguments,
		},
	}

	aggregated := &model.InternalLLMResponse{Object: "chat.completion"}
	for _, event := range events {
		payload, err := json.Marshal(event)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}
		chunk, err := outbound.TransformStream(context.Background(), payload)
		if err != nil {
			t.Fatalf("TransformStream() error = %v", err)
		}
		if chunk != nil {
			model.MergeStreamingResponseAggregate(aggregated, chunk)
		}
	}

	if len(aggregated.Choices) != 1 || aggregated.Choices[0].Message == nil {
		t.Fatalf("aggregated response malformed: %#v", aggregated)
	}
	toolCalls := aggregated.Choices[0].Message.ToolCalls
	if len(toolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(toolCalls))
	}
	if toolCalls[0].Index != 0 {
		t.Fatalf("tool call index = %d, want 0", toolCalls[0].Index)
	}
	if toolCalls[0].ID != "call-weather" {
		t.Fatalf("tool call id = %q, want call-weather", toolCalls[0].ID)
	}
	if toolCalls[0].Function.Name != "weather_lookup" {
		t.Fatalf("tool call name = %q, want weather_lookup", toolCalls[0].Function.Name)
	}
	if toolCalls[0].Function.Arguments != arguments {
		t.Fatalf("tool call arguments = %q, want %q", toolCalls[0].Function.Arguments, arguments)
	}
}

func TestResponseOutboundTransformStreamSkipsRedundantToolCallDoneAfterDelta(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 1
	itemID := "item-call-2"
	arguments := "{\"zone\":\"Asia/Tokyo\"}"
	name := "time_lookup"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-tool-done-dedupe",
			Model:  "gpt-4.1",
			Status: &status,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() created error = %v", err)
	}
	if _, err := outbound.TransformStream(context.Background(), createdPayload); err != nil {
		t.Fatalf("TransformStream() created error = %v", err)
	}

	addedPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.added",
		OutputIndex: outputIndex,
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "function_call",
			CallID: "call-time",
			Name:   name,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() added error = %v", err)
	}
	addedChunk, err := outbound.TransformStream(context.Background(), addedPayload)
	if err != nil {
		t.Fatalf("TransformStream() added error = %v", err)
	}
	if addedChunk == nil {
		t.Fatal("added chunk = nil, want initial tool-call chunk")
	}

	deltaPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.function_call_arguments.delta",
		OutputIndex: outputIndex,
		ItemID:      &itemID,
		CallID:      "call-time",
		Delta:       arguments,
	})
	if err != nil {
		t.Fatalf("json.Marshal() delta error = %v", err)
	}
	deltaChunk, err := outbound.TransformStream(context.Background(), deltaPayload)
	if err != nil {
		t.Fatalf("TransformStream() delta error = %v", err)
	}
	if deltaChunk == nil {
		t.Fatal("delta chunk = nil, want arguments delta chunk")
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.done",
		OutputIndex: outputIndex,
		Item: &ResponsesItem{
			ID:        itemID,
			Type:      "function_call",
			CallID:    "call-time",
			Name:      name,
			Arguments: arguments,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() done error = %v", err)
	}
	doneChunk, err := outbound.TransformStream(context.Background(), donePayload)
	if err != nil {
		t.Fatalf("TransformStream() done error = %v", err)
	}
	if doneChunk != nil {
		t.Fatalf("done chunk = %#v, want nil to avoid duplicate aggregation", doneChunk)
	}
}

func TestResponseOutboundTransformStreamRestoresToolCallFromDoneWithOnlyCallID(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 2
	arguments := "{\"city\":\"Tokyo\"}"
	name := "weather_lookup"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-tool-callid-only",
			Model:  "gpt-4.1",
			Status: &status,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() created error = %v", err)
	}
	if _, err := outbound.TransformStream(context.Background(), createdPayload); err != nil {
		t.Fatalf("TransformStream() created error = %v", err)
	}

	addedPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.added",
		OutputIndex: outputIndex,
		Item: &ResponsesItem{
			Type:   "function_call",
			CallID: "call-weather",
			Name:   name,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() added error = %v", err)
	}
	addedChunk, err := outbound.TransformStream(context.Background(), addedPayload)
	if err != nil {
		t.Fatalf("TransformStream() added error = %v", err)
	}
	if addedChunk == nil {
		t.Fatal("added chunk = nil, want initial tool-call chunk")
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.function_call_arguments.done",
		OutputIndex: outputIndex,
		CallID:      "call-weather",
		Arguments:   arguments,
	})
	if err != nil {
		t.Fatalf("json.Marshal() done error = %v", err)
	}
	doneChunk, err := outbound.TransformStream(context.Background(), donePayload)
	if err != nil {
		t.Fatalf("TransformStream() done error = %v", err)
	}
	if doneChunk == nil || len(doneChunk.Choices) != 1 || doneChunk.Choices[0].Delta == nil {
		t.Fatalf("done chunk = %#v, want fallback tool-call chunk", doneChunk)
	}
	toolCalls := doneChunk.Choices[0].Delta.ToolCalls
	if len(toolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(toolCalls))
	}
	if toolCalls[0].Index != 0 {
		t.Fatalf("tool call index = %d, want 0", toolCalls[0].Index)
	}
	if toolCalls[0].ID != "" {
		t.Fatalf("tool call id = %q, want empty because id was already seen", toolCalls[0].ID)
	}
	if toolCalls[0].Function.Name != "" {
		t.Fatalf("tool call name = %q, want empty because name was already seen", toolCalls[0].Function.Name)
	}
	if toolCalls[0].Function.Arguments != arguments {
		t.Fatalf("tool call arguments = %q, want %q", toolCalls[0].Function.Arguments, arguments)
	}
	if len(outbound.toolCallIndexByCallID) != 1 || outbound.toolCallIndexByCallID["call-weather"] != 0 {
		t.Fatalf("toolCallIndexByCallID = %#v, want call-weather -> 0", outbound.toolCallIndexByCallID)
	}
	if len(outbound.toolCallIndexByItemID) != 0 {
		t.Fatalf("toolCallIndexByItemID = %#v, want empty because no item_id was ever provided", outbound.toolCallIndexByItemID)
	}
	state := outbound.toolCallState(0)
	if !state.NameSeen || !state.ArgumentsSeen || state.CallID != "call-weather" {
		t.Fatalf("tool call state = %#v, want call_id + name + arguments tracked", state)
	}
	if state.ItemID != "" {
		t.Fatalf("tool call item id = %q, want empty", state.ItemID)
	}
}

func TestResponseOutboundTransformStreamSkipsToolCallDoneWithItemIDOnly(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 3
	itemID := "item-call-only"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-itemid-only",
			Model:  "gpt-4.1",
			Status: &status,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() created error = %v", err)
	}
	if _, err := outbound.TransformStream(context.Background(), createdPayload); err != nil {
		t.Fatalf("TransformStream() created error = %v", err)
	}

	addedPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.added",
		OutputIndex: outputIndex,
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "function_call",
			CallID: "call-weather",
			Name:   "weather_lookup",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() added error = %v", err)
	}
	addedChunk, err := outbound.TransformStream(context.Background(), addedPayload)
	if err != nil {
		t.Fatalf("TransformStream() added error = %v", err)
	}
	if addedChunk == nil {
		t.Fatal("added chunk = nil, want initial tool-call chunk")
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.done",
		OutputIndex: outputIndex,
		Item: &ResponsesItem{
			ID:   itemID,
			Type: "function_call",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() done error = %v", err)
	}
	doneChunk, err := outbound.TransformStream(context.Background(), donePayload)
	if err != nil {
		t.Fatalf("TransformStream() done error = %v", err)
	}
	if doneChunk != nil {
		t.Fatalf("done chunk = %#v, want nil for item_id-only done event", doneChunk)
	}
}

func TestResponseOutboundTransformStreamSkipsToolCallArgumentsDoneWithItemIDOnly(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 4
	itemID := "item-args-only"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-args-itemid-only",
			Model:  "gpt-4.1",
			Status: &status,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() created error = %v", err)
	}
	if _, err := outbound.TransformStream(context.Background(), createdPayload); err != nil {
		t.Fatalf("TransformStream() created error = %v", err)
	}

	addedPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.added",
		OutputIndex: outputIndex,
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "function_call",
			CallID: "call-args",
			Name:   "args_lookup",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() added error = %v", err)
	}
	addedChunk, err := outbound.TransformStream(context.Background(), addedPayload)
	if err != nil {
		t.Fatalf("TransformStream() added error = %v", err)
	}
	if addedChunk == nil {
		t.Fatal("added chunk = nil, want initial tool-call chunk")
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.function_call_arguments.done",
		OutputIndex: outputIndex,
		ItemID:      &itemID,
	})
	if err != nil {
		t.Fatalf("json.Marshal() done error = %v", err)
	}
	doneChunk, err := outbound.TransformStream(context.Background(), donePayload)
	if err != nil {
		t.Fatalf("TransformStream() done error = %v", err)
	}
	if doneChunk != nil {
		t.Fatalf("done chunk = %#v, want nil for item_id-only arguments.done event", doneChunk)
	}
}

func TestResponseOutboundTransformStreamRestoresMessageFromDoneWithoutDelta(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 5
	itemID := "item-message-only"
	text := "Recovered from done"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-done-only",
			Model:  "gpt-4.1",
			Status: &status,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() created error = %v", err)
	}
	if _, err := outbound.TransformStream(context.Background(), createdPayload); err != nil {
		t.Fatalf("TransformStream() created error = %v", err)
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.done",
		OutputIndex: outputIndex,
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
			Content: &ResponsesInput{Items: []ResponsesItem{{
				Type: "output_text",
				Text: lo.ToPtr(text),
			}}},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() done error = %v", err)
	}

	doneChunk, err := outbound.TransformStream(context.Background(), donePayload)
	if err != nil {
		t.Fatalf("TransformStream() done error = %v", err)
	}
	if doneChunk == nil || len(doneChunk.Choices) != 1 || doneChunk.Choices[0].Delta == nil {
		t.Fatalf("done chunk = %#v, want assistant text fallback chunk", doneChunk)
	}
	if doneChunk.Choices[0].Delta.Content.Content == nil || *doneChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("done chunk content = %#v, want %q", doneChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamSkipsMessageDoneAfterTextDelta(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 6
	itemID := "item-message-delta"
	text := "Already streamed"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-delta",
			Model:  "gpt-4.1",
			Status: &status,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() created error = %v", err)
	}
	if _, err := outbound.TransformStream(context.Background(), createdPayload); err != nil {
		t.Fatalf("TransformStream() created error = %v", err)
	}

	deltaPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.delta",
		OutputIndex:  outputIndex,
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Delta:        text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() delta error = %v", err)
	}
	deltaChunk, err := outbound.TransformStream(context.Background(), deltaPayload)
	if err != nil {
		t.Fatalf("TransformStream() delta error = %v", err)
	}
	if deltaChunk == nil || len(deltaChunk.Choices) != 1 || deltaChunk.Choices[0].Delta == nil {
		t.Fatalf("delta chunk = %#v, want assistant text delta chunk", deltaChunk)
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.done",
		OutputIndex: outputIndex,
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
			Content: &ResponsesInput{Items: []ResponsesItem{{
				Type: "output_text",
				Text: lo.ToPtr(text),
			}}},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() done error = %v", err)
	}
	doneChunk, err := outbound.TransformStream(context.Background(), donePayload)
	if err != nil {
		t.Fatalf("TransformStream() done error = %v", err)
	}
	if doneChunk != nil {
		t.Fatalf("done chunk = %#v, want nil to avoid duplicate assistant text", doneChunk)
	}
}

func TestResponseOutboundTransformStreamRestoresOutputTextDoneWithoutDelta(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 7
	itemID := "item-output-text-only"
	text := "Recovered from output_text.done"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-output-text-done-only",
			Model:  "gpt-4.1",
			Status: &status,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() created error = %v", err)
	}
	if _, err := outbound.TransformStream(context.Background(), createdPayload); err != nil {
		t.Fatalf("TransformStream() created error = %v", err)
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  outputIndex,
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() done error = %v", err)
	}

	doneChunk, err := outbound.TransformStream(context.Background(), donePayload)
	if err != nil {
		t.Fatalf("TransformStream() done error = %v", err)
	}
	if doneChunk == nil || len(doneChunk.Choices) != 1 || doneChunk.Choices[0].Delta == nil {
		t.Fatalf("done chunk = %#v, want assistant text fallback chunk", doneChunk)
	}
	if doneChunk.Choices[0].Delta.Content.Content == nil || *doneChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("done chunk content = %#v, want %q", doneChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamSkipsOutputTextDoneAfterTextDelta(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 8
	itemID := "item-output-text-delta"
	text := "Already streamed via delta"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-output-text-delta",
			Model:  "gpt-4.1",
			Status: &status,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() created error = %v", err)
	}
	if _, err := outbound.TransformStream(context.Background(), createdPayload); err != nil {
		t.Fatalf("TransformStream() created error = %v", err)
	}

	deltaPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.delta",
		OutputIndex:  outputIndex,
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Delta:        text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() delta error = %v", err)
	}
	deltaChunk, err := outbound.TransformStream(context.Background(), deltaPayload)
	if err != nil {
		t.Fatalf("TransformStream() delta error = %v", err)
	}
	if deltaChunk == nil || len(deltaChunk.Choices) != 1 || deltaChunk.Choices[0].Delta == nil {
		t.Fatalf("delta chunk = %#v, want assistant text delta chunk", deltaChunk)
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  outputIndex,
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() done error = %v", err)
	}
	doneChunk, err := outbound.TransformStream(context.Background(), donePayload)
	if err != nil {
		t.Fatalf("TransformStream() done error = %v", err)
	}
	if doneChunk != nil {
		t.Fatalf("done chunk = %#v, want nil to avoid duplicate assistant text", doneChunk)
	}
}

func TestResponseOutboundTransformStreamMapsIncompleteAndFailedTerminalEvents(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		eventType        string
		status           string
		incompleteReason string
		wantFinishReason string
	}{
		{name: "incomplete", eventType: "response.incomplete", status: "incomplete", wantFinishReason: "length"},
		{name: "content_filter", eventType: "response.incomplete", status: "incomplete", incompleteReason: "content_filter", wantFinishReason: "content_filter"},
		{name: "failed", eventType: "response.failed", status: "failed", wantFinishReason: "error"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outbound := &ResponseOutbound{}
			payload, err := json.Marshal(ResponsesStreamEvent{
				Type: tt.eventType,
				Response: &ResponsesResponse{
					ID:                "resp-terminal",
					Model:             "gpt-4.1",
					Status:            &tt.status,
					IncompleteDetails: lo.Ternary(tt.incompleteReason != "", &ResponsesIncompleteDetails{Reason: tt.incompleteReason}, nil),
					Usage: &ResponsesUsage{
						InputTokens:  2,
						OutputTokens: 1,
						TotalTokens:  3,
					},
				},
			})
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			chunk, err := outbound.TransformStream(context.Background(), payload)
			if err != nil {
				t.Fatalf("TransformStream() error = %v", err)
			}
			if chunk == nil || len(chunk.Choices) != 1 {
				t.Fatalf("chunk = %#v, want one terminal choice", chunk)
			}
			if chunk.Choices[0].FinishReason == nil || *chunk.Choices[0].FinishReason != tt.wantFinishReason {
				t.Fatalf("finish reason = %#v, want %s", chunk.Choices[0].FinishReason, tt.wantFinishReason)
			}
			if chunk.Usage == nil || chunk.Usage.TotalTokens != 3 {
				t.Fatalf("usage = %#v, want total tokens 3", chunk.Usage)
			}
		})
	}
}
