package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type Volume struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Status    string `json:"status"`
	ServerID  int    `json:"server_id"`
	CreatedAt string `json:"created_at"`
}

func NewVolumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "volume",
		Aliases: []string{"vol", "volumes"},
		Short:   "Manage volumes",
		Long:    "Create and manage block storage volumes.",
	}

	cmd.AddCommand(newVolumeListCmd())
	cmd.AddCommand(newVolumeCreateCmd())
	cmd.AddCommand(newVolumeGetCmd())
	cmd.AddCommand(newVolumeDeleteCmd())
	cmd.AddCommand(newVolumeAttachCmd())
	cmd.AddCommand(newVolumeDetachCmd())
	cmd.AddCommand(newVolumeResizeCmd())

	return cmd
}

func newVolumeListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all volumes",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/volumes")
			if err != nil {
				return err
			}

			var volumes []Volume
			if err := json.Unmarshal(resp.Data, &volumes); err != nil {
				return fmt.Errorf("failed to parse volumes: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(volumes, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(volumes) == 0 {
				fmt.Println("No volumes found")
				return nil
			}

			fmt.Printf("%-6s %-25s %-10s %-12s %-10s\n", "ID", "NAME", "SIZE(GB)", "STATUS", "SERVER")
			fmt.Println(strings.Repeat("-", 70))
			for _, v := range volumes {
				serverStr := "-"
				if v.ServerID > 0 {
					serverStr = fmt.Sprintf("%d", v.ServerID)
				}
				fmt.Printf("%-6d %-25s %-10d %-12s %-10s\n", v.ID, truncate(v.Name, 25), v.Size, v.Status, serverStr)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newVolumeCreateCmd() *cobra.Command {
	var name string
	var size, datacenter int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new volume",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"name":          name,
				"size":          size,
				"datacenter_id": datacenter,
			}

			resp, err := client.Post("/v1/cloud/volumes", body)
			if err != nil {
				return err
			}

			var volume Volume
			if err := json.Unmarshal(resp.Data, &volume); err != nil {
				return fmt.Errorf("failed to parse volume: %w", err)
			}

			fmt.Printf("Volume created successfully!\n")
			fmt.Printf("ID: %d\n", volume.ID)
			fmt.Printf("Name: %s\n", volume.Name)
			fmt.Printf("Size: %d GB\n", volume.Size)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Volume name")
	cmd.Flags().IntVar(&size, "size", 10, "Volume size in GB")
	cmd.Flags().IntVar(&datacenter, "datacenter", 1, "Datacenter ID")

	cmd.MarkFlagRequired("name")

	return cmd
}

func newVolumeGetCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get [volume-id]",
		Short: "Get volume details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/volumes/" + args[0])
			if err != nil {
				return err
			}

			var volume Volume
			if err := json.Unmarshal(resp.Data, &volume); err != nil {
				return fmt.Errorf("failed to parse volume: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(volume, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			fmt.Printf("ID:        %d\n", volume.ID)
			fmt.Printf("Name:      %s\n", volume.Name)
			fmt.Printf("Size:      %d GB\n", volume.Size)
			fmt.Printf("Status:    %s\n", volume.Status)
			fmt.Printf("Server ID: %d\n", volume.ServerID)
			fmt.Printf("Created:   %s\n", volume.CreatedAt)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newVolumeDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [volume-id]",
		Short: "Delete a volume",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete volume %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete("/v1/cloud/volumes/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("Volume deleted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func newVolumeAttachCmd() *cobra.Command {
	var serverID int

	cmd := &cobra.Command{
		Use:   "attach [volume-id]",
		Short: "Attach volume to server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/cloud/volumes/attach", map[string]interface{}{
				"volume_id": args[0],
				"server_id": serverID,
			})
			if err != nil {
				return err
			}

			fmt.Println("Volume attached successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&serverID, "server", 0, "Server ID to attach to")
	cmd.MarkFlagRequired("server")

	return cmd
}

func newVolumeDetachCmd() *cobra.Command {
	var serverID int

	cmd := &cobra.Command{
		Use:   "detach [volume-id]",
		Short: "Detach volume from server",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/cloud/volumes/detach", map[string]interface{}{
				"volume_id": args[0],
				"server_id": serverID,
			})
			if err != nil {
				return err
			}

			fmt.Println("Volume detached successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&serverID, "server", 0, "Server ID to detach from")
	cmd.MarkFlagRequired("server")

	return cmd
}

func newVolumeResizeCmd() *cobra.Command {
	var size int

	cmd := &cobra.Command{
		Use:   "resize [volume-id]",
		Short: "Resize volume",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Put("/v1/cloud/volumes/"+args[0], map[string]interface{}{
				"size": size,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Volume resized to %d GB\n", size)
			return nil
		},
	}

	cmd.Flags().IntVar(&size, "size", 0, "New size in GB")
	cmd.MarkFlagRequired("size")

	return cmd
}
