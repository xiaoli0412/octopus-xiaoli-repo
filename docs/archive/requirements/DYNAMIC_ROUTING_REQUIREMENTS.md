# Octopus Dynamic Routing Requirements

> Current scope (2026-04-23 mainline update): this document describes the dynamic-routing behavior that is now wired in the repository: `dynamic_routing_mode`, `dynamic_routing_health_enabled`, the five `race_*_budget` settings, the relay recommendation / fallback / conservative mode layer, and the daily `dynamic summary scan` pipeline. It also includes the already-wired `dynamic_routing_learning_enabled` local AI-learning track.
>
> Current implementation note: `shadow-ai`, `hybrid`, `metrics-only`, `strict-mechanism`, and `incident-safe` are now implemented as runtime modes backed by existing stats / route-target policy / circuit / probe signals. The “AI” naming here refers first to a local explainable recommendation and learning layer, not an external model service. The external AI Automation Center may generate suggestions and AI Profiles, but it must not directly overwrite dynamic-routing configuration.

---

## 1. Current Product Goal

The current goal is to:

1. make the failover / race path mode-aware through a runtime switch layer
2. constrain concurrent racing through global-to-probe budgets
3. allow `hybrid` to adopt runtime recommendations only when confidence is high enough
4. let `shadow-ai` and `metrics-only` emit recommendation audit without mutating the live path
5. provide a daily summary so operators can understand current mode, effective mode, and routing-pool shape
6. provide a local AI-learning loop that learns at `(channel, key, model)` granularity and only affects runtime recommendation scoring

---

## 2. Current Shipped Scope

The shipped scope is limited to:

- runtime mode: `dynamic_routing_mode`
- runtime health toggle: `dynamic_routing_health_enabled`
- race budget: `race_global_budget`
- race budget: `race_group_budget`
- race budget: `race_channel_budget`
- race budget: `race_key_budget`
- race budget: `race_probe_budget`
- relay recommendation and mode-decision layer
- relay-log dynamic-routing audit fields
- daily `dynamic summary scan`
- stats API exposure for the dynamic-routing summary
- settings-page controls and summary surface for mode, health, budgets, and summary

The currently wired scope also includes:

- setting: `dynamic_routing_learning_enabled`
- learning state: `DynamicRouteLearningState`
- management API: `GET /api/v1/dynamic-routing/learning`
- management API: `POST /api/v1/dynamic-routing/learning/reset`
- relay writeback after completed attempts
- `hybrid` scoring integration
- settings-page learning summary, switch, and reset action

These capabilities exist to constrain, observe, and selectively adjust the current relay race / fallback path.

---

## 3. Runtime Behavior Requirements

### 3.1 Health Toggle

- When `dynamic_routing_health_enabled=false`, relay must fall back to the group's default race settings and skip dynamic health tuning.
- When `dynamic_routing_health_enabled=true`, relay may adjust `RaceAfterFails` and `RaceConcurrency` using route-target policy.
- These adjustments must remain request-local runtime decisions and must not be persisted back into user configuration.

### 3.2 Race Budgets

- Every race probe must be budget-constrained.
- Budgets must cover at least the global, group, channel, key, and probe levels.
- When budget is exhausted, the system must skip the affected race candidate instead of probing anyway.

### 3.3 Runtime Boundaries

- Dynamic health tuning may only change runtime thresholds and concurrency for the failover / race path.
- Recommendation logic may only reorder candidates for the current request when `hybrid` adopts the recommendation.
- Dynamic health tuning must not persist new thresholds.
- Dynamic-routing audit must record current mode, effective mode, decision type, fallback reason, and recommended order.

### 3.4 Runtime Modes

