use serde_json::{json, Value};
use std::ffi::c_int;

/// Transform an OpenAI chat completion request JSON for the target provider.
/// Targets: openai, anthropic, gemini, volcengine.
pub fn transform_openai_chat_request(json: &str, target: &str) -> Result<String, c_int> {
    let mut req: Value = serde_json::from_str(json).map_err(|_| -5)?;
    let target = target.to_ascii_lowercase();
    match target.as_str() {
        "openai" | "volcengine" => normalize_openai_chat_request(&mut req),
        "anthropic" => to_anthropic_chat_request(&req),
        "gemini" => to_gemini_chat_request(&req),
        _ => Err(-8),
    }
}

/// Transform an OpenAI chat completion response JSON for the target provider.
pub fn transform_openai_response(json: &str, target: &str) -> Result<String, c_int> {
    let resp: Value = serde_json::from_str(json).map_err(|_| -5)?;
    let target = target.to_ascii_lowercase();
    match target.as_str() {
        "openai" | "volcengine" | "anthropic" | "gemini" => {
            Ok(serde_json::to_string(&resp).map_err(|_| -7)?)
        }
        _ => Err(-8),
    }
}

/// Transform an OpenAI embedding request JSON for the target provider.
pub fn transform_embedding_request(json: &str, target: &str) -> Result<String, c_int> {
    let mut req: Value = serde_json::from_str(json).map_err(|_| -5)?;
    let target = target.to_ascii_lowercase();
    match target.as_str() {
        "openai" | "volcengine" => normalize_openai_embedding_request(&mut req),
        "gemini" => to_gemini_embedding_request(&req),
        "anthropic" => Err(-9), // Anthropic does not expose embeddings via this API.
        _ => Err(-8),
    }
}

fn normalize_openai_chat_request(req: &mut Value) -> Result<String, c_int> {
    // Ensure developer role is converted to system for upstream compatibility.
    if let Some(messages) = req.get_mut("messages").and_then(|m| m.as_array_mut()) {
        for msg in messages.iter_mut() {
            if let Some(role) = msg.get("role").and_then(|r| r.as_str()) {
                if role == "developer" {
                    msg["role"] = json!("system");
                }
            }
        }
    }
    // Ensure streaming requests request usage inclusion.
    if req.get("stream").and_then(|v| v.as_bool()).unwrap_or(false) {
        let mut opts = req.get("stream_options").cloned().unwrap_or_else(|| json!({}));
        if opts.get("include_usage").is_none() {
            opts["include_usage"] = json!(true);
        }
        req["stream_options"] = opts;
    }
    serde_json::to_string(req).map_err(|_| -7)
}

fn to_anthropic_chat_request(req: &Value) -> Result<String, c_int> {
    let mut out = json!({});
    out["model"] = req.get("model").cloned().unwrap_or(json!(""));

    let mut system_parts = Vec::new();
    let mut messages = Vec::new();
    if let Some(arr) = req.get("messages").and_then(|m| m.as_array()) {
        for msg in arr {
            let role = msg.get("role").and_then(|r| r.as_str()).unwrap_or("user");
            if role == "system" {
                if let Some(content) = extract_text_content(msg.get("content")) {
                    system_parts.push(json!({"type": "text", "text": content}));
                }
                continue;
            }
            let mut anth_msg = json!({"role": role});
            if let Some(content) = msg.get("content") {
                anth_msg["content"] = clone_content_for_anthropic(content);
            }
            if let Some(tc_id) = msg.get("tool_call_id").and_then(|v| v.as_str()) {
                anth_msg["tool_call_id"] = json!(tc_id);
            }
            messages.push(anth_msg);
        }
    }
    if !system_parts.is_empty() {
        out["system"] = json!(system_parts);
    }
    out["messages"] = json!(messages);

    if let Some(v) = req.get("max_tokens").and_then(|v| v.as_i64()) {
        out["max_tokens"] = json!(v);
    } else {
        out["max_tokens"] = json!(4096);
    }
    copy_if_present(req, &mut out, "temperature");
    copy_if_present(req, &mut out, "top_p");
    copy_if_present(req, &mut out, "stream");

    if let Some(tools) = req.get("tools") {
        out["tools"] = convert_tools_to_anthropic(tools);
    }
    if let Some(tc) = req.get("tool_choice") {
        out["tool_choice"] = convert_tool_choice_to_anthropic(tc);
    }

    serde_json::to_string(&out).map_err(|_| -7)
}

