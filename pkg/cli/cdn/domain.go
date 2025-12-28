package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
	"github.com/mizbancloud/cli/pkg/types"
)

type Domain struct {
	ID                 int                  `json:"id"`
	Name               string               `json:"name"`
	Domain             string               `json:"domain"`
	Status             string               `json:"status"`
	Plan               string               `json:"plan"`
	PlanDisplayName    string               `json:"plan_display_name"`
	WAFEnabled         types.NumericBool    `json:"waf-enabled"`
	DNSSECEnabled      types.NumericBool    `json:"dnssec_enabled"`
	H3Enabled          types.NumericBool    `json:"h3_enabled"`
	SupportsWebsocket  types.NumericBool    `json:"supports_websocket"`
	Nameservers        *Nameserver          `json:"nameservers,omitempty"`
	CurrentNameservers *CurrentNameserver   `json:"current_nameservers,omitempty"`
	AddedAt            string               `json:"added_at"`
	CreatedAt          string               `json:"created_at"`
	UpdatedAt          string               `json:"updated_at"`
}

type Nameserver struct {
	NS1 string               `json:"ns1"`
	NS2 string               `json:"ns2"`
	IP1 types.FlexibleString `json:"ip1"`
	IP2 types.FlexibleString `json:"ip2"`
}

type CurrentNameserver struct {
	NS1 string `json:"ns1"`
	NS2 string `json:"ns2"`
}

func NewDomainCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "domain",
		Aliases: []string{"domains"},
		Short:   "Manage CDN domains",
		Long:    "Add and manage domains in the CDN service.",
	}

	cmd.AddCommand(newDomainListCmd())
	cmd.AddCommand(newDomainAddCmd())
	cmd.AddCommand(newDomainGetCmd())
	cmd.AddCommand(newDomainDeleteCmd())
	cmd.AddCommand(newDomainUsageCmd())
	cmd.AddCommand(newDomainWhoisCmd())
	cmd.AddCommand(newDomainReportsCmd())
	cmd.AddCommand(newDomainRedirectModeCmd())

	return cmd
}

func newDomainListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all domains",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cdn/ng/domains")
			if err != nil {
				return err
			}

			var domains []Domain
			if err := json.Unmarshal(resp.Data, &domains); err != nil {
				return fmt.Errorf("failed to parse domains: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(domains, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(domains) == 0 {
				fmt.Println("No domains found")
				return nil
			}

			fmt.Printf("%-6s %-30s %-12s %-15s %-6s\n", "ID", "DOMAIN", "STATUS", "PLAN", "WAF")
			fmt.Println(strings.Repeat("-", 75))
			for _, d := range domains {
				waf := "No"
				if d.WAFEnabled.Bool() {
					waf = "Yes"
				}
				domainName := d.Name
				if domainName == "" {
					domainName = d.Domain
				}
				fmt.Printf("%-6d %-30s %-12s %-15s %-6s\n",
					d.ID, truncate(domainName, 30), d.Status, d.PlanDisplayName, waf)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newDomainAddCmd() *cobra.Command {
	var domain string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			resp, err := client.Post("/v1/cdn/ng/domains", map[string]string{"domain": domain})
			if err != nil {
				return err
			}

			var result Domain
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				return fmt.Errorf("failed to parse domain: %w", err)
			}

			fmt.Printf("Domain added successfully!\n")
			fmt.Printf("ID: %d\n", result.ID)
			fmt.Printf("Domain: %s\n", result.Name)
			fmt.Printf("Status: %s\n", result.Status)
			if result.Nameservers != nil {
				fmt.Println("\nNameservers (point your domain to these):")
				fmt.Printf("  - %s\n", result.Nameservers.NS1)
				fmt.Printf("  - %s\n", result.Nameservers.NS2)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&domain, "domain", "", "Domain name to add")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDomainGetCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get [domain-id]",
		Short: "Get domain details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cdn/ng/domains/" + args[0])
			if err != nil {
				return err
			}

			var domain Domain
			if err := json.Unmarshal(resp.Data, &domain); err != nil {
				return fmt.Errorf("failed to parse domain: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(domain, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			domainName := domain.Name
			if domainName == "" {
				domainName = domain.Domain
			}
			fmt.Printf("ID:          %d\n", domain.ID)
			fmt.Printf("Domain:      %s\n", domainName)
			fmt.Printf("Status:      %s\n", domain.Status)
			fmt.Printf("Plan:        %s (%s)\n", domain.Plan, domain.PlanDisplayName)
			fmt.Printf("WAF:         %v\n", domain.WAFEnabled.Bool())
			fmt.Printf("DNSSEC:      %v\n", domain.DNSSECEnabled.Bool())
			fmt.Printf("HTTP/3:      %v\n", domain.H3Enabled.Bool())
			fmt.Printf("WebSocket:   %v\n", domain.SupportsWebsocket.Bool())
			fmt.Printf("Added:       %s\n", domain.AddedAt)
			if domain.CurrentNameservers != nil {
				fmt.Println("Current Nameservers:")
				fmt.Printf("  - NS1: %s\n", domain.CurrentNameservers.NS1)
				fmt.Printf("  - NS2: %s\n", domain.CurrentNameservers.NS2)
			}
			if domain.Nameservers != nil {
				fmt.Println("Target Nameservers:")
				fmt.Printf("  - NS1: %s\n", domain.Nameservers.NS1)
				fmt.Printf("  - NS2: %s\n", domain.Nameservers.NS2)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newDomainDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [domain-id]",
		Short: "Delete a domain",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete domain %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete("/v1/cdn/ng/domains/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("Domain deleted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func newDomainUsageCmd() *cobra.Command {
	var period string

	cmd := &cobra.Command{
		Use:   "usage [domain-id]",
		Short: "Get domain traffic usage",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cdn/ng/domains/" + args[0] + "/usage")
			if err != nil {
				return err
			}

			var usage struct {
				Traffic   int64 `json:"traffic"`
				Requests  int64 `json:"requests"`
				Bandwidth int64 `json:"bandwidth"`
			}
			if err := json.Unmarshal(resp.Data, &usage); err != nil {
				return fmt.Errorf("failed to parse usage: %w", err)
			}

			fmt.Printf("Traffic:   %s\n", formatBytes(usage.Traffic))
			fmt.Printf("Requests:  %d\n", usage.Requests)
			fmt.Printf("Bandwidth: %s/s\n", formatBytes(usage.Bandwidth))

			return nil
		},
	}

	cmd.Flags().StringVar(&period, "period", "day", "Time period (hour/day/week/month)")

	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}

func formatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func newDomainWhoisCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "whois [domain-id]",
		Short: "Get domain WHOIS information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cdn/ng/domains/" + args[0] + "/whois")
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var whois struct {
				Registrar     string `json:"registrar"`
				CreationDate  string `json:"creation_date"`
				ExpiryDate    string `json:"expiry_date"`
				Nameservers   []string `json:"nameservers"`
				Status        string `json:"status"`
			}
			if err := json.Unmarshal(resp.Data, &whois); err != nil {
				// If parsing fails, just print raw data
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Printf("WHOIS Information\n")
			fmt.Printf("=================\n")
			fmt.Printf("Registrar:     %s\n", whois.Registrar)
			fmt.Printf("Created:       %s\n", whois.CreationDate)
			fmt.Printf("Expires:       %s\n", whois.ExpiryDate)
			fmt.Printf("Status:        %s\n", whois.Status)
			if len(whois.Nameservers) > 0 {
				fmt.Printf("Nameservers:\n")
				for _, ns := range whois.Nameservers {
					fmt.Printf("  - %s\n", ns)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newDomainReportsCmd() *cobra.Command {
	var domainID int
	var period string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "reports",
		Short: "Get domain traffic reports",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/reports", domainID), map[string]interface{}{
				"period": period,
			})
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var reports struct {
				TotalTraffic   int64 `json:"total_traffic"`
				TotalRequests  int64 `json:"total_requests"`
				CacheHitRatio  float64 `json:"cache_hit_ratio"`
				BandwidthPeak  int64 `json:"bandwidth_peak"`
			}
			if err := json.Unmarshal(resp.Data, &reports); err != nil {
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Printf("Domain Reports (%s)\n", period)
			fmt.Printf("====================\n")
			fmt.Printf("Total Traffic:   %s\n", formatBytes(reports.TotalTraffic))
			fmt.Printf("Total Requests:  %d\n", reports.TotalRequests)
			fmt.Printf("Cache Hit Ratio: %.2f%%\n", reports.CacheHitRatio*100)
			fmt.Printf("Bandwidth Peak:  %s/s\n", formatBytes(reports.BandwidthPeak))

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&period, "period", "day", "Time period (hour/day/week/month)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDomainRedirectModeCmd() *cobra.Command {
	var domainID int
	var mode string

	cmd := &cobra.Command{
		Use:   "redirect-mode",
		Short: "Set domain redirect mode",
		Long: `Set redirect mode for the domain:
  - none:  No redirect
  - www:   Redirect to www subdomain
  - naked: Redirect to naked domain (without www)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/redirect-mode", domainID), map[string]interface{}{
				"mode": mode,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Redirect mode set to: %s\n", mode)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&mode, "mode", "none", "Redirect mode (none/www/naked)")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("mode")

	return cmd
}
