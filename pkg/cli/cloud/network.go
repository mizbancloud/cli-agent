package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type PrivateNetwork struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	CIDR      string `json:"cidr"`
	Gateway   string `json:"gateway"`
	Servers   []int  `json:"servers"`
	CreatedAt string `json:"created_at"`
}

func NewNetworkCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "network",
		Aliases: []string{"net", "networks"},
		Short:   "Manage private networks",
		Long:    "Create and manage private networks for server connectivity.",
	}

	cmd.AddCommand(newNetworkListCmd())
	cmd.AddCommand(newNetworkCreateCmd())
	cmd.AddCommand(newNetworkDeleteCmd())
	cmd.AddCommand(newNetworkAttachCmd())
	cmd.AddCommand(newNetworkDetachCmd())

	return cmd
}

func newNetworkListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all private networks",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/private-networks")
			if err != nil {
				return err
			}

			var networks []PrivateNetwork
			if err := json.Unmarshal(resp.Data, &networks); err != nil {
				return fmt.Errorf("failed to parse networks: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(networks, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(networks) == 0 {
				fmt.Println("No private networks found")
				return nil
			}

			fmt.Printf("%-6s %-20s %-18s %-15s %-10s\n", "ID", "NAME", "CIDR", "GATEWAY", "SERVERS")
			fmt.Println(strings.Repeat("-", 75))
			for _, n := range networks {
				fmt.Printf("%-6d %-20s %-18s %-15s %-10d\n", n.ID, truncate(n.Name, 20), n.CIDR, n.Gateway, len(n.Servers))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newNetworkCreateCmd() *cobra.Command {
	var name, cidr string
	var datacenter int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new private network",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"name":          name,
				"cidr":          cidr,
				"datacenter_id": datacenter,
			}

			resp, err := client.Post("/v1/cloud/private-networks", body)
			if err != nil {
				return err
			}

			var network PrivateNetwork
			if err := json.Unmarshal(resp.Data, &network); err != nil {
				return fmt.Errorf("failed to parse network: %w", err)
			}

			fmt.Printf("Private network created successfully!\n")
			fmt.Printf("ID: %d\n", network.ID)
			fmt.Printf("Name: %s\n", network.Name)
			fmt.Printf("CIDR: %s\n", network.CIDR)
			fmt.Printf("Gateway: %s\n", network.Gateway)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Network name")
	cmd.Flags().StringVar(&cidr, "cidr", "10.0.0.0/24", "Network CIDR (e.g., 10.0.0.0/24)")
	cmd.Flags().IntVar(&datacenter, "datacenter", 1, "Datacenter ID")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newNetworkDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [network-id]",
		Short: "Delete a private network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete network %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete("/v1/cloud/private-networks/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("Private network deleted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func newNetworkAttachCmd() *cobra.Command {
	var serverID int
	var ip string

	cmd := &cobra.Command{
		Use:   "attach [network-id]",
		Short: "Attach server to private network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"network_id": args[0],
				"server_id":  serverID,
			}
			if ip != "" {
				body["ip"] = ip
			}

			_, err := client.Post("/v1/cloud/private-networks/attach", body)
			if err != nil {
				return err
			}

			fmt.Println("Server attached to network successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&serverID, "server", 0, "Server ID")
	cmd.Flags().StringVar(&ip, "ip", "", "Specific IP address (optional)")

	cmd.MarkFlagRequired("server")

	return cmd
}

func newNetworkDetachCmd() *cobra.Command {
	var serverID int

	cmd := &cobra.Command{
		Use:   "detach [network-id]",
		Short: "Detach server from private network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/cloud/private-networks/detach", map[string]interface{}{
				"network_id": args[0],
				"server_id":  serverID,
			})
			if err != nil {
				return err
			}

			fmt.Println("Server detached from network successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&serverID, "server", 0, "Server ID")
	cmd.MarkFlagRequired("server")

	return cmd
}
