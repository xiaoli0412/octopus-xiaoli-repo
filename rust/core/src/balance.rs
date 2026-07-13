use serde::{Deserialize, Serialize};
use serde_json::json;
use std::ffi::c_int;

#[derive(Debug, Clone, Default, Deserialize, Serialize)]
pub struct Candidate {
    pub id: i64,
    #[serde(default)]
    pub weight: i64,
    #[serde(default)]
    pub latency: i64,
    #[serde(default)]
    pub priority: i64,
    #[serde(default = "default_true")]
    pub healthy: bool,
    #[serde(default)]
    pub circuit_state: String,
}

#[repr(C)]
pub struct BalanceCandidateFFI {
    pub id: i64,
    pub weight: i64,
    pub latency: i64,
    pub priority: i64,
    pub healthy: i32,
    pub circuit_state: i32,
}

fn default_true() -> bool {
    true
}

impl Candidate {
    fn is_available(&self) -> bool {
        self.healthy && self.circuit_state != "open"
    }
}

impl BalanceCandidateFFI {
    fn is_available(&self) -> bool {
        self.healthy != 0 && self.circuit_state != 2
    }
}

/// Select a candidate according to the strategy.
/// Returns a JSON string `{"id": ..., "next_index": ...}`.
pub fn balance_select(candidates_json: &str, strategy: &str, current_idx: c_int) -> Result<String, c_int> {
    let candidates: Vec<Candidate> = serde_json::from_str(candidates_json).map_err(|_| -5)?;
    if candidates.is_empty() {
        return Err(-10);
    }

    let available: Vec<&Candidate> = candidates.iter().filter(|c| c.is_available()).collect();
    if available.is_empty() {
        return Err(-11);
    }

    let strategy = strategy.to_ascii_lowercase();
    let (selected_id, next_index) = match strategy.as_str() {
        "weighted" => select_weighted(&available, current_idx),
        "round_robin" => select_round_robin(&available, current_idx),
        "random" => select_random(&available),
        "failover" => select_failover(&available),
        "least_latency" => select_least_latency(&available),
        "health_aware" => select_health_aware(&available),
        _ => return Err(-8),
    };

    Ok(json!({"id": selected_id, "next_index": next_index}).to_string())
}

fn circuit_severity_rank(state: &str) -> i32 {
    match state {
        "open" => 2,
        "half-open" => 1,
        _ => 0,
    }
}

fn select_least_latency(candidates: &[&Candidate]) -> (i64, c_int) {
    let mut sorted: Vec<&Candidate> = candidates.to_vec();
    sorted.sort_by(|a, b| {
        a.latency
            .cmp(&b.latency)
            .then_with(|| a.id.cmp(&b.id))
    });
    (sorted[0].id, 0)
}

fn select_health_aware(candidates: &[&Candidate]) -> (i64, c_int) {
    let mut sorted: Vec<&Candidate> = candidates.to_vec();
    sorted.sort_by(|a, b| {
        circuit_severity_rank(&a.circuit_state)
            .cmp(&circuit_severity_rank(&b.circuit_state))
            .then_with(|| a.latency.cmp(&b.latency))
            .then_with(|| a.id.cmp(&b.id))
    });
    (sorted[0].id, 0)
}

fn select_weighted(candidates: &[&Candidate], current_idx: c_int) -> (i64, c_int) {
    let mut sequence: Vec<i64> = Vec::new();
    for c in candidates {
        let w = if c.weight <= 0 { 1 } else { c.weight };
        for _ in 0..w {
            sequence.push(c.id);
        }
    }
    if sequence.is_empty() {
        return (candidates[0].id, current_idx);
    }
    let idx = if current_idx < 0 {
        0
    } else {
        (current_idx as usize) % sequence.len()
    };
    (sequence[idx], (idx + 1) as c_int)
}

fn select_round_robin(candidates: &[&Candidate], current_idx: c_int) -> (i64, c_int) {
    let mut sorted: Vec<&Candidate> = candidates.to_vec();
    sorted.sort_by(|a, b| {
        a.priority
            .cmp(&b.priority)
            .then_with(|| a.latency.cmp(&b.latency))
            .then_with(|| a.id.cmp(&b.id))
    });
    let idx = if current_idx < 0 {
        0
    } else {
        (current_idx as usize) % sorted.len()
    };
    (sorted[idx].id, (idx + 1) as c_int)
}

fn select_random(candidates: &[&Candidate]) -> (i64, c_int) {
    use rand::seq::SliceRandom;
    let mut rng = rand::thread_rng();
    let c = candidates.choose(&mut rng).unwrap();
    (c.id, 0)
}

fn select_failover(candidates: &[&Candidate]) -> (i64, c_int) {
    let mut sorted: Vec<&Candidate> = candidates.to_vec();
    sorted.sort_by(|a, b| {
        a.priority
            .cmp(&b.priority)
            .then_with(|| a.latency.cmp(&b.latency))
            .then_with(|| a.id.cmp(&b.id))
    });
    (sorted[0].id, 0)
}

// =============================================================================
// Binary FFI variant
// =============================================================================

pub fn balance_select_v2(
    candidates: &[BalanceCandidateFFI],
    strategy: &str,
    current_idx: c_int,
) -> Result<(i64, c_int), c_int> {
    if candidates.is_empty() {
        return Err(-10);
    }
    let available: Vec<&BalanceCandidateFFI> =
        candidates.iter().filter(|c| c.is_available()).collect();
    if available.is_empty() {
        return Err(-11);
    }

    let strategy = strategy.to_ascii_lowercase();
    let (selected_id, next_index) = match strategy.as_str() {
        "weighted" => select_weighted_ffi(&available, current_idx),
        "round_robin" => select_round_robin_ffi(&available, current_idx),
        "random" => select_random_ffi(&available),
        "failover" => select_failover_ffi(&available),
        "least_latency" => select_least_latency_ffi(&available),
        "health_aware" => select_health_aware_ffi(&available),
        _ => return Err(-8),
    };
    Ok((selected_id, next_index))
}

