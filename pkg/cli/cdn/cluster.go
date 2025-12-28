package cdn

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
	"github.com/mizbancloud/cli/pkg/types"
)

type ClusterPool struct {
	ID                      int               `json:"id"`
	DomainID                int               `json:"domain_id"`
	Name                    string            `json:"name"`
	Port                    int               `json:"port"`
	Description             string            `json:"description"`
	Method                  string            `json:"method"`
	HashKey                 string            `json:"hash_key,omitempty"`
	ErrorReporting          types.NumericBool `json:"error_reporting"`
	MonitoringProtocol      string            `json:"monitoring_protocol,omitempty"`
	MonitoringPort          int               `json:"monitoring_port,omitempty"`
	MonitoringMethod        string            `json:"monitoring_method,omitempty"`
	MonitoringErrorReporting types.NumericBool `json:"monitoring_error_reporting"`
	Servers                 []ClusterServer   `json:"servers,omitempty"`
	CreatedAt               string            `json:"created_at"`
	UpdatedAt               string            `json:"updated_at"`
}

type ClusterServer struct {
	ID         int    `json:"id"`
	PoolID     int    `json:"pool_id"`
	Address    string `json:"address"`
	Weight     int    `json:"weight"`
	HostHeader string `json:"host_header,omitempty"`
	Port       int    `json:"port"`
	Priority   int    `json:"priority"`
	Protocol   string `json:"protocol"`
}

func NewClusterCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "cluster",
		Aliases: []string{"clusters", "pool", "pools"},
		Short:   "Manage load balancer clusters",
		Long:    "Configure load balancer pools and servers for your domains.",
	}

	cmd.AddCommand(newClusterListCmd())
	cmd.AddCommand(newClusterAddCmd())
	cmd.AddCommand(newClusterUpdateCmd())
	cmd.AddCommand(newClusterDeleteCmd())
	cmd.AddCommand(newClusterServerCmd())
	cmd.AddCommand(newClusterAssignmentsCmd())
	cmd.AddCommand(newClusterAssignCmd())
	cmd.AddCommand(newClusterUnassignCmd())

	return cmd
}

func newClusterListCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cluster pools",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster", domainID))
			if err != nil {
				return err
			}

			var pools []ClusterPool
			if err := json.Unmarshal(resp.Data, &pools); err != nil {
				return fmt.Errorf("failed to parse clusters: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(pools, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(pools) == 0 {
				fmt.Println("No cluster pools found")
				return nil
			}

			for _, p := range pools {
				fmt.Printf("Pool: %s (ID: %d)\n", p.Name, p.ID)
				fmt.Printf("  Method: %-15s  Port: %-6d  Error Reporting: %v\n", p.Method, p.Port, p.ErrorReporting.Bool())

				// Show monitoring status
				monitoring := "off"
				if p.MonitoringProtocol != "" {
					monitoring = strings.ToLower(p.MonitoringProtocol)
					if p.MonitoringPort > 0 {
						monitoring = fmt.Sprintf("%s:%d", monitoring, p.MonitoringPort)
					}
				}
				fmt.Printf("  Monitoring: %s\n", monitoring)

				if p.Description != "" {
					fmt.Printf("  Description: %s\n", p.Description)
				}

				if len(p.Servers) > 0 {
					fmt.Println("  Servers:")
					fmt.Printf("    %-8s %-25s %-8s %-8s %-10s %-10s\n", "ID", "ADDRESS", "PORT", "WEIGHT", "PROTOCOL", "STATUS")
					fmt.Printf("    %s\n", strings.Repeat("-", 80))
					for _, s := range p.Servers {
						status := "active"
						if s.Priority == -1 {
							status = "backup"
						}
						fmt.Printf("    %-8d %-25s %-8d %-8d %-10s %-10s\n",
							s.ID, s.Address, s.Port, s.Weight, s.Protocol, status)
					}
				} else {
					fmt.Println("  Servers: (none)")
				}
				fmt.Println()
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

type ClusterAssignment struct {
	ClusterID   int    `json:"cluster_id"`
	ClusterName string `json:"cluster_name"`
	PathID      int    `json:"path_id"`
	Path        string `json:"path"`
}

func newClusterAssignmentsCmd() *cobra.Command {
	var domainID int
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "assignments",
		Short: "List all cluster assignments",
		Long:  "List all cluster to path assignments for a domain.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster/assignments", domainID))
			if err != nil {
				return err
			}

			if jsonOutput {
				fmt.Println(string(resp.Data))
				return nil
			}

			var assignments []ClusterAssignment
			if err := json.Unmarshal(resp.Data, &assignments); err != nil {
				return fmt.Errorf("failed to parse assignments: %w", err)
			}

			if len(assignments) == 0 {
				fmt.Println("No cluster assignments found")
				return nil
			}

			fmt.Printf("%-12s %-20s %-10s %-30s\n", "CLUSTER ID", "CLUSTER NAME", "PATH ID", "PATH")
			fmt.Println(strings.Repeat("-", 75))
			for _, a := range assignments {
				fmt.Printf("%-12d %-20s %-10d %-30s\n",
					a.ClusterID, truncate(a.ClusterName, 20), a.PathID, truncate(a.Path, 30))
			}

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")
	cmd.MarkFlagRequired("domain")

	return cmd
}

