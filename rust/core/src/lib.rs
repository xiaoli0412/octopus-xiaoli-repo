use std::ffi::{c_char, c_int, CStr, CString};
use std::fmt;

use serde::de::{self, SeqAccess, Visitor};
use serde::{Deserialize, Deserializer, Serialize, Serializer};

pub mod balance;
pub mod stats_agg;
pub mod stream_buffer;
pub mod transform;

// =============================================================================
// C ABI helpers
// =============================================================================

fn to_str<'a>(ptr: *const c_char) -> Result<&'a str, c_int> {
    if ptr.is_null() {
        return Err(-1);
    }
    unsafe { CStr::from_ptr(ptr) }.to_str().map_err(|_| -2)
}

fn write_output(s: String, out: *mut *mut c_char) -> c_int {
    if out.is_null() {
        return -3;
    }
    match CString::new(s) {
        Ok(c) => {
            unsafe { *out = c.into_raw() };
            0
        }
        Err(_) => -4,
    }
}

// =============================================================================
// Tokenizer
// =============================================================================

fn encoding_for_model(model: &str) -> &'static str {
    let m = model.to_ascii_lowercase();
    if m.contains("cl100k")
        || m.contains("text-embedding")
        || (m.contains("gpt-4") && !m.contains("gpt-4o"))
        || m.contains("gpt-3.5")
    {
        "cl100k_base"
    } else {
        "o200k_base"
    }
}

/// Count tokens for `text` using the encoding implied by `model`.
/// Returns a non-negative count on success, negative error code otherwise.
#[no_mangle]
pub extern "C" fn octopus_tokenizer_count(text: *const c_char, model: *const c_char) -> c_int {
    let text = match to_str(text) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let model = match to_str(model) {
        Ok(s) => s,
        Err(e) => return e,
    };

    let enc = encoding_for_model(model);
    let tokens = match enc {
        "cl100k_base" => tiktoken_rs::cl100k_base_singleton().encode_with_special_tokens(text),
        _ => tiktoken_rs::o200k_base_singleton().encode_with_special_tokens(text),
    };

    if tokens.len() > i32::MAX as usize {
        return -4;
    }
    tokens.len() as c_int
}

/// Free a C string previously returned by this library.
///
/// # Safety
/// `s` must be a pointer previously returned by this library, or null.
#[no_mangle]
pub unsafe extern "C" fn octopus_free_string(s: *mut c_char) {
    if !s.is_null() {
        drop(CString::from_raw(s));
    }
}

// =============================================================================
// JSON field extraction helpers
// =============================================================================

