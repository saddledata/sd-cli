package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/saddledata/sd-cli/internal/api"
	"github.com/saddledata/sd-cli/internal/config"
	"github.com/spf13/cobra"
)

type DiscoveredField struct {
	Name string `json:"name"`
	Type string `json:"type"`
}

type ValidateDriftRequest struct {
	FlowID         string            `json:"flow_id"`
	ProposedSchema []DiscoveredField `json:"proposed_schema"`
	EntityName     string            `json:"entity_name"`
}

type DriftImpact struct {
	Type           string   `json:"type"`
	Message        string   `json:"message"`
	AffectedFields []string `json:"affected_fields"`
}

type ValidateDriftResponse struct {
	HasDrift         bool          `json:"has_drift"`
	DriftType        string        `json:"drift_type"`
	Impacts          []DriftImpact `json:"impacts"`
	SuggestedActions []string      `json:"suggested_actions"`
}

var (
	impactFlowID   string
	impactEntity   string
	impactSchema   string
)

var checkImpactCmd = &cobra.Command{
	Use:   "check-impact",
	Short: "Check the impact of a proposed schema change on a flow",
	Long:  `Simulates a schema change and returns an impact report showing potential breaking changes or evolutions.`,
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

		client := api.NewClient(ctx)

		// Read proposed schema from file
		schemaData, err := os.ReadFile(impactSchema)
		if err != nil {
			return fmt.Errorf("error reading schema file: %w", err)
		}

		var proposedSchema []DiscoveredField
		if err := json.Unmarshal(schemaData, &proposedSchema); err != nil {
			return fmt.Errorf("error parsing schema JSON: %w", err)
		}

		req := ValidateDriftRequest{
			FlowID:         impactFlowID,
			EntityName:     impactEntity,
			ProposedSchema: proposedSchema,
		}

		body, _ := json.Marshal(req)
		respBody, err := client.Post("/iac/validate-drift", body)
		if err != nil {
			return fmt.Errorf("impact check failed: %w", err)
		}

		var resp ValidateDriftResponse
		if err := json.Unmarshal(respBody, &resp); err != nil {
			return fmt.Errorf("error parsing response: %w", err)
		}

		printImpactReport(resp)
		return nil
	},
}

func printImpactReport(resp ValidateDriftResponse) {
	if !resp.HasDrift {
		fmt.Println("✅ No drift detected. Your changes are compatible with the current flow configuration.")
		return
	}

	fmt.Printf("⚠️  DRIFT DETECTED [%s]\n", resp.DriftType)
	fmt.Println("--------------------------------------------------------------------------------")
	
	for _, impact := range resp.Impacts {
		prefix := "  "
		if impact.Type == "breaking" {
			prefix = "  ❌ BREAKING: "
		} else {
			prefix = "  ℹ️  EVOLUTION: "
		}
		
		fmt.Printf("%s%s\n", prefix, impact.Message)
		if len(impact.AffectedFields) > 0 {
			fmt.Printf("     Fields: %v\n", impact.AffectedFields)
		}
		fmt.Println()
	}

	if len(resp.SuggestedActions) > 0 {
		fmt.Println("Suggested Actions:")
		for _, action := range resp.SuggestedActions {
			fmt.Printf("  - %s\n", action)
		}
	}
}

func init() {
	RootCmd.AddCommand(checkImpactCmd)

	checkImpactCmd.Flags().StringVar(&impactFlowID, "flow-id", "", "ID of the flow to check")
	checkImpactCmd.Flags().StringVar(&impactEntity, "entity", "", "Entity name (table) to check")
	checkImpactCmd.Flags().StringVar(&impactSchema, "schema-file", "", "Path to JSON file containing proposed schema")

	checkImpactCmd.MarkFlagRequired("flow-id")
	checkImpactCmd.MarkFlagRequired("entity")
	checkImpactCmd.MarkFlagRequired("schema-file")
}
