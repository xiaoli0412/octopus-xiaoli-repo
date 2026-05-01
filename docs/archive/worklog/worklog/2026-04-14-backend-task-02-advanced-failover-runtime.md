# Backend Task 02 - Advanced Failover Runtime Hardening

## Goal

Harden the runtime path for advanced failover so the system moves from simple retry toward the canonical plan's 360-second failover window, conditional racing escalation, and route-target-aware runtime decisions.

## Completed

### Group failover configuration chain

Added and fully wired the following fields:

- `failover_window_sec`
- `race_after_fails`
- `race_concurrency`

These now exist across:

- model definition
- update request
- operation layer
- frontend types
- create/edit UI
- migration (`008.go`)

### Relay runtime controls

The relay runtime now includes:

- failover window deadline derived from `failover_window_sec`
- escalation trigger based on `race_after_fails`
- first-stage racing only in `GroupModeFailover`
- source-type-aware default racing gate
- model-level racing gate via `billing_mode` and `probe_policy`

### Phase-1 bounded racing fallback

Implemented a conservative non-stream racing path:

- only for non-streaming requests
- bounded by `race_concurrency`
- skips channels with no key / unsupported type / circuit-tripped state
- race winner returns an internal response to the main goroutine
- main goroutine writes the final response safely

### Winner integration back into system state

When a race winner is selected, runtime now updates:

- key total cost
- channel stats success
- breaker success record
- sticky routing state
- attempts record (`race fallback winner`)

## Files changed

- `internal/model/group.go`
- `internal/op/group.go`
- `internal/db/migrate/008.go`
- `internal/relay/relay.go`
- `internal/relay/type.go`
- `internal/relay/balancer/iterator.go`
- `web/src/api/endpoints/group.ts`
- `web/src/components/modules/group/Editor.tsx`
- `web/src/components/modules/group/Create.tsx`
- `web/src/components/modules/group/Card.tsx`
- `web/public/locale/zh-Hans.json`
- `web/public/locale/en.json`
- `web/public/locale/zh-Hant.json`

## Runtime behavior after this task

### What works now

- multi-round retry with delay
- bounded failover window (default 360s)
- escalation after consecutive failures
- paid/metered targets blocked from racing by default
- `per_request` targets blocked from racing
- `per_token` targets only race when probe policy is concurrent and concurrency limit > 1
- non-stream race probing with safe main-goroutine response write-back

### What is intentionally NOT done yet

- full user-priority-based winner arbitration
- streaming race fallback
- complete route-target runtime policy matrix
- sophisticated bounded concurrent hedging policy

## Risks / follow-up

- racing still uses a phase-1 winner rule, not the final canonical-plan arbitration rule
- attempts for race candidates can still be enriched further
- `source_type` + `billing_mode` + `probe_policy` runtime combination is still incomplete

## Next step

Continue runtime hardening in two directions:

1. Separate `fill_priority` / `priority_order` semantics more clearly in key selection runtime
2. Continue improving failover racing toward canonical-plan arbitration behavior without breaking non-stream safety
