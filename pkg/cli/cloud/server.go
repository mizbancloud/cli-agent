package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type Server struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	CPU          int    `json:"cpu"`
	RAM          int    `json:"ram"`
	Storage      int    `json:"storage"`
	OS           string `json:"os"`
	PublicIP     string `json:"public_ip"`
	PrivateIP    string `json:"private_ip"`
	DatacenterID int    `json:"datacenter_id"`
	CreatedAt    string `json:"created_at"`
}

func NewServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"srv", "servers"},
		Short:   "Manage cloud servers",
		Long:    "Create, manage, and monitor your cloud servers.",
	}

	cmd.AddCommand(newServerListCmd())
	cmd.AddCommand(newServerCreateCmd())
	cmd.AddCommand(newServerGetCmd())
	cmd.AddCommand(newServerDeleteCmd())
	cmd.AddCommand(newServerPowerCmd())
	cmd.AddCommand(newServerRenameCmd())
	cmd.AddCommand(newServerVNCCmd())
	cmd.AddCommand(newServerLogsCmd())
	cmd.AddCommand(newServerRebuildCmd())
	cmd.AddCommand(newServerReportsCmd())
	cmd.AddCommand(newServerRescueCmd())

	return cmd
}

func newServerListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all servers",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/servers")
			if err != nil {
				return err
			}

			var servers []Server
			if err := json.Unmarshal(resp.Data, &servers); err != nil {
				return fmt.Errorf("failed to parse servers: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(servers, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(servers) == 0 {
				fmt.Println("No servers found")
				return nil
			}

			fmt.Printf("%-6s %-20s %-12s %-6s %-8s %-18s %-12s\n",
				"ID", "NAME", "STATUS", "CPU", "RAM", "IP", "OS")
			fmt.Println(strings.Repeat("-", 90))
			for _, s := range servers {
				fmt.Printf("%-6d %-20s %-12s %-6d %-8d %-18s %-12s\n",
					s.ID, truncate(s.Name, 20), s.Status, s.CPU, s.RAM, s.PublicIP, truncate(s.OS, 12))
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newServerCreateCmd() *cobra.Command {
	var name, os string
	var cpu, ram, storage, datacenter int
	var sshKeyID int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new server",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"name":          name,
				"os":            os,
				"cpu":           cpu,
				"ram":           ram,
				"storage":       storage,
				"datacenter_id": datacenter,
			}
			if sshKeyID > 0 {
				body["ssh_key_id"] = sshKeyID
			}

			resp, err := client.Post("/v1/cloud/servers", body)
			if err != nil {
				return err
			}

			var server Server
			if err := json.Unmarshal(resp.Data, &server); err != nil {
				return fmt.Errorf("failed to parse server: %w", err)
			}

			fmt.Printf("Server created successfully!\n")
			fmt.Printf("ID: %d\n", server.ID)
			fmt.Printf("Name: %s\n", server.Name)
			fmt.Printf("Status: %s\n", server.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Server name")
	cmd.Flags().StringVar(&os, "os", "", "Operating system (e.g., ubuntu-22.04)")
	cmd.Flags().IntVar(&cpu, "cpu", 1, "Number of CPU cores")
	cmd.Flags().IntVar(&ram, "ram", 1024, "RAM in MB")
	cmd.Flags().IntVar(&storage, "storage", 20, "Storage in GB")
	cmd.Flags().IntVar(&datacenter, "datacenter", 1, "Datacenter ID")
	cmd.Flags().IntVar(&sshKeyID, "ssh-key", 0, "SSH key ID")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("os")

	return cmd
}

func newServerGetCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get [server-id]",
		Short: "Get server details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/servers/" + args[0])
			if err != nil {
				return err
			}

			var server Server
			if err := json.Unmarshal(resp.Data, &server); err != nil {
				return fmt.Errorf("failed to parse server: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(server, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			fmt.Printf("ID:         %d\n", server.ID)
			fmt.Printf("Name:       %s\n", server.Name)
			fmt.Printf("Status:     %s\n", server.Status)
			fmt.Printf("CPU:        %d cores\n", server.CPU)
			fmt.Printf("RAM:        %d MB\n", server.RAM)
			fmt.Printf("Storage:    %d GB\n", server.Storage)
			fmt.Printf("OS:         %s\n", server.OS)
			fmt.Printf("Public IP:  %s\n", server.PublicIP)
			fmt.Printf("Private IP: %s\n", server.PrivateIP)
			fmt.Printf("Created:    %s\n", server.CreatedAt)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newServerDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [server-id]",
		Short: "Delete a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete server %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete("/v1/cloud/servers/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("Server deleted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func newServerPowerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "power",
		Short: "Power management commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "on [server-id]",
		Short: "Power on server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put("/v1/cloud/servers/"+args[0]+"/power/on", nil)
			if err != nil {
				return err
			}
			fmt.Println("Server powering on...")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "off [server-id]",
		Short: "Power off server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put("/v1/cloud/servers/"+args[0]+"/power/off", nil)
			if err != nil {
				return err
			}
			fmt.Println("Server powering off...")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "reboot [server-id]",
		Short: "Reboot server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put("/v1/cloud/servers/"+args[0]+"/power/reboot", nil)
			if err != nil {
				return err
			}
			fmt.Println("Server rebooting...")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "restart [server-id]",
		Short: "Restart server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put("/v1/cloud/servers/"+args[0]+"/power/restart", nil)
			if err != nil {
				return err
			}
			fmt.Println("Server restarting...")
			return nil
		},
	})

	return cmd
}

func newServerRenameCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "rename [server-id]",
		Short: "Rename a server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/cloud/servers/"+args[0]+"/rename", map[string]string{
				"name": name,
			})
			if err != nil {
				return err
			}
			fmt.Printf("Server renamed to %s\n", name)
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "New server name")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newServerVNCCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "vnc [server-id]",
		Short: "Get VNC console URL",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/servers/" + args[0] + "/access/vnc")
			if err != nil {
				return err
			}

			var vnc struct {
				URL string `json:"url"`
			}
			if err := json.Unmarshal(resp.Data, &vnc); err != nil {
				return fmt.Errorf("failed to parse VNC info: %w", err)
			}

			fmt.Printf("VNC Console URL: %s\n", vnc.URL)
			return nil
		},
	}
}

func newServerLogsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logs [server-id]",
		Short: "Get server logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/servers/" + args[0] + "/logs")
			if err != nil {
				return err
			}

			var logs []struct {
				Action    string `json:"action"`
				Status    string `json:"status"`
				CreatedAt string `json:"created_at"`
			}
			if err := json.Unmarshal(resp.Data, &logs); err != nil {
				return fmt.Errorf("failed to parse logs: %w", err)
			}

			if len(logs) == 0 {
				fmt.Println("No logs found")
				return nil
			}

			fmt.Printf("%-20s %-15s %-25s\n", "ACTION", "STATUS", "DATE")
			fmt.Println(strings.Repeat("-", 60))
			for _, log := range logs {
				fmt.Printf("%-20s %-15s %-25s\n", log.Action, log.Status, log.CreatedAt)
			}

			return nil
		},
	}
}

func newServerReportsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "reports [server-id]",
		Short: "Get server performance reports",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/servers/" + args[0] + "/reports")
			if err != nil {
				return err
			}

			fmt.Println(string(resp.Data))
			return nil
		},
	}
}

func newServerRebuildCmd() *cobra.Command {
	var os string

	cmd := &cobra.Command{
		Use:   "rebuild [server-id]",
		Short: "Rebuild server with new OS",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put("/v1/cloud/servers/"+args[0]+"/rebuild/software", map[string]string{
				"os": os,
			})
			if err != nil {
				return err
			}

			fmt.Println("Server rebuild initiated...")
			return nil
		},
	}

	cmd.Flags().StringVar(&os, "os", "", "New operating system")
	cmd.MarkFlagRequired("os")

	return cmd
}

func newServerRescueCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rescue",
		Short: "Rescue mode commands",
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "enable [server-id]",
		Short: "Enable rescue mode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/cloud/servers/"+args[0]+"/rescue", nil)
			if err != nil {
				return err
			}
			fmt.Println("Rescue mode enabled")
			return nil
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "disable [server-id]",
		Short: "Disable rescue mode",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/cloud/servers/"+args[0]+"/unrescue", nil)
			if err != nil {
				return err
			}
			fmt.Println("Rescue mode disabled")
			return nil
		},
	})

	return cmd
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
