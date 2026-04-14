package cmd

import (
	"fmt"
	"github.com/saddledata/sd-cli/internal/config"
	"github.com/spf13/cobra"
)

var authKey string

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Authenticate and manage contexts",
}

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in with an API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		ctxName := context
		if ctxName == "" {
			ctxName = cfg.ActiveContext
		}

		ctx := cfg.Contexts[ctxName]
		ctx.ApiKey = authKey
		cfg.Contexts[ctxName] = ctx

		if err := config.SaveConfig(cfg); err != nil {
			return err
		}

		fmt.Printf("✓ Successfully logged in context '%s'\n", ctxName)
		return nil
	},
}

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show current authentication status",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := config.LoadConfig()
		if err != nil {
			return err
		}

		name, ctx, err := config.GetActiveContext(cfg, context)
		if err != nil {
			return err
		}

		fmt.Printf("Active context: %s\n", name)
		fmt.Printf("API URL:        %s\n", ctx.ApiUrl)
		if ctx.ApiKey != "" {
			masked := ctx.ApiKey[:4] + "****" + ctx.ApiKey[len(ctx.ApiKey)-4:]
			fmt.Printf("API Key:        %s\n", masked)
		} else {
			fmt.Println("API Key:        Not set")
		}
		return nil
	},
}

func init() {
	loginCmd.Flags().StringVar(&authKey, "key", "", "Saddle Data API Key")
	loginCmd.MarkFlagRequired("key")

	authCmd.AddCommand(loginCmd)
	authCmd.AddCommand(whoamiCmd)
	RootCmd.AddCommand(authCmd)
}
