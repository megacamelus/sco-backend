package stdout

import (
	"context"
	"encoding/json"
	"log/slog"
)

// Stdout provide our basic publishing.
type Stdout struct {
	log *slog.Logger
}

// NewStdout initializes stdout for publishing metrics.
func NewStdout(log *slog.Logger) *Stdout {
	return &Stdout{log}
}

// Publish publishers for writing to stdout.
func (s *Stdout) Publish(data map[string]any) error {
	ctx := context.Background()

	rawJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	var d map[string]any
	if err := json.Unmarshal(rawJSON, &d); err != nil {
		return err
	}

	// Add heap value into the data set.
	memStats, ok := (d["memstats"]).(map[string]any)
	if ok {
		d["heap"] = memStats["Alloc"]
	}

	// Remove unnecessary keys.
	delete(d, "memstats")
	delete(d, "cmdline")

	out, err := json.MarshalIndent(d, "", "    ")
	if err != nil {
		return err
	}
	s.log.InfoContext(ctx, "stdout", "data", string(out))

	return nil
}
