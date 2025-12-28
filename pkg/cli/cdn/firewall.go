package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type FirewallRule struct {
	ID      int    `json:"id"`
	Type    string `json:"type"`
	Value   string `json:"value"`
	Action  string `json:"action"`
}

type FirewallConfigs struct {
	IPRules      []FirewallRule `json:"ip_rules"`
	CountryRules []FirewallRule `json:"country_rules"`
}

func NewAccessRulesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "access-rules",
		Aliases: []string{"acl", "ip-access"},
		Short:   "Manage IP/Country access rules",
		Long:    "Configure IP and country-based access rules for your domains.",
	}

	cmd.AddCommand(newFirewallStatusCmd())
	cmd.AddCommand(newFirewallAddIPCmd())
	cmd.AddCommand(newFirewallRemoveIPCmd())
	cmd.AddCommand(newFirewallAddCountryCmd())
	cmd.AddCommand(newFirewallRemoveCountryCmd())

	return cmd
}

func newFirewallStatusCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get firewall rules",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var configs FirewallConfigs
			if err := json.Unmarshal(resp.Data, &configs); err != nil {
				return fmt.Errorf("failed to parse configs: %w", err)
			}

			fmt.Printf("Firewall Rules\n")
			fmt.Printf("==============\n\n")

			fmt.Printf("IP Rules:\n")
			if len(configs.IPRules) == 0 {
				fmt.Println("  (none)")
			} else {
				fmt.Printf("  %-8s %-20s %-12s\n", "ID", "IP/CIDR", "ACTION")
				fmt.Printf("  %s\n", strings.Repeat("-", 45))
				for _, r := range configs.IPRules {
					fmt.Printf("  %-8d %-20s %-12s\n", r.ID, r.Value, r.Action)
				}
			}

			fmt.Printf("\nCountry Rules:\n")
			if len(configs.CountryRules) == 0 {
				fmt.Println("  (none)")
			} else {
				fmt.Printf("  %-8s %-10s %-12s\n", "ID", "COUNTRY", "ACTION")
				fmt.Printf("  %s\n", strings.Repeat("-", 35))
				for _, r := range configs.CountryRules {
					fmt.Printf("  %-8d %-10s %-12s\n", r.ID, r.Value, r.Action)
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newFirewallAddIPCmd() *cobra.Command {
	var domainID int
	var ip, action string

	cmd := &cobra.Command{
		Use:   "add-ip",
		Short: "Add IP rule",
		Long: `Add an IP-based firewall rule:
  Actions:
    - block:     Block requests from this IP
    - allow:     Allow requests from this IP (whitelist)
    - challenge: Show captcha challenge`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID), map[string]interface{}{
				"type":   "ip",
				"ip":     ip,
				"action": action,
			})
			if err != nil {
				return err
			}

			fmt.Printf("IP rule added: %s -> %s\n", ip, action)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address or CIDR range")
	cmd.Flags().StringVar(&action, "action", "block", "Action (block/allow/challenge)")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("ip")

	return cmd
}

func newFirewallRemoveIPCmd() *cobra.Command {
	var domainID int
	var ip string

	cmd := &cobra.Command{
		Use:   "remove-ip",
		Short: "Remove IP rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID), map[string]interface{}{
				"type":   "ip",
				"ip":     ip,
				"action": "remove",
			})
			if err != nil {
				return err
			}

			fmt.Printf("IP rule removed: %s\n", ip)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&ip, "ip", "", "IP address")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("ip")

	return cmd
}

func newFirewallAddCountryCmd() *cobra.Command {
	var domainID int
	var country, action string

	cmd := &cobra.Command{
		Use:   "add-country",
		Short: "Add country rule",
		Long: `Add a country-based firewall rule:
  Actions:
    - block:     Block requests from this country
    - allow:     Allow requests from this country (whitelist)
    - challenge: Show captcha challenge`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID), map[string]interface{}{
				"type":    "country",
				"country": country,
				"action":  action,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Country rule added: %s -> %s\n", country, action)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&country, "country", "", "Country code (e.g., US, DE, IR)")
	cmd.Flags().StringVar(&action, "action", "block", "Action (block/allow/challenge)")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("country")

	return cmd
}

func newFirewallRemoveCountryCmd() *cobra.Command {
	var domainID int
	var country string

	cmd := &cobra.Command{
		Use:   "remove-country",
		Short: "Remove country rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/firewall", domainID), map[string]interface{}{
				"type":    "country",
				"country": country,
				"action":  "remove",
			})
			if err != nil {
				return err
			}

			fmt.Printf("Country rule removed: %s\n", country)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&country, "country", "", "Country code")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("country")

	return cmd
}
