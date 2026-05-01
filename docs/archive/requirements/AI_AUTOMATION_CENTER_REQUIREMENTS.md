# Octopus AI Automation Center Requirements

> Current scope (2026-04-23 mainline planning pass): this document defines the `AI Automation Center + AI Profile dual-track configuration + Dynamic Routing AI Learning` mainline. It is the English requirements entry point for later implementation, review, modification, and search.
>
> Execution rule: this document must stay aligned with [LLM-Gateway-Refactor-Plan.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/planning/LLM-Gateway-Refactor-Plan.zh-CN.md), [USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/requirements/USER_CONTEXT_REQUIREMENTS_AND_WORKFLOW.zh-CN.md), and [DETAILED_EXECUTION_WORKFLOW.zh-CN.md](/D:/GPT-codex/octopus_repo/docs/archive/status/DETAILED_EXECUTION_WORKFLOW.zh-CN.md). If implementation needs to diverge, update the docs first, then change code.

---

## 1. Product Goal

The AI Automation Center should become the unified entry point for all AI-assisted operations in Octopus, rather than scattering buttons across channels, groups, model pricing, or settings.

Current goals:

- Add a top-level `AI Automation` section next to Home, Channels, Groups, Models/Pricing, Logs, and Settings.
- Let users customize the AI channel, model, `base_url`, and API key used by automation tasks.
- When no custom configuration exists, default to the local Octopus OpenAI-compatible endpoint, derived from the current service URL when possible and falling back to `http://127.0.0.1:8080/v1`.
- Support automatic model-list discovery and recommend free, recently successful, low-latency models by default.
- Support natural-language task input, built-in prompt templates, and user-defined prompts.
- Let AI generate grouping, channel recognition, price recognition, and model classification suggestions, then save them as independent AI Profiles.
- Add a settings-page switch that lets users choose `Manual Configuration` or a specific `AI Generated Profile`.
- Keep manual configuration and AI-generated profiles side by side. AI output must not overwrite manual configuration.
- Treat dynamic routing as a dedicated exception path: dynamic routing keeps local mechanism controls plus an AI learning switch. AI learning only affects runtime recommendations and must not permanently rewrite groups, channels, or priorities.

---

## 2. Core Principles

### 2.1 Dual Track Preservation

- Manual user configuration is the primary asset and must always be preserved.
- AI-generated configuration is saved as independent AI Profiles and must not silently overwrite source tables.
- `channels`, `groups`, `group_items`, `llm_infos`, and `route_target_overrides` must not be directly overwritten by AI tasks.
- Activating an AI Profile only changes the selected read source. It must not delete, overwrite, or reorder manual configuration.

### 2.2 Explicit Switching

- Settings must expose an explicit `Manual Configuration / AI Generated Profile` switch.
- Selecting AI-generated mode must require selecting a concrete AI Profile.
- If an AI Profile is missing, invalid, incomplete for the target model, or fails validation, runtime must fall back to manual configuration.
- Phase 1 only allows saving, previewing, and switching profiles. Silent overwrite or direct apply-to-manual behavior is not allowed by default.

### 2.3 Explainable AI Suggestions

- AI output must include a human summary, structured suggestions, confidence, risks, and evidence.
- AI-generated results must be reviewable, versioned, and comparable.
- Future apply-to-manual behavior must require diff, confirmation, audit, and rollback.

### 2.4 Price Must Not Drive Business Routing

- The default “free first, high success rate, low latency” model selection rule only selects the model used to run AI automation tasks.
- That rule must not affect business request routing order.
- Business routing must continue to respect explicit user configuration and dynamic-routing hard rules.

---

## 3. AI Automation Center Page Requirements

The AI Automation Center must include at least these areas:

- Current AI model status: endpoint, model, source, and whether the local default is active.
- AI channel configuration: custom `base_url`, API key, channel type, and model.
- Model discovery: fetch model lists and display source, availability, free/paid tendency, recent success rate, and average latency.
- Task cards: grouping, channel recognition, price recognition, model classification, configuration health check, and dynamic-routing explanation.
- Natural-language input: let users describe tasks directly, such as “generate grouping suggestions from these channels”.
- Prompt templates: built-in templates plus user-defined prompt and work-requirement additions.
- Task progress: collecting context, selecting model, calling AI, parsing output, generating profile, saving result.
- Result panel: AI summary, structured profile, risks, actionable items, and next steps. Current multi-AI execution fans out into multiple frontend-launched lane tasks rather than a backend orchestrator graph.
- Task history: input, model used, creation time, status, result, and linked profile.

---

## 4. AI Profile Requirements

