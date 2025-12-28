package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type SSLCertificate struct {
	ID        int    `json:"id"`
	Type      string `json:"type"`
	Status    string `json:"status"`
	ExpiresAt string `json:"expires_at"`
	Domains   []string `json:"domains"`
	CreatedAt string `json:"created_at"`
}

type SSLConfigs struct {
	TLSVersion       string `json:"tls_version"`
	HTTPSRedirect    bool   `json:"https_redirect"`
	HSTSEnabled      bool   `json:"hsts_enabled"`
	HSTSMaxAge       int    `json:"hsts_max_age"`
	HSTSSubdomains   bool   `json:"hsts_include_subdomains"`
	HSTSPreload      bool   `json:"hsts_preload"`
	BackendProtocol  string `json:"backend_protocol"`
	HTTP3Enabled     bool   `json:"h3_enabled"`
	CSPOverride      bool   `json:"csp_override"`
}

func NewSSLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ssl",
		Aliases: []string{"https", "certificate", "cert"},
		Short:   "Manage SSL certificates",
		Long:    "Request and manage SSL certificates for your domains.",
	}

	cmd.AddCommand(newSSLListCmd())
	cmd.AddCommand(newSSLStatusCmd())
	cmd.AddCommand(newSSLInfoCmd())
	cmd.AddCommand(newSSLRequestFreeCmd())
	cmd.AddCommand(newSSLAddCustomCmd())
	cmd.AddCommand(newSSLDeleteCmd())
	cmd.AddCommand(newSSLAttachCmd())
	cmd.AddCommand(newSSLDetachCmd())
	cmd.AddCommand(newSSLAttachDefaultCmd())
	cmd.AddCommand(newSSLDetachDefaultCmd())
	cmd.AddCommand(newSSLSettingsCmd())

	return cmd
}

func newSSLInfoCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Get SSL certificate info for domain",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/ssl/get-info", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var info struct {
				HasSSL      bool   `json:"has_ssl"`
				Issuer      string `json:"issuer"`
				ValidFrom   string `json:"valid_from"`
				ValidTo     string `json:"valid_to"`
				Domains     []string `json:"domains"`
				Fingerprint string `json:"fingerprint"`
			}
			if err := json.Unmarshal(resp.Data, &info); err != nil {
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Printf("SSL Certificate Info\n")
			fmt.Printf("====================\n")
			fmt.Printf("Has SSL:     %v\n", info.HasSSL)
			if info.HasSSL {
				fmt.Printf("Issuer:      %s\n", info.Issuer)
				fmt.Printf("Valid From:  %s\n", info.ValidFrom)
				fmt.Printf("Valid To:    %s\n", info.ValidTo)
				fmt.Printf("Fingerprint: %s\n", info.Fingerprint)
				if len(info.Domains) > 0 {
					fmt.Printf("Domains:     %s\n", strings.Join(info.Domains, ", "))
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

func newSSLAttachDefaultCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "attach-default",
		Short: "Attach default MizbanCloud SSL certificate",
		Long:  "Attach the default MizbanCloud shared SSL certificate to your domain.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/attach-default", domainID), nil)
			if err != nil {
				return err
			}

			fmt.Println("Default SSL certificate attached successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLDetachDefaultCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "detach-default",
		Short: "Detach default MizbanCloud SSL certificate",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/detach-default", domainID), nil)
			if err != nil {
				return err
			}

			fmt.Println("Default SSL certificate detached successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLStatusCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get SSL/HTTPS settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/ssl/get-configs", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var configs SSLConfigs
			if err := json.Unmarshal(resp.Data, &configs); err != nil {
				return fmt.Errorf("failed to parse configs: %w", err)
			}

			fmt.Printf("SSL/HTTPS Settings\n")
			fmt.Printf("==================\n")
			fmt.Printf("TLS Version:       %s\n", configs.TLSVersion)
			fmt.Printf("HTTPS Redirect:    %v\n", configs.HTTPSRedirect)
			fmt.Printf("Backend Protocol:  %s\n", configs.BackendProtocol)
			fmt.Printf("HTTP/3 (QUIC):     %v\n", configs.HTTP3Enabled)
			fmt.Printf("CSP Override:      %v\n", configs.CSPOverride)
			fmt.Printf("\nHSTS:\n")
			fmt.Printf("  Enabled:         %v\n", configs.HSTSEnabled)
			if configs.HSTSEnabled {
				fmt.Printf("  Max Age:         %d seconds\n", configs.HSTSMaxAge)
				fmt.Printf("  Subdomains:      %v\n", configs.HSTSSubdomains)
				fmt.Printf("  Preload:         %v\n", configs.HSTSPreload)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLAttachCmd() *cobra.Command {
	var domainID, certID int
	var recordIDs []int

	cmd := &cobra.Command{
		Use:   "attach",
		Short: "Attach SSL certificate to DNS records",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/attach", domainID), map[string]interface{}{
				"certificate_id": certID,
				"record_ids":     recordIDs,
			})
			if err != nil {
				return err
			}

			fmt.Println("SSL certificate attached successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&certID, "cert", 0, "Certificate ID")
	cmd.Flags().IntSliceVar(&recordIDs, "records", nil, "DNS record IDs to attach")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("cert")
	cmd.MarkFlagRequired("records")

	return cmd
}

func newSSLDetachCmd() *cobra.Command {
	var domainID int
	var recordIDs []int

	cmd := &cobra.Command{
		Use:   "detach",
		Short: "Detach SSL certificate from DNS records",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/detach", domainID), map[string]interface{}{
				"record_ids": recordIDs,
			})
			if err != nil {
				return err
			}

			fmt.Println("SSL certificate detached successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntSliceVar(&recordIDs, "records", nil, "DNS record IDs to detach")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("records")

	return cmd
}

func newSSLListCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List SSL certificates",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/ssl", domainID))
			if err != nil {
				return err
			}

			var certs []SSLCertificate
			if err := json.Unmarshal(resp.Data, &certs); err != nil {
				return fmt.Errorf("failed to parse certificates: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(certs, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(certs) == 0 {
				fmt.Println("No SSL certificates found")
				return nil
			}

			fmt.Printf("%-6s %-12s %-12s %-25s %-30s\n", "ID", "TYPE", "STATUS", "EXPIRES", "DOMAINS")
			fmt.Println(strings.Repeat("-", 90))
			for _, c := range certs {
				domains := strings.Join(c.Domains, ", ")
				fmt.Printf("%-6d %-12s %-12s %-25s %-30s\n",
					c.ID, c.Type, c.Status, c.ExpiresAt, truncate(domains, 30))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLRequestFreeCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "request-free",
		Short: "Request free Let's Encrypt SSL certificate",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/ssl/free", domainID), nil)
			if err != nil {
				return err
			}

			fmt.Println("SSL certificate request submitted successfully!")
			fmt.Println("The certificate will be issued within a few minutes.")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLAddCustomCmd() *cobra.Command {
	var domainID int
	var certificate, privateKey, chain string

	cmd := &cobra.Command{
		Use:   "add-custom",
		Short: "Add custom SSL certificate",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"certificate": certificate,
				"private_key": privateKey,
			}
			if chain != "" {
				body["chain"] = chain
			}

			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/ssl/add", domainID), body)
			if err != nil {
				return err
			}

			fmt.Println("Custom SSL certificate added successfully!")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&certificate, "cert", "", "Certificate PEM content")
	cmd.Flags().StringVar(&privateKey, "key", "", "Private key PEM content")
	cmd.Flags().StringVar(&chain, "chain", "", "Certificate chain PEM content (optional)")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("cert")
	cmd.MarkFlagRequired("key")

	return cmd
}

func newSSLDeleteCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "delete [cert-id]",
		Short: "Delete SSL certificate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/ssl/%s", domainID, args[0]))
			if err != nil {
				return err
			}

			fmt.Println("SSL certificate deleted successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage SSL settings",
	}

	cmd.AddCommand(newSSLTLSVersionCmd())
	cmd.AddCommand(newSSLHSTSCmd())
	cmd.AddCommand(newSSLRedirectCmd())
	cmd.AddCommand(newSSLBackendProtocolCmd())
	cmd.AddCommand(newSSLH3Cmd())
	cmd.AddCommand(newSSLCSPOverrideCmd())

	return cmd
}

