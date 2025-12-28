package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type WAFRule struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type WAFLayer struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

type WAFGroup struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Enabled bool   `json:"enabled"`
}

func NewWAFCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "waf",
		Short: "Manage Web Application Firewall",
		Long:  "Configure WAF rules and settings for your domains.",
	}

	cmd.AddCommand(newWAFStatusCmd())
	cmd.AddCommand(newWAFEnableCmd())
	cmd.AddCommand(newWAFDisableCmd())
	cmd.AddCommand(newWAFLayersCmd())
	cmd.AddCommand(newWAFRulesCmd())
	cmd.AddCommand(newWAFGroupsCmd())
	cmd.AddCommand(newWAFFirewallCmd())

	return cmd
}

func newWAFStatusCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get WAF status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/waf", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var status struct {
				Enabled bool   `json:"enabled"`
				Mode    string `json:"mode"`
			}
			if err := json.Unmarshal(resp.Data, &status); err != nil {
				return fmt.Errorf("failed to parse status: %w", err)
			}

			enabledStr := "Disabled"
			if status.Enabled {
				enabledStr = "Enabled"
			}

			fmt.Printf("WAF Status: %s\n", enabledStr)
			fmt.Printf("Mode:       %s\n", status.Mode)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newWAFEnableCmd() *cobra.Command {
	var domainID int
	var mode string

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable WAF",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put(fmt.Sprintf("/v1/cdn/ng/domains/%d/waf", domainID), map[string]interface{}{
				"enabled": true,
				"mode":    mode,
			})
			if err != nil {
				return err
			}

			fmt.Printf("WAF enabled (mode: %s)\n", mode)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&mode, "mode", "block", "WAF mode (block/simulate)")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newWAFDisableCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable WAF",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put(fmt.Sprintf("/v1/cdn/ng/domains/%d/waf", domainID), map[string]interface{}{
				"enabled": false,
			})
			if err != nil {
				return err
			}

			fmt.Println("WAF disabled")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newWAFLayersCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "layers",
		Short: "List WAF layers",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/waf/layers", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var layers []WAFLayer
			if err := json.Unmarshal(resp.Data, &layers); err != nil {
				return fmt.Errorf("failed to parse layers: %w", err)
			}

			if len(layers) == 0 {
				fmt.Println("No WAF layers found")
				return nil
			}

			fmt.Printf("%-20s %-30s %-10s\n", "ID", "NAME", "ENABLED")
			fmt.Println(strings.Repeat("-", 65))
			for _, l := range layers {
				enabled := "No"
				if l.Enabled {
					enabled = "Yes"
				}
				fmt.Printf("%-20s %-30s %-10s\n", l.ID, truncate(l.Name, 30), enabled)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newWAFRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rules",
		Short: "Manage WAF rules",
	}

	cmd.AddCommand(newWAFRulesListCmd())
	cmd.AddCommand(newWAFRulesDisabledCmd())
	cmd.AddCommand(newWAFRuleToggleCmd())

	return cmd
}

func newWAFRulesListCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List WAF rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/waf/rules", domainID))
			if err != nil {
				return err
			}

			var rules []WAFRule
			if err := json.Unmarshal(resp.Data, &rules); err != nil {
				return fmt.Errorf("failed to parse rules: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(rules, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(rules) == 0 {
				fmt.Println("No WAF rules found")
				return nil
			}

			fmt.Printf("%-20s %-30s %-10s\n", "ID", "NAME", "ENABLED")
			fmt.Println(strings.Repeat("-", 65))
			for _, r := range rules {
				enabled := "No"
				if r.Enabled {
					enabled = "Yes"
				}
				fmt.Printf("%-20s %-30s %-10s\n", r.ID, truncate(r.Name, 30), enabled)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newWAFRulesDisabledCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "disabled",
		Short: "List disabled WAF rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/waf/disabled-rules", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var rules []WAFRule
			if err := json.Unmarshal(resp.Data, &rules); err != nil {
				return fmt.Errorf("failed to parse rules: %w", err)
			}

			if len(rules) == 0 {
				fmt.Println("No disabled WAF rules")
				return nil
			}

			fmt.Printf("%-20s %-40s\n", "ID", "NAME")
			fmt.Println(strings.Repeat("-", 65))
			for _, r := range rules {
				fmt.Printf("%-20s %-40s\n", r.ID, truncate(r.Name, 40))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newWAFRuleToggleCmd() *cobra.Command {
	var domainID int
	var ruleID string
	var enabled bool

	cmd := &cobra.Command{
		Use:   "toggle",
		Short: "Enable/disable a WAF rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put(fmt.Sprintf("/v1/cdn/ng/domains/%d/waf/switch-rule", domainID), map[string]interface{}{
				"rule_id": ruleID,
				"enabled": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Printf("WAF rule %s enabled\n", ruleID)
			} else {
				fmt.Printf("WAF rule %s disabled\n", ruleID)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&ruleID, "rule", "", "Rule ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable/disable rule")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("rule")

	return cmd
}

func newWAFGroupsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "groups",
		Short: "Manage WAF rule groups",
	}

	cmd.AddCommand(newWAFGroupToggleCmd())

	return cmd
}

func newWAFGroupToggleCmd() *cobra.Command {
	var domainID int
	var groupID string
	var enabled bool

	cmd := &cobra.Command{
		Use:   "toggle",
		Short: "Enable/disable a WAF rule group",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put(fmt.Sprintf("/v1/cdn/ng/domains/%d/waf/switch-group", domainID), map[string]interface{}{
				"group_id": groupID,
				"enabled":  enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Printf("WAF group %s enabled\n", groupID)
			} else {
				fmt.Printf("WAF group %s disabled\n", groupID)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&groupID, "group", "", "Group ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable/disable group")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("group")

	return cmd
}

func newWAFFirewallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firewall",
		Short: "Manage IP/country firewall rules",
	}

	cmd.AddCommand(newWAFBlockIPCmd())
	cmd.AddCommand(newWAFUnblockIPCmd())
	cmd.AddCommand(newWAFBlockCountryCmd())
	cmd.AddCommand(newWAFUnblockCountryCmd())

	return cmd
}

func newWAFBlockIPCmd() *cobra.Command {
	var domainID int
	var ip, action string

	cmd := &cobra.Command{
		Use:   "block-ip",
		Short: "Block an IP address",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID), map[string]interface{}{
				"ip":     ip,
				"action": action,
			})
			if err != nil {
				return err
			}

			fmt.Printf("IP %s added with action: %s\n", ip, action)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address or CIDR")
	cmd.Flags().StringVar(&action, "action", "block", "Action (block/allow/challenge)")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("ip")

	return cmd
}

func newWAFUnblockIPCmd() *cobra.Command {
	var domainID int
	var ip string

	cmd := &cobra.Command{
		Use:   "unblock-ip",
		Short: "Remove IP from firewall",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID), map[string]interface{}{
				"ip":     ip,
				"action": "remove",
			})
			if err != nil {
				return err
			}

			fmt.Printf("IP %s removed from firewall\n", ip)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("ip")

	return cmd
}

func newWAFBlockCountryCmd() *cobra.Command {
	var domainID int
	var country string

	cmd := &cobra.Command{
		Use:   "block-country",
		Short: "Block a country",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID), map[string]interface{}{
				"country": country,
				"action":  "block",
			})
			if err != nil {
				return err
			}

			fmt.Printf("Country %s blocked\n", country)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&country, "country", "", "Country code (e.g., US, DE)")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("country")

	return cmd
}

func newWAFUnblockCountryCmd() *cobra.Command {
	var domainID int
	var country string

	cmd := &cobra.Command{
		Use:   "unblock-country",
		Short: "Unblock a country",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID), map[string]interface{}{
				"country": country,
				"action":  "remove",
			})
			if err != nil {
				return err
			}

			fmt.Printf("Country %s unblocked\n", country)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&country, "country", "", "Country code")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("country")

	return cmd
}

