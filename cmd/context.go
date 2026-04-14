package cmd

import (
	"fmt"
	"github.com/saddledata/sd-cli/internal/config"
	"github.com/spf13/cobra"
)

var contextCmd = &cobra.Command{
	Use:   "context",
	Short: "Manage CLI contexts",
}

var contextListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all contexts",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		fmt.Println("NAME        API URL")
		for name, ctx := range cfg.Contexts {
			activeMark := " "
			if name == cfg.ActiveContext {
				activeMark = "*"
			}
			fmt.Printf("%s %-10s %s\n", activeMark, name, ctx.ApiUrl)
		}
		return nil
	},
}

var contextUseCmd = &cobra.Command{
	Use:   "use [context]",
	Short: "Switch the active context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		name := args[0]
		if _, ok := cfg.Contexts[name]; !ok {
			return fmt.Errorf("context '%s' not found", name)
		}

		cfg.ActiveContext = name
		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Printf("✓ Switched to context '%s'\n", name)
		return nil
	},
}

var contextSetCmd = &cobra.Command{
	Use:   "set [context]",
	Short: "Set or update a context",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		name := args[0]
		apiUrl, _ := cmd.Flags().GetString("api-url")
		insecure, _ := cmd.Flags().GetBool("insecure")

		ctx := cfg.Contexts[name]
		if apiUrl != "" {
			ctx.ApiUrl = apiUrl
		}
		if cmd.Flags().Changed("insecure") {
			ctx.InsecureSkipVerify = insecure
		}
		cfg.Contexts[name] = ctx

		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Printf("✓ Context '%s' updated\n", name)
		return nil
	},
}

func init() {
	contextSetCmd.Flags().String("api-url", "", "The Base API URL for this context")
	contextSetCmd.Flags().Bool("insecure", false, "Skip TLS certificate verification (useful for local development)")

	contextCmd.AddCommand(contextListCmd)
	contextCmd.AddCommand(contextUseCmd)
	contextCmd.AddCommand(contextSetCmd)
	RootCmd.AddCommand(contextCmd)
}
