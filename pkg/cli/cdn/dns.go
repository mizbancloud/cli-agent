package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type DNSRecord struct {
	ID       int    `json:"id"`
	Type     string `json:"type"`
	Name     string `json:"name"`
	Content  string `json:"content"`
	TTL      int    `json:"ttl"`
	Priority int    `json:"priority,omitempty"`
	Port     int    `json:"port,omitempty"`
	Protocol string `json:"protocol,omitempty"`
	Proxy    string `json:"proxy"`
}

func NewDNSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage DNS records",
		Long:  "Create and manage DNS records for your domains.",
	}

	cmd.AddCommand(newDNSListCmd())
	cmd.AddCommand(newDNSGetCmd())
	cmd.AddCommand(newDNSAddCmd())
	cmd.AddCommand(newDNSUpdateCmd())
	cmd.AddCommand(newDNSDeleteCmd())
	cmd.AddCommand(newDNSProxiableCmd())
	cmd.AddCommand(newDNSImportCmd())
	cmd.AddCommand(newDNSExportCmd())
	cmd.AddCommand(newDNSFetchRecordsCmd())
	cmd.AddCommand(newDNSCustomNSCmd())
	cmd.AddCommand(newDNSDNSSECCmd())

	return cmd
}

func newDNSListCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns", domainID))
			if err != nil {
				return err
			}

			var records []DNSRecord
			if err := json.Unmarshal(resp.Data, &records); err != nil {
				return fmt.Errorf("failed to parse records: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(records, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(records) == 0 {
				fmt.Println("No DNS records found")
				return nil
			}

			fmt.Printf("%-6s %-8s %-25s %-40s %-8s %-10s %-8s\n", "ID", "TYPE", "NAME", "CONTENT", "TTL", "PROTOCOL", "PROXIED")
			fmt.Println(strings.Repeat("-", 115))
			for _, r := range records {
				proxied := "No"
				if r.Proxy == "ACTIVE" {
					proxied = "Yes"
				}
				// Show protocol with port if not default
				protocol := r.Protocol
				if protocol == "" || protocol == "DEFAULT" {
					protocol = "-"
				}
				if r.Port > 0 {
					protocol = fmt.Sprintf("%s:%d", protocol, r.Port)
				}
				fmt.Printf("%-6d %-8s %-25s %-40s %-8d %-10s %-8s\n",
					r.ID, r.Type, truncate(r.Name, 25), truncate(r.Content, 40), r.TTL, protocol, proxied)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSGetCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get [record-id]",
		Short: "Get a single DNS record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/%s", domainID, args[0]))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var record DNSRecord
			if err := json.Unmarshal(resp.Data, &record); err != nil {
				return fmt.Errorf("failed to parse record: %w", err)
			}

			fmt.Printf("DNS Record Details\n")
			fmt.Printf("==================\n")
			fmt.Printf("ID:       %d\n", record.ID)
			fmt.Printf("Type:     %s\n", record.Type)
			fmt.Printf("Name:     %s\n", record.Name)
			fmt.Printf("Content:  %s\n", record.Content)
			fmt.Printf("TTL:      %d\n", record.TTL)
			if record.Priority > 0 {
				fmt.Printf("Priority: %d\n", record.Priority)
			}
			if record.Port > 0 {
				fmt.Printf("Port:     %d\n", record.Port)
			}
			if record.Protocol != "" && record.Protocol != "DEFAULT" {
				fmt.Printf("Protocol: %s\n", record.Protocol)
			}
			fmt.Printf("Proxied:  %s\n", record.Proxy)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSProxiableCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "proxiable",
		Short: "List proxiable DNS records",
		Long:  "List DNS records that can be proxied through CDN (includes trashed records).",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/proxiable", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var records []DNSRecord
			if err := json.Unmarshal(resp.Data, &records); err != nil {
				return fmt.Errorf("failed to parse records: %w", err)
			}

			if len(records) == 0 {
				fmt.Println("No proxiable DNS records found")
				return nil
			}

			fmt.Printf("%-6s %-8s %-25s %-40s %-8s\n", "ID", "TYPE", "NAME", "CONTENT", "PROXIED")
			fmt.Println(strings.Repeat("-", 95))
			for _, r := range records {
				proxied := "No"
				if r.Proxy == "ACTIVE" {
					proxied = "Yes"
				}
				fmt.Printf("%-6d %-8s %-25s %-40s %-8s\n",
					r.ID, r.Type, truncate(r.Name, 25), truncate(r.Content, 40), proxied)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSAddCmd() *cobra.Command {
	var domainID, ttl, priority, port int
	var recordType, name, destination, protocol string
	var proxy bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a DNS record",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"type":        recordType,
				"name":        name,
				"destination": destination,
				"ttl":         ttl,
				"protocol":    protocol,
				"proxy":       proxy,
			}
			if priority > 0 {
				body["priority"] = priority
			}
			if port > 0 {
				body["port"] = port
			}

			resp, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns", domainID), body)
			if err != nil {
				return err
			}

			var record DNSRecord
			if err := json.Unmarshal(resp.Data, &record); err != nil {
				return fmt.Errorf("failed to parse record: %w", err)
			}

			fmt.Printf("DNS record added successfully!\n")
			fmt.Printf("ID: %d\n", record.ID)
			fmt.Printf("Type: %s\n", record.Type)
			fmt.Printf("Name: %s\n", record.Name)
			fmt.Printf("Content: %s\n", record.Content)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&recordType, "type", "", "Record type (A, AAAA, CNAME, MX, TXT, etc.)")
	cmd.Flags().StringVar(&name, "name", "", "Record name (@ for root)")
	cmd.Flags().StringVar(&destination, "destination", "", "Record destination/value")
	cmd.Flags().IntVar(&ttl, "ttl", 3600, "TTL in seconds")
	cmd.Flags().IntVar(&priority, "priority", 0, "Priority (for MX records)")
	cmd.Flags().IntVar(&port, "port", 0, "Port (for proxied records with custom port)")
	cmd.Flags().StringVar(&protocol, "protocol", "DEFAULT", "Protocol (DEFAULT/HTTPS/HTTP)")
	cmd.Flags().BoolVar(&proxy, "proxy", false, "Enable CDN proxy")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("destination")

	return cmd
}

