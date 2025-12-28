package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type Firewall struct {
	ID        int            `json:"id"`
	Name      string         `json:"name"`
	Rules     []FirewallRule `json:"rules"`
	Servers   []int          `json:"servers"`
	CreatedAt string         `json:"created_at"`
}

type FirewallRule struct {
	ID        int    `json:"id"`
	Direction string `json:"direction"`
	Protocol  string `json:"protocol"`
	PortMin   int    `json:"port_min"`
	PortMax   int    `json:"port_max"`
	RemoteIP  string `json:"remote_ip"`
}

func NewFirewallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "firewall",
		Aliases: []string{"fw", "sg"},
		Short:   "Manage firewalls (security groups)",
		Long:    "Create and manage firewall rules for your servers.",
	}

	cmd.AddCommand(newFirewallListCmd())
	cmd.AddCommand(newFirewallCreateCmd())
	cmd.AddCommand(newFirewallDeleteCmd())
	cmd.AddCommand(newFirewallRuleCmd())
	cmd.AddCommand(newFirewallAttachCmd())
	cmd.AddCommand(newFirewallDetachCmd())

	return cmd
}

func newFirewallListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all firewalls",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/firewall")
			if err != nil {
				return err
			}

			var firewalls []Firewall
			if err := json.Unmarshal(resp.Data, &firewalls); err != nil {
				return fmt.Errorf("failed to parse firewalls: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(firewalls, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(firewalls) == 0 {
				fmt.Println("No firewalls found")
				return nil
			}

			fmt.Printf("%-6s %-25s %-10s %-10s\n", "ID", "NAME", "RULES", "SERVERS")
			fmt.Println(strings.Repeat("-", 55))
			for _, f := range firewalls {
				fmt.Printf("%-6d %-25s %-10d %-10d\n", f.ID, truncate(f.Name, 25), len(f.Rules), len(f.Servers))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newFirewallCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new firewall",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			resp, err := client.Post("/v1/cloud/firewall", map[string]string{"name": name})
			if err != nil {
				return err
			}

			var firewall Firewall
			if err := json.Unmarshal(resp.Data, &firewall); err != nil {
				return fmt.Errorf("failed to parse firewall: %w", err)
			}

			fmt.Printf("Firewall created successfully!\n")
			fmt.Printf("ID: %d\n", firewall.ID)
			fmt.Printf("Name: %s\n", firewall.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Firewall name")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newFirewallDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [firewall-id]",
		Short: "Delete a firewall",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete firewall %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete("/v1/cloud/firewall/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("Firewall deleted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func newFirewallRuleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rule",
		Short: "Manage firewall rules",
	}

	cmd.AddCommand(newFirewallRuleAddCmd())
	cmd.AddCommand(newFirewallRuleDeleteCmd())

	return cmd
}

func newFirewallRuleAddCmd() *cobra.Command {
	var firewallID int
	var direction, protocol, remoteIP string
	var portMin, portMax int

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a firewall rule",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"firewall_id": firewallID,
				"direction":   direction,
				"protocol":    protocol,
				"port_min":    portMin,
				"port_max":    portMax,
				"remote_ip":   remoteIP,
			}

			_, err := client.Post("/v1/cloud/firewall/rule", body)
			if err != nil {
				return err
			}

			fmt.Println("Firewall rule added successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&firewallID, "firewall", 0, "Firewall ID")
	cmd.Flags().StringVar(&direction, "direction", "ingress", "Rule direction (ingress/egress)")
	cmd.Flags().StringVar(&protocol, "protocol", "tcp", "Protocol (tcp/udp/icmp)")
	cmd.Flags().IntVar(&portMin, "port-min", 0, "Minimum port")
	cmd.Flags().IntVar(&portMax, "port-max", 0, "Maximum port (default: same as port-min)")
	cmd.Flags().StringVar(&remoteIP, "remote-ip", "0.0.0.0/0", "Remote IP CIDR")

	cmd.MarkFlagRequired("firewall")
	cmd.MarkFlagRequired("port-min")

	return cmd
}

func newFirewallRuleDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [rule-id]",
		Short: "Delete a firewall rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Delete("/v1/cloud/firewall/rule/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("Firewall rule deleted successfully")
			return nil
		},
	}
}

func newFirewallAttachCmd() *cobra.Command {
	var serverID int

	cmd := &cobra.Command{
		Use:   "attach [firewall-id]",
		Short: "Attach firewall to server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/cloud/firewall/attach", map[string]interface{}{
				"firewall_id": args[0],
				"server_id":   serverID,
			})
			if err != nil {
				return err
			}

			fmt.Println("Firewall attached successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&serverID, "server", 0, "Server ID")
	cmd.MarkFlagRequired("server")

	return cmd
}

func newFirewallDetachCmd() *cobra.Command {
	var serverID int

	cmd := &cobra.Command{
		Use:   "detach [firewall-id]",
		Short: "Detach firewall from server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/cloud/firewall/detach", map[string]interface{}{
				"firewall_id": args[0],
				"server_id":   serverID,
			})
			if err != nil {
				return err
			}

			fmt.Println("Firewall detached successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&serverID, "server", 0, "Server ID")
	cmd.MarkFlagRequired("server")

	return cmd
}