fn select_weighted_ffi(candidates: &[&BalanceCandidateFFI], current_idx: c_int) -> (i64, c_int) {
    let mut sequence: Vec<i64> = Vec::new();
    for c in candidates {
        let w = if c.weight <= 0 { 1 } else { c.weight };
        for _ in 0..w {
            sequence.push(c.id);
        }
    }
    if sequence.is_empty() {
        return (candidates[0].id, current_idx);
    }
    let idx = if current_idx < 0 {
        0
    } else {
        (current_idx as usize) % sequence.len()
    };
    (sequence[idx], (idx + 1) as c_int)
}

fn select_round_robin_ffi(candidates: &[&BalanceCandidateFFI], current_idx: c_int) -> (i64, c_int) {
    let mut sorted: Vec<&BalanceCandidateFFI> = candidates.to_vec();
    sorted.sort_by(|a, b| {
        a.priority
            .cmp(&b.priority)
            .then_with(|| a.latency.cmp(&b.latency))
            .then_with(|| a.id.cmp(&b.id))
    });
    let idx = if current_idx < 0 {
        0
    } else {
        (current_idx as usize) % sorted.len()
    };
    (sorted[idx].id, (idx + 1) as c_int)
}

fn select_random_ffi(candidates: &[&BalanceCandidateFFI]) -> (i64, c_int) {
    use rand::seq::SliceRandom;
    let mut rng = rand::thread_rng();
    let c = candidates.choose(&mut rng).unwrap();
    (c.id, 0)
}

fn select_failover_ffi(candidates: &[&BalanceCandidateFFI]) -> (i64, c_int) {
    let mut sorted: Vec<&BalanceCandidateFFI> = candidates.to_vec();
    sorted.sort_by(|a, b| {
        a.priority
            .cmp(&b.priority)
            .then_with(|| a.latency.cmp(&b.latency))
            .then_with(|| a.id.cmp(&b.id))
    });
    (sorted[0].id, 0)
}

fn select_least_latency_ffi(candidates: &[&BalanceCandidateFFI]) -> (i64, c_int) {
    let mut sorted: Vec<&BalanceCandidateFFI> = candidates.to_vec();
    sorted.sort_by(|a, b| {
        a.latency
            .cmp(&b.latency)
            .then_with(|| a.id.cmp(&b.id))
    });
    (sorted[0].id, 0)
}

fn select_health_aware_ffi(candidates: &[&BalanceCandidateFFI]) -> (i64, c_int) {
    let mut sorted: Vec<&BalanceCandidateFFI> = candidates.to_vec();
    sorted.sort_by(|a, b| {
        a.circuit_state
            .cmp(&b.circuit_state)
            .then_with(|| a.latency.cmp(&b.latency))
            .then_with(|| a.id.cmp(&b.id))
    });
    (sorted[0].id, 0)
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::Value;

    fn make_candidates() -> Vec<Candidate> {
        vec![
            Candidate { id: 1, weight: 2, latency: 10, priority: 1, healthy: true, circuit_state: "closed".into() },
            Candidate { id: 2, weight: 1, latency: 20, priority: 2, healthy: true, circuit_state: "closed".into() },
            Candidate { id: 3, weight: 1, latency: 5, priority: 0, healthy: false, circuit_state: "open".into() },
        ]
    }

    #[test]
    fn weighted_cycles() {
        let cands = make_candidates();
        let json = serde_json::to_string(&cands).unwrap();
        let r0 = balance_select(&json, "weighted", 0).unwrap();
        let v0: Value = serde_json::from_str(&r0).unwrap();
        assert_eq!(v0["id"], 1);
        assert_eq!(v0["next_index"], 1);

        let r1 = balance_select(&json, "weighted", 1).unwrap();
        let v1: Value = serde_json::from_str(&r1).unwrap();
        assert_eq!(v1["id"], 1);

        let r2 = balance_select(&json, "weighted", 2).unwrap();
        let v2: Value = serde_json::from_str(&r2).unwrap();
        assert_eq!(v2["id"], 2);
    }

    #[test]
    fn round_robin_respects_priority() {
        let cands = make_candidates();
        let json = serde_json::to_string(&cands).unwrap();
        let r = balance_select(&json, "round_robin", 0).unwrap();
        let v: Value = serde_json::from_str(&r).unwrap();
        assert_eq!(v["id"], 1); // priority 1, latency 10
    }

    #[test]
    fn failover_picks_best() {
        let cands = make_candidates();
        let json = serde_json::to_string(&cands).unwrap();
        let r = balance_select(&json, "failover", 0).unwrap();
        let v: Value = serde_json::from_str(&r).unwrap();
        assert_eq!(v["id"], 1);
    }

    #[test]
    fn random_returns_available() {
        let cands = make_candidates();
        let json = serde_json::to_string(&cands).unwrap();
        let r = balance_select(&json, "random", 0).unwrap();
        let v: Value = serde_json::from_str(&r).unwrap();
        let id = v["id"].as_i64().unwrap();
        assert!(id == 1 || id == 2);
    }

    #[test]
    fn unhealthy_excluded() {
        let mut cands = make_candidates();
        cands[0].healthy = false;
        let json = serde_json::to_string(&cands).unwrap();
        let r = balance_select(&json, "failover", 0).unwrap();
        let v: Value = serde_json::from_str(&r).unwrap();
        assert_eq!(v["id"], 2);
    }
}