/// Extract the top-level `model` field from a request/response JSON object.
/// Returns a JSON string `{"model":"..."}` (or `{}` if absent) through `out`.
#[no_mangle]
pub extern "C" fn octopus_json_extract_model(json: *const c_char, out: *mut *mut c_char) -> c_int {
    let json = match to_str(json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let value: serde_json::Value = match serde_json::from_str(json) {
        Ok(v) => v,
        Err(_) => return -5,
    };
    let model = value.get("model").and_then(|v| v.as_str()).unwrap_or("");
    let out_json = serde_json::json!({"model": model});
    write_output(out_json.to_string(), out)
}

/// Extract the top-level `usage` object from a response JSON object.
/// Returns a JSON string with `prompt_tokens`, `completion_tokens`, `total_tokens`.
#[no_mangle]
pub extern "C" fn octopus_json_extract_usage(json: *const c_char, out: *mut *mut c_char) -> c_int {
    let json = match to_str(json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let value: serde_json::Value = match serde_json::from_str(json) {
        Ok(v) => v,
        Err(_) => return -5,
    };

    let usage_obj = if let Some(u) = value.get("usage") {
        serde_json::json!({
            "prompt_tokens": u.get("prompt_tokens").and_then(|v| v.as_i64()).unwrap_or(0),
            "completion_tokens": u.get("completion_tokens").and_then(|v| v.as_i64()).unwrap_or(0),
            "total_tokens": u.get("total_tokens").and_then(|v| v.as_i64()).unwrap_or(0),
        })
    } else {
        serde_json::json!({})
    };
    write_output(usage_obj.to_string(), out)
}

// =============================================================================
// SSE aggregate
// =============================================================================

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct ImageURL {
    url: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    detail: Option<String>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct MessageContentPart {
    #[serde(rename = "type")]
    ty: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    text: Option<String>,
    #[serde(rename = "image_url", skip_serializing_if = "Option::is_none")]
    image_url: Option<ImageURL>,
    #[serde(rename = "input_audio", skip_serializing_if = "Option::is_none")]
    audio: Option<serde_json::Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    file: Option<serde_json::Value>,
}

#[derive(Debug, Clone, Default, PartialEq)]
struct MessageContent {
    content: Option<String>,
    multiple_content: Option<Vec<MessageContentPart>>,
}

impl MessageContent {
    fn is_empty(&self) -> bool {
        self.content.is_none() && self.multiple_content.as_ref().is_none_or(|v| v.is_empty())
    }
}

impl Serialize for MessageContent {
    fn serialize<S>(&self, serializer: S) -> Result<S::Ok, S::Error>
    where
        S: Serializer,
    {
        if let Some(parts) = &self.multiple_content {
            if parts.len() == 1 && parts[0].ty == "text" && parts[0].text.is_some() {
                return parts[0].text.serialize(serializer);
            }
            return parts.serialize(serializer);
        }
        self.content.serialize(serializer)
    }
}

impl<'de> Deserialize<'de> for MessageContent {
    fn deserialize<D>(deserializer: D) -> Result<Self, D::Error>
    where
        D: Deserializer<'de>,
    {
        struct ContentVisitor;

        impl<'de> Visitor<'de> for ContentVisitor {
            type Value = MessageContent;

            fn expecting(&self, formatter: &mut fmt::Formatter) -> fmt::Result {
                formatter.write_str("a string or an array of content parts")
            }

            fn visit_str<E>(self, value: &str) -> Result<Self::Value, E>
            where
                E: de::Error,
            {
                Ok(MessageContent {
                    content: Some(value.to_string()),
                    multiple_content: None,
                })
            }

            fn visit_string<E>(self, value: String) -> Result<Self::Value, E>
            where
                E: de::Error,
            {
                Ok(MessageContent {
                    content: Some(value),
                    multiple_content: None,
                })
            }

            fn visit_seq<A>(self, mut seq: A) -> Result<Self::Value, A::Error>
            where
                A: SeqAccess<'de>,
            {
                let mut parts = Vec::new();
                while let Some(part) = seq.next_element::<MessageContentPart>()? {
                    parts.push(part);
                }
                Ok(MessageContent {
                    content: None,
                    multiple_content: Some(parts),
                })
            }
        }

        deserializer.deserialize_any(ContentVisitor)
    }
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct FunctionCall {
    #[serde(skip_serializing_if = "String::is_empty")]
    name: String,
    arguments: String,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct ToolCall {
    #[serde(skip_serializing_if = "String::is_empty")]
    id: String,
    #[serde(rename = "type", skip_serializing_if = "String::is_empty")]
    ty: String,
    function: FunctionCall,
    index: i32,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct Message {
    #[serde(skip_serializing_if = "String::is_empty")]
    role: String,
    #[serde(skip_serializing_if = "MessageContent::is_empty")]
    content: MessageContent,
    #[serde(skip_serializing_if = "Option::is_none")]
    name: Option<String>,
    #[serde(skip_serializing_if = "String::is_empty")]
    refusal: String,
    #[serde(rename = "tool_call_id", skip_serializing_if = "Option::is_none")]
    tool_call_id: Option<String>,
    #[serde(rename = "tool_calls", skip_serializing_if = "Vec::is_empty")]
    tool_calls: Vec<ToolCall>,
    #[serde(rename = "reasoning_content", skip_serializing_if = "Option::is_none")]
    reasoning_content: Option<String>,
    #[serde(
        rename = "reasoning_signature",
        skip_serializing_if = "Option::is_none"
    )]
    reasoning_signature: Option<String>,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    images: Vec<MessageContentPart>,
    #[serde(skip_serializing_if = "Option::is_none")]
    audio: Option<serde_json::Value>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct TokenLogprob {
    token: String,
    logprob: f64,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    bytes: Vec<i32>,
    #[serde(rename = "top_logprobs", skip_serializing_if = "Vec::is_empty")]
    top_logprobs: Vec<TopLogprob>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct TopLogprob {
    token: String,
    logprob: f64,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    bytes: Vec<i32>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct LogprobsContent {
    content: Vec<TokenLogprob>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct Choice {
    index: i32,
    #[serde(skip_serializing_if = "Option::is_none")]
    message: Option<Message>,
    #[serde(skip_serializing_if = "Option::is_none")]
    delta: Option<Message>,
    #[serde(rename = "finish_reason", skip_serializing_if = "Option::is_none")]
    finish_reason: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    logprobs: Option<LogprobsContent>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct PromptTokensDetails {
    audio_tokens: i64,
    cached_tokens: i64,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct CompletionTokensDetails {
    audio_tokens: i64,
    reasoning_tokens: i64,
    accepted_prediction_tokens: i64,
    rejected_prediction_tokens: i64,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct Usage {
    prompt_tokens: i64,
    completion_tokens: i64,
    total_tokens: i64,
    #[serde(
        rename = "prompt_tokens_details",
        skip_serializing_if = "Option::is_none"
    )]
    prompt_tokens_details: Option<PromptTokensDetails>,
    #[serde(
        rename = "completion_tokens_details",
        skip_serializing_if = "Option::is_none"
    )]
    completion_tokens_details: Option<CompletionTokensDetails>,
}

#[derive(Debug, Clone, Default, Serialize, Deserialize, PartialEq)]
#[serde(default)]
struct InternalLLMResponse {
    id: String,
    #[serde(skip_serializing_if = "Vec::is_empty")]
    choices: Vec<Choice>,
    #[serde(rename = "embedding_data", skip_serializing_if = "Vec::is_empty")]
    embedding_data: Vec<serde_json::Value>,
    object: String,
    created: i64,
    model: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    usage: Option<Usage>,
    #[serde(
        rename = "system_fingerprint",
        skip_serializing_if = "String::is_empty"
    )]
    system_fingerprint: String,
    #[serde(rename = "service_tier", skip_serializing_if = "String::is_empty")]
    service_tier: String,
    #[serde(skip_serializing_if = "Option::is_none")]
    error: Option<serde_json::Value>,
}

fn append_text_content(dst: &mut MessageContent, text: &str) {
    if text.is_empty() {
        return;
    }
    if let Some(parts) = &mut dst.multiple_content {
        if let Some(last) = parts.last_mut() {
            if last.ty == "text" {
                if let Some(t) = &mut last.text {
                    t.push_str(text);
                    return;
                }
            }
        }
        parts.push(MessageContentPart {
            ty: "text".to_string(),
            text: Some(text.to_string()),
            ..Default::default()
        });
    } else if let Some(existing) = &mut dst.content {
        existing.push_str(text);
    } else {
        dst.content = Some(text.to_string());
    }
}

fn promote_text_content(dst: &mut MessageContent) {
    if let Some(text) = dst.content.take() {
        if !text.is_empty() {
            dst.multiple_content
                .get_or_insert_with(Vec::new)
                .push(MessageContentPart {
                    ty: "text".to_string(),
                    text: Some(text),
                    ..Default::default()
                });
        }
    }
}

fn append_content_parts(dst: &mut MessageContent, parts: &[MessageContentPart]) {
    if parts.is_empty() {
        return;
    }
    if dst.multiple_content.is_none() && parts.iter().all(|p| p.ty == "text") {
        for p in parts {
            if let Some(t) = &p.text {
                append_text_content(dst, t);
            }
        }
        return;
    }
    promote_text_content(dst);
    dst.multiple_content
        .get_or_insert_with(Vec::new)
        .extend_from_slice(parts);
}

fn merge_message(dst: &mut Message, src: &Message) {
    if !src.role.is_empty() {
        dst.role = src.role.clone();
    }
    if let Some(name) = &src.name {
        if !name.is_empty() {
            dst.name = Some(name.clone());
        }
    }
    if let Some(content) = &src.content.content {
        append_text_content(&mut dst.content, content);
    }
    if let Some(parts) = &src.content.multiple_content {
        append_content_parts(&mut dst.content, parts);
    }
    if !src.refusal.is_empty() {
        dst.refusal.push_str(&src.refusal);
    }
    if let Some(reasoning) = &src.reasoning_content {
        dst.reasoning_content
            .get_or_insert_with(String::new)
            .push_str(reasoning);
    }
    if let Some(signature) = &src.reasoning_signature {
        dst.reasoning_signature
            .get_or_insert_with(String::new)
            .push_str(signature);
    }
    for delta in &src.tool_calls {
        if let Some(tc) = dst.tool_calls.iter_mut().find(|t| t.index == delta.index) {
            if !delta.id.is_empty() {
                tc.id = delta.id.clone();
            }
            if !delta.ty.is_empty() {
                tc.ty = delta.ty.clone();
            }
            tc.function.name.push_str(&delta.function.name);
            tc.function.arguments.push_str(&delta.function.arguments);
        } else {
            dst.tool_calls.push(delta.clone());
        }
    }
    if !src.images.is_empty() {
        append_content_parts(&mut dst.content, &src.images);
    }
    if let Some(audio) = &src.audio {
        dst.audio = Some(audio.clone());
    }
}

fn merge_logprobs(dst: &mut LogprobsContent, src: &LogprobsContent) {
    for token in &src.content {
        let mut merged = TokenLogprob {
            token: token.token.clone(),
            logprob: token.logprob,
            ..Default::default()
        };
        if !token.bytes.is_empty() {
            merged.bytes = token.bytes.clone();
        }
        if !token.top_logprobs.is_empty() {
            merged.top_logprobs = token
                .top_logprobs
                .iter()
                .map(|t| TopLogprob {
                    token: t.token.clone(),
                    logprob: t.logprob,
                    bytes: if t.bytes.is_empty() {
                        Vec::new()
                    } else {
                        t.bytes.clone()
                    },
                })
                .collect();
        }
        dst.content.push(merged);
    }
}

fn ensure_choice_index(resp: &mut InternalLLMResponse, index: i32) -> usize {
    for (i, choice) in resp.choices.iter().enumerate() {
        if choice.index == index {
            return i;
        }
    }
    let idx = resp.choices.len();
    resp.choices.push(Choice {
        index,
        message: Some(Message::default()),
        ..Default::default()
    });
    idx
}

fn merge_streaming_response_aggregate(
    result: &mut InternalLLMResponse,
    chunk: &InternalLLMResponse,
) {
    if !chunk.id.is_empty() {
        result.id = chunk.id.clone();
    }
    if chunk.created != 0 {
        result.created = chunk.created;
    }
    if !chunk.model.is_empty() {
        result.model = chunk.model.clone();
    }
    if !chunk.system_fingerprint.is_empty() {
        result.system_fingerprint = chunk.system_fingerprint.clone();
    }
    if !chunk.service_tier.is_empty() {
        result.service_tier = chunk.service_tier.clone();
    }
    if let Some(usage) = &chunk.usage {
        result.usage = Some(usage.clone());
    }
    if let Some(error) = &chunk.error {
        result.error = Some(error.clone());
    }

    for choice in &chunk.choices {
        let idx = ensure_choice_index(result, choice.index);
        let existing_choice = &mut result.choices[idx];
        if existing_choice.message.is_none() {
            existing_choice.message = Some(Message::default());
        }

        if let Some(msg) = &choice.message {
            merge_message(existing_choice.message.as_mut().unwrap(), msg);
        }
        if choice.delta.is_some() && choice.delta != choice.message {
            if let Some(delta) = &choice.delta {
                merge_message(existing_choice.message.as_mut().unwrap(), delta);
            }
        }

        if let Some(fr) = &choice.finish_reason {
            existing_choice.finish_reason = Some(fr.clone());
        }
        if let Some(lp) = &choice.logprobs {
            let dst = existing_choice
                .logprobs
                .get_or_insert_with(LogprobsContent::default);
            merge_logprobs(dst, lp);
        }
    }
}

/// Merge a streaming chunk JSON into an aggregate response JSON.
/// On success returns 0 and writes the updated aggregate JSON through `out`.
#[no_mangle]
pub extern "C" fn octopus_sse_aggregate(
    aggregate: *const c_char,
    chunk: *const c_char,
    out: *mut *mut c_char,
) -> c_int {
    let agg_str = match to_str(aggregate) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let chunk_str = match to_str(chunk) {
        Ok(s) => s,
        Err(e) => return e,
    };

    let mut aggregate: InternalLLMResponse = match serde_json::from_str(agg_str) {
        Ok(v) => v,
        Err(_) => return -5,
    };
    let chunk: InternalLLMResponse = match serde_json::from_str(chunk_str) {
        Ok(v) => v,
        Err(_) => return -6,
    };

    merge_streaming_response_aggregate(&mut aggregate, &chunk);

    match serde_json::to_string(&aggregate) {
        Ok(s) => write_output(s, out),
        Err(_) => -7,
    }
}

// =============================================================================
// FFI re-exports for protocol transform
// =============================================================================

/// Transform an OpenAI chat completion request JSON for the target provider.
/// `target` is one of: openai, anthropic, gemini, volcengine.
/// On success returns 0 and writes the transformed JSON through `out`.
#[no_mangle]
pub extern "C" fn octopus_transform_openai_chat_request(
    json: *const c_char,
    target: *const c_char,
    out: *mut *mut c_char,
) -> c_int {
    let json = match to_str(json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let target = match to_str(target) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match transform::transform_openai_chat_request(json, target) {
        Ok(s) => write_output(s, out),
        Err(e) => e,
    }
}

/// Transform an OpenAI chat completion response JSON for the target provider.
#[no_mangle]
pub extern "C" fn octopus_transform_openai_response(
    json: *const c_char,
    target: *const c_char,
    out: *mut *mut c_char,
) -> c_int {
    let json = match to_str(json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let target = match to_str(target) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match transform::transform_openai_response(json, target) {
        Ok(s) => write_output(s, out),
        Err(e) => e,
    }
}

/// Transform an OpenAI embedding request JSON for the target provider.
#[no_mangle]
pub extern "C" fn octopus_transform_embedding_request(
    json: *const c_char,
    target: *const c_char,
    out: *mut *mut c_char,
) -> c_int {
    let json = match to_str(json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let target = match to_str(target) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match transform::transform_embedding_request(json, target) {
        Ok(s) => write_output(s, out),
        Err(e) => e,
    }
}

// =============================================================================
// FFI re-exports for load balancer
// =============================================================================

/// Select a candidate according to the given strategy.
/// `candidates_json` is an array of objects with id, weight, latency, priority,
/// healthy and circuit_state fields. `strategy` is weighted/round_robin/random/failover.
/// `current_idx` is the current round-robin/weighted cursor.
/// On success returns 0 and writes `{"id":..., "next_index":...}` through `out`.
#[no_mangle]
pub extern "C" fn octopus_balance_select(
    candidates_json: *const c_char,
    strategy: *const c_char,
    current_idx: c_int,
    out: *mut *mut c_char,
) -> c_int {
    let candidates = match to_str(candidates_json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let strategy = match to_str(strategy) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match balance::balance_select(candidates, strategy, current_idx) {
        Ok(s) => write_output(s, out),
        Err(e) => e,
    }
}

/// High-performance binary variant that avoids JSON round-trips.
#[no_mangle]
pub unsafe extern "C" fn octopus_balance_select_v2(
    candidates: *const balance::BalanceCandidateFFI,
    candidates_len: c_int,
    strategy: *const c_char,
    current_idx: c_int,
    out_id: *mut i64,
    out_next_index: *mut c_int,
) -> c_int {
    if candidates.is_null() || out_id.is_null() || out_next_index.is_null() {
        return -1;
    }
    let len = candidates_len as usize;
    let slice = std::slice::from_raw_parts(candidates, len);
    let strategy = match to_str(strategy) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match balance::balance_select_v2(slice, strategy, current_idx) {
        Ok((id, next)) => {
            *out_id = id;
            *out_next_index = next;
            0
        }
        Err(e) => e,
    }
}

// =============================================================================
// FFI re-exports for stream buffer
// =============================================================================

/// Create a new thread-safe stream buffer. Returns an opaque handle.
#[no_mangle]
pub extern "C" fn octopus_stream_buffer_create() -> *mut stream_buffer::StreamBufferHandle {
    stream_buffer::buffer_create()
}

/// Feed a new chunk into the stream buffer. Returns a JSON array of complete
/// events extracted so far through `out`.
#[no_mangle]
pub extern "C" fn octopus_stream_buffer_feed(
    handle: *mut stream_buffer::StreamBufferHandle,
    chunk: *const c_char,
    out: *mut *mut c_char,
) -> c_int {
    if handle.is_null() {
        return -1;
    }
    let chunk = match to_str(chunk) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match stream_buffer::buffer_feed(handle, chunk) {
        Ok(s) => write_output(s, out),
        Err(e) => e,
    }
}

/// Take all complete events currently buffered without feeding new data.
#[no_mangle]
pub extern "C" fn octopus_stream_buffer_take(
    handle: *mut stream_buffer::StreamBufferHandle,
    out: *mut *mut c_char,
) -> c_int {
    if handle.is_null() {
        return -1;
    }
    match stream_buffer::buffer_take(handle) {
        Ok(s) => write_output(s, out),
        Err(e) => e,
    }
}

/// Free a stream buffer previously created by octopus_stream_buffer_create.
#[no_mangle]
pub unsafe extern "C" fn octopus_stream_buffer_free(
    handle: *mut stream_buffer::StreamBufferHandle,
) {
    stream_buffer::buffer_free(handle);
}

/// Stateless SSE event extractor.
/// Returns a JSON array of complete events through `out_events` and the number
/// of consumed bytes through `out_consumed_bytes`.
#[no_mangle]
pub extern "C" fn octopus_stream_extract_events(
    data: *const c_char,
    out_events: *mut *mut c_char,
    out_consumed_bytes: *mut c_int,
) -> c_int {
    if out_events.is_null() || out_consumed_bytes.is_null() {
        return -1;
    }
    let data = match to_str(data) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match stream_buffer::extract_events_stateless(data) {
        Ok((json, consumed)) => {
            unsafe { *out_consumed_bytes = consumed as c_int };
            write_output(json, out_events)
        }
        Err(e) => e,
    }
}

/// Find the byte ranges of complete SSE events in `data`.
/// `data` points to `len` bytes of valid UTF-8 text.
/// Returns an opaque boundary descriptor through `out_boundaries`.
#[no_mangle]
pub extern "C" fn octopus_stream_find_event_boundaries(
    data: *const c_char,
    len: c_int,
    out_boundaries: *mut *mut stream_buffer::OctopusEventBoundaries,
) -> c_int {
    if out_boundaries.is_null() || data.is_null() {
        return -1;
    }
    if len < 0 {
        return -2;
    }
    let bytes = unsafe { std::slice::from_raw_parts(data as *const u8, len as usize) };
    let s = match std::str::from_utf8(bytes) {
        Ok(s) => s,
        Err(_) => return -2,
    };
    unsafe { *out_boundaries = stream_buffer::find_event_boundaries(s) };
    0
}

/// Free a boundary descriptor previously returned by
/// `octopus_stream_find_event_boundaries`.
#[no_mangle]
pub unsafe extern "C" fn octopus_stream_boundaries_free(
    boundaries: *mut stream_buffer::OctopusEventBoundaries,
) {
    stream_buffer::free_event_boundaries(boundaries);
}

/// High-performance variant of the SSE boundary finder.
/// `starts` and `ends` must point to arrays of at least `max_events` entries.
/// On success writes the number of events into `out_count` and the number of
/// consumed bytes into `out_consumed`. Returns -15 if `max_events` is too small.
#[no_mangle]
pub extern "C" fn octopus_stream_find_event_boundaries_ex(
    data: *const c_char,
    len: c_int,
    starts: *mut c_int,
    ends: *mut c_int,
    max_events: c_int,
    out_count: *mut c_int,
    out_consumed: *mut c_int,
) -> c_int {
    if data.is_null()
        || starts.is_null()
        || ends.is_null()
        || out_count.is_null()
        || out_consumed.is_null()
    {
        return -1;
    }
    if len < 0 || max_events < 0 {
        return -2;
    }
    unsafe {
        stream_buffer::fill_event_boundaries_bytes(
            data as *const u8,
            len as usize,
            starts,
            ends,
            max_events as usize,
            out_count,
            out_consumed,
        )
    }
}

// =============================================================================
// FFI re-exports for stats aggregation
// =============================================================================

/// Merge a delta StatsMetrics JSON into an existing aggregate JSON.
#[no_mangle]
pub extern "C" fn octopus_stats_merge_hourly(
    existing_json: *const c_char,
    delta_json: *const c_char,
    out: *mut *mut c_char,
) -> c_int {
    let existing = match to_str(existing_json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    let delta = match to_str(delta_json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match stats_agg::stats_merge_hourly(existing, delta) {
        Ok(s) => write_output(s, out),
        Err(e) => e,
    }
}

/// Compute a quantile for an array of f64 values.
/// `quantile` must be in [0.0, 1.0].
#[no_mangle]
pub extern "C" fn octopus_stats_quantile(
    values_json: *const c_char,
    quantile: f64,
    out: *mut *mut c_char,
) -> c_int {
    let values = match to_str(values_json) {
        Ok(s) => s,
        Err(e) => return e,
    };
    match stats_agg::stats_quantile(values, quantile) {
        Ok(s) => write_output(s, out),
        Err(e) => e,
    }
}

// =============================================================================
// Unit tests
// =============================================================================

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn tokenizer_counts_known_text() {
        let count = octopus_tokenizer_count(
            CString::new("hello world").unwrap().as_ptr(),
            CString::new("gpt-4o").unwrap().as_ptr(),
        );
        assert!(count >= 0);
    }

    #[test]
    fn extract_model_and_usage() {
        let json = CString::new(r#"{"model":"gpt-4o","usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}"#).unwrap();
        let mut out: *mut c_char = std::ptr::null_mut();
        assert_eq!(octopus_json_extract_model(json.as_ptr(), &mut out), 0);
        unsafe { octopus_free_string(out) };

        out = std::ptr::null_mut();
        assert_eq!(octopus_json_extract_usage(json.as_ptr(), &mut out), 0);
        unsafe { octopus_free_string(out) };
    }
}