func newSSLTLSVersionCmd() *cobra.Command {
	var domainID int
	var minVersion string

	cmd := &cobra.Command{
		Use:   "tls-version",
		Short: "Set minimum TLS version",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/ssl/tls-version", domainID), map[string]interface{}{
				"min_version": minVersion,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Minimum TLS version set to %s\n", minVersion)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&minVersion, "version", "1.2", "Minimum TLS version (1.0, 1.1, 1.2, 1.3)")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLHSTSCmd() *cobra.Command {
	var domainID int
	var enabled bool
	var maxAge int
	var includeSubdomains, preload bool

	cmd := &cobra.Command{
		Use:   "hsts",
		Short: "Configure HSTS settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/hsts", domainID), map[string]interface{}{
				"enabled":            enabled,
				"max_age":            maxAge,
				"include_subdomains": includeSubdomains,
				"preload":            preload,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("HSTS enabled successfully")
			} else {
				fmt.Println("HSTS disabled successfully")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable HSTS")
	cmd.Flags().IntVar(&maxAge, "max-age", 31536000, "Max age in seconds")
	cmd.Flags().BoolVar(&includeSubdomains, "include-subdomains", false, "Include subdomains")
	cmd.Flags().BoolVar(&preload, "preload", false, "Enable preload")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLRedirectCmd() *cobra.Command {
	var domainID int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "redirect",
		Short: "Enable/disable HTTPS redirect",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/redirect", domainID), map[string]interface{}{
				"enabled": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("HTTPS redirect enabled")
			} else {
				fmt.Println("HTTPS redirect disabled")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable HTTPS redirect")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLBackendProtocolCmd() *cobra.Command {
	var domainID int
	var protocol string

	cmd := &cobra.Command{
		Use:   "backend-protocol",
		Short: "Set default backend protocol",
		Long: `Set default protocol for connecting to origin:
  - http:  Connect to origin via HTTP
  - https: Connect to origin via HTTPS (default)
  - auto:  Auto-detect based on DNS record settings`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/backend-protocol", domainID), map[string]interface{}{
				"protocol": protocol,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Backend protocol set to: %s\n", protocol)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&protocol, "protocol", "https", "Protocol (http/https/auto)")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLH3Cmd() *cobra.Command {
	var domainID int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "h3",
		Short: "Enable/disable HTTP/3 (QUIC)",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/h3", domainID), map[string]interface{}{
				"enabled": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("HTTP/3 (QUIC) enabled")
			} else {
				fmt.Println("HTTP/3 (QUIC) disabled")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable HTTP/3")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newSSLCSPOverrideCmd() *cobra.Command {
	var domainID int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "csp-override",
		Short: "Enable/disable Content Security Policy override",
		Long:  "When enabled, CDN will modify CSP headers to allow CDN resources.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/https/csp-override", domainID), map[string]interface{}{
				"enabled": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("CSP override enabled")
			} else {
				fmt.Println("CSP override disabled")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable CSP override")
	cmd.MarkFlagRequired("domain")

	return cmd
}
