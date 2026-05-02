package openai

import (
	"context"
	"encoding/json"
	"fmt"
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
			OutputIndex: lo.ToPtr(outputIndex),
			Item: &ResponsesItem{
				ID:     itemID,
				Type:   "function_call",
				CallID: "call-weather",
				Name:   "weather_lookup",
			},
		},
		{
			Type:        "response.function_call_arguments.done",
			OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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

func TestResponseOutboundTransformStreamRestoresToolCallFromDoneWithoutPriorState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 4
	arguments := "{\"city\":\"Tokyo\"}"
	name := "weather_lookup"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-toolcall-done-only",
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
		OutputIndex: lo.ToPtr(outputIndex),
		Item: &ResponsesItem{
			Type:      "function_call",
			CallID:    "call-weather",
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
	if toolCalls[0].ID != "call-weather" {
		t.Fatalf("tool call id = %q, want call-weather on first done-only fallback", toolCalls[0].ID)
	}
	if toolCalls[0].Function.Name != name {
		t.Fatalf("tool call name = %q, want %q", toolCalls[0].Function.Name, name)
	}
	if toolCalls[0].Function.Arguments != arguments {
		t.Fatalf("tool call arguments = %q, want %q", toolCalls[0].Function.Arguments, arguments)
	}
	if len(outbound.toolCallIndexByCallID) != 1 || outbound.toolCallIndexByCallID["call-weather"] != 0 {
		t.Fatalf("toolCallIndexByCallID = %#v, want call-weather -> 0", outbound.toolCallIndexByCallID)
	}
	state := outbound.toolCallState(0)
	if state.CallID != "call-weather" || !state.NameSeen || !state.ArgumentsSeen {
		t.Fatalf("tool call state = %#v, want first done-only fallback to record call_id + name + arguments", state)
	}
	if state.AddedSeen {
		t.Fatalf("tool call state = %#v, want AddedSeen=false for done-only fallback", state)
	}

	lateAddedPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.added",
		OutputIndex: lo.ToPtr(outputIndex),
		Item: &ResponsesItem{
			Type:   "function_call",
			CallID: "call-weather",
			Name:   name,
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() late added error = %v", err)
	}
	lateAddedChunk, err := outbound.TransformStream(context.Background(), lateAddedPayload)
	if err != nil {
		t.Fatalf("TransformStream() late added error = %v", err)
	}
	if lateAddedChunk != nil {
		t.Fatalf("late added chunk = %#v, want nil after done-only fallback already recorded call_id + name", lateAddedChunk)
	}
	if !outbound.toolCallState(0).AddedSeen {
		t.Fatalf("tool call state = %#v, want AddedSeen=true after late added marker", outbound.toolCallState(0))
	}
}

func TestResponseOutboundTransformStreamRestoresToolCallArgumentsDoneWithOnlyCallIDWithoutPriorState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 5
	arguments := "{\"zone\":\"Asia/Tokyo\"}"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-toolcall-args-done-only",
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
		Type:        "response.function_call_arguments.done",
		OutputIndex: lo.ToPtr(outputIndex),
		CallID:      "call-time",
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
	if toolCalls[0].ID != "call-time" {
		t.Fatalf("tool call id = %q, want call-time on first arguments.done fallback", toolCalls[0].ID)
	}
	if toolCalls[0].Function.Name != "" {
		t.Fatalf("tool call name = %q, want empty because name was unavailable", toolCalls[0].Function.Name)
	}
	if toolCalls[0].Function.Arguments != arguments {
		t.Fatalf("tool call arguments = %q, want %q", toolCalls[0].Function.Arguments, arguments)
	}
	if len(outbound.toolCallIndexByCallID) != 1 || outbound.toolCallIndexByCallID["call-time"] != 0 {
		t.Fatalf("toolCallIndexByCallID = %#v, want call-time -> 0", outbound.toolCallIndexByCallID)
	}
	state := outbound.toolCallState(0)
	if state.CallID != "call-time" || state.NameSeen || !state.ArgumentsSeen {
		t.Fatalf("tool call state = %#v, want first arguments.done fallback to record call_id + arguments only", state)
	}
	if state.AddedSeen {
		t.Fatalf("tool call state = %#v, want AddedSeen=false for arguments.done-only fallback", state)
	}
}

