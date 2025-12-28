package auth

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type Profile struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
	NationalID  string `json:"national_id"`
	TFAEnabled  bool   `json:"tfa_enabled"`
}

func NewProfileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "profile",
		Short: "Manage your profile",
		Long:  "View and manage your MizbanCloud profile settings.",
	}

	cmd.AddCommand(newProfileShowCmd())
	cmd.AddCommand(newProfileUpdateCmd())
	cmd.AddCommand(newAPIKeysCmd())

	return cmd
}

func newProfileShowCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "show",
		Short: "Show profile information",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/auth/profile")
			if err != nil {
				return err
			}

			var profile Profile
			if err := json.Unmarshal(resp.Data, &profile); err != nil {
				return fmt.Errorf("failed to parse profile: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(profile, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			fmt.Printf("Name:        %s\n", profile.Name)
			fmt.Printf("Email:       %s\n", profile.Email)
			fmt.Printf("Phone:       %s\n", profile.PhoneNumber)
			fmt.Printf("2FA Enabled: %v\n", profile.TFAEnabled)

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newProfileUpdateCmd() *cobra.Command {
	var name, phone string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update profile information",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]string{}
			if name != "" {
				body["name"] = name
			}
			if phone != "" {
				body["phone_number"] = phone
			}

			if len(body) == 0 {
				return fmt.Errorf("no fields to update")
			}

			_, err := client.Put("/v1/auth/profile", body)
			if err != nil {
				return err
			}

			fmt.Println("Profile updated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Update name")
	cmd.Flags().StringVar(&phone, "phone", "", "Update phone number")

	return cmd
}

func newAPIKeysCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "api-keys",
		Short: "Manage API keys",
	}

	cmd.AddCommand(newAPIKeyListCmd())
	cmd.AddCommand(newAPIKeyCreateCmd())
	cmd.AddCommand(newAPIKeyDeleteCmd())

	return cmd
}

func newAPIKeyListCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List API keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/auth/api-token")
			if err != nil {
				return err
			}

			var keys []struct {
				ID        int    `json:"id"`
				Name      string `json:"name"`
				Token     string `json:"token"`
				CreatedAt string `json:"created_at"`
			}
			if err := json.Unmarshal(resp.Data, &keys); err != nil {
				return fmt.Errorf("failed to parse keys: %w", err)
			}

			if len(keys) == 0 {
				fmt.Println("No API keys found")
				return nil
			}

			fmt.Printf("%-5s %-20s %-40s %-20s\n", "ID", "NAME", "TOKEN", "CREATED")
			fmt.Println(strings.Repeat("-", 90))
			for _, key := range keys {
				fmt.Printf("%-5d %-20s %-40s %-20s\n", key.ID, key.Name, key.Token[:20]+"...", key.CreatedAt)
			}

			return nil
		},
	}
}

func newAPIKeyCreateCmd() *cobra.Command {
	var name string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create new API key",
		RunE: func(cmd *cobra.Command, args []string) error {
			if name == "" {
				return fmt.Errorf("name is required")
			}

			client := api.NewClient()
			resp, err := client.Post("/v1/auth/api-token", map[string]string{"name": name})
			if err != nil {
				return err
			}

			var key struct {
				Token string `json:"token"`
			}
			if err := json.Unmarshal(resp.Data, &key); err != nil {
				return fmt.Errorf("failed to parse response: %w", err)
			}

			fmt.Printf("API key created successfully!\nToken: %s\n", key.Token)
			fmt.Println("\nWarning: Save this token now. You won't be able to see it again!")

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "Name for the API key")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newAPIKeyDeleteCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "delete [key-id]",
		Short: "Delete an API key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Delete("/v1/auth/api-token/" + args[0])
			if err != nil {
				return err
			}

			fmt.Println("API key deleted successfully")
			return nil
		},
	}
}
