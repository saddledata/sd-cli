package cmd

import (
	"fmt"
	"os"

	piihound "github.com/saddledata/pii-hound/cmd"
	"github.com/spf13/cobra"
)

var scanCmd = &cobra.Command{
	Use:   "scan [targets...]",
	Short: "Scan files or databases for PII (pii-hound integration)",
	Long: `Scan local files (JSON, CSV, Parquet, etc.) or databases (Postgres, MySQL, Snowflake)
for PII and sensitive data using the pii-hound engine.

Example:
  sd scan ./data/*.json
  sd scan postgres://user:pass@localhost:5432/dbname
`,
	Run: func(cmd *cobra.Command, args []string) {
		// We delegate to pii-hound's scan command logic.
		// Since we've imported pii-hound/cmd, we can try to reuse their Scan command.
		
		// Note: Directly calling piihound.ScanCmd.Run might be tricky because of flag initialization.
		// A cleaner way is to re-wrap their internal engine, but for now, let's try to 
		// just invoke their command if it's exported and accessible.
		
		// If we can't easily re-use the command, we'll import the engine and implement the scan logic here.
		// Let's check pii-hound/cmd/scan.go to see if it's exported.
		
		fmt.Println("Running PII scan...")
		
		// As a simple first pass, we can just say we are wrapping it.
		// In a real implementation, we'd either link the logic or execute the binary.
		// Since we have the code, let's try to link it.
		
		piihound.RootCmd.SetArgs(append([]string{"scan"}, args...))
		if err := piihound.RootCmd.Execute(); err != nil {
			fmt.Fprintf(os.Stderr, "Scan failed: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	// We want to inherit flags from pii-hound's scan command if possible.
	// For now, we'll just add it to Root.
	RootCmd.AddCommand(scanCmd)
}
