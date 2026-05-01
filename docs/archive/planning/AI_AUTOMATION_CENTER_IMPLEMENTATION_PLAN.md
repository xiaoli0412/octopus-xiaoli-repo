# Octopus AI Automation Center Implementation Plan

> Current scope (2026-04-23 mainline planning pass): this document is the English implementation plan for `AI Automation Center + AI Profile dual-track configuration + Dynamic Routing AI Learning`. Requirements are defined in [AI_AUTOMATION_CENTER_REQUIREMENTS.md](/D:/GPT-codex/octopus_repo/docs/AI_AUTOMATION_CENTER_REQUIREMENTS.md).
>
> Execution rule: implementation must first align with the canonical plan, user-context ledger, detailed workflow, and dynamic-routing docs. Phase 1 must not overwrite manual user configuration or trigger silent migration.

---

## 1. Implementation Goal

This mainline introduces AI automation into Octopus in stages:

1. Sync documentation and mainline planning first.
2. Add backend models, settings, and task framework for AI automation.
3. Add the top-level `AI Automation` page and settings source switch.
4. Implement local AI learning for dynamic routing.
5. Expand into AI grouping, channel recognition, price recognition, audit, and rollback.

Key invariants:

- AI Profiles and manual configuration are preserved as separate tracks.
- AI-generated output must not silently overwrite `channels`, `groups`, `group_items`, `llm_infos`, or `route_target_overrides`.
- Dynamic-routing AI learning only affects runtime recommendations and must not permanently rewrite user configuration.

---

## 2. Phases

### Phase H1: Documentation And Mainline Sync

Goal: write this mainline into all key project docs.

Tasks:

- Add Chinese and English requirements docs.
- Add Chinese and English implementation plans.
- Add a worklog.
- Update the canonical plan.
- Update the user-context ledger with requirements 54-64.
- Update the detailed workflow with Phase H.
- Update current status, frontend mainline status, and environment/next plan docs.
- Update Chinese and English dynamic-routing requirements and implementation docs, removing the old “online learning is a non-goal” position.

Acceptance:

- Key docs contain `AI Automation`, `AI Profile`, `dynamic_routing_learning_enabled`, and “must not overwrite user configuration” wording.

### Phase H2: Backend Data Models

Goal: define AI automation persistence structures.

New models:

- `AIAutomationConfig`
- `AITask`
- `AITaskStep`
- `AIPromptTemplate`
- `AIProfile`
- `AIProfileVersion`
- `DynamicRouteLearningState`

New settings:

- `ai_automation_enabled`
- `ai_automation_base_url`
- `ai_automation_model`
- `ai_automation_use_local_default`
- `config_source_mode`
- `active_ai_profile_id`
- `dynamic_routing_learning_enabled`

Acceptance:

- Migrations are repeatable.
- Defaults are safe.
- Existing data is not affected.

### Phase H3: Backend APIs And Task Framework

Goal: provide APIs for AI config, tasks, templates, profiles, and dynamic-learning state.

New APIs:

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

Requirements:

- AI tasks can start as synchronous short tasks or lightweight asynchronous status updates.
- Task steps must drive the frontend progress bar.
- Profile activation must only update source-selection settings and must not overwrite manual tables.

### Phase H4: Frontend Top-Level Section

Goal: add the top-level `AI Automation` entry and page shell.

Page modules:

- AI model status card.
- AI endpoint / API key / model configuration.
- Automatic model-list discovery.
- Task-type cards.
- Natural-language input.
- Built-in and custom prompt areas.
- Task progress bar.
- Result preview and AI Profile save area.
- Task history list.

Acceptance:

- Navigation can open the page.
- The page preserves the existing deep-green, rounded, progressive-disclosure product style.
- Chinese UI does not leak raw enums or i18n keys.

### Phase H5: AI Profile Dual-Track Switch

Goal: add a settings-page configuration source switch.

Requirements:

- Support `manual` and `ai_profile`.
- Selecting `ai_profile` requires selecting a profile.
- Invalid profiles fall back to `manual`.
- UI clearly explains that AI profiles do not overwrite manual configuration.

Tests:

- `manual -> ai_profile -> manual` preserves all manual configuration.
- Invalid profiles do not break runtime availability.

### Phase H6: Dynamic Routing AI Learning

Goal: implement local, explainable, route-target-level dynamic-routing learning.

Data granularity:

- `(channel_id, channel_key_id, model_name)`

Learning fields:

- `sample_count`
- `success_ewma`
- `failure_ewma`
- `latency_ewma_ms`
- `fallback_ewma`
- `race_win_ewma`
- `score`
- `confidence`
- `last_status_code`
- `last_outcome`
- `last_outcome_at`
- `updated_at`

Integration points:

