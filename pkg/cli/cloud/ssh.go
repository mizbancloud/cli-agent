package cloud

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type SSHKey struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Fingerprint string `json:"fingerprint"`
	PublicKey   string `json:"public_key"`
	CreatedAt   string `json:"created_at"`
}

func NewSSHCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ssh-key",
		Aliases: []string{"ssh", "sshkey"},
		Short:   "Manage SSH keys",
		Long:    "Add and manage SSH keys for server access.",
	}

	cmd.AddCommand(newSSHListCmd())
	cmd.AddCommand(newSSHAddCmd())
	cmd.AddCommand(newSSHGetCmd())
	cmd.AddCommand(newSSHDeleteCmd())
	cmd.AddCommand(newSSHGenerateCmd())

	return cmd
}

func newSSHListCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all SSH keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/ssh")
			if err != nil {
				return err
			}

			var keys []SSHKey
			if err := json.Unmarshal(resp.Data, &keys); err != nil {
				return fmt.Errorf("failed to parse SSH keys: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(keys, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(keys) == 0 {
				fmt.Println("No SSH keys found")
				return nil
			}

			fmt.Printf("%-6s %-20s %-50s\n", "ID", "NAME", "FINGERPRINT")
			fmt.Println(strings.Repeat("-", 80))
			for _, k := range keys {
				fmt.Printf("%-6d %-20s %-50s\n", k.ID, truncate(k.Name, 20), k.Fingerprint)
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newSSHAddCmd() *cobra.Command {
	var name, publicKey string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new SSH key",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]string{
				"name":       name,
				"public_key": publicKey,
			}

			resp, err := client.Post("/v1/cloud/ssh", body)
			if err != nil {
				return err
			}

			var key SSHKey
			if err := json.Unmarshal(resp.Data, &key); err != nil {
				return fmt.Errorf("failed to parse SSH key: %w", err)
			}

			fmt.Printf("SSH key added successfully!\n")
			fmt.Printf("ID: %d\n", key.ID)
			fmt.Printf("Name: %s\n", key.Name)
			fmt.Printf("Fingerprint: %s\n", key.Fingerprint)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Key name")
	cmd.Flags().StringVar(&publicKey, "key", "", "Public key content")

	cmd.MarkFlagRequired("name")
	cmd.MarkFlagRequired("key")

	return cmd
}

func newSSHGetCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get [key-id]",
		Short: "Get SSH key details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/cloud/ssh/" + args[0])
			if err != nil {
				return err
			}

			var key SSHKey
			if err := json.Unmarshal(resp.Data, &key); err != nil {
				return fmt.Errorf("failed to parse SSH key: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(key, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			fmt.Printf("ID:          %d\n", key.ID)
			fmt.Printf("Name:        %s\n", key.Name)
			fmt.Printf("Fingerprint: %s\n", key.Fingerprint)
			fmt.Printf("Public Key:\n%s\n", key.PublicKey)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newSSHDeleteCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "delete [key-id]",
		Short: "Delete an SSH key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete SSH key %s? (yes/no): ", args[0])
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete("/v1/cloud/ssh/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("SSH key deleted successfully")
			return nil
		},
	}

	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	return cmd
}

func newSSHGenerateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a new SSH key pair",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			resp, err := client.Get("/v1/cloud/ssh/random")
			if err != nil {
				return err
			}

			var result struct {
				PrivateKey string `json:"private_key"`
				PublicKey  string `json:"public_key"`
				ID         int    `json:"id"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				return fmt.Errorf("failed to parse SSH key: %w", err)
			}

			fmt.Printf("SSH key pair generated successfully!\n")
			fmt.Printf("ID: %d\n\n", result.ID)
			fmt.Println("Private Key (save this securely):")
			fmt.Println(result.PrivateKey)
			fmt.Println("\nPublic Key:")
			fmt.Println(result.PublicKey)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Key name")
	cmd.MarkFlagRequired("name")

	return cmd
}
