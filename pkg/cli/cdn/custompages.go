package cdn

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
)

type CustomPages struct {
	Error403 string `json:"error_403"`
	Error404 string `json:"error_404"`
	Error500 string `json:"error_500"`
	Error502 string `json:"error_502"`
	Error503 string `json:"error_503"`
	Error504 string `json:"error_504"`
}

func NewCustomPagesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "custom-pages",
		Aliases: []string{"pages", "error-pages"},
		Short:   "Manage custom error pages",
		Long:    "Configure custom HTML pages for CDN error responses.",
	}

	cmd.AddCommand(newCustomPagesGetCmd())
	cmd.AddCommand(newCustomPagesSetCmd())
	cmd.AddCommand(newCustomPagesDeleteCmd())

	return cmd
}

func newCustomPagesGetCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get custom error pages",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/custom-pages", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var pages CustomPages
			if err := json.Unmarshal(resp.Data, &pages); err != nil {
				return fmt.Errorf("failed to parse pages: %w", err)
			}

			fmt.Printf("Custom Error Pages\n")
			fmt.Printf("==================\n")
			printPageStatus("403 Forbidden", pages.Error403)
			printPageStatus("404 Not Found", pages.Error404)
			printPageStatus("500 Internal Server Error", pages.Error500)
			printPageStatus("502 Bad Gateway", pages.Error502)
			printPageStatus("503 Service Unavailable", pages.Error503)
			printPageStatus("504 Gateway Timeout", pages.Error504)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func printPageStatus(name, content string) {
	status := "Default"
	if content != "" {
		status = "Custom"
	}
	fmt.Printf("%-30s %s\n", name+":", status)
}

func newCustomPagesSetCmd() *cobra.Command {
	var domainID int
	var errorCode int
	var htmlContent string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set custom error page",
		Long: `Set custom HTML content for an error page.
Supported error codes: 403, 404, 500, 502, 503, 504`,
		RunE: func(cmd *cobra.Command, args []string) error {
			validCodes := map[int]bool{403: true, 404: true, 500: true, 502: true, 503: true, 504: true}
			if !validCodes[errorCode] {
				return fmt.Errorf("invalid error code: %d (valid: 403, 404, 500, 502, 503, 504)", errorCode)
			}

			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/custom-pages", domainID), map[string]interface{}{
				"error_code": errorCode,
				"content":    htmlContent,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Custom page for error %d set successfully\n", errorCode)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&errorCode, "code", 0, "Error code (403, 404, 500, 502, 503, 504)")
	cmd.Flags().StringVar(&htmlContent, "html", "", "HTML content for the error page")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("code")
	cmd.MarkFlagRequired("html")

	return cmd
}

func newCustomPagesDeleteCmd() *cobra.Command {
	var domainID int
	var errorCode int

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete custom error page (restore default)",
		RunE: func(cmd *cobra.Command, args []string) error {
			validCodes := map[int]bool{403: true, 404: true, 500: true, 502: true, 503: true, 504: true}
			if !validCodes[errorCode] {
				return fmt.Errorf("invalid error code: %d", errorCode)
			}

			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/custom-pages", domainID))
			if err != nil {
				return err
			}

			fmt.Printf("Custom page for error %d deleted (restored to default)\n", errorCode)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&errorCode, "code", 0, "Error code")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("code")

	return cmd
}
