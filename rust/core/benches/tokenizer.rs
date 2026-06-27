use std::ffi::CString;

use criterion::{black_box, criterion_group, criterion_main, Criterion};

fn criterion_benchmark(c: &mut Criterion) {
    let text = CString::new(
        "The quick brown fox jumps over the lazy dog. ".repeat(200),
    )
    .unwrap();
    let model = CString::new("gpt-4o").unwrap();

    c.bench_function("octopus_tokenizer_count", |b| {
        b.iter(|| {
            black_box(octopus_core::octopus_tokenizer_count(
                text.as_ptr(),
                model.as_ptr(),
            ))
        })
    });

    let json = CString::new(r#"{"model":"gpt-4o","usage":{"prompt_tokens":10,"completion_tokens":5,"total_tokens":15}}"#).unwrap();

    c.bench_function("octopus_json_extract_model", |b| {
        b.iter(|| {
            let mut out = std::ptr::null_mut();
            black_box(octopus_core::octopus_json_extract_model(
                json.as_ptr(),
                &mut out,
            ));
            if !out.is_null() {
                unsafe { octopus_core::octopus_free_string(out) };
            }
        })
    });

    c.bench_function("octopus_json_extract_usage", |b| {
        b.iter(|| {
            let mut out = std::ptr::null_mut();
            black_box(octopus_core::octopus_json_extract_usage(
                json.as_ptr(),
                &mut out,
            ));
            if !out.is_null() {
                unsafe { octopus_core::octopus_free_string(out) };
            }
        })
    });

    let aggregate = CString::new(r#"{"object":"chat.completion"}"#).unwrap();
    let chunk = CString::new(r#"{"object":"chat.completion.chunk","choices":[{"index":0,"delta":{"role":"assistant","content":"hello world"}}]}"#).unwrap();

    c.bench_function("octopus_sse_aggregate", |b| {
        b.iter(|| {
            let mut out = std::ptr::null_mut();
            black_box(octopus_core::octopus_sse_aggregate(
                aggregate.as_ptr(),
                chunk.as_ptr(),
                &mut out,
            ));
            if !out.is_null() {
                unsafe { octopus_core::octopus_free_string(out) };
            }
        })
    });
}

criterion_group!(benches, criterion_benchmark);
criterion_main!(benches);
