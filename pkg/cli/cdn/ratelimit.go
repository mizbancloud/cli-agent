package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
	"github.com/mizbancloud/cli/pkg/types"
)

type RateLimitSettings struct {
	DomainID       int               `json:"domain_id"`
	Enabled        types.NumericBool `json:"enabled"`
	Limit          int               `json:"limit"`
	Block          int               `json:"block"`
	AllowMethods   []string          `json:"allow_methods"`
	Whitelist      []string          `json:"whitelist"`
	AllowCountries []string          `json:"allow_countries"`
	CreatedAt      string            `json:"created_at"`
	UpdatedAt      string            `json:"updated_at"`
}

func NewRateLimitCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ratelimit",
		Aliases: []string{"rate-limit", "rl"},
		Short:   "Manage rate limiting",
		Long:    "Configure rate limiting settings for your domains.",
	}

	cmd.AddCommand(newRateLimitStatusCmd())
	cmd.AddCommand(newRateLimitSetCmd())
	cmd.AddCommand(newRateLimitEnableCmd())
	cmd.AddCommand(newRateLimitDisableCmd())

	return cmd
}

func newRateLimitStatusCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get rate limit status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/ratelimit", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var settings RateLimitSettings
			if err := json.Unmarshal(resp.Data, &settings); err != nil {
				return fmt.Errorf("failed to parse settings: %w", err)
			}

			fmt.Printf("Rate Limit Settings\n")
			fmt.Printf("===================\n")
			fmt.Printf("Enabled:           %v\n", settings.Enabled.Bool())
			fmt.Printf("Request Limit:     %d req/s\n", settings.Limit)
			fmt.Printf("Block Duration:    %d seconds\n", settings.Block)

			if len(settings.AllowMethods) > 0 {
				fmt.Printf("Allowed Methods:   %s\n", strings.Join(settings.AllowMethods, ", "))
			} else {
				fmt.Printf("Allowed Methods:   (all)\n")
			}

			if len(settings.Whitelist) > 0 {
				fmt.Printf("Whitelisted IPs:   %s\n", strings.Join(settings.Whitelist, ", "))
			} else {
				fmt.Printf("Whitelisted IPs:   (none)\n")
			}

			if len(settings.AllowCountries) > 0 {
				fmt.Printf("Allowed Countries: %s\n", strings.Join(settings.AllowCountries, ", "))
			} else {
				fmt.Printf("Allowed Countries: (all)\n")
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newRateLimitSetCmd() *cobra.Command {
	var domainID int
	var enabled bool
	var requestCount, blockTime int
	var methods, ips, countries []string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set rate limit configuration",
		Long: `Set rate limit configuration:
  --request-count: Max requests per second (1-1000)
  --block-time:    Block duration in seconds (1-1000)
  --methods:       Whitelisted HTTP methods (GET,POST,PUT,DELETE,etc.)
  --ips:           Whitelisted IP addresses (comma-separated)
  --countries:     Whitelisted country codes (comma-separated, e.g., US,DE)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"mode":          enabled,
				"request_count": requestCount,
				"block_time":    blockTime,
				"methods":       methods,
				"ips":           ips,
				"countries":     countries,
			}

			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/ratelimit", domainID), body)
			if err != nil {
				return err
			}

			enabledStr := "enabled"
			if !enabled {
				enabledStr = "disabled"
			}

			fmt.Printf("Rate limiting %s\n", enabledStr)
			fmt.Printf("Request limit: %d req/s\n", requestCount)
			fmt.Printf("Block duration: %d seconds\n", blockTime)

			if len(methods) > 0 {
				fmt.Printf("Whitelisted methods: %s\n", strings.Join(methods, ", "))
			}
			if len(ips) > 0 {
				fmt.Printf("Whitelisted IPs: %s\n", strings.Join(ips, ", "))
			}
			if len(countries) > 0 {
				fmt.Printf("Whitelisted countries: %s\n", strings.Join(countries, ", "))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable/disable rate limiting")
	cmd.Flags().IntVar(&requestCount, "request-count", 100, "Max requests per second (1-1000)")
	cmd.Flags().IntVar(&blockTime, "block-time", 60, "Block duration in seconds (1-1000)")
	cmd.Flags().StringSliceVar(&methods, "methods", []string{}, "Whitelisted HTTP methods")
	cmd.Flags().StringSliceVar(&ips, "ips", []string{}, "Whitelisted IP addresses")
	cmd.Flags().StringSliceVar(&countries, "countries", []string{}, "Whitelisted country codes")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newRateLimitEnableCmd() *cobra.Command {
	var domainID int
	var requestCount, blockTime int

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable rate limiting",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"mode":          true,
				"request_count": requestCount,
				"block_time":    blockTime,
				"methods":       []string{},
				"ips":           []string{},
				"countries":     []string{},
			}

			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/ratelimit", domainID), body)
			if err != nil {
				return err
			}

			fmt.Println("Rate limiting enabled")
			fmt.Printf("Request limit: %d req/s\n", requestCount)
			fmt.Printf("Block duration: %d seconds\n", blockTime)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&requestCount, "request-count", 100, "Max requests per second")
	cmd.Flags().IntVar(&blockTime, "block-time", 60, "Block duration in seconds")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newRateLimitDisableCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable rate limiting",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			// Get current settings to preserve them
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/ratelimit", domainID))
			if err != nil {
				return err
			}

			var settings RateLimitSettings
			if err := json.Unmarshal(resp.Data, &settings); err != nil {
				return fmt.Errorf("failed to parse settings: %w", err)
			}

			body := map[string]interface{}{
				"mode":          false,
				"request_count": settings.Limit,
				"block_time":    settings.Block,
				"methods":       settings.AllowMethods,
				"ips":           settings.Whitelist,
				"countries":     settings.AllowCountries,
			}

			_, err = client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/ratelimit", domainID), body)
			if err != nil {
				return err
			}

			fmt.Println("Rate limiting disabled")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}
