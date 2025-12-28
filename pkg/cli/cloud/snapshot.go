package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type Snapshot struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Size      int    `json:"size"`
	Status    string `json:"status"`
	ServerID  int    `json:"server_id"`
	CreatedAt string `json:"created_at"`
}

func NewSnapshotCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "snapshot",
		Aliases: []string{"snap", "snapshots"},
		Short:   "Manage snapshots",
		Long:    "Create and manage server snapshots.",
	}

	cmd.AddCommand(newSnapshotListCmd())
	cmd.AddCommand(newSnapshotCreateCmd())
	cmd.AddCommand(newSnapshotGetCmd())
	cmd.AddCommand(newSnapshotDeleteCmd())

	return cmd
}

func newSnapshotListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all snapshots",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/snapshots")
			if err != nil {
				return err
			}

			var snapshots []Snapshot
			if err := json.Unmarshal(resp.Data, &snapshots); err != nil {
				return fmt.Errorf("failed to parse snapshots: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(snapshots, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(snapshots) == 0 {
				fmt.Println("No snapshots found")
				return nil
			}

			fmt.Printf("%-6s %-25s %-10s %-12s %-20s\n", "ID", "NAME", "SIZE(GB)", "STATUS", "CREATED")
			fmt.Println(strings.Repeat("-", 80))
			for _, s := range snapshots {
				fmt.Printf("%-6d %-25s %-10d %-12s %-20s\n", s.ID, truncate(s.Name, 25), s.Size, s.Status, s.CreatedAt)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newSnapshotCreateCmd() *cobra.Command {
	var name string
	var serverID int

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new snapshot",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"name":      name,
				"server_id": serverID,
			}

			resp, err := client.Post("/v1/cloud/snapshots", body)
			if err != nil {
				return err
			}

			var snapshot Snapshot
			if err := json.Unmarshal(resp.Data, &snapshot); err != nil {
				return fmt.Errorf("failed to parse snapshot: %w", err)
			}

			fmt.Printf("Snapshot created successfully!\n")
			fmt.Printf("ID: %d\n", snapshot.ID)
			fmt.Printf("Name: %s\n", snapshot.Name)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Snapshot name")
	cmd.Flags().IntVar(&serverID, "server", 0, "Server ID to snapshot")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("server")

	return cmd
}

func newSnapshotGetCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get [snapshot-id]",
		Short: "Get snapshot details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/snapshots/" + args[0])
			if err != nil {
				return err
			}

			var snapshot Snapshot
			if err := json.Unmarshal(resp.Data, &snapshot); err != nil {
				return fmt.Errorf("failed to parse snapshot: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(snapshot, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			fmt.Printf("ID:        %d\n", snapshot.ID)
			fmt.Printf("Name:      %s\n", snapshot.Name)
			fmt.Printf("Size:      %d GB\n", snapshot.Size)
			fmt.Printf("Status:    %s\n", snapshot.Status)
			fmt.Printf("Server ID: %d\n", snapshot.ServerID)
			fmt.Printf("Created:   %s\n", snapshot.CreatedAt)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newSnapshotDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [snapshot-id]",
		Short: "Delete a snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete snapshot %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete("/v1/cloud/snapshots/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("Snapshot deleted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}