func newDNSUpdateCmd() *cobra.Command {
	var domainID, recordID, ttl, priority, port int
	var recordType, name, destination, protocol string
	var proxy bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a DNS record",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"record_id":   recordID,
				"type":        recordType,
				"name":        name,
				"destination": destination,
				"ttl":         ttl,
				"protocol":    protocol,
				"proxy":       proxy,
			}
			if priority > 0 {
				body["priority"] = priority
			}
			if port > 0 {
				body["port"] = port
			}

			_, err := client.Put(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/%d", domainID, recordID), body)
			if err != nil {
				return err
			}

			fmt.Println("DNS record updated successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&recordID, "record", 0, "Record ID")
	cmd.Flags().StringVar(&recordType, "type", "", "Record type")
	cmd.Flags().StringVar(&name, "name", "", "Record name")
	cmd.Flags().StringVar(&destination, "destination", "", "Record destination/value")
	cmd.Flags().IntVar(&ttl, "ttl", 3600, "TTL in seconds")
	cmd.Flags().IntVar(&priority, "priority", 0, "Priority (for MX records)")
	cmd.Flags().IntVar(&port, "port", 0, "Port (for proxied records with custom port)")
	cmd.Flags().StringVar(&protocol, "protocol", "DEFAULT", "Protocol (DEFAULT/HTTPS/HTTP)")
	cmd.Flags().BoolVar(&proxy, "proxy", false, "Enable CDN proxy")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("record")

	return cmd
}

func newDNSDeleteCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "delete [record-id]",
		Short: "Delete a DNS record",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/%s", domainID, args[0]))
			if err != nil {
				return err
			}

			fmt.Println("DNS record deleted successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSImportCmd() *cobra.Command {
	var domainID int
	var zone string

	cmd := &cobra.Command{
		Use:   "import",
		Short: "Import DNS zone file",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/import", domainID), map[string]interface{}{
				"zone": zone,
			})
			if err != nil {
				return err
			}

			fmt.Println("DNS zone imported successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&zone, "zone", "", "Zone file content")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("zone")

	return cmd
}

func newDNSExportCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export DNS zone file",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/export", domainID))
			if err != nil {
				return err
			}

			var result struct {
				Zone string `json:"zone"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				return fmt.Errorf("failed to parse zone: %w", err)
			}

			fmt.Println(result.Zone)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSFetchRecordsCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "fetch-records",
		Short: "Fetch DNS records from authoritative nameservers",
		Long:  "Automatically discover and import DNS records from the current authoritative nameservers.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/fetch-records", domainID), nil)
			if err != nil {
				return err
			}

			var result struct {
				Records []DNSRecord `json:"records"`
				Count   int         `json:"count"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				fmt.Println("DNS records fetched successfully")
				return nil
			}

			fmt.Printf("Fetched %d DNS records from authoritative nameservers\n", result.Count)
			if len(result.Records) > 0 {
				fmt.Printf("\n%-6s %-8s %-25s %-40s\n", "ID", "TYPE", "NAME", "CONTENT")
				fmt.Println(strings.Repeat("-", 85))
				for _, r := range result.Records {
					fmt.Printf("%-6d %-8s %-25s %-40s\n",
						r.ID, r.Type, truncate(r.Name, 25), truncate(r.Content, 40))
				}
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSCustomNSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom-ns",
		Short: "Manage custom nameservers",
		Long:  "Configure custom nameservers (vanity nameservers) for your domain.",
	}

	cmd.AddCommand(newDNSCustomNSGetCmd())
	cmd.AddCommand(newDNSCustomNSSetCmd())
	cmd.AddCommand(newDNSCustomNSDeleteCmd())

	return cmd
}

func newDNSCustomNSGetCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get custom nameserver configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/custom-ns", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var ns struct {
				NS1     string `json:"ns1"`
				NS2     string `json:"ns2"`
				Enabled bool   `json:"enabled"`
			}
			if err := json.Unmarshal(resp.Data, &ns); err != nil {
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Printf("Custom Nameservers\n")
			fmt.Printf("==================\n")
			fmt.Printf("Enabled: %v\n", ns.Enabled)
			if ns.NS1 != "" {
				fmt.Printf("NS1:     %s\n", ns.NS1)
			}
			if ns.NS2 != "" {
				fmt.Printf("NS2:     %s\n", ns.NS2)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSCustomNSSetCmd() *cobra.Command {
	var domainID int
	var ns1, ns2 string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set custom nameservers",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/custom-ns", domainID), map[string]interface{}{
				"ns1": ns1,
				"ns2": ns2,
			})
			if err != nil {
				return err
			}

			fmt.Println("Custom nameservers configured successfully")
			fmt.Printf("NS1: %s\n", ns1)
			fmt.Printf("NS2: %s\n", ns2)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&ns1, "ns1", "", "Primary nameserver")
	cmd.Flags().StringVar(&ns2, "ns2", "", "Secondary nameserver")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("ns1")
	cmd.MarkFlagRequired("ns2")

	return cmd
}

func newDNSCustomNSDeleteCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove custom nameservers",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/custom-ns", domainID))
			if err != nil {
				return err
			}

			fmt.Println("Custom nameservers removed successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSDNSSECCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dnssec",
		Short: "Manage DNSSEC settings",
		Long:  "Enable or disable DNSSEC (DNS Security Extensions) for your domain.",
	}

	cmd.AddCommand(newDNSSECStatusCmd())
	cmd.AddCommand(newDNSSECEnableCmd())
	cmd.AddCommand(newDNSSECDisableCmd())

	return cmd
}

func newDNSSECStatusCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get DNSSEC status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/dnssec", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var dnssec struct {
				Enabled   bool   `json:"enabled"`
				Algorithm string `json:"algorithm"`
				DS        string `json:"ds"`
				KeyTag    int    `json:"key_tag"`
				DigestType string `json:"digest_type"`
				Digest    string `json:"digest"`
			}
			if err := json.Unmarshal(resp.Data, &dnssec); err != nil {
				fmt.Println(string(resp.Data))
				return nil
			}

			fmt.Printf("DNSSEC Configuration\n")
			fmt.Printf("====================\n")
			fmt.Printf("Enabled:     %v\n", dnssec.Enabled)
			if dnssec.Enabled {
				fmt.Printf("Algorithm:   %s\n", dnssec.Algorithm)
				fmt.Printf("Key Tag:     %d\n", dnssec.KeyTag)
				fmt.Printf("Digest Type: %s\n", dnssec.DigestType)
				if dnssec.DS != "" {
					fmt.Printf("\nDS Record:\n%s\n", dnssec.DS)
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

func newDNSSECEnableCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "enable",
		Short: "Enable DNSSEC",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/dnssec", domainID), map[string]interface{}{
				"enabled": true,
			})
			if err != nil {
				return err
			}

			var dnssec struct {
				DS string `json:"ds"`
			}
			if err := json.Unmarshal(resp.Data, &dnssec); err == nil && dnssec.DS != "" {
				fmt.Println("DNSSEC enabled successfully!")
				fmt.Println("\nAdd this DS record to your registrar:")
				fmt.Println(dnssec.DS)
			} else {
				fmt.Println("DNSSEC enabled successfully")
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDNSSECDisableCmd() *cobra.Command {
	var domainID int

	cmd := &cobra.Command{
		Use:   "disable",
		Short: "Disable DNSSEC",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/dns/dnssec", domainID), map[string]interface{}{
				"enabled": false,
			})
			if err != nil {
				return err
			}

			fmt.Println("DNSSEC disabled successfully")
			fmt.Println("Remember to remove the DS record from your registrar")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.MarkFlagRequired("domain")

	return cmd
}