func newClusterAddCmd() *cobra.Command {
	var domainID, port int
	var name, method, description, hashKey string
	var errorReporting bool

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a new cluster pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"name":            name,
				"port":            port,
				"method":          method,
				"description":     description,
				"error_reporting": errorReporting,
			}
			if hashKey != "" {
				body["hash_key"] = hashKey
			}

			resp, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster", domainID), body)
			if err != nil {
				return err
			}

			var pool ClusterPool
			if err := json.Unmarshal(resp.Data, &pool); err != nil {
				return fmt.Errorf("failed to parse cluster: %w", err)
			}

			fmt.Printf("Cluster pool created successfully!\n")
			fmt.Printf("ID: %d\n", pool.ID)
			fmt.Printf("Name: %s\n", pool.Name)
			fmt.Printf("Method: %s\n", pool.Method)

			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().StringVar(&name, "name", "", "Pool name")
	cmd.Flags().IntVar(&port, "port", 443, "Backend port")
	cmd.Flags().StringVar(&method, "method", "roundrobin", "Load balancing method (roundrobin/leastconn/iphash)")
	cmd.Flags().StringVar(&description, "description", "", "Pool description")
	cmd.Flags().StringVar(&hashKey, "hash-key", "", "Hash key for iphash method")
	cmd.Flags().BoolVar(&errorReporting, "error-reporting", true, "Enable error reporting")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("name")

	return cmd
}

func newClusterUpdateCmd() *cobra.Command {
	var domainID, clusterID, port int
	var name, method, description, hashKey string
	var errorReporting bool

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a cluster pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"name":            name,
				"port":            port,
				"method":          method,
				"description":     description,
				"error_reporting": errorReporting,
			}
			if hashKey != "" {
				body["hash_key"] = hashKey
			}

			_, err := client.Put(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster/%d", domainID, clusterID), body)
			if err != nil {
				return err
			}

			fmt.Println("Cluster pool updated successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&clusterID, "cluster", 0, "Cluster ID")
	cmd.Flags().StringVar(&name, "name", "", "Pool name")
	cmd.Flags().IntVar(&port, "port", 443, "Backend port")
	cmd.Flags().StringVar(&method, "method", "roundrobin", "Load balancing method")
	cmd.Flags().StringVar(&description, "description", "", "Pool description")
	cmd.Flags().StringVar(&hashKey, "hash-key", "", "Hash key for iphash method")
	cmd.Flags().BoolVar(&errorReporting, "error-reporting", true, "Enable error reporting")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("cluster")

	return cmd
}

func newClusterDeleteCmd() *cobra.Command {
	var domainID, clusterID int
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a cluster pool",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to delete cluster %d? (yes/no): ", clusterID)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster/%d", domainID, clusterID))
			if err != nil {
				return err
			}

			fmt.Println("Cluster pool deleted successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&clusterID, "cluster", 0, "Cluster ID")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("cluster")

	return cmd
}

func newClusterServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "server",
		Aliases: []string{"servers"},
		Short:   "Manage cluster servers",
	}

	cmd.AddCommand(newClusterServerAddCmd())
	cmd.AddCommand(newClusterServerDeleteCmd())

	return cmd
}

