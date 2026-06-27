use std::ffi::c_int;
use std::sync::Mutex;

/// Opaque handle for a thread-safe stream buffer.
pub struct StreamBufferHandle {
    inner: Mutex<StreamBuffer>,
}

struct StreamBuffer {
    data: String,
}

impl StreamBufferHandle {
    fn new() -> Self {
        Self {
            inner: Mutex::new(StreamBuffer { data: String::new() }),
        }
    }
}

/// Create a new stream buffer handle.
pub fn buffer_create() -> *mut StreamBufferHandle {
    let handle = Box::new(StreamBufferHandle::new());
    Box::into_raw(handle)
}

/// Feed a chunk into the buffer and return a JSON array of complete events.
pub fn buffer_feed(handle: *mut StreamBufferHandle, chunk: &str) -> Result<String, c_int> {
    let guard = unsafe { &*handle };
    let mut inner = guard.inner.lock().map_err(|_| -12)?;
    inner.data.push_str(chunk);
    let (events, remainder) = extract_events(&inner.data);
    inner.data = remainder;
    Ok(serde_json::to_string(&events).map_err(|_| -7)?)
}

/// Take all complete events currently buffered without feeding new data.
pub fn buffer_take(handle: *mut StreamBufferHandle) -> Result<String, c_int> {
    let guard = unsafe { &*handle };
    let mut inner = guard.inner.lock().map_err(|_| -12)?;
    let (events, remainder) = extract_events(&inner.data);
    inner.data = remainder;
    Ok(serde_json::to_string(&events).map_err(|_| -7)?)
}

/// Free a stream buffer handle.
pub unsafe fn buffer_free(handle: *mut StreamBufferHandle) {
    if !handle.is_null() {
        drop(Box::from_raw(handle));
    }
}

/// Extract complete SSE events from `data` without maintaining internal state.
/// Returns a JSON array of complete events and the number of bytes consumed.
pub fn extract_events_stateless(data: &str) -> Result<(String, usize), c_int> {
    let (events_json, consumed) = extract_events_json(data);
    Ok((events_json, consumed))
}

/// Extract complete SSE events from the buffered data.
/// An event is a sequence of lines ending with a blank line (`\n\n`).
/// Returns (complete_events, remaining_partial_data).
fn extract_events(data: &str) -> (Vec<String>, String) {
    let mut events = Vec::new();
    let mut start = 0;
    while let Some(pos) = data[start..].find("\n\n") {
        let end = start + pos;
        let event_text = data[start..end].trim();
        if !event_text.is_empty() {
            events.push(event_text.to_string());
        }
        start = end + 2;
    }
    (events, data[start..].to_string())
}

/// Find the trimmed byte ranges of complete SSE events in `data`.
/// Returns a vector of (start, end) byte offsets and the number of bytes that
/// can be safely discarded (up to and including the last complete event).
fn find_event_ranges(data: &str) -> (Vec<(usize, usize)>, usize) {
    find_event_ranges_bytes(data.as_bytes())
}

/// Byte-oriented variant that does not require UTF-8 validation.
fn find_event_ranges_bytes(data: &[u8]) -> (Vec<(usize, usize)>, usize) {
    let mut ranges = Vec::new();
    let mut start = 0;
    while let Some(pos) = find_double_lf(&data[start..]) {
        let end = start + pos;
        let slice = &data[start..end];
        let trimmed_start = trim_ascii_ws_start(slice);
        let trimmed_end = trim_ascii_ws_end(slice);
        if trimmed_start < trimmed_end {
            ranges.push((start + trimmed_start, start + trimmed_end));
        }
        start = end + 2;
    }
    (ranges, start)
}

fn find_double_lf(data: &[u8]) -> Option<usize> {
    if data.len() < 2 {
        return None;
    }
    let mut i = 0;
    while i + 1 < data.len() {
        if data[i] == b'\n' && data[i + 1] == b'\n' {
            return Some(i);
        }
        i += 1;
    }
    None
}

fn trim_ascii_ws_start(data: &[u8]) -> usize {
    let mut i = 0;
    while i < data.len() && data[i].is_ascii_whitespace() {
        i += 1;
    }
    i
}

fn trim_ascii_ws_end(data: &[u8]) -> usize {
    let mut i = data.len();
    while i > 0 && data[i - 1].is_ascii_whitespace() {
        i -= 1;
    }
    i
}

/// Fill caller-provided arrays with the byte ranges of complete SSE events.
/// Returns 0 on success, -15 if `max_events` is too small. Writes the number
/// of events into `out_count` and the consumed byte count into `out_consumed`.
pub unsafe fn fill_event_boundaries_bytes(
    data: *const u8,
    len: usize,
    starts: *mut c_int,
    ends: *mut c_int,
    max_events: usize,
    out_count: *mut c_int,
    out_consumed: *mut c_int,
) -> c_int {
    let bytes = std::slice::from_raw_parts(data, len);
    let mut count = 0usize;
    let mut start = 0usize;
    let mut i = 0usize;
    while i + 1 < bytes.len() {
        if bytes[i] == b'\n' && bytes[i + 1] == b'\n' {
            let end = i;
            let slice = &bytes[start..end];
            let trimmed_start = trim_ascii_ws_start(slice);
            let trimmed_end = trim_ascii_ws_end(slice);
            if trimmed_start < trimmed_end {
                if count >= max_events {
                    return -15;
                }
                *starts.add(count) = (start + trimmed_start) as c_int;
                *ends.add(count) = (start + trimmed_end) as c_int;
                count += 1;
            }
            start = end + 2;
            i = start;
            continue;
        }
        i += 1;
    }
    *out_count = count as c_int;
    *out_consumed = start as c_int;
    0
}

