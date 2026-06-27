package rustbridge

import "os"

const (
	envDisableRustTokenizer = "OCTOPUS_RUST_TOKENIZER"
	envDisableRustTransform = "OCTOPUS_RUST_TRANSFORM"
	envDisableRustBalancer  = "OCTOPUS_RUST_BALANCER"
	envDisableRustStream    = "OCTOPUS_RUST_STREAM"
	envDisableRustStats     = "OCTOPUS_RUST_STATS"
)

// envEnabled reports whether a Rust feature switch is enabled.
// The feature is enabled unless the env var is explicitly set to "0", "false",
// "FALSE" or "False".
func envEnabled(env string) bool {
	v := os.Getenv(env)
	return v != "0" && v != "false" && v != "FALSE" && v != "False"
}