func TestResponseOutboundTransformStreamTracksToolCallDoneIdentityFromTopLevelItemID(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 3
	itemID := "item-toolcall-top-level-itemid"
	arguments := "{\"city\":\"Seoul\"}"
	name := "weather_lookup"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-toolcall-top-level-itemid",
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		Type:        "response.output_item.done",
		OutputIndex: lo.ToPtr(outputIndex),
		ItemID:      &itemID,
		Item: &ResponsesItem{
			Type:      "function_call",
			CallID:    "call-weather",
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
	if doneChunk == nil || len(doneChunk.Choices) != 1 || doneChunk.Choices[0].Delta == nil {
		t.Fatalf("done chunk = %#v, want fallback tool-call chunk", doneChunk)
	}
	toolCalls := doneChunk.Choices[0].Delta.ToolCalls
	if len(toolCalls) != 1 {
		t.Fatalf("tool call count = %d, want 1", len(toolCalls))
	}
	if toolCalls[0].Function.Arguments != arguments {
		t.Fatalf("tool call arguments = %q, want %q", toolCalls[0].Function.Arguments, arguments)
	}
	if idx, ok := outbound.toolCallIndexByItemID[itemID]; !ok || idx != 0 {
		t.Fatalf("toolCallIndexByItemID = %#v, want %q -> 0 after top-level item_id binding", outbound.toolCallIndexByItemID, itemID)
	}

	duplicatePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:      "response.function_call_arguments.done",
		ItemID:    &itemID,
		Arguments: arguments,
	})
	if err != nil {
		t.Fatalf("json.Marshal() duplicate error = %v", err)
	}
	duplicateChunk, err := outbound.TransformStream(context.Background(), duplicatePayload)
	if err != nil {
		t.Fatalf("TransformStream() duplicate error = %v", err)
	}
	if duplicateChunk != nil {
		t.Fatalf("duplicate chunk = %#v, want nil after top-level item_id state was recorded", duplicateChunk)
	}
	if outbound.nextToolCallIndex != 1 {
		t.Fatalf("nextToolCallIndex = %d, want 1 without duplicate identity allocation", outbound.nextToolCallIndex)
	}
	state := outbound.toolCallState(0)
	if state.ItemID != itemID || !state.NameSeen || !state.ArgumentsSeen || state.CallID != "call-weather" {
		t.Fatalf("tool call state = %#v, want item_id + call_id + name + arguments tracked on the first tool-call state", state)
	}
}

func TestResponseOutboundTransformStreamTracksToolCallAddedIdentityFromTopLevelItemID(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	itemID := "item-toolcall-added-top-level-itemid"
	arguments := "{\"city\":\"Busan\"}"
	name := "weather_lookup"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-toolcall-added-top-level-itemid",
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
		Type:   "response.output_item.added",
		ItemID: &itemID,
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
	if addedChunk == nil || len(addedChunk.Choices) != 1 || addedChunk.Choices[0].Delta == nil {
		t.Fatalf("added chunk = %#v, want initial tool-call chunk", addedChunk)
	}
	if idx, ok := outbound.toolCallIndexByItemID[itemID]; !ok || idx != 0 {
		t.Fatalf("toolCallIndexByItemID = %#v, want %q -> 0 after top-level item_id added event", outbound.toolCallIndexByItemID, itemID)
	}

	donePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:      "response.function_call_arguments.done",
		ItemID:    &itemID,
		Arguments: arguments,
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
		t.Fatalf("tool call index = %d, want 0 reusing top-level item_id binding from added event", toolCalls[0].Index)
	}
	if toolCalls[0].Function.Arguments != arguments {
		t.Fatalf("tool call arguments = %q, want %q", toolCalls[0].Function.Arguments, arguments)
	}
	if outbound.nextToolCallIndex != 1 {
		t.Fatalf("nextToolCallIndex = %d, want 1 without duplicate identity allocation", outbound.nextToolCallIndex)
	}
	state := outbound.toolCallState(0)
	if state.ItemID != itemID || !state.NameSeen || !state.ArgumentsSeen || state.CallID != "call-weather" {
		t.Fatalf("tool call state = %#v, want item_id + call_id + name + arguments tracked on the first tool-call state", state)
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex:  lo.ToPtr(outputIndex),
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
		OutputIndex: lo.ToPtr(outputIndex),
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
		OutputIndex:  lo.ToPtr(outputIndex),
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
		OutputIndex:  lo.ToPtr(outputIndex),
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
		OutputIndex:  lo.ToPtr(outputIndex),
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

func TestResponseOutboundTransformStreamSkipsOutputTextDoneWithoutTextWithoutPoisoningState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 8
	itemID := "item-output-text-empty-done"
	text := "Recovered after empty output_text.done"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-output-text-empty-done",
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

	ignoredPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
	})
	if err != nil {
		t.Fatalf("json.Marshal() ignored error = %v", err)
	}
	ignoredChunk, err := outbound.TransformStream(context.Background(), ignoredPayload)
	if err != nil {
		t.Fatalf("TransformStream() ignored error = %v", err)
	}
	if ignoredChunk != nil {
		t.Fatalf("ignored chunk = %#v, want nil for empty output_text.done", ignoredChunk)
	}

	validPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() valid error = %v", err)
	}
	validChunk, err := outbound.TransformStream(context.Background(), validPayload)
	if err != nil {
		t.Fatalf("TransformStream() valid error = %v", err)
	}
	if validChunk == nil || len(validChunk.Choices) != 1 || validChunk.Choices[0].Delta == nil {
		t.Fatalf("valid chunk = %#v, want assistant text fallback chunk", validChunk)
	}
	if validChunk.Choices[0].Delta.Content.Content == nil || *validChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("valid chunk content = %#v, want %q", validChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamRestoresContentPartDoneWithoutDelta(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 9
	itemID := "item-content-part-done-only"
	text := "Recovered from content_part.done"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-content-part-done-only",
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
		Type:         "response.content_part.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Part: &ResponsesContentPart{
			Type: "output_text",
			Text: lo.ToPtr(text),
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

func TestResponseOutboundTransformStreamRestoresOutputTextDoneWithItemIDOnly(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	itemID := "item-output-text-itemid-only"
	text := "Recovered from output_text.done with item_id only"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-output-text-itemid-only",
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

	donePayload2, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		ItemID:       lo.ToPtr("item-output-text-itemid-only-2"),
		ContentIndex: lo.ToPtr(0),
		Text:         "Recovered from second item_id-only done",
	})
	if err != nil {
		t.Fatalf("json.Marshal() second done error = %v", err)
	}
	doneChunk2, err := outbound.TransformStream(context.Background(), donePayload2)
	if err != nil {
		t.Fatalf("TransformStream() second done error = %v", err)
	}
	if doneChunk2 == nil || len(doneChunk2.Choices) != 1 || doneChunk2.Choices[0].Delta == nil {
		t.Fatalf("second done chunk = %#v, want second assistant text fallback chunk", doneChunk2)
	}
	if doneChunk2.Choices[0].Delta.Content.Content == nil || *doneChunk2.Choices[0].Delta.Content.Content != "Recovered from second item_id-only done" {
		t.Fatalf("second done chunk content = %#v, want second item text", doneChunk2.Choices[0].Delta.Content.Content)
	}
}