- `strict-mechanism`: must stay on the deterministic mechanism path and ignore recommendation ordering and recommendation-layer runtime tuning.
- `metrics-only`: must compute recommendation output and audit, but not change the live candidate order or apply recommendation-layer runtime tuning.
- `shadow-ai`: must compute a shadow recommendation and audit, but not change the live candidate order or apply recommendation-layer runtime tuning.
- `hybrid`: must adopt recommendation order when runtime confidence is sufficient and fall back automatically when it is not.
- `incident-safe`: must force a conservative path and tighten live racing behavior.

### 3.5 Local AI Learning

- Local AI learning stores state at `(channel_id, channel_key_id, model_name)` granularity.
- Learning inputs include success, failure, status code, latency, fallback, race winner, and last outcome time.
- Learning outputs include `score` and `confidence`, used as supporting signals in `hybrid` recommendation scoring.
- `shadow-ai` and `metrics-only` may record learning and audit output, but must not change live candidate order.
- `strict-mechanism` must remain deterministic.
- `incident-safe` must stay conservative and must not allow the learning layer to aggressively promote risky targets.
- When `dynamic_routing_learning_enabled=false`, stored learning state must not participate in recommendation scoring.
- The learning layer must not write back to `group_items`, overwrite user priority, or permanently reorder user configuration.

---

## 4. Configuration Requirements

The current dynamic-routing surface only requires these settings:

- `dynamic_routing_mode`: choose the active dynamic-routing mode
- `dynamic_routing_health_enabled`: enable or disable runtime dynamic-health tuning
- `race_global_budget`: global race budget limit
- `race_group_budget`: per-group race budget limit
- `race_channel_budget`: per-channel race budget limit
- `race_key_budget`: per-key race budget limit
- `race_probe_budget`: per-probe race budget limit
- `dynamic_routing_learning_enabled`: enable or disable local dynamic-routing AI learning in recommendation scoring

Docs, migrations, backend validation, and frontend settings wiring must stay aligned on this setting set.

---

## 5. Daily Summary Scan Requirements

The `dynamic summary scan` must:

- run once per day
- read current mode and effective mode context
- read the dynamic-health switch state
- summarize channel count, enabled channel count, group count, and failover-group count
- summarize key source-type distribution
- emit an explicit `basis`, currently `daily_summary_scan_no_runtime_mutation`
- emit an explainable `skipped` status and message when health is disabled

Its responsibility is limited to summary and observability:

- it does not persist new dynamic thresholds
- it does not rewrite user routing order
- it does not persist a new runtime decision back into saved settings

---

## 6. Observability and Surface Requirements

The current implementation must provide:

- a management API endpoint for reading the dynamic-routing summary
- a settings-page entry for mode, health toggle, and budgets
- a settings-page summary surface for the `dynamic summary scan`
- readable UI rendering for status, mode, effective mode, decision, basis, and recent execution details

If the summary exists without any API or UI consumer, it should not be counted as a completed loop.

---

## 7. Explicit Non-Goals

The following are not dynamic-routing requirements and must not be documented as shipped behavior or release gates:

- external model-backed recommendation service
- persisted threshold auto-tuning
- automatic reordering of user-configured routes
- generic AI Profile override of dynamic-routing configuration

The following has moved from non-goal to the current official target:

- a local, explainable AI-learning loop at `(channel, key, model)` granularity
- runtime learning scores controlled by `dynamic_routing_learning_enabled`

---

## 8. Acceptance Criteria

The current dynamic-routing scope is only complete when all of the following are true:

1. Defaults, migrations, and validation include `dynamic_routing_mode`, `dynamic_routing_health_enabled`, and all `race_*_budget` settings.
2. Relay runtime actually consumes mode, health toggle, and budgets, rather than leaving them as doc-only or DB-only fields.
3. The five runtime modes all have explainable behavioral differences and test coverage.
4. The daily `dynamic summary scan` produces summary only and does not persist new thresholds or permanently reorder routes.
5. The summary is readable through the management API and consumed by the settings page.
6. `dynamic_routing_learning_enabled` already controls whether learning state participates in recommendation scoring.
7. Tests prove learning recommendations do not write back to `group_items` or overwrite user priority.

