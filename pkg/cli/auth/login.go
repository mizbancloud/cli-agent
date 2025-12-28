package auth

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/mizbancloud/cli/pkg/api"
	"github.com/mizbancloud/cli/pkg/config"
)

func NewLoginCmd() *cobra.Command {
	var token, apiURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Login to MizbanCloud",
		Long:  "Authenticate with MizbanCloud using your API token or credentials.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.GetConfig()

			// Set custom API URL if provided
			if apiURL != "" {
				cfg.BaseURL = apiURL
			}

			if token == "" {
				fmt.Print("Enter your API token: ")
				byteToken, err := term.ReadPassword(int(syscall.Stdin))
				if err != nil {
					reader := bufio.NewReader(os.Stdin)
					token, _ = reader.ReadString('\n')
					token = strings.TrimSpace(token)
				} else {
					token = string(byteToken)
				}
				fmt.Println()
			}

			if token == "" {
				return fmt.Errorf("token cannot be empty")
			}

			cfg.Token = token

			client := api.NewClient()
			resp, err := client.Get("/v1/auth/profile")
			if err != nil {
				return fmt.Errorf("login failed: %w", err)
			}

			if err := cfg.Save(); err != nil {
				return fmt.Errorf("failed to save config: %w", err)
			}

			var profile struct {
				Name  string `json:"name"`
				Email string `json:"email"`
			}
			if resp.Data != nil {
				profile, _ = api.ParseData[struct {
					Name  string `json:"name"`
					Email string `json:"email"`
				}](resp)
			}

			fmt.Printf("Successfully logged in as %s (%s)\n", profile.Name, profile.Email)
			return nil
		},
	}

	cmd.Flags().StringVarP(&token, "token", "t", "", "API token")
	cmd.Flags().StringVar(&apiURL, "url", "", "API base URL (e.g., http://127.0.0.1:8003/api/v1)")

	return cmd
}

func NewLogoutCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Logout from MizbanCloud",
		Long:  "Clear saved credentials and logout from MizbanCloud.",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg := config.GetConfig()
			if err := cfg.Logout(); err != nil {
				return fmt.Errorf("failed to logout: %w", err)
			}
			fmt.Println("Successfully logged out")
			return nil
		},
	}
}
