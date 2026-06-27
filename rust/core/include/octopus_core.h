#ifndef OCTOPUS_CORE_H
#define OCTOPUS_CORE_H

#ifdef __cplusplus
extern "C" {
#endif

/* Count tokens for `text` using the encoding implied by `model`.
 * Returns a non-negative count on success, negative error code otherwise. */
int octopus_tokenizer_count(const char *text, const char *model);

/* Free a C string previously returned by this library. */
void octopus_free_string(char *s);

/* Extract the top-level `model` field from a JSON request/response.
 * On success returns 0 and writes a JSON string `{"model":"..."}` through *out. */
int octopus_json_extract_model(const char *json, char **out);

/* Extract the top-level `usage` object from a JSON response.
 * On success returns 0 and writes a JSON string through *out. */
int octopus_json_extract_usage(const char *json, char **out);

/* Merge a streaming SSE chunk JSON into an aggregate response JSON.
 * On success returns 0 and writes the updated aggregate JSON through *out. */
int octopus_sse_aggregate(const char *aggregate, const char *chunk, char **out);

/* Transform an OpenAI chat completion request JSON for the target provider.
 * `target` is one of: openai, anthropic, gemini, volcengine.
 * On success returns 0 and writes the transformed JSON through *out. */
int octopus_transform_openai_chat_request(const char *json, const char *target, char **out);

/* Transform an OpenAI chat completion response JSON for the target provider.
 * On success returns 0 and writes the transformed JSON through *out. */
int octopus_transform_openai_response(const char *json, const char *target, char **out);

/* Transform an OpenAI embedding request JSON for the target provider.
 * On success returns 0 and writes the transformed JSON through *out. */
int octopus_transform_embedding_request(const char *json, const char *target, char **out);

/* Select a candidate according to the given strategy.
 * `candidates_json` is an array of objects with id, weight, latency, priority,
 * healthy and circuit_state fields. `strategy` is weighted/round_robin/random/failover.
 * `current_idx` is the current round-robin/weighted cursor.
 * On success returns 0 and writes `{"id":..., "next_index":...}` through *out. */
int octopus_balance_select(const char *candidates_json, const char *strategy, int current_idx, char **out);

/* Opaque candidate descriptor for the binary balance selector. */
typedef struct {
    long long id;
    long long weight;
    long long latency;
    long long priority;
    int healthy;
    int circuit_open;
} OctopusBalanceCandidate;

/* High-performance binary variant of octopus_balance_select.
 * Avoids JSON round-trips by accepting a packed candidate array.
 * On success returns 0 and writes the selected id and next cursor. */
int octopus_balance_select_v2(const OctopusBalanceCandidate *candidates, int candidates_len,
                              const char *strategy, int current_idx,
                              long long *out_id, int *out_next_index);

/* Opaque handle for a thread-safe stream buffer. */
typedef struct OctopusStreamBufferHandle OctopusStreamBufferHandle;

/* Create a new thread-safe stream buffer. Returns an opaque handle. */
OctopusStreamBufferHandle *octopus_stream_buffer_create(void);

/* Feed a new chunk into the stream buffer. Returns a JSON array of complete
 * events extracted so far through `out`. */
int octopus_stream_buffer_feed(OctopusStreamBufferHandle *handle, const char *chunk, char **out);

/* Take all complete events currently buffered without feeding new data. */
int octopus_stream_buffer_take(OctopusStreamBufferHandle *handle, char **out);

/* Free a stream buffer previously created by octopus_stream_buffer_create. */
void octopus_stream_buffer_free(OctopusStreamBufferHandle *handle);

/* Stateless event extractor for SSE streams.
 * Parses `data` and returns a JSON array of complete events through *out_events
 * and the number of consumed bytes through *out_consumed_bytes.
 * The unconsumed suffix is `data + *out_consumed_bytes`.
 * On success returns 0. */
int octopus_stream_extract_events(const char *data, char **out_events, int *out_consumed_bytes);

/* Opaque descriptor for SSE event boundaries. */
typedef struct OctopusEventBoundaries OctopusEventBoundaries;

/* High-performance boundary finder for SSE streams.
 * `data` points to `len` bytes of valid UTF-8 text.
 * Finds the trimmed byte ranges of complete events in `data` and returns an
 * opaque descriptor through `out_boundaries`. The descriptor must be freed with
 * `octopus_stream_boundaries_free`. On success returns 0. */
int octopus_stream_find_event_boundaries(const char *data, int len,
                                         OctopusEventBoundaries **out_boundaries);

/* Free a boundary descriptor previously returned by
 * `octopus_stream_find_event_boundaries`. */
void octopus_stream_boundaries_free(OctopusEventBoundaries *boundaries);

/* High-performance variant of the SSE boundary finder.
 * `starts` and `ends` must point to arrays of at least `max_events` entries.
 * On success writes the number of events into `out_count` and the number of
 * consumed bytes into `out_consumed`. Returns -15 if `max_events` is too small. */
int octopus_stream_find_event_boundaries_ex(const char *data, int len,
                                            int *starts, int *ends, int max_events,
                                            int *out_count, int *out_consumed);

/* Merge a delta StatsMetrics JSON into an existing aggregate JSON.
 * On success returns 0 and writes the merged JSON through *out. */
int octopus_stats_merge_hourly(const char *existing_json, const char *delta_json, char **out);

/* Compute a quantile for an array of f64 values.
 * `quantile` must be in [0.0, 1.0].
 * On success returns 0 and writes `{"quantile":..., "value":...}` through *out. */
int octopus_stats_quantile(const char *values_json, double quantile, char **out);

#ifdef __cplusplus
}
#endif

#endif /* OCTOPUS_CORE_H */
