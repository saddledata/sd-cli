package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/saddledata/sd-cli/internal/api"
	"github.com/saddledata/sd-cli/internal/config"
	"github.com/spf13/cobra"
)

type NullTime struct {
	Time  time.Time `json:"Time"`
	Valid bool      `json:"Valid"`
}

type NullString struct {
	String string `json:"String"`
	Valid  bool   `json:"Valid"`
}

type Flow struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	LastRunStatus  NullString `json:"last_run_status"`
	LastRunAt      NullTime   `json:"last_run_at"`
	Schedule       string     `json:"schedule"`
}

var (
	followLogs bool
)

// ... flowListCmd ...

var flowCmd = &cobra.Command{
	Use:   "flow",
	Short: "Manage and monitor data flows",
}

var flowListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all flows",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		_, ctx, err := config.GetActiveContext(cfg, context)
		if err != nil {
			return err
		}

		client := api.NewClient(ctx)
		resp, err := client.Get("/flows")
		if err != nil {
			return err
		}

		var flows []Flow
		if err := json.Unmarshal(resp, &flows); err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
		fmt.Fprintln(w, "ID\tNAME\tSTATUS\tLAST RUN STATUS\tLAST RUN AT\tSCHEDULE")
		for _, f := range flows {
			lastRunAt := "never"
			if f.LastRunAt.Valid {
				lastRunAt = f.LastRunAt.Time.Format("2006-01-02 15:04")
			}
			lastRunStatus := "-"
			if f.LastRunStatus.Valid {
				lastRunStatus = f.LastRunStatus.String
			}
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", f.ID, f.Name, f.Status, lastRunStatus, lastRunAt, f.Schedule)
		}
		w.Flush()
		return nil
	},
}

var flowLogsCmd = &cobra.Command{
	Use:   "logs [flow-id]",
	Short: "Show logs for the latest run of a flow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if followLogs {
			return runMonitor(args[0])
		}
		return printLatestLogs(args[0])
	},
}

var flowMonitorCmd = &cobra.Command{
	Use:   "monitor [flow-id]",
	Short: "Stream logs for a flow in real-time",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		return runMonitor(args[0])
	},
}

func printLatestLogs(flowID string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	_, ctx, err := config.GetActiveContext(cfg, context)
	if err != nil {
		return err
	}
	client := api.NewClient(ctx)

	resp, err := client.Get(fmt.Sprintf("/flows/%s/details", flowID))
	if err != nil {
		return err
	}

	var details struct {
		RunHistory []struct {
			ID uint `json:"id"`
		} `json:"run_history"`
	}
	if err := json.Unmarshal(resp, &details); err != nil {
		return err
	}

	if len(details.RunHistory) == 0 {
		fmt.Println("No run history found for this flow.")
		return nil
	}

	lastRunID := details.RunHistory[0].ID
	logResp, err := client.Get(fmt.Sprintf("/flows/%s/runs/%d/logs", flowID, lastRunID))
	if err != nil {
		return err
	}

	var logs struct {
		Logs string `json:"logs"`
	}
	if err := json.Unmarshal(logResp, &logs); err != nil {
		return err
	}

	fmt.Println(logs.Logs)
	return nil
}

func runMonitor(flowID string) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return err
	}
	_, ctx, err := config.GetActiveContext(cfg, context)
	if err != nil {
		return err
	}
	client := api.NewClient(ctx)

	fmt.Printf("Streaming logs for flow %s (Ctrl+C to stop)...\n", flowID)

	var lastPrintedSize int
	var currentRunID uint
	var currentRunFinished bool

	finishedStatuses := map[string]bool{
		"success":        true,
		"failed":         true,
		"cancelled":      true,
		"drift_detected": true,
	}

	for {
		// 1. Get flow details to find latest run ID and its status
		resp, err := client.Get(fmt.Sprintf("/flows/%s/details", flowID))
		if err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		var details struct {
			RunHistory []struct {
				ID     uint   `json:"id"`
				Status string `json:"status"`
			} `json:"run_history"`
		}
		if err := json.Unmarshal(resp, &details); err != nil {
			time.Sleep(5 * time.Second)
			continue
		}

		if len(details.RunHistory) == 0 {
			time.Sleep(5 * time.Second)
			continue
		}

		latestRun := details.RunHistory[0]

		// If it's a new run, reset trackers
		if latestRun.ID != currentRunID {
			if currentRunID != 0 {
				fmt.Printf("\n--- New Run Detected: %d ---\n", latestRun.ID)
			}
			currentRunID = latestRun.ID
			lastPrintedSize = 0
			currentRunFinished = false
		}

		// 2. Fetch and print logs
		if !currentRunFinished || lastPrintedSize == 0 {
			logResp, err := client.Get(fmt.Sprintf("/flows/%s/runs/%d/logs", flowID, currentRunID))
			if err == nil {
				var logs struct {
					Logs string `json:"logs"`
				}
				if err := json.Unmarshal(logResp, &logs); err == nil {
					if len(logs.Logs) > lastPrintedSize {
						fmt.Print(logs.Logs[lastPrintedSize:])
						lastPrintedSize = len(logs.Logs)
					}
				}
			}
		}

		// 3. Handle finish status
		if finishedStatuses[latestRun.Status] && !currentRunFinished {
			fmt.Printf("\n--- Run %d Finished: %s ---\n", currentRunID, latestRun.Status)
			currentRunFinished = true
			fmt.Println("Waiting for next run...")
		}

		time.Sleep(2 * time.Second)
	}
}

var flowRunCmd = &cobra.Command{
	Use:   "run [flow-id]",
	Short: "Trigger a manual sync of a flow",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		flowID := args[0]
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		_, ctx, err := config.GetActiveContext(cfg, context)
		if err != nil {
			return err
		}

		client := api.NewClient(ctx)
		_, err = client.Post(fmt.Sprintf("/flows/%s/run", flowID), []byte("{}"))
		if err != nil {
			return err
		}

		fmt.Printf("✓ Flow %s trigger successfully\n", flowID)
		return nil
	},
}

func init() {
	flowLogsCmd.Flags().BoolVarP(&followLogs, "follow", "f", false, "Stream logs in real-time")

	flowCmd.AddCommand(flowListCmd)
	flowCmd.AddCommand(flowLogsCmd)
	flowCmd.AddCommand(flowMonitorCmd)
	flowCmd.AddCommand(flowRunCmd)
	RootCmd.AddCommand(flowCmd)
}
