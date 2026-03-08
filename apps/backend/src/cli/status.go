// src/cli/status.go
package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/lucas/oraculo/apps/backend/src/db"
	"github.com/lucas/oraculo/apps/backend/src/output"
	"github.com/spf13/cobra"
)

func newStatusCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show project dashboard",
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()

			// Check whether a project exists.
			if _, err := os.Stat(".oraculo"); os.IsNotExist(err) {
				fmt.Fprintln(w, "No Oraculo project found. Run a tools command to get started.")
				return nil
			}

			database, err := db.Open()
			if err != nil {
				return err
			}
			defer database.Close()

			fmt.Fprintln(w, "=== Oraculo Status ===")

			// --- Epics table ---
			epics, err := db.NewEpicStore(database).List()
			if err != nil {
				return fmt.Errorf("list epics: %w", err)
			}

			fmt.Fprintln(w)
			fmt.Fprintln(w, "Epics")
			epicHeaders := []string{"Name", "Status"}
			var epicRows [][]string
			for _, e := range epics {
				epicRows = append(epicRows, []string{e.Name, string(e.ApprovalStatus)})
			}
			if len(epicRows) == 0 {
				fmt.Fprintln(w, "(none)")
			} else {
				if err := output.WriteTable(w, epicHeaders, epicRows); err != nil {
					return err
				}
			}

			// --- Task summary ---
			type statusCount struct {
				Status string
				Count  int
			}
			rows, err := database.Conn().Query(
				"SELECT status, COUNT(*) FROM tasks GROUP BY status ORDER BY status",
			)
			if err != nil {
				return fmt.Errorf("query task counts: %w", err)
			}
			defer rows.Close()

			counts := map[string]int{
				"pending":     0,
				"in_progress": 0,
				"completed":   0,
				"failed":      0,
			}
			for rows.Next() {
				var sc statusCount
				if err := rows.Scan(&sc.Status, &sc.Count); err != nil {
					return fmt.Errorf("scan task count: %w", err)
				}
				counts[sc.Status] = sc.Count
			}
			if err := rows.Err(); err != nil {
				return fmt.Errorf("iterate task counts: %w", err)
			}

			fmt.Fprintln(w)
			fmt.Fprintln(w, "Tasks")
			taskHeaders := []string{"Status", "Count"}
			taskRows := [][]string{
				{"pending", strconv.Itoa(counts["pending"])},
				{"in_progress", strconv.Itoa(counts["in_progress"])},
				{"completed", strconv.Itoa(counts["completed"])},
				{"failed", strconv.Itoa(counts["failed"])},
			}
			if err := output.WriteTable(w, taskHeaders, taskRows); err != nil {
				return err
			}

			// --- Pending approvals ---
			pendingApprovals, err := db.NewApprovalStore(database).List(true)
			if err != nil {
				return fmt.Errorf("list pending approvals: %w", err)
			}

			fmt.Fprintln(w)
			fmt.Fprintf(w, "Pending Approvals: %d\n", len(pendingApprovals))

			return nil
		},
	}
}