fn to_gemini_chat_request(req: &Value) -> Result<String, c_int> {
    let mut out = json!({});
    out["model"] = req.get("model").cloned().unwrap_or(json!(""));

    let mut contents = Vec::new();
    if let Some(arr) = req.get("messages").and_then(|m| m.as_array()) {
        for msg in arr {
            let role = msg.get("role").and_then(|r| r.as_str()).unwrap_or("user");
            let gemini_role = if role == "assistant" { "model" } else { role };
            let mut entry = json!({"role": gemini_role});
            if let Some(content) = msg.get("content") {
                entry["parts"] = clone_content_for_gemini(content);
            }
            contents.push(entry);
        }
    }
    out["contents"] = json!(contents);

    let mut gen_config = json!({});
    copy_if_present_rename(req, &mut gen_config, "temperature", "temperature");
    copy_if_present_rename(req, &mut gen_config, "top_p", "topP");
    if let Some(v) = req.get("max_tokens").and_then(|v| v.as_i64()) {
        gen_config["maxOutputTokens"] = json!(v);
    }
    out["generationConfig"] = gen_config;

    if let Some(tools) = req.get("tools") {
        out["tools"] = convert_tools_to_gemini(tools);
    }

    serde_json::to_string(&out).map_err(|_| -7)
}

fn normalize_openai_embedding_request(req: &mut Value) -> Result<String, c_int> {
    serde_json::to_string(req).map_err(|_| -7)
}

fn to_gemini_embedding_request(req: &Value) -> Result<String, c_int> {
    let mut out = json!({});
    out["model"] = req.get("model").cloned().unwrap_or(json!(""));
    if let Some(input) = req.get("input") {
        if let Some(text) = input.as_str() {
            out["content"] = json!({"parts": [{"text": text}]});
        } else if let Some(arr) = input.as_array() {
            let mut parts = Vec::new();
            for item in arr {
                if let Some(s) = item.as_str() {
                    parts.push(json!({"text": s}));
                }
            }
            out["content"] = json!({"parts": parts});
        }
    }
    serde_json::to_string(&out).map_err(|_| -7)
}

fn extract_text_content(content: Option<&Value>) -> Option<String> {
    let content = content?;
    if let Some(s) = content.as_str() {
        return Some(s.to_string());
    }
    if let Some(arr) = content.as_array() {
        let mut text = String::new();
        for part in arr {
            if let Some(s) = part.get("text").and_then(|t| t.as_str()) {
                text.push_str(s);
            }
        }
        if !text.is_empty() {
            return Some(text);
        }
    }
    None
}

fn clone_content_for_anthropic(content: &Value) -> Value {
    if content.is_string() || content.is_array() {
        content.clone()
    } else {
        json!("")
    }
}

fn clone_content_for_gemini(content: &Value) -> Value {
    if let Some(s) = content.as_str() {
        return json!([{"text": s}]);
    }
    if let Some(arr) = content.as_array() {
        let mut parts = Vec::new();
        for part in arr {
            let ty = part.get("type").and_then(|t| t.as_str()).unwrap_or("text");
            if ty == "text" {
                if let Some(text) = part.get("text").and_then(|t| t.as_str()) {
                    parts.push(json!({"text": text}));
                }
            } else if ty == "image_url" {
                if let Some(url) = part.get("image_url").and_then(|u| u.get("url")).and_then(|u| u.as_str()) {
                    parts.push(json!({"inlineData": {"data": url.strip_prefix("data:").unwrap_or(url)}}));
                }
            }
        }
        return json!(parts);
    }
    json!([])
}