- Record success, latency, and token/cost references after successful relay attempts.
- Record failures, status codes, and error classes after failed relay attempts.
- Record race winners for race fallback.
- Recommendation scoring reads learning state and contributes to `hybrid` scoring.

Hard rules:

- When `dynamic_routing_learning_enabled` is disabled, learning must not participate in scoring.
- Do not modify user groups or priorities.
- `strict-mechanism` stays deterministic.
- `shadow-ai` and `metrics-only` may record and audit, but must not change live ordering.

### Phase H7: AI Grouping And Channel Recognition MVP

Goal: first batch of AI Profile generation.

Scope:

- Grouping suggestions.
- Channel type recognition.
- Model classification suggestions.
- Configuration health checks.

Requirements:

- Generate profiles only.
- Support preview, explanation, save, and switch.
- Do not silently apply changes to manual configuration.

### Phase H8: Price Recognition And Smart Grouping Enhancements

Goal: expand AI automation tasks.

Scope:

- Price and billing-mode recognition.
- Source-type suggestions.
- Canonical-name suggestions.
- More complete grouping strategy templates.

Requirements:

- Price recognition generates suggestions only and must not affect business routing order.
- Writing manual price rules must be handled by a separate diff / confirm / rollback flow.

### Phase H9: Audit, Diff, And Rollback Enhancements

Goal: prepare safety boundaries for future selective application of AI suggestions into manual configuration.

Scope:

- Profile diff.
- Selective apply.
- Audit log.
- Rollback point.
- Risk confirmation.

This capability is not enabled by default in Phase 1.

---

## 3. File Targets

Expected backend files:

- `internal/model/ai_automation.go`
- `internal/op/ai_automation.go`
- `internal/server/handlers/ai_automation.go`
- `internal/model/dynamic_route_learning.go`
- `internal/op/dynamic_route_learning.go`
- `internal/db/migrate/012.go`
- `internal/relay/dynamic_learning.go`

Expected frontend files:

- `web/src/components/modules/ai-automation/index.tsx`
- `web/src/api/endpoints/ai-automation.ts`
- `web/src/components/modules/setting/ConfigSource.tsx`
- `web/src/components/modules/setting/DynamicRouting.tsx`
- `web/src/route/config.tsx`
- `web/src/components/modules/navbar/nav-store.ts`
- `web/public/locale/*.json`

Docs:

- `docs/AI_AUTOMATION_CENTER_REQUIREMENTS.zh-CN.md`
- `docs/AI_AUTOMATION_CENTER_REQUIREMENTS.md`
- `docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.zh-CN.md`
- `docs/AI_AUTOMATION_CENTER_IMPLEMENTATION_PLAN.md`

---

## 4. Verification Plan

Documentation phase:

- Search for `AI 自动化`, `AI Profile`, `dynamic_routing_learning_enabled`, and `不覆盖用户配置`.
- Confirm Chinese and English AI automation docs exist in pairs.
- Confirm dynamic-routing docs no longer list online learning as a non-goal.

Backend phase:

- `go test ./internal/model ./internal/op ./internal/server/handlers ./internal/relay/...`
- Repeatable migration tests.
- API handler tests.
- Profile activation must not overwrite source table tests.
- Dynamic learning switch and scoring tests.

Frontend phase:

- `pnpm --dir web exec tsc --noEmit`
- Locale consistency checks.
- AI automation page component tests.
- Settings configuration-source switch tests.
- Dynamic-routing learning switch tests.

End-to-end phase:

- Create an AI task and generate a profile.
- Activate the profile, then switch back to manual configuration.
- Confirm manual configuration remains untouched.
- Confirm relay dynamic-routing learning records and recommendation scores are explainable.

---

## 5. Risks And Rollback

- Risk: AI Profile semantics are misunderstood as overwriting source config. Rollback: enforce preview / activate dual track in UI and API; never write source tables.
- Risk: default free-model selection leaks into business routing. Rollback: docs and tests clarify it only selects the AI automation execution model.
- Risk: dynamic learning changes user priority. Rollback: read only at recommendation-score level, do not write group items or priority.
- Risk: task center scope becomes too large. Rollback: before Phase H7, implement only the task framework and dynamic-routing learning mainline.

---

## 6. Phase 1 Completion Criteria

Phase 1 is complete only when:

1. Documentation mainline is fully synced.
2. AI Automation section exists.
3. AI config, local default, and model discovery are available.
4. Natural-language task input and progress steps are available.
5. AI Profiles can be saved, previewed, activated, and reverted.
6. Settings can manually switch between manual configuration and AI Profile.
7. Dynamic-routing AI learning state can be recorded, queried, reset, and controlled by the switch.
8. Key tests and worklogs are complete.