/// Build a JSON array of complete SSE events directly from `data` without
/// allocating per-event strings. Returns (json_array, consumed_bytes).
fn extract_events_json(data: &str) -> (String, usize) {
    let mut ranges: Vec<(usize, usize)> = Vec::new();
    let mut start = 0;
    while let Some(pos) = data[start..].find("\n\n") {
        let end = start + pos;
        // Find non-whitespace bounds within the event block.
        let slice = &data[start..end];
        let trimmed_start = slice.find(|c: char| !c.is_whitespace()).unwrap_or(0);
        let trimmed_end = slice.rfind(|c: char| !c.is_whitespace()).map(|i| i + 1).unwrap_or(0);
        if trimmed_start < trimmed_end {
            ranges.push((start + trimmed_start, start + trimmed_end));
        }
        start = end + 2;
    }

    let mut json = String::with_capacity(data.len());
    json.push('[');
    for (i, (s, e)) in ranges.iter().enumerate() {
        if i > 0 {
            json.push(',');
        }
        json.push('"');
        for ch in data[*s..*e].chars() {
            match ch {
                '"' => json.push_str("\\\""),
                '\\' => json.push_str("\\\\"),
                '\n' => json.push_str("\\n"),
                '\r' => json.push_str("\\r"),
                '\t' => json.push_str("\\t"),
                c if (c as u32) < 0x20 => {
                    json.push_str(&format!("\\u{:04x}", c as u32));
                }
                c => json.push(c),
            }
        }
        json.push('"');
    }
    json.push(']');
    (json, start)
}

/// Opaque C descriptor for SSE event boundaries.
#[repr(C)]
pub struct OctopusEventBoundaries {
    pub starts: *mut c_int,
    pub ends: *mut c_int,
    pub count: c_int,
    pub consumed_bytes: c_int,
}

/// Find the trimmed byte ranges of complete SSE events in `data`.
/// Returns an opaque boundary descriptor that must be freed with
/// `octopus_stream_boundaries_free`.
pub fn find_event_boundaries(data: &str) -> *mut OctopusEventBoundaries {
    let (ranges, consumed) = find_event_ranges(data);
    let count = ranges.len();
    let (starts, ends): (Vec<c_int>, Vec<c_int>) = ranges
        .into_iter()
        .map(|(s, e)| (s as c_int, e as c_int))
        .unzip();

    let mut starts_box = starts.into_boxed_slice();
    let mut ends_box = ends.into_boxed_slice();

    let boundaries = Box::new(OctopusEventBoundaries {
        starts: starts_box.as_mut_ptr(),
        ends: ends_box.as_mut_ptr(),
        count: count as c_int,
        consumed_bytes: consumed as c_int,
    });

    std::mem::forget(starts_box);
    std::mem::forget(ends_box);

    Box::into_raw(boundaries)
}

/// Free a boundary descriptor previously returned by `find_event_boundaries`.
pub unsafe fn free_event_boundaries(b: *mut OctopusEventBoundaries) {
    if b.is_null() {
        return;
    }
    let count = (*b).count as usize;
    if count > 0 {
        drop(Vec::from_raw_parts((*b).starts, count, count));
        drop(Vec::from_raw_parts((*b).ends, count, count));
    }
    drop(Box::from_raw(b));
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn extract_simple_sse() {
        let (events, rem) = extract_events("data: hello\n\ndata: world\n\n");
        assert_eq!(events.len(), 2);
        assert!(rem.is_empty());
    }

    #[test]
    fn partial_event_held() {
        let (events, rem) = extract_events("data: partial");
        assert!(events.is_empty());
        assert_eq!(rem, "data: partial");
    }

    #[test]
    fn buffer_feed_sequence() {
        let h = buffer_create();
        let r1 = buffer_feed(h, "data: a\n\ndata: b").unwrap();
        let r2 = buffer_feed(h, "\n\ndata: c\n\n").unwrap();
        unsafe { buffer_free(h) };

        let v1: Vec<String> = serde_json::from_str(&r1).unwrap();
        let v2: Vec<String> = serde_json::from_str(&r2).unwrap();
        assert_eq!(v1.len(), 1);
        assert_eq!(v2.len(), 2);
    }

    #[test]
    fn buffer_take_returns_pending() {
        let h = buffer_create();
        buffer_feed(h, "data: x\n\ndata: y\n\n").unwrap();
        let r = buffer_take(h).unwrap();
        unsafe { buffer_free(h) };
        let v: Vec<String> = serde_json::from_str(&r).unwrap();
        assert_eq!(v.len(), 2);
    }
}