An AI Profile is the independent saved unit for AI-generated configuration.

Required profile domains:

- `grouping`: AI-generated group plan.
- `channel_recognition`: AI channel-recognition result.
- `price_recognition`: AI price and billing recognition result.
- `model_classification`: AI model classification and canonical-name suggestion.
- `config_health`: AI configuration health-check result.

Each AI Profile must store at least:

- Name, domain, status, version.
- Source task ID.
- Original natural-language input.
- AI endpoint and model used.
- Structured content.
- Confidence.
- Risk explanation.
- Created and updated timestamps.
- Whether it is currently active.

AI Profile structured content must be previewable by the frontend and validated by the backend. It must not be stored only as unparseable prose. Typed payload is now the primary consumption contract, while legacy `content_json` remains an audit and compatibility fallback layer.

---

## 5. Settings Switch Requirements

Settings must add a configuration-source area.

It must support:

- Selecting `Manual Configuration`.
- Selecting `AI Generated Profile`.
- Selecting a concrete active AI Profile when AI-generated mode is enabled.
- Showing current source, profile name, update time, confidence, and risks.
- Providing an explicit fallback action to manual configuration.
- Clearly explaining that AI profiles do not overwrite manual configuration.

Forbidden behavior:

- Do not delete manual configuration when switching to an AI Profile.
- Do not rewrite `channels`, `groups`, `group_items`, or other source tables when switching profiles.
- Do not allow an invalid AI Profile to create an unexplained runtime state.

---

## 6. Dynamic Routing AI Learning Track

Dynamic routing is not part of the generic AI Profile override system.

Dynamic routing keeps two tracks:

- Local mechanism: existing `shadow-ai`, `hybrid`, `metrics-only`, `strict-mechanism`, and `incident-safe` modes, plus stats / route-target policy / circuit / probe signals.
- AI learning: local online-learning state at `(channel, key, model)` granularity, tracking success rate, failure rate, latency, fallback, race winner, score, and confidence.

The dynamic-routing settings area must add:

- `dynamic_routing_learning_enabled` switch.
- Learning-state summary.
- Learning-state inspection entry.
- Learning-state reset action.

Hard rules:

- AI learning only affects runtime recommendation order in modes that allow it.
- AI learning must not write back to `group_items`.
- AI learning must not overwrite user priority.
- AI learning must not permanently rewrite channels, groups, or key configuration.
- When learning is disabled, learning data may remain stored but must not participate in recommendation scoring.

---

## 7. Settings And API Requirements

New settings:

- `ai_automation_enabled`
- `ai_automation_base_url`
- `ai_automation_model`
- `ai_automation_use_local_default`
- `config_source_mode`
- `active_ai_profile_id`
- `dynamic_routing_learning_enabled`

New management APIs:

- `GET /api/v1/ai/config`
- `POST /api/v1/ai/config`
- `POST /api/v1/ai/models/fetch`
- `GET /api/v1/ai/prompt-templates`
- `POST /api/v1/ai/prompt-templates`
- `POST /api/v1/ai/tasks`
- `GET /api/v1/ai/tasks/:id`
- `POST /api/v1/ai/tasks/:id/cancel`
- `GET /api/v1/ai/profiles`
- `GET /api/v1/ai/profiles/:id`
- `POST /api/v1/ai/profiles/:id/activate`
- `GET /api/v1/dynamic-routing/learning`
- `POST /api/v1/dynamic-routing/learning/reset`

---

## 8. Acceptance Criteria

Minimum completion criteria for this mainline:

1. The top-level `AI Automation` section exists as the unified task entry point.
2. AI automation configuration supports local default and custom endpoint / model.
3. Model lists can be fetched automatically and default recommendation evidence is visible.
4. Natural-language tasks, prompt templates, user-defined prompts, and progress steps are available.
5. AI output is saved as AI Profiles and does not overwrite manual configuration.
6. Settings can explicitly switch between manual configuration and AI Profile.
7. Invalid AI Profiles fall back to manual configuration.
8. Dynamic-routing AI learning has an independent switch and learning-state surface.
9. Dynamic-routing AI learning only affects runtime recommendations and never permanently rewrites user configuration.
10. Docs, backend APIs, frontend entry points, tests, and worklogs stay aligned on this mainline.

---

## 9. Explicit Non-Goals

Phase 1 does not include:

- Silent AI overwrite of manual configuration.
- One-click apply without diff, confirmation, and rollback.
- Using generic AI Profiles as dynamic-routing override sources.
- Letting price affect business routing order.
- Depending on an external model service for dynamic-routing learning. Dynamic-routing learning must be local and explainable.
