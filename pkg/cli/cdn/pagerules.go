package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type PageRulePath struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Priority int    `json:"priority"`
}

type PageRule struct {
	ID       int         `json:"id"`
	PathID   int         `json:"path_id"`
	Type     string      `json:"type"`
	Settings interface{} `json:"settings"`
}

func NewPageRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "page-rules",
		Aliases: []string{"rules", "paths"},
		Short:   "Manage page rules",
		Long:    "Configure path-based rules for caching, security, and more.",
	}

	cmd.AddCommand(newPageRulesListCmd())
	cmd.AddCommand(newPageRulesAddPathCmd())
	cmd.AddCommand(newPageRulesDeletePathCmd())
	cmd.AddCommand(newPageRulesSetCmd())
	cmd.AddCommand(newPageRulesDeleteRuleCmd())

	return cmd
}

func newPageRulesListCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool
	var ruleType string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List page rules",
		Long: `List page rules. Optionally filter by type:
  - all:       All rules (default)
  - waf:       WAF rules
  - ratelimit: Rate limit rules
  - ddos:      DDoS rules
  - firewall:  Firewall rules`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			endpoint := fmt.Sprintf("/v1/cdn/ng/domains/%d/paths", domainID)
			if ruleType != "" && ruleType != "all" {
				endpoint = fmt.Sprintf("/v1/cdn/ng/domains/%d/paths/%s", domainID, ruleType)
			}

			resp, err := client.Get(endpoint)
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var paths []PageRulePath
			if err := json.Unmarshal(resp.Data, &paths); err != nil {
				return fmt.Errorf("failed to parse paths: %w", err)
			}

			if len(paths) == 0 {
				fmt.Println("No page rules found")
				return nil
			}

			fmt.Printf("%-8s %-40s %-10s\n", "ID", "PATH", "PRIORITY")
			fmt.Println(strings.Repeat("-", 60))
			for _, p := range paths {
				fmt.Printf("%-8d %-40s %-10d\n", p.ID, truncate(p.Path, 40), p.Priority)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.Flags().StringVar(&ruleType, "type", "all", "Rule type (all/waf/ratelimit/ddos/firewall)")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newPageRulesAddPathCmd() *cobra.Command {
	var domainID int
	var path string
	var priority int

	cmd := &cobra.Command{
		Use:   "add-path",
		Short: "Add a page rule path",
		Long: `Add a new path for page rules. Use glob patterns:
  - /api/*      Matches /api/anything
  - /images/**  Matches /images/any/nested/path
  - *.js        Matches any .js file`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/paths", domainID), map[string]interface{}{
				"path":     path,
				"priority": priority,
			})
			if err != nil {
				return err
			}

			var result PageRulePath
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				fmt.Println("Path added successfully")
				return nil
			}

			fmt.Printf("Path added successfully!\n")
			fmt.Printf("ID: %d\n", result.ID)
			fmt.Printf("Path: %s\n", result.Path)
			fmt.Printf("Priority: %d\n", result.Priority)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&path, "path", "", "URL path pattern")
	cmd.Flags().IntVar(&priority, "priority", 10, "Rule priority (lower = higher priority)")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("path")

	return cmd
}

func newPageRulesDeletePathCmd() *cobra.Command {
	var domainID int
	var force bool

	cmd := &cobra.Command{
		Use:   "delete-path [path-id]",
		Short: "Delete a page rule path",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete path %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/paths/%s", domainID, args[0]))
			if err != nil {
				return err
			}

			fmt.Println("Path deleted successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newPageRulesSetCmd() *cobra.Command {
	var domainID int
	var pathID int
	var ruleType string
	var settings string

	cmd := &cobra.Command{
		Use:   "set-rule",
		Short: "Set a rule for a path",
		Long: `Set a rule for a specific path. Rule types:
  - cache:     Cache settings
  - waf:       WAF settings
  - ratelimit: Rate limiting
  - ddos:      DDoS protection
  - firewall:  Firewall rules

Settings should be JSON format.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var settingsMap map[string]interface{}
			if err := json.Unmarshal([]byte(settings), &settingsMap); err != nil {
				return fmt.Errorf("invalid settings JSON: %w", err)
			}

			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/paths/%d/rules", domainID, pathID), map[string]interface{}{
				"type":     ruleType,
				"settings": settingsMap,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Rule '%s' set for path %d successfully\n", ruleType, pathID)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&pathID, "path", 0, "Path ID")
	cmd.Flags().StringVar(&ruleType, "type", "", "Rule type (cache/waf/ratelimit/ddos/firewall)")
	cmd.Flags().StringVar(&settings, "settings", "{}", "Rule settings as JSON")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("type")

	return cmd
}

func newPageRulesDeleteRuleCmd() *cobra.Command {
	var domainID int
	var pathID int
	var ruleType string

	cmd := &cobra.Command{
		Use:   "delete-rule",
		Short: "Delete a rule from a path",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/paths/%d/rules/%s", domainID, pathID, ruleType))
			if err != nil {
				return err
			}

			fmt.Printf("Rule '%s' deleted from path %d\n", ruleType, pathID)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&pathID, "path", 0, "Path ID")
	cmd.Flags().StringVar(&ruleType, "type", "", "Rule type")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("path")
	cmd.MarkFlagRequired("type")

	return cmd
}
