use serde::{Deserialize, Serialize};
use serde_json::json;
use std::ffi::c_int;

#[derive(Debug, Clone, Default, Deserialize, Serialize)]
struct StatsMetrics {
    #[serde(default)]
    input_token: i64,
    #[serde(default)]
    output_token: i64,
    #[serde(default)]
    input_cost: f64,
    #[serde(default)]
    output_cost: f64,
    #[serde(default)]
    wait_time: i64,
    #[serde(default)]
    request_success: i64,
    #[serde(default)]
    request_failed: i64,
}

impl StatsMetrics {
    fn merge(&mut self, delta: &StatsMetrics) {
        self.input_token += delta.input_token;
        self.output_token += delta.output_token;
        self.input_cost += delta.input_cost;
        self.output_cost += delta.output_cost;
        self.wait_time += delta.wait_time;
        self.request_success += delta.request_success;
        self.request_failed += delta.request_failed;
    }
}

/// Merge an existing aggregate StatsMetrics JSON with a delta JSON.
pub fn stats_merge_hourly(existing_json: &str, delta_json: &str) -> Result<String, c_int> {
    let mut existing: StatsMetrics = serde_json::from_str(existing_json).map_err(|_| -5)?;
    let delta: StatsMetrics = serde_json::from_str(delta_json).map_err(|_| -6)?;
    existing.merge(&delta);
    serde_json::to_string(&existing).map_err(|_| -7)
}

/// Compute a quantile for an array of f64 values.
pub fn stats_quantile(values_json: &str, quantile: f64) -> Result<String, c_int> {
    if !(0.0..=1.0).contains(&quantile) {
        return Err(-13);
    }
    let values: Vec<f64> = serde_json::from_str(values_json).map_err(|_| -5)?;
    if values.is_empty() {
        return Ok(json!({"quantile": quantile, "value": null}).to_string());
    }
    let mut sorted = values.clone();
    sorted.sort_by(|a, b| a.partial_cmp(b).unwrap_or(std::cmp::Ordering::Equal));

    let n = sorted.len();
    let idx_f = quantile * (n - 1) as f64;
    let lower = idx_f.floor() as usize;
    let upper = idx_f.ceil() as usize;
    let frac = idx_f - lower as f64;

    let value = if lower == upper {
        sorted[lower]
    } else {
        sorted[lower] * (1.0 - frac) + sorted[upper] * frac
    };

    Ok(json!({"quantile": quantile, "value": value}).to_string())
}

#[cfg(test)]
mod tests {
    use super::*;
    use serde_json::Value;

    #[test]
    fn merge_adds_fields() {
        let existing = json!({
            "input_token": 10,
            "output_token": 5,
            "input_cost": 1.0,
            "output_cost": 0.5,
            "wait_time": 100,
            "request_success": 1,
            "request_failed": 0
        });
        let delta = json!({
            "input_token": 3,
            "output_token": 2,
            "input_cost": 0.3,
            "output_cost": 0.2,
            "wait_time": 50,
            "request_success": 0,
            "request_failed": 1
        });
        let out = stats_merge_hourly(&existing.to_string(), &delta.to_string()).unwrap();
        let v: Value = serde_json::from_str(&out).unwrap();
        assert_eq!(v["input_token"], 13);
        assert_eq!(v["output_token"], 7);
        assert_eq!(v["request_failed"], 1);
    }

    #[test]
    fn quantile_median() {
        let values = json!([1.0, 3.0, 2.0, 4.0, 5.0]);
        let out = stats_quantile(&values.to_string(), 0.5).unwrap();
        let v: Value = serde_json::from_str(&out).unwrap();
        assert_eq!(v["value"], 3.0);
    }

    #[test]
    fn quantile_empty() {
        let values = json!([]);
        let out = stats_quantile(&values.to_string(), 0.5).unwrap();
        let v: Value = serde_json::from_str(&out).unwrap();
        assert!(v["value"].is_null());
    }
}
