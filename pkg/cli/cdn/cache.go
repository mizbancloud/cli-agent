package cdn

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
	"github.com/mizbancloud/cli/pkg/types"
)

type CacheSettings struct {
	CacheMode         string            `json:"cache_mode"`
	CacheTTL          int               `json:"cache_ttl"`
	DeveloperMode     types.NumericBool `json:"developer_mode"`
	AlwaysOnline      types.NumericBool `json:"always_online"`
	CacheCookies      types.NumericBool `json:"cache_cookies"`
	BrowserCacheMode  string            `json:"browser_cache_mode"`
	BrowserCacheTTL   int               `json:"browser_cache_ttl"`
	ErrorsCacheTTL    int               `json:"errors_cache_ttl"`
	MinifyHTML        types.NumericBool `json:"minify_html"`
	MinifyCSS         types.NumericBool `json:"minify_css"`
	MinifyJS          types.NumericBool `json:"minify_js"`
	ImageOptimization types.NumericBool `json:"image_optimization"`
}

func NewCacheCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cache",
		Short: "Manage cache settings",
		Long:  "Configure caching settings for your domains.",
	}

	cmd.AddCommand(newCacheStatusCmd())
	cmd.AddCommand(newCachePurgeCmd())
	cmd.AddCommand(newCacheModeCmd())
	cmd.AddCommand(newCacheDeveloperModeCmd())
	cmd.AddCommand(newCacheAlwaysOnlineCmd())
	cmd.AddCommand(newCacheCookiesCmd())
	cmd.AddCommand(newCacheSettingsCmd())

	return cmd
}

func newCacheStatusCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "status",
		Short: "Get cache settings",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var settings CacheSettings
			if err := json.Unmarshal(resp.Data, &settings); err != nil {
				return fmt.Errorf("failed to parse settings: %w", err)
			}

			fmt.Printf("Cache Settings\n")
			fmt.Printf("==============\n")
			fmt.Printf("Edge Cache:\n")
			fmt.Printf("  Mode:           %s\n", settings.CacheMode)
			fmt.Printf("  TTL:            %d seconds\n", settings.CacheTTL)
			fmt.Printf("  Developer Mode: %v\n", settings.DeveloperMode.Bool())
			fmt.Printf("  Always Online:  %v\n", settings.AlwaysOnline.Bool())
			fmt.Printf("  Cache Cookies:  %v\n", settings.CacheCookies.Bool())
			fmt.Printf("\nBrowser Cache:\n")
			fmt.Printf("  Mode:           %s\n", settings.BrowserCacheMode)
			fmt.Printf("  TTL:            %d seconds\n", settings.BrowserCacheTTL)
			fmt.Printf("\nError Cache TTL:  %d seconds\n", settings.ErrorsCacheTTL)
			fmt.Printf("\nMinification:\n")
			fmt.Printf("  HTML:           %v\n", settings.MinifyHTML.Bool())
			fmt.Printf("  CSS:            %v\n", settings.MinifyCSS.Bool())
			fmt.Printf("  JS:             %v\n", settings.MinifyJS.Bool())
			fmt.Printf("\nImage Optimization: %v\n", settings.ImageOptimization.Bool())

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheModeCmd() *cobra.Command {
	var domainID int
	var mode string

	cmd := &cobra.Command{
		Use:   "mode",
		Short: "Set cache mode",
		Long: `Set edge cache mode:
  - standard:   Standard caching
  - aggressive: Aggressive caching (cache more content)
  - no-cache:   Disable caching`,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache/edge/change-mode", domainID), map[string]interface{}{
				"mode": mode,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Cache mode set to: %s\n", mode)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&mode, "mode", "aggressive", "Cache mode (standard/aggressive/no-cache)")
	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("mode")

	return cmd
}

func newCacheAlwaysOnlineCmd() *cobra.Command {
	var domainID int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "always-online",
		Short: "Enable/disable always online mode",
		Long:  "When enabled, serves cached content when origin is unavailable.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache/edge/always-online", domainID), map[string]interface{}{
				"enabled": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("Always online mode enabled")
			} else {
				fmt.Println("Always online mode disabled")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable always online")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheCookiesCmd() *cobra.Command {
	var domainID int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "cache-cookies",
		Short: "Enable/disable cookie caching",
		Long:  "When enabled, caches content even when cookies are present.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache/edge/cache-cookies", domainID), map[string]interface{}{
				"enabled": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("Cookie caching enabled")
			} else {
				fmt.Println("Cookie caching disabled")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable cookie caching")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCachePurgeCmd() *cobra.Command {
	var domainID int
	var urls []string
	var all bool

	cmd := &cobra.Command{
		Use:   "purge",
		Short: "Purge cached content",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"domain_id": domainID,
			}

			if all {
				body["purge_all"] = true
			} else if len(urls) > 0 {
				body["urls"] = urls
			} else {
				return fmt.Errorf("specify --all or --url")
			}

			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache/edge/purge-cache", domainID), body)
			if err != nil {
				return err
			}

			if all {
				fmt.Println("All cache purged successfully")
			} else {
				fmt.Printf("Purged %d URL(s) successfully\n", len(urls))
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringSliceVar(&urls, "url", nil, "URLs to purge (can be specified multiple times)")
	cmd.Flags().BoolVar(&all, "all", false, "Purge all cache")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheSettingsCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "settings",
		Short: "Manage cache settings",
	}

	cmd.AddCommand(newCacheTTLCmd())
	cmd.AddCommand(newCacheBrowserCmd())
	cmd.AddCommand(newCacheErrorsTTLCmd())
	cmd.AddCommand(newCacheMinifyCmd())
	cmd.AddCommand(newCacheImageCmd())

	return cmd
}

func newCacheTTLCmd() *cobra.Command {
	var domainID, ttl int
	var mode string

	cmd := &cobra.Command{
		Use:   "ttl",
		Short: "Set cache TTL",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache/edge/change-ttl", domainID), map[string]interface{}{
				"mode": mode,
				"ttl":  ttl,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Cache TTL set to %d seconds (mode: %s)\n", ttl, mode)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&mode, "mode", "aggressive", "Cache mode (standard/aggressive/no-cache)")
	cmd.Flags().IntVar(&ttl, "ttl", 86400, "TTL in seconds")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheBrowserCmd() *cobra.Command {
	var domainID, ttl int
	var mode string

	cmd := &cobra.Command{
		Use:   "browser",
		Short: "Set browser cache TTL",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache/browser/change-mode", domainID), map[string]interface{}{
				"mode": mode,
				"ttl":  ttl,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Browser cache set (mode: %s, TTL: %d)\n", mode, ttl)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&mode, "mode", "respect", "Mode (respect/override)")
	cmd.Flags().IntVar(&ttl, "ttl", 86400, "TTL in seconds")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheMinifyCmd() *cobra.Command {
	var domainID int
	var html, css, js bool

	cmd := &cobra.Command{
		Use:   "minify",
		Short: "Configure minification",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/acceleration/assets/minify", domainID), map[string]interface{}{
				"html": html,
				"css":  css,
				"js":   js,
			})
			if err != nil {
				return err
			}

			fmt.Println("Minification settings updated")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&html, "html", false, "Minify HTML")
	cmd.Flags().BoolVar(&css, "css", false, "Minify CSS")
	cmd.Flags().BoolVar(&js, "js", false, "Minify JavaScript")

	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheErrorsTTLCmd() *cobra.Command {
	var domainID, ttl int

	cmd := &cobra.Command{
		Use:   "errors-ttl",
		Short: "Set error responses cache TTL",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache/errors/cache-ttl", domainID), map[string]interface{}{
				"ttl": ttl,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Error responses cache TTL set to %d seconds\n", ttl)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&ttl, "ttl", 300, "TTL in seconds")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheImageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "image",
		Short: "Configure image optimization",
	}

	cmd.AddCommand(newCacheImageWebpCmd())
	cmd.AddCommand(newCacheImageResizeCmd())

	return cmd
}

func newCacheImageWebpCmd() *cobra.Command {
	var domainID int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "webp",
		Short: "Enable/disable WebP conversion",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/acceleration/images/optimize", domainID), map[string]interface{}{
				"webp": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("WebP conversion enabled")
			} else {
				fmt.Println("WebP conversion disabled")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable WebP")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheImageResizeCmd() *cobra.Command {
	var domainID int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "resize",
		Short: "Enable/disable image resizing",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/acceleration/images/resize", domainID), map[string]interface{}{
				"enabled": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("Image resizing enabled")
			} else {
				fmt.Println("Image resizing disabled")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable image resizing")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newCacheDeveloperModeCmd() *cobra.Command {
	var domainID int
	var enabled bool

	cmd := &cobra.Command{
		Use:   "dev-mode",
		Short: "Enable/disable developer mode",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cache/edge/developer-mode", domainID), map[string]interface{}{
				"enabled": enabled,
			})
			if err != nil {
				return err
			}

			if enabled {
				fmt.Println("Developer mode enabled (cache bypassed)")
			} else {
				fmt.Println("Developer mode disabled")
			}
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&enabled, "enabled", true, "Enable developer mode")

	cmd.MarkFlagRequired("domain")

	return cmd
}
