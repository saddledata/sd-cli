package cmd

import (
	"fmt"
	"os"

	"github.com/saddledata/sd-cli/internal/api"
	"github.com/saddledata/sd-cli/internal/config"
	"github.com/spf13/cobra"
)

var assetID string
var rawFile string

var gatewayCmd = &cobra.Command{
	Use:   "gateway",
	Short: "Test AI validation and LLM extraction",
}

var gatewayValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a raw LLM payload against a data asset schema",
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

		payload, err := os.ReadFile(rawFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		// Prepare the request body for the gateway
		// We can reuse the same JSON structure as the frontend
		requestBody := fmt.Sprintf(`{"asset_id": "%s", "raw_output": %q}`, assetID, string(payload))

		fmt.Printf("Validating payload against asset %s...\n", assetID)
		resp, err := client.Post("/gateway/validate", []byte(requestBody))
		if err != nil {
			return err
		}

		fmt.Println("✓ Validation completed")
		fmt.Printf("%s\n", string(resp))
		return nil
	},
}

func init() {
	gatewayValidateCmd.Flags().StringVar(&assetID, "asset", "", "The ID of the Data Asset (locked schema)")
	gatewayValidateCmd.Flags().StringVar(&rawFile, "file", "", "Path to the raw LLM output text file")
	gatewayValidateCmd.MarkFlagRequired("asset")
	gatewayValidateCmd.MarkFlagRequired("file")

	gatewayCmd.AddCommand(gatewayValidateCmd)
	RootCmd.AddCommand(gatewayCmd)
}
