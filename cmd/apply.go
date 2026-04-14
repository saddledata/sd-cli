package cmd

import (
	"fmt"
	"os"

	"github.com/saddledata/sd-cli/internal/api"
	"github.com/saddledata/sd-cli/internal/config"
	"github.com/spf13/cobra"
)

var applyFile string

var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply a declarative configuration to your data stack",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		_, ctx, err := config.GetActiveContext(cfg, context)
		if err != nil {
			return err
		}

		if ctx.ApiKey == "" {
			return fmt.Errorf("not logged in. Use 'sd auth login' to authenticate")
		}

		data, err := os.ReadFile(applyFile)
		if err != nil {
			return fmt.Errorf("failed to read file: %w", err)
		}

		client := api.NewClient(ctx)
		fmt.Printf("Applying configuration from %s to %s...\n", applyFile, ctx.ApiUrl)

		resp, err := client.Post("/iac/apply", data)
		if err != nil {
			return err
		}

		fmt.Println("✓ Successfully applied configuration")
		fmt.Printf("%s\n", string(resp))
		return nil
	},
}

func init() {
	applyCmd.Flags().StringVarP(&applyFile, "file", "f", "", "The YAML configuration file to apply")
	applyCmd.MarkFlagRequired("file")

	RootCmd.AddCommand(applyCmd)
}