func newClusterServerAddCmd() *cobra.Command {
	var domainID, clusterID, port, weight, priority int
	var address, hostHeader, protocol string

	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add a server to cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]interface{}{
				"address":  address,
				"port":     port,
				"weight":   weight,
				"priority": priority,
				"protocol": protocol,
			}
			if hostHeader != "" {
				body["host_header"] = hostHeader
			}

			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster/%d/servers", domainID, clusterID), body)
			if err != nil {
				return err
			}

			fmt.Printf("Server %s added to cluster successfully\n", address)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&clusterID, "cluster", 0, "Cluster ID")
	cmd.Flags().StringVar(&address, "address", "", "Server address (IP or hostname)")
	cmd.Flags().IntVar(&port, "port", 443, "Server port")
	cmd.Flags().IntVar(&weight, "weight", 100, "Server weight (1-100)")
	cmd.Flags().IntVar(&priority, "priority", 1, "Server priority")
	cmd.Flags().StringVar(&protocol, "protocol", "HTTPS", "Protocol (HTTP/HTTPS)")
	cmd.Flags().StringVar(&hostHeader, "host-header", "", "Custom host header")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("cluster")
	cmd.MarkFlagRequired("address")

	return cmd
}

func newClusterServerDeleteCmd() *cobra.Command {
	var domainID, clusterID, serverID int
	var force bool

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Remove a server from cluster",
		RunE: func(cmd *cobra.Command, args []string) error {
			if !force {
				fmt.Printf("Are you sure you want to remove server %d? (yes/no): ", serverID)
				var confirm string
				fmt.Scanln(&confirm)
				if confirm != "yes" {
					fmt.Println("Aborted")
					return nil
				}
			}

			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster/%d/servers/%d", domainID, clusterID, serverID))
			if err != nil {
				return err
			}

			fmt.Println("Server removed from cluster successfully")
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&clusterID, "cluster", 0, "Cluster ID")
	cmd.Flags().IntVar(&serverID, "server", 0, "Server ID")
	cmd.Flags().BoolVarP(&force, "force", "f", false, "Skip confirmation")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("cluster")
	cmd.MarkFlagRequired("server")

	return cmd
}

func newClusterAssignCmd() *cobra.Command {
	var domainID, clusterID, pathID int

	cmd := &cobra.Command{
		Use:   "assign",
		Short: "Assign cluster to a path",
		Long:  "Assign a cluster pool to handle requests for a specific path/page rule.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster/%d/assign", domainID, clusterID), map[string]interface{}{
				"path_id": pathID,
			})
			if err != nil {
				return err
			}

			fmt.Printf("Cluster %d assigned to path %d successfully\n", clusterID, pathID)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&clusterID, "cluster", 0, "Cluster ID")
	cmd.Flags().IntVar(&pathID, "path", 0, "Path ID to assign cluster to")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("cluster")
	cmd.MarkFlagRequired("path")

	return cmd
}

func newClusterUnassignCmd() *cobra.Command {
	var domainID, clusterID, pathID int

	cmd := &cobra.Command{
		Use:   "unassign",
		Short: "Unassign cluster from a path",
		Long:  "Remove cluster assignment from a specific path/page rule.",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Delete(fmt.Sprintf("/v1/cdn/ng/domains/%d/cluster/%d/assign/%d", domainID, clusterID, pathID))
			if err != nil {
				return err
			}

			fmt.Printf("Cluster %d unassigned from path %d successfully\n", clusterID, pathID)
			return nil
		},
	}

	cmd.Flags().IntVar(&domainID, "domain", 0, "Domain ID")
	cmd.Flags().IntVar(&clusterID, "cluster", 0, "Cluster ID")
	cmd.Flags().IntVar(&pathID, "path", 0, "Path ID to unassign cluster from")

	cmd.MarkFlagRequired("domain")
	cmd.MarkFlagRequired("cluster")
	cmd.MarkFlagRequired("path")

	return cmd
}
