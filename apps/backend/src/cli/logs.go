// apps/backend/src/cli/logs.go
package cli

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/spf13/cobra"

	"github.com/lucas/oraculo/apps/backend/src/config"
)

func newLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs",
		Short: "Stream logs from the running Oraculo instance",
		RunE:  runLogs,
	}
}

func runLogs(cmd *cobra.Command, _ []string) error {
	cfg, err := config.Read()
	if err != nil {
		return err
	}

	port := cfg.Port
	if port == 0 {
		port = 3100
	}

	url := fmt.Sprintf("http://localhost:%d/logs", port)
	req, err := http.NewRequestWithContext(cmd.Context(), http.MethodGet, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "text/event-stream")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Errorf("connect to Oraculo instance at :%d: %w", port, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from /logs", resp.StatusCode)
	}

	w := cmd.OutOrStdout()
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		if out := parseLine(scanner.Text()); out != "" {
			fmt.Fprintln(w, out)
		}
	}
	return scanner.Err()
}

// parseLine extracts and formats a log entry from an SSE data line.
// Returns empty string for non-data lines.
func parseLine(line string) string {
	if !strings.HasPrefix(line, "data: ") {
		return ""
	}
	raw := strings.TrimPrefix(line, "data: ")
	var e struct {
		Time  string `json:"time"`
		Level string `json:"level"`
		Msg   string `json:"msg"`
	}
	if err := json.Unmarshal([]byte(raw), &e); err != nil {
		return raw // not JSON: print as-is
	}
	return fmt.Sprintf("%s %s %s", e.Time, e.Level, e.Msg)
}
