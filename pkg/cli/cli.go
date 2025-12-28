package cli

import (
	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/cli/auth"
	"github.com/mizbancloud/cli/pkg/cli/cdn"
	"github.com/mizbancloud/cli/pkg/cli/cloud"
	"github.com/mizbancloud/cli/pkg/cli/ticket"
	"github.com/mizbancloud/cli/pkg/config"
)

func NewRootCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:     "mizban",
		Short:   "MizbanCloud CLI - Manage your cloud infrastructure",
		Long:    "MizbanCloud CLI is a command-line tool for managing MizbanCloud services including Cloud (IaaS), CDN, and Support.",
		Version: config.Version,
		Run: func(cmd *cobra.Command, args []string) {
			cmd.Help()
		},
	}

	// Auth commands
	rootCmd.AddCommand(auth.NewLoginCmd())
	rootCmd.AddCommand(auth.NewLogoutCmd())
	rootCmd.AddCommand(auth.NewProfileCmd())

	// Cloud commands
	rootCmd.AddCommand(cloud.NewServerCmd())
	rootCmd.AddCommand(cloud.NewVolumeCmd())
	rootCmd.AddCommand(cloud.NewSnapshotCmd())
	rootCmd.AddCommand(cloud.NewSSHCmd())
	rootCmd.AddCommand(cloud.NewFirewallCmd())
	rootCmd.AddCommand(cloud.NewNetworkCmd())

	// CDN commands
	rootCmd.AddCommand(cdn.NewDomainCmd())
	rootCmd.AddCommand(cdn.NewDNSCmd())
	rootCmd.AddCommand(cdn.NewSSLCmd())
	rootCmd.AddCommand(cdn.NewCacheCmd())
	rootCmd.AddCommand(cdn.NewWAFCmd())
	rootCmd.AddCommand(cdn.NewClusterCmd())
	rootCmd.AddCommand(cdn.NewDDoSCmd())
	rootCmd.AddCommand(cdn.NewRateLimitCmd())
	rootCmd.AddCommand(cdn.NewAccessRulesCmd())
	rootCmd.AddCommand(cdn.NewCustomPagesCmd())
	rootCmd.AddCommand(cdn.NewPageRulesCmd())
	rootCmd.AddCommand(cdn.NewLogForwarderCmd())
	rootCmd.AddCommand(cdn.NewPlansCmd())

	// Ticket commands
	rootCmd.AddCommand(ticket.NewTicketCmd())

	return rootCmd
}
