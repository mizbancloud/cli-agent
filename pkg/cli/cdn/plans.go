package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type CDNPlan struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Traffic     int64  `json:"traffic"`
	Price       int64  `json:"price"`
	Features    []string `json:"features"`
}

func NewPlansCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "plan",
		Aliases: []string{"plans"},
		Short:   "View CDN plans",
		Long:    "View available CDN plans.",
	}

	cmd.AddCommand(newPlansListCmd())

	return cmd
}

func newPlansListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available CDN plans",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cdn/ng/plans")
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var plans []CDNPlan
			if err := json.Unmarshal(resp.Data, &plans); err != nil {
				return fmt.Errorf("failed to parse plans: %w", err)
			}

			if len(plans) == 0 {
				fmt.Println("No plans available")
				return nil
			}

			fmt.Printf("%-6s %-15s %-20s %-15s %-15s\n", "ID", "NAME", "DISPLAY NAME", "TRAFFIC", "PRICE")
			fmt.Println(strings.Repeat("-", 75))
			for _, p := range plans {
				traffic := formatBytes(p.Traffic)
				price := fmt.Sprintf("%d Toman", p.Price)
				if p.Price == 0 {
					price = "Free"
				}
				fmt.Printf("%-6d %-15s %-20s %-15s %-15s\n",
					p.ID, p.Name, truncate(p.DisplayName, 20), traffic, price)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}
