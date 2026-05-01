# Octopus Dynamic Routing Implementation Plan

> Current scope (2026-04-23 mainline update): this plan covers the dynamic-routing path now wired in the codebase: `dynamic_routing_mode`, `dynamic_routing_health_enabled`, the five `race_*_budget` settings, the relay recommendation / fallback / conservative mode layer, the daily `dynamic summary scan`, the dynamic-routing summary API, and the matching settings-page controls and summary surface. It also includes the already-wired `dynamic_routing_learning_enabled` local AI-learning track.
>
> Current implementation note: `shadow-ai`, `hybrid`, `metrics-only`, `strict-mechanism`, and `incident-safe` are now implemented and tested. The “AI” naming here refers first to a local explainable recommendation and learning layer, not an external model service. The external AI Automation Center may generate suggestions and AI Profiles, but it must not directly overwrite dynamic-routing configuration.

---

## 1. Current Target Deliverables

The active implementation path delivers five concrete capabilities:

1. persist, validate, and surface the dynamic-routing mode, dynamic-health toggle, and race-budget settings
2. make relay actually consume those settings inside the failover / race path
3. provide explainable runtime differences across the five dynamic-routing modes
4. produce a daily `dynamic summary scan` with no runtime mutation
5. expose that summary through the management API and the settings page
6. include local AI-learning state that learns at `(channel, key, model)` granularity and participates only in allowed runtime recommendations

---

## 2. Module Breakdown

### 2.1 Settings and Migration Module

Responsibilities:

- provide defaults for `dynamic_routing_mode`, `dynamic_routing_health_enabled`, and the five `race_*_budget` settings
- validate those settings as a boolean or non-negative integers
- expose them through frontend setting constants and the settings page

### 2.2 Relay Runtime Tuning Module

Responsibilities:

- when health is enabled, adjust `RaceAfterFails` and `RaceConcurrency` using route-target policy
- when health is disabled, fall back to the group's original defaults
- keep the adjustment request-local and never persist it back into configuration

### 2.3 Mode Decision and Audit Module

Responsibilities:

- resolve `dynamic_routing_mode`
- produce recommendation order and confidence from existing runtime signals
- decide whether the current request adopts recommendation, records shadow audit only, records metrics only, stays deterministic without recommendation-layer tuning, or enters incident-safe mode
- persist mode, effective mode, decision type, fallback reason, and recommendation order into relay-log audit fields

### 2.4 Race Budget Module

Responsibilities:

- enforce separate budgets for the global, group, channel, key, and probe levels
- reject additional race candidates when the relevant budget is exhausted
- integrate directly with the existing failover / race fallback path

### 2.5 Daily Summary Scan Module

Responsibilities:

- run `dynamic summary scan` once per day
- summarize current mode, effective mode, and decision posture
- summarize channels, groups, failover groups, and key source-type distribution
- emit `status`, `message`, `basis`, and timestamps
- explicitly keep the basis at `daily_summary_scan_no_runtime_mutation`

### 2.6 Summary API and Settings UI Module

Responsibilities:

- expose the dynamic-routing summary through the management API
- show mode, health toggle, budget inputs, and summary panel in settings
- render readable status strings for `ok`, `skipped`, and `error`

### 2.7 Local AI Learning Module

Responsibilities:

- add the `dynamic_routing_learning_enabled` setting
- add `DynamicRouteLearningState` at `(channel_id, channel_key_id, model_name)` granularity
- update learning state after relay attempts with success, failure, status code, latency, fallback, and race-winner signals
- compute `score` and `confidence`, then use them as supporting signals in `hybrid` scoring
- record and audit in `shadow-ai` and `metrics-only` without changing live order
- keep `strict-mechanism` deterministic
- retain learning data when disabled, but exclude it from scoring
- explicitly forbid writing back to `group_items`, priority, channel config, or user-defined order

---

## 3. Current Request and Task Flow

### 3.1 Relay Request Path

1. request enters relay
2. `dynamic_routing_mode` and `dynamic_routing_health_enabled` are read
3. recommendation order and confidence are derived from existing runtime signals
4. the request adopts recommendation, shadow audit, metrics audit, deterministic mechanism, or incident-safe posture according to the current mode
5. when health is enabled, runtime `RaceAfterFails` / `RaceConcurrency` are derived from route-target policy
6. `race_*_budget` is acquired before race probing starts
7. if budget is available, race probing continues; if not, the candidate is skipped
8. mode-decision audit is written into relay log fields
9. if `dynamic_routing_learning_enabled=true`, local learning state is updated after completed attempts
10. later scoring may read learning state for `hybrid` recommendations, without writing back user configuration

### 3.2 Background Summary Path

1. the daily scheduler triggers `dynamic summary scan`
2. the current mode and health switch are read; when disabled, an explainable summary is still written
3. when enabled, channels and groups are listed
4. an in-memory summary is produced
5. the management API reads that summary
6. the settings page consumes and displays it

---

## 4. Current Configuration List

The current implementation scope only includes:

- `dynamic_routing_mode`
- `dynamic_routing_health_enabled`
- `race_global_budget`
- `race_group_budget`
- `race_channel_budget`
- `race_key_budget`
- `race_probe_budget`

The currently wired scope also includes:

- `dynamic_routing_learning_enabled`

This dynamic-routing track still does not add external-model service flags. Learning must be local and explainable.

---

## 5. Current Delivery Phases

### Phase A

Add defaults, migration coverage, and validation for the mode, health toggle, and race budgets.

### Phase B

Wire the mode layer, health toggle, and budgets into the relay failover / race runtime path.

### Phase C

Implement relay audit plus the daily `dynamic summary scan` as summary-only behavior, with no runtime mutation.

### Phase D

Expose the summary and settings through the management API and the settings page.

### Phase E

Local AI-learning state and the `dynamic_routing_learning_enabled` switch are already wired; learning scores already feed `hybrid` recommendation scoring, and the learning layer does not overwrite user configuration.

---

## 6. Future Backlog And Promoted Targets

The following remain backlog items and are not current dynamic-routing commitments:

- external model-backed recommendation service
- persisted threshold auto-tuning
- generic AI Profile override of dynamic-routing configuration

The following has been promoted to the current official target:

- local AI-learning loop
- `dynamic_routing_learning_enabled` setting
- `(channel, key, model)` learning state
- learning score support for `hybrid` recommendation scoring
- tests proving learning recommendations do not write back `group_items` or overwrite priority

If work resumes in these areas later, it should start from a new plan grounded in the then-current code reality rather than restoring the old wording.

---

## 7. Current Minimum Release Gate

Before shipping the current phase, all of the following should be true:

1. `dynamic_routing_mode`, `dynamic_routing_health_enabled`, and all `race_*_budget` settings have defaults, migration coverage, validation, and frontend constant wiring.
2. Relay runtime actually consumes those settings and produces explainable skip behavior when budgets are exhausted.
3. The five runtime modes all have explainable behavioral differences.
4. `dynamic summary scan` is registered as a daily task and explicitly remains summary-only, without persisting thresholds or rewriting user route order.
5. The management API can read the dynamic-routing summary.
6. The settings page can display mode, health toggle, budget settings, and summary result.
7. Settings already control `dynamic_routing_learning_enabled`.
8. Learning state can already be queried and reset, and disabled learning does not participate in scoring.
9. Tests prove learning recommendations do not write back `group_items` or overwrite priority.