fn convert_tools_to_anthropic(tools: &Value) -> Value {
    let Some(arr) = tools.as_array() else {
        return json!([]);
    };
    let mut out = Vec::new();
    for tool in arr {
        if let Some(func) = tool.get("function") {
            let mut converted = json!({"type": "function"});
            converted["name"] = func.get("name").cloned().unwrap_or(json!(""));
            if let Some(desc) = func.get("description") {
                converted["description"] = desc.clone();
            }
            if let Some(params) = func.get("parameters") {
                converted["input_schema"] = params.clone();
            }
            out.push(converted);
        }
    }
    json!(out)
}

fn convert_tool_choice_to_anthropic(tc: &Value) -> Value {
    if let Some(s) = tc.as_str() {
        if s == "none" {
            return json!({"type": "none"});
        }
        if s == "auto" {
            return json!({"type": "auto"});
        }
        if s == "required" {
            return json!({"type": "any"});
        }
    }
    if let Some(obj) = tc.as_object() {
        if let Some(func) = obj.get("function") {
            if let Some(name) = func.get("name").and_then(|n| n.as_str()) {
                return json!({"type": "tool", "name": name});
            }
        }
    }
    tc.clone()
}

fn convert_tools_to_gemini(tools: &Value) -> Value {
    let Some(arr) = tools.as_array() else {
        return json!([]);
    };
    let mut out = Vec::new();
    for tool in arr {
        if let Some(func) = tool.get("function") {
            let mut declaration = json!({});
            declaration["name"] = func.get("name").cloned().unwrap_or(json!(""));
            if let Some(desc) = func.get("description") {
                declaration["description"] = desc.clone();
            }
            if let Some(params) = func.get("parameters") {
                declaration["parameters"] = params.clone();
            }
            out.push(json!({"functionDeclarations": [declaration]}));
        }
    }
    json!(out)
}

fn copy_if_present(src: &Value, dst: &mut Value, key: &str) {
    if let Some(v) = src.get(key) {
        dst[key] = v.clone();
    }
}

fn copy_if_present_rename(src: &Value, dst: &mut Value, src_key: &str, dst_key: &str) {
    if let Some(v) = src.get(src_key) {
        dst[dst_key] = v.clone();
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn openai_chat_passes_through() {
        let json = r#"{"model":"gpt-4o","messages":[{"role":"developer","content":"hi"}],"stream":true}"#;
        let out = transform_openai_chat_request(json, "openai").unwrap();
        assert!(out.contains("\"role\":\"system\""));
        assert!(out.contains("include_usage"));
    }

    #[test]
    fn anthropic_chat_converts_system() {
        let json = r#"{"model":"claude-3","messages":[{"role":"system","content":"sys"},{"role":"user","content":"hello"}],"temperature":0.5}"#;
        let out = transform_openai_chat_request(json, "anthropic").unwrap();
        let v: Value = serde_json::from_str(&out).unwrap();
        assert!(v.get("system").is_some());
        assert_eq!(v["messages"].as_array().unwrap().len(), 1);
    }

    #[test]
    fn gemini_chat_converts_contents() {
        let json = r#"{"model":"gemini-1.5","messages":[{"role":"user","content":"hello"}],"max_tokens":100}"#;
        let out = transform_openai_chat_request(json, "gemini").unwrap();
        let v: Value = serde_json::from_str(&out).unwrap();
        assert!(v.get("contents").is_some());
        assert!(v.get("generationConfig").is_some());
    }

    #[test]
    fn embedding_gemini_conversion() {
        let json = r#"{"model":"models/gemini-embedding","input":"hello"}"#;
        let out = transform_embedding_request(json, "gemini").unwrap();
        let v: Value = serde_json::from_str(&out).unwrap();
        assert!(v.get("content").is_some());
    }
}
