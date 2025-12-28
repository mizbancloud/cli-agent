package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
	"github.com/mizbancloud/cli/pkg/types"
)

type LogForwarder struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Type        string            `json:"type"`
	Endpoint    string            `json:"endpoint"`
	Enabled     types.NumericBool `json:"enabled"`
	CreatedAt   string            `json:"created_at"`
}

func NewLogForwarderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "log-forwarder",
		Aliases: []string{"logs", "log-forwarders"},
		Short:   "Manage log forwarding",
		Long:    "Configure log forwarding to external services (Elasticsearch, S3, etc.).",
	}

	cmd.AddCommand(newLogForwarderListCmd())
	cmd.AddCommand(newLogForwarderAddCmd())
	cmd.AddCommand(newLogForwarderUpdateCmd())
	cmd.AddCommand(newLogForwarderDeleteCmd())

	return cmd
}

func newLogForwarderListCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List log forwarders",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/log-forwarders", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var forwarders []LogForwarder
			if err := json.Unmarshal(resp.Data, &forwarders); err != nil {
				return fmt.Errorf("failed to parse forwarders: %w", err)
			}

			if len(forwarders) == 0 {
				fmt.Println("No log forwarders configured")
				return nil
			}

			fmt.Printf("%-6s %-20s %-15s %-35s %-8s\n", "ID", "NAME", "TYPE", "ENDPOINT", "ENABLED")
			fmt.Println(strings.Repeat("-", 90))
			for _, f := range forwarders {
				enabled := "No"
				if f.Enabled.Bool() {
					enabled = "Yes"
				}
				fmt.Printf("%-6d %-20s %-15s %-35s %-8s\n",
					f.ID, truncate(f.Name, 20), f.Type, truncate(f.Endpoint, 35), enabled)
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newLogForwarderAddCmd() *cobra.Command {
	var domainID int
	var name, forwarderType, endpoint string
	var enabled bool
	var config string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a log forwarder",
		Long: `Add a new log forwarder. Supported types:
  - elasticsearch: Forward logs to Elasticsearch
  - s3:           Forward logs to S3-compatible storage
  - http:         Forward logs via HTTP webhook
  - datadog:      Forward logs to Datadog`,
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]interface{}{
				"name":     name,
				"type":     forwarderType,
				"endpoint": endpoint,
				"enabled":  enabled,
			}

			if config != "" {
				var configMap map[string]interface{}
				if err := json.Unmarshal([]byte(config), &configMap); err != nil {
					return fmt.Errorf("invalid config JSON: %w", err)
				}
				body["config"] = configMap
			}

			client := api.NewClient()
			resp, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/log-forwarders", domainID), body)
			if err != nil {
				return err
			}

			var forwarder LogForwarder
			if err := json.Unmarshal(resp.Data, &forwarder); err != nil {
				fmt.Println("Log forwarder added successfully")
				return nil
			}

			fmt.Printf("Log forwarder added successfully!\n")
			fmt.Printf("ID: %d\n", forwarder.ID)
			fmt.Printf("Name: %s\n", forwarder.Name)
			fmt.Printf("Type: %s\n", forwarder.Type)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&name, "name", "", "Forwarder name")
	cmd.Flags().StringVar(&forwarderType, "type", "", "Forwarder type (elasticsearch/s3/http/datadog)")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "Destination endpoint URL")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable forwarder")
	cmd.Flags().StringVar(&config, "config", "", "Additional config as JSON")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("type")
	cmd.MarkFlagRequired("endpoint")

	return cmd
}

func newLogForwarderUpdateCmd() *cobra.Command {
	var domainID, forwarderID int
	var name, endpoint string
	var enabled bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a log forwarder",
		RunE: func(cmd *cobra.Command, args []string) error {
			body := map[string]interface{}{}
			if name != "" {
				body["name"] = name
			}
			if endpoint != "" {
				body["endpoint"] = endpoint
			}
			body["enabled"] = enabled

			client := api.NewClient()
			_, err := client.Put(fmt.Sprintf("/v1/cdn/ng/domains/%d/log-forwarders/%d", domainID, forwarderID), body)
			if err != nil {
				return err
			}

			fmt.Println("Log forwarder updated successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&forwarderID, "forwarder", 0, "Forwarder ID")
	cmd.Flags().StringVar(&name, "name", "", "Forwarder name")
	cmd.Flags().StringVar(&endpoint, "endpoint", "", "Destination endpoint URL")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable forwarder")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("forwarder")

	return cmd
}

func newLogForwarderDeleteCmd() *cobra.Command {
	var domainID int
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [forwarder-id]",
		Short: "Delete a log forwarder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete forwarder %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/log-forwarders/%s", domainID, args[0]))
			if err != nil {
				return err
			}

			fmt.Println("Log forwarder deleted successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")
	cmd.MarkFlagRequired("domain")

	return cmd
}
