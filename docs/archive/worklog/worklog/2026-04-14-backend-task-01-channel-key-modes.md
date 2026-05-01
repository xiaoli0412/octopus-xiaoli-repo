# Backend Task 01 - Channel Key Modes and Routing Policy

## Goal

Complete the backend/data/configuration chain for channel-level key management modes and key routing policies, so `classified / pooled` and `round_robin / fill_priority / priority_order` are no longer just plan concepts, but real persisted and configurable system capabilities.

## Completed

### Data model

- Added `KeyManagementMode` with:
  - `classified`
  - `pooled`
- Added `KeyRoutingPolicy` with:
  - `round_robin`
  - `fill_priority`
  - `priority_order`
- Added `Channel.KeyManagementMode`
- Added `Channel.KeyRoutingPolicy`

### Database migration

- Added migration `internal/db/migrate/009.go`
  - `channels.key_management_mode`
  - `channels.key_routing_policy`

### Backend update flow

- `internal/op/channel.go` now persists updates for:
  - `key_management_mode`
  - `key_routing_policy`
- Removed duplicated update branches in `ChannelUpdate`

### Frontend/API configuration chain

- `web/src/api/endpoints/channel.ts`
  - `Channel` includes both fields
  - `CreateChannelRequest` includes both fields
  - `UpdateChannelRequest` includes both fields
- `web/src/components/modules/channel/Create.tsx`
  - default values set to `pooled` + `round_robin`
  - submit payload includes both fields
- `web/src/components/modules/channel/CardContent.tsx`
  - edit state includes both fields
  - diff/update payload includes both fields
- `web/src/components/modules/channel/Form.tsx`
  - form state supports both fields
  - visible Select controls added for both fields

### Runtime - phase 1

- `internal/model/channel.go`
  - `GetChannelKeyForModel()` now distinguishes:
    - `classified`: filter by `allowed_models`
    - `pooled`: shared key pool for the channel model set
  - `fill_priority` / `priority_order` phase-1 behavior:
    - pick the first available key in ordered key list
  - `round_robin` phase-1 behavior:
    - rotate on ordered eligible keys

## Files changed

- `internal/model/channel.go`
- `internal/op/channel.go`
- `internal/db/migrate/009.go`
- `web/src/api/endpoints/channel.ts`
- `web/src/components/modules/channel/Create.tsx`
- `web/src/components/modules/channel/CardContent.tsx`
- `web/src/components/modules/channel/Form.tsx`

## Risks / Not finished

- `priority_order` and `fill_priority` still share the same first-stage runtime behavior
- `pooled` shared model set semantics are still basic; future work should refine shared model selection and UI presentation
- left-panel `provider -> key -> model` grouped display is not implemented yet
- no feature flag yet for key management mode rollout

## Next step

Move to advanced failover runtime:

- reinforce 360-second failover window behavior
- continue bounded racing fallback
- connect `source_type`, `billing_mode`, and `probe_policy` more deeply into runtime decisions
