package cdn

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
	"github.com/mizbancloud/cli/pkg/types"
)

type DDoSSettings struct {
	Mode              string            `json:"mode"`
	CaptchaModule     string            `json:"captcha_module"`
	CookieTTL         int               `json:"cookie_ttl"`
	JsTTL             int               `json:"js_ttl"`
	CaptchaTTL        int               `json:"captcha_ttl"`
	UnderAttack       types.NumericBool `json:"under_attack"`
	JsChallenge       types.NumericBool `json:"js_challenge"`
	CaptchaChallenge  types.NumericBool `json:"captcha_challenge"`
}

func NewDDoSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ddos",
		Short: "Manage DDoS protection",
		Long:  "Configure DDoS protection settings for your domains.",
	}

	cmd.AddCommand(newDDoSStatusCmd())
	cmd.AddCommand(newDDoSModeCmd())
	cmd.AddCommand(newDDoSCaptchaCmd())
	cmd.AddCommand(newDDoSTTLCmd())

	return cmd
}

func newDDoSStatusCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get DDoS protection status",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/ddos", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var settings DDoSSettings
			if err := json.Unmarshal(resp.Data, &settings); err != nil {
				return fmt.Errorf("failed to parse settings: %w", err)
			}

			fmt.Printf("DDoS Protection Settings\n")
			fmt.Printf("========================\n")
			fmt.Printf("Mode:              %s\n", settings.Mode)
			fmt.Printf("Under Attack:      %v\n", settings.UnderAttack.Bool())
			fmt.Printf("JS Challenge:      %v\n", settings.JsChallenge.Bool())
			fmt.Printf("Captcha Challenge: %v\n", settings.CaptchaChallenge.Bool())
			fmt.Printf("Captcha Module:    %s\n", settings.CaptchaModule)
			fmt.Printf("\nTTL Settings:\n")
			fmt.Printf("  Cookie TTL:      %d seconds\n", settings.CookieTTL)
			fmt.Printf("  JS TTL:          %d seconds\n", settings.JsTTL)
			fmt.Printf("  Captcha TTL:     %d seconds\n", settings.CaptchaTTL)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newDDoSModeCmd() *cobra.Command {
	var domainID int
	var mode string

	cmd := &cobra.Command{
		Use:   "mode",
		Short: "Set DDoS protection mode",
		Long: `Set DDoS protection mode:
  - off:          Protection disabled
  - normal:       Standard protection
  - high:         High protection level
  - under_attack: Maximum protection (use when under attack)`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/ddos", domainID), map[string]interface{}{
				"mode": mode,
			})
			if err != nil {
				return err
			}

			modeDesc := map[string]string{
				"off":          "Protection disabled",
				"normal":       "Standard protection",
				"high":         "High protection",
				"under_attack": "Maximum protection (Under Attack mode)",
			}

			fmt.Printf("DDoS protection mode set to: %s\n", mode)
			if desc, ok := modeDesc[mode]; ok {
				fmt.Printf("Description: %s\n", desc)
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&mode, "mode", "normal", "Protection mode (off/normal/high/under_attack)")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("mode")

	return cmd
}

func newDDoSCaptchaCmd() *cobra.Command {
	var domainID int
	var module string

	cmd := &cobra.Command{
		Use:   "captcha",
		Short: "Set captcha module",
		Long: `Set captcha module type:
  - recaptcha:  Google reCAPTCHA
  - hcaptcha:   hCaptcha
  - turnstile:  Cloudflare Turnstile`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/ddos/captcha-module", domainID), map[string]interface{}{
				"module": module,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Captcha module set to: %s\n", module)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&module, "module", "recaptcha", "Captcha module (recaptcha/hcaptcha/turnstile)")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("module")

	return cmd
}

func newDDoSTTLCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ttl",
		Short: "Set challenge TTL values",
	}

	cmd.AddCommand(newDDoSCookieTTLCmd())
	cmd.AddCommand(newDDoSJsTTLCmd())
	cmd.AddCommand(newDDoSCaptchaTTLCmd())

	return cmd
}

func newDDoSCookieTTLCmd() *cobra.Command {
	var domainID, ttl int

	cmd := &cobra.Command{
		Use:   "cookie",
		Short: "Set cookie challenge TTL",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/ddos/set-ttl/cookie", domainID), map[string]interface{}{
				"ttl": ttl,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Cookie challenge TTL set to: %d seconds\n", ttl)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&ttl, "ttl", 3600, "TTL in seconds")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("ttl")

	return cmd
}

func newDDoSJsTTLCmd() *cobra.Command {
	var domainID, ttl int

	cmd := &cobra.Command{
		Use:   "js",
		Short: "Set JavaScript challenge TTL",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/ddos/set-ttl/js", domainID), map[string]interface{}{
				"ttl": ttl,
			})
			if err != nil {
				return err
			}

			fmt.Printf("JavaScript challenge TTL set to: %d seconds\n", ttl)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&ttl, "ttl", 3600, "TTL in seconds")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("ttl")

	return cmd
}

func newDDoSCaptchaTTLCmd() *cobra.Command {
	var domainID, ttl int

	cmd := &cobra.Command{
		Use:   "captcha",
		Short: "Set captcha challenge TTL",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/ddos/set-ttl/captcha", domainID), map[string]interface{}{
				"ttl": ttl,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Captcha challenge TTL set to: %d seconds\n", ttl)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&ttl, "ttl", 3600, "TTL in seconds")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("ttl")

	return cmd
}