func TestResponseOutboundTransformStreamSkipsContentPartDoneAfterTextDelta(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 10
	itemID := "item-content-part-delta"
	text := "Already streamed before content_part.done"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-content-part-delta",
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
		OutputIndex:  lo.ToPtr(outputIndex),
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
		Type:         "response.content_part.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Part: &ResponsesContentPart{
			Type: "output_text",
			Text: lo.ToPtr(text),
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

func TestResponseOutboundTransformStreamRestoresContentPartDoneWithItemIDOnly(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-content-part-itemid-only",
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

	tests := []struct {
		name   string
		itemID string
		text   string
	}{
		{name: "first", itemID: "item-content-part-itemid-only-1", text: "Recovered from first content_part.done item_id only"},
		{name: "second", itemID: "item-content-part-itemid-only-2", text: "Recovered from second content_part.done item_id only"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			donePayload, err := json.Marshal(ResponsesStreamEvent{
				Type:         "response.content_part.done",
				ItemID:       lo.ToPtr(tt.itemID),
				ContentIndex: lo.ToPtr(0),
				Part: &ResponsesContentPart{
					Type: "output_text",
					Text: lo.ToPtr(tt.text),
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
			if doneChunk.Choices[0].Delta.Content.Content == nil || *doneChunk.Choices[0].Delta.Content.Content != tt.text {
				t.Fatalf("done chunk content = %#v, want %q", doneChunk.Choices[0].Delta.Content.Content, tt.text)
			}
		})
	}
}

func TestResponseOutboundTransformStreamDeduplicatesOutputTextDoneAndContentPartDoneByContentIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		firstEvent  string
		secondEvent string
	}{
		{name: "output_text_then_content_part", firstEvent: "response.output_text.done", secondEvent: "response.content_part.done"},
		{name: "content_part_then_output_text", firstEvent: "response.content_part.done", secondEvent: "response.output_text.done"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outbound := &ResponseOutbound{}
			status := "in_progress"
			outputIndex := 17
			itemID := "item-message-cross-dedup-" + tt.name
			text := "deduped message text"

			createdPayload, err := json.Marshal(ResponsesStreamEvent{
				Type: "response.created",
				Response: &ResponsesResponse{
					ID:     "resp-message-cross-dedup-" + tt.name,
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

			buildPayload := func(eventType string) ([]byte, error) {
				switch eventType {
				case "response.output_text.done":
					return json.Marshal(ResponsesStreamEvent{
						Type:         eventType,
						OutputIndex:  lo.ToPtr(outputIndex),
						ItemID:       &itemID,
						ContentIndex: lo.ToPtr(0),
						Text:         text,
					})
				case "response.content_part.done":
					return json.Marshal(ResponsesStreamEvent{
						Type:         eventType,
						OutputIndex:  lo.ToPtr(outputIndex),
						ItemID:       &itemID,
						ContentIndex: lo.ToPtr(0),
						Part: &ResponsesContentPart{
							Type: "output_text",
							Text: lo.ToPtr(text),
						},
					})
				default:
					return nil, fmt.Errorf("unexpected event type %q", eventType)
				}
			}

			firstPayload, err := buildPayload(tt.firstEvent)
			if err != nil {
				t.Fatalf("json.Marshal() first error = %v", err)
			}
			firstChunk, err := outbound.TransformStream(context.Background(), firstPayload)
			if err != nil {
				t.Fatalf("TransformStream() first error = %v", err)
			}
			if firstChunk == nil || len(firstChunk.Choices) != 1 || firstChunk.Choices[0].Delta == nil {
				t.Fatalf("first chunk = %#v, want assistant text fallback chunk", firstChunk)
			}
			if firstChunk.Choices[0].Delta.Content.Content == nil || *firstChunk.Choices[0].Delta.Content.Content != text {
				t.Fatalf("first chunk content = %#v, want %q", firstChunk.Choices[0].Delta.Content.Content, text)
			}

			secondPayload, err := buildPayload(tt.secondEvent)
			if err != nil {
				t.Fatalf("json.Marshal() second error = %v", err)
			}
			secondChunk, err := outbound.TransformStream(context.Background(), secondPayload)
			if err != nil {
				t.Fatalf("TransformStream() second error = %v", err)
			}
			if secondChunk != nil {
				t.Fatalf("second chunk = %#v, want nil after first %s event already recorded content_index", secondChunk, tt.firstEvent)
			}
			if !outbound.messageState(lo.ToPtr(outputIndex), &itemID).contentSeen(lo.ToPtr(0)) {
				t.Fatalf("message state did not record content index after %s then %s", tt.firstEvent, tt.secondEvent)
			}
		})
	}
}

func TestResponseOutboundTransformStreamTracksMixedDoneEventsPerDistinctContentIndex(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		firstEvent  string
		secondEvent string
	}{
		{name: "output_text_then_content_part", firstEvent: "response.output_text.done", secondEvent: "response.content_part.done"},
		{name: "content_part_then_output_text", firstEvent: "response.content_part.done", secondEvent: "response.output_text.done"},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			outbound := &ResponseOutbound{}
			status := "in_progress"
			outputIndex := 18
			itemID := "item-message-mixed-distinct-index-" + tt.name
			firstText := "first content index"
			secondText := "second content index"

			createdPayload, err := json.Marshal(ResponsesStreamEvent{
				Type: "response.created",
				Response: &ResponsesResponse{
					ID:     "resp-message-mixed-distinct-index-" + tt.name,
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

			buildPayload := func(eventType string, contentIndex int, text string) ([]byte, error) {
				switch eventType {
				case "response.output_text.done":
					return json.Marshal(ResponsesStreamEvent{
						Type:         eventType,
						OutputIndex:  lo.ToPtr(outputIndex),
						ItemID:       &itemID,
						ContentIndex: lo.ToPtr(contentIndex),
						Text:         text,
					})
				case "response.content_part.done":
					return json.Marshal(ResponsesStreamEvent{
						Type:         eventType,
						OutputIndex:  lo.ToPtr(outputIndex),
						ItemID:       &itemID,
						ContentIndex: lo.ToPtr(contentIndex),
						Part: &ResponsesContentPart{
							Type: "output_text",
							Text: lo.ToPtr(text),
						},
					})
				default:
					return nil, fmt.Errorf("unexpected event type %q", eventType)
				}
			}

			firstPayload, err := buildPayload(tt.firstEvent, 0, firstText)
			if err != nil {
				t.Fatalf("json.Marshal() first error = %v", err)
			}
			firstChunk, err := outbound.TransformStream(context.Background(), firstPayload)
			if err != nil {
				t.Fatalf("TransformStream() first error = %v", err)
			}
			if firstChunk == nil || len(firstChunk.Choices) != 1 || firstChunk.Choices[0].Delta == nil {
				t.Fatalf("first chunk = %#v, want assistant text fallback chunk", firstChunk)
			}
			if firstChunk.Choices[0].Delta.Content.Content == nil || *firstChunk.Choices[0].Delta.Content.Content != firstText {
				t.Fatalf("first chunk content = %#v, want %q", firstChunk.Choices[0].Delta.Content.Content, firstText)
			}

			secondPayload, err := buildPayload(tt.secondEvent, 1, secondText)
			if err != nil {
				t.Fatalf("json.Marshal() second error = %v", err)
			}
			secondChunk, err := outbound.TransformStream(context.Background(), secondPayload)
			if err != nil {
				t.Fatalf("TransformStream() second error = %v", err)
			}
			if secondChunk == nil || len(secondChunk.Choices) != 1 || secondChunk.Choices[0].Delta == nil {
				t.Fatalf("second chunk = %#v, want assistant text fallback chunk", secondChunk)
			}
			if secondChunk.Choices[0].Delta.Content.Content == nil || *secondChunk.Choices[0].Delta.Content.Content != secondText {
				t.Fatalf("second chunk content = %#v, want %q", secondChunk.Choices[0].Delta.Content.Content, secondText)
			}

			messageDonePayload, err := json.Marshal(ResponsesStreamEvent{
				Type:        "response.output_item.done",
				OutputIndex: lo.ToPtr(outputIndex),
				ItemID:      &itemID,
				Item: &ResponsesItem{
					ID:     itemID,
					Type:   "message",
					Role:   "assistant",
					Status: lo.ToPtr("completed"),
					Content: &ResponsesInput{Items: []ResponsesItem{
						{Type: "output_text", Text: lo.ToPtr(firstText)},
						{Type: "output_text", Text: lo.ToPtr(secondText)},
					}},
				},
			})
			if err != nil {
				t.Fatalf("json.Marshal() message done error = %v", err)
			}
			messageDoneChunk, err := outbound.TransformStream(context.Background(), messageDonePayload)
			if err != nil {
				t.Fatalf("TransformStream() message done error = %v", err)
			}
			if messageDoneChunk != nil {
				t.Fatalf("message done chunk = %#v, want nil after both content indexes were recorded", messageDoneChunk)
			}

			if !outbound.messageState(lo.ToPtr(outputIndex), &itemID).contentSeen(lo.ToPtr(0)) {
				t.Fatalf("message state did not record content index 0 after %s", tt.firstEvent)
			}
			if !outbound.messageState(lo.ToPtr(outputIndex), &itemID).contentSeen(lo.ToPtr(1)) {
				t.Fatalf("message state did not record content index 1 after %s", tt.secondEvent)
			}
		})
	}
}

func TestResponseOutboundTransformStreamSkipsNonOutputTextContentPartDoneWithoutPoisoningState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	itemID := "item-content-part-summary-only"
	text := "Recovered after ignored summary_text part"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-content-part-summary-only",
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

	ignoredPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.content_part.done",
		ItemID:       lo.ToPtr(itemID),
		ContentIndex: lo.ToPtr(0),
		Part: &ResponsesContentPart{
			Type: "summary_text",
			Text: lo.ToPtr("internal reasoning summary"),
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() ignored error = %v", err)
	}
	ignoredChunk, err := outbound.TransformStream(context.Background(), ignoredPayload)
	if err != nil {
		t.Fatalf("TransformStream() ignored error = %v", err)
	}
	if ignoredChunk != nil {
		t.Fatalf("ignored chunk = %#v, want nil for non-output_text content part", ignoredChunk)
	}

	validPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.content_part.done",
		ItemID:       lo.ToPtr(itemID),
		ContentIndex: lo.ToPtr(1),
		Part: &ResponsesContentPart{
			Type: "output_text",
			Text: lo.ToPtr(text),
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() valid error = %v", err)
	}
	validChunk, err := outbound.TransformStream(context.Background(), validPayload)
	if err != nil {
		t.Fatalf("TransformStream() valid error = %v", err)
	}
	if validChunk == nil || len(validChunk.Choices) != 1 || validChunk.Choices[0].Delta == nil {
		t.Fatalf("valid chunk = %#v, want assistant text fallback chunk", validChunk)
	}
	if validChunk.Choices[0].Delta.Content.Content == nil || *validChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("valid chunk content = %#v, want %q", validChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamSkipsContentPartDoneWithoutTextWithoutPoisoningState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	itemID := "item-content-part-nil-text"
	text := "Recovered after nil-text content part"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-content-part-nil-text",
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

	ignoredPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.content_part.done",
		ItemID:       lo.ToPtr(itemID),
		ContentIndex: lo.ToPtr(0),
		Part: &ResponsesContentPart{
			Type: "output_text",
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() ignored error = %v", err)
	}
	ignoredChunk, err := outbound.TransformStream(context.Background(), ignoredPayload)
	if err != nil {
		t.Fatalf("TransformStream() ignored error = %v", err)
	}
	if ignoredChunk != nil {
		t.Fatalf("ignored chunk = %#v, want nil for nil-text output_text content part", ignoredChunk)
	}

	validPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.content_part.done",
		ItemID:       lo.ToPtr(itemID),
		ContentIndex: lo.ToPtr(1),
		Part: &ResponsesContentPart{
			Type: "output_text",
			Text: lo.ToPtr(text),
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() valid error = %v", err)
	}
	validChunk, err := outbound.TransformStream(context.Background(), validPayload)
	if err != nil {
		t.Fatalf("TransformStream() valid error = %v", err)
	}
	if validChunk == nil || len(validChunk.Choices) != 1 || validChunk.Choices[0].Delta == nil {
		t.Fatalf("valid chunk = %#v, want assistant text fallback chunk", validChunk)
	}
	if validChunk.Choices[0].Delta.Content.Content == nil || *validChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("valid chunk content = %#v, want %q", validChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamSkipsContentPartDoneWithEmptyTextWithoutPoisoningState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	itemID := "item-content-part-empty-text"
	text := "Recovered after empty-string content part"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-content-part-empty-text",
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

	ignoredPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.content_part.done",
		ItemID:       lo.ToPtr(itemID),
		ContentIndex: lo.ToPtr(0),
		Part: &ResponsesContentPart{
			Type: "output_text",
			Text: lo.ToPtr(""),
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() ignored error = %v", err)
	}
	ignoredChunk, err := outbound.TransformStream(context.Background(), ignoredPayload)
	if err != nil {
		t.Fatalf("TransformStream() ignored error = %v", err)
	}
	if ignoredChunk != nil {
		t.Fatalf("ignored chunk = %#v, want nil for empty-string output_text content part", ignoredChunk)
	}

	validPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.content_part.done",
		ItemID:       lo.ToPtr(itemID),
		ContentIndex: lo.ToPtr(1),
		Part: &ResponsesContentPart{
			Type: "output_text",
			Text: lo.ToPtr(text),
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() valid error = %v", err)
	}
	validChunk, err := outbound.TransformStream(context.Background(), validPayload)
	if err != nil {
		t.Fatalf("TransformStream() valid error = %v", err)
	}
	if validChunk == nil || len(validChunk.Choices) != 1 || validChunk.Choices[0].Delta == nil {
		t.Fatalf("valid chunk = %#v, want assistant text fallback chunk", validChunk)
	}
	if validChunk.Choices[0].Delta.Content.Content == nil || *validChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("valid chunk content = %#v, want %q", validChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamRestoresMessageDoneWithItemIDOnly(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	itemID := "item-message-itemid-only"
	text := "Recovered from message.done with item_id only"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-itemid-only",
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
		Type: "response.output_item.done",
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

	donePayload2, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.output_item.done",
		Item: &ResponsesItem{
			ID:     "item-message-itemid-only-2",
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
			Content: &ResponsesInput{Items: []ResponsesItem{{
				Type: "output_text",
				Text: lo.ToPtr("Recovered from second message item_id-only done"),
			}}},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() second done error = %v", err)
	}
	doneChunk2, err := outbound.TransformStream(context.Background(), donePayload2)
	if err != nil {
		t.Fatalf("TransformStream() second done error = %v", err)
	}
	if doneChunk2 == nil || len(doneChunk2.Choices) != 1 || doneChunk2.Choices[0].Delta == nil {
		t.Fatalf("second done chunk = %#v, want second assistant text fallback chunk", doneChunk2)
	}
	if doneChunk2.Choices[0].Delta.Content.Content == nil || *doneChunk2.Choices[0].Delta.Content.Content != "Recovered from second message item_id-only done" {
		t.Fatalf("second done chunk content = %#v, want second item text", doneChunk2.Choices[0].Delta.Content.Content)
	}
}

func TestResponseOutboundTransformStreamTracksMessageDoneIdentityFromTopLevelItemID(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	itemID := "item-message-top-level-itemid"
	text := "Recovered from message.done with top-level item_id"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-top-level-itemid",
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
		Type:   "response.output_item.done",
		ItemID: &itemID,
		Item: &ResponsesItem{
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

	duplicatePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() duplicate error = %v", err)
	}
	duplicateChunk, err := outbound.TransformStream(context.Background(), duplicatePayload)
	if err != nil {
		t.Fatalf("TransformStream() duplicate error = %v", err)
	}
	if duplicateChunk != nil {
		t.Fatalf("duplicate chunk = %#v, want nil after top-level item_id state was recorded", duplicateChunk)
	}
}

func TestResponseOutboundTransformStreamSkipsEmptyMessageDoneWithTopLevelItemIDWithoutPoisoningState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	itemID := "item-message-top-level-itemid-empty"
	text := "Recovered after empty top-level item_id message.done"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-top-level-itemid-empty",
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

	ignoredPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:   "response.output_item.done",
		ItemID: &itemID,
		Item: &ResponsesItem{
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() ignored error = %v", err)
	}
	ignoredChunk, err := outbound.TransformStream(context.Background(), ignoredPayload)
	if err != nil {
		t.Fatalf("TransformStream() ignored error = %v", err)
	}
	if ignoredChunk != nil {
		t.Fatalf("ignored chunk = %#v, want nil for empty top-level item_id message.done", ignoredChunk)
	}

	validPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() valid error = %v", err)
	}
	validChunk, err := outbound.TransformStream(context.Background(), validPayload)
	if err != nil {
		t.Fatalf("TransformStream() valid error = %v", err)
	}
	if validChunk == nil || len(validChunk.Choices) != 1 || validChunk.Choices[0].Delta == nil {
		t.Fatalf("valid chunk = %#v, want assistant text fallback chunk", validChunk)
	}
	if validChunk.Choices[0].Delta.Content.Content == nil || *validChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("valid chunk content = %#v, want %q", validChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamSkipsMessageDoneWithoutTextWithoutPoisoningState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 12
	itemID := "item-message-empty-done"
	text := "Recovered after empty message.done"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-empty-done",
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

	ignoredPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.done",
		OutputIndex: lo.ToPtr(outputIndex),
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() ignored error = %v", err)
	}
	ignoredChunk, err := outbound.TransformStream(context.Background(), ignoredPayload)
	if err != nil {
		t.Fatalf("TransformStream() ignored error = %v", err)
	}
	if ignoredChunk != nil {
		t.Fatalf("ignored chunk = %#v, want nil for empty message output_item.done", ignoredChunk)
	}

	validPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() valid error = %v", err)
	}
	validChunk, err := outbound.TransformStream(context.Background(), validPayload)
	if err != nil {
		t.Fatalf("TransformStream() valid error = %v", err)
	}
	if validChunk == nil || len(validChunk.Choices) != 1 || validChunk.Choices[0].Delta == nil {
		t.Fatalf("valid chunk = %#v, want assistant text fallback chunk", validChunk)
	}
	if validChunk.Choices[0].Delta.Content.Content == nil || *validChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("valid chunk content = %#v, want %q", validChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamSkipsMessageDoneWithEmptyOutputTextWithoutPoisoningState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 13
	itemID := "item-message-empty-output-text"
	text := "Recovered after empty-string message output_text"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-empty-output-text",
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

	ignoredPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.done",
		OutputIndex: lo.ToPtr(outputIndex),
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
			Content: &ResponsesInput{Items: []ResponsesItem{{
				Type: "output_text",
				Text: lo.ToPtr(""),
			}}},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() ignored error = %v", err)
	}
	ignoredChunk, err := outbound.TransformStream(context.Background(), ignoredPayload)
	if err != nil {
		t.Fatalf("TransformStream() ignored error = %v", err)
	}
	if ignoredChunk != nil {
		t.Fatalf("ignored chunk = %#v, want nil for empty-string message output_item.done", ignoredChunk)
	}

	validPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() valid error = %v", err)
	}
	validChunk, err := outbound.TransformStream(context.Background(), validPayload)
	if err != nil {
		t.Fatalf("TransformStream() valid error = %v", err)
	}
	if validChunk == nil || len(validChunk.Choices) != 1 || validChunk.Choices[0].Delta == nil {
		t.Fatalf("valid chunk = %#v, want assistant text fallback chunk", validChunk)
	}
	if validChunk.Choices[0].Delta.Content.Content == nil || *validChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("valid chunk content = %#v, want %q", validChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamSkipsMessageDoneWithNonTextContentWithoutPoisoningState(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 14
	itemID := "item-message-non-text-content"
	text := "Recovered after non-text message content"
	imageURL := "https://example.com/image.png"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-non-text-content",
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

	ignoredPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.done",
		OutputIndex: lo.ToPtr(outputIndex),
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
			Content: &ResponsesInput{Items: []ResponsesItem{{
				Type:     "input_image",
				ImageURL: lo.ToPtr(imageURL),
			}}},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() ignored error = %v", err)
	}
	ignoredChunk, err := outbound.TransformStream(context.Background(), ignoredPayload)
	if err != nil {
		t.Fatalf("TransformStream() ignored error = %v", err)
	}
	if ignoredChunk != nil {
		t.Fatalf("ignored chunk = %#v, want nil for non-text message output_item.done", ignoredChunk)
	}

	validPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(1),
		Text:         text,
	})
	if err != nil {
		t.Fatalf("json.Marshal() valid error = %v", err)
	}
	validChunk, err := outbound.TransformStream(context.Background(), validPayload)
	if err != nil {
		t.Fatalf("TransformStream() valid error = %v", err)
	}
	if validChunk == nil || len(validChunk.Choices) != 1 || validChunk.Choices[0].Delta == nil {
		t.Fatalf("valid chunk = %#v, want assistant text fallback chunk", validChunk)
	}
	if validChunk.Choices[0].Delta.Content.Content == nil || *validChunk.Choices[0].Delta.Content.Content != text {
		t.Fatalf("valid chunk content = %#v, want %q", validChunk.Choices[0].Delta.Content.Content, text)
	}
}

func TestResponseOutboundTransformStreamTracksMessageDonePerContentIndex(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 16
	itemID := "item-message-partial-delta"
	firstPart := "first part"
	secondPart := "second part"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-partial-delta",
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
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Delta:        firstPart,
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
		OutputIndex: lo.ToPtr(outputIndex),
		ItemID:      &itemID,
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
			Content: &ResponsesInput{Items: []ResponsesItem{
				{Type: "output_text", Text: lo.ToPtr(firstPart)},
				{Type: "output_text", Text: lo.ToPtr(secondPart)},
			}},
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
	if doneChunk.Choices[0].Delta.Content.Content == nil || *doneChunk.Choices[0].Delta.Content.Content != secondPart {
		t.Fatalf("done chunk content = %#v, want %q", doneChunk.Choices[0].Delta.Content.Content, secondPart)
	}
	if !outbound.messageState(lo.ToPtr(outputIndex), &itemID).contentSeen(lo.ToPtr(0)) {
		t.Fatalf("message state did not record first content index after delta")
	}
	if !outbound.messageState(lo.ToPtr(outputIndex), &itemID).contentSeen(lo.ToPtr(1)) {
		t.Fatalf("message state did not record second content index after fallback")
	}
}

func TestResponseOutboundTransformStreamSkipsOutputItemDoneAfterOutputTextDoneByContentIndex(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 19
	itemID := "item-message-output-text-then-item-done"
	firstText := "first part"
	secondText := "second part"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-output-text-then-item-done",
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

	firstPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         firstText,
	})
	if err != nil {
		t.Fatalf("json.Marshal() first error = %v", err)
	}
	firstChunk, err := outbound.TransformStream(context.Background(), firstPayload)
	if err != nil {
		t.Fatalf("TransformStream() first error = %v", err)
	}
	if firstChunk == nil || len(firstChunk.Choices) != 1 || firstChunk.Choices[0].Delta == nil {
		t.Fatalf("first chunk = %#v, want assistant text fallback chunk", firstChunk)
	}
	if firstChunk.Choices[0].Delta.Content.Content == nil || *firstChunk.Choices[0].Delta.Content.Content != firstText {
		t.Fatalf("first chunk content = %#v, want %q", firstChunk.Choices[0].Delta.Content.Content, firstText)
	}

	secondPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:        "response.output_item.done",
		OutputIndex: lo.ToPtr(outputIndex),
		ItemID:      &itemID,
		Item: &ResponsesItem{
			ID:     itemID,
			Type:   "message",
			Role:   "assistant",
			Status: lo.ToPtr("completed"),
			Content: &ResponsesInput{Items: []ResponsesItem{
				{Type: "output_text", Text: lo.ToPtr(firstText)},
				{Type: "output_text", Text: lo.ToPtr(secondText)},
			}},
		},
	})
	if err != nil {
		t.Fatalf("json.Marshal() second error = %v", err)
	}
	secondChunk, err := outbound.TransformStream(context.Background(), secondPayload)
	if err != nil {
		t.Fatalf("TransformStream() second error = %v", err)
	}
	if secondChunk == nil || len(secondChunk.Choices) != 1 || secondChunk.Choices[0].Delta == nil {
		t.Fatalf("second chunk = %#v, want assistant text fallback chunk", secondChunk)
	}
	if secondChunk.Choices[0].Delta.Content.Content == nil || *secondChunk.Choices[0].Delta.Content.Content != secondText {
		t.Fatalf("second chunk content = %#v, want %q", secondChunk.Choices[0].Delta.Content.Content, secondText)
	}
	if !outbound.messageState(lo.ToPtr(outputIndex), &itemID).contentSeen(lo.ToPtr(0)) {
		t.Fatalf("message state did not record first content index after output_text.done")
	}
	if !outbound.messageState(lo.ToPtr(outputIndex), &itemID).contentSeen(lo.ToPtr(1)) {
		t.Fatalf("message state did not record second content index after output_item.done")
	}
}

func TestResponseOutboundTransformStreamTracksMessageTextDonePerContentIndex(t *testing.T) {
	t.Parallel()

	outbound := &ResponseOutbound{}
	status := "in_progress"
	outputIndex := 15
	itemID := "item-message-content-index"
	text0 := "first part"
	text1 := "second part"

	createdPayload, err := json.Marshal(ResponsesStreamEvent{
		Type: "response.created",
		Response: &ResponsesResponse{
			ID:     "resp-message-content-index",
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

	firstPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(0),
		Text:         text0,
	})
	if err != nil {
		t.Fatalf("json.Marshal() first error = %v", err)
	}
	firstChunk, err := outbound.TransformStream(context.Background(), firstPayload)
	if err != nil {
		t.Fatalf("TransformStream() first error = %v", err)
	}
	if firstChunk == nil || len(firstChunk.Choices) != 1 || firstChunk.Choices[0].Delta == nil {
		t.Fatalf("first chunk = %#v, want assistant text fallback chunk", firstChunk)
	}
	if firstChunk.Choices[0].Delta.Content.Content == nil || *firstChunk.Choices[0].Delta.Content.Content != text0 {
		t.Fatalf("first chunk content = %#v, want %q", firstChunk.Choices[0].Delta.Content.Content, text0)
	}

	secondPayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(1),
		Text:         text1,
	})
	if err != nil {
		t.Fatalf("json.Marshal() second error = %v", err)
	}
	secondChunk, err := outbound.TransformStream(context.Background(), secondPayload)
	if err != nil {
		t.Fatalf("TransformStream() second error = %v", err)
	}
	if secondChunk == nil || len(secondChunk.Choices) != 1 || secondChunk.Choices[0].Delta == nil {
		t.Fatalf("second chunk = %#v, want assistant text fallback chunk", secondChunk)
	}
	if secondChunk.Choices[0].Delta.Content.Content == nil || *secondChunk.Choices[0].Delta.Content.Content != text1 {
		t.Fatalf("second chunk content = %#v, want %q", secondChunk.Choices[0].Delta.Content.Content, text1)
	}

	duplicatePayload, err := json.Marshal(ResponsesStreamEvent{
		Type:         "response.output_text.done",
		OutputIndex:  lo.ToPtr(outputIndex),
		ItemID:       &itemID,
		ContentIndex: lo.ToPtr(1),
		Text:         text1,
	})
	if err != nil {
		t.Fatalf("json.Marshal() duplicate error = %v", err)
	}
	duplicateChunk, err := outbound.TransformStream(context.Background(), duplicatePayload)
	if err != nil {
		t.Fatalf("TransformStream() duplicate error = %v", err)
	}
	if duplicateChunk != nil {
		t.Fatalf("duplicate chunk = %#v, want nil for repeated content_index", duplicateChunk)
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
