package ticket

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mizbancloud/cli/pkg/api"
	"github.com/mizbancloud/cli/pkg/types"
)

type Ticket struct {
	ID           int               `json:"id"`
	Subject      string            `json:"subject"`
	Status       string            `json:"status"`
	Priority     string            `json:"priority"`
	Department   string            `json:"department"`
	DepartmentID int               `json:"department_id"`
	UserID       int               `json:"user_id"`
	IsClosed     types.NumericBool `json:"is_closed"`
	CreatedAt    string            `json:"created_at"`
	UpdatedAt    string            `json:"updated_at"`
}

type TicketReply struct {
	ID        int               `json:"id"`
	TicketID  int               `json:"ticket_id"`
	Message   string            `json:"message"`
	Content   string            `json:"content"`
	Author    string            `json:"author"`
	UserID    int               `json:"user_id"`
	IsStaff   types.NumericBool `json:"is_staff"`
	CreatedAt string            `json:"created_at"`
}

func NewTicketCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "ticket",
		Aliases: []string{"tickets", "support"},
		Short:   "Manage support tickets",
		Long:    "Create and manage support tickets.",
	}

	cmd.AddCommand(newTicketListCmd())
	cmd.AddCommand(newTicketCreateCmd())
	cmd.AddCommand(newTicketGetCmd())
	cmd.AddCommand(newTicketReplyCmd())
	cmd.AddCommand(newTicketCloseCmd())
	cmd.AddCommand(newTicketDepartmentsCmd())

	return cmd
}

func newTicketListCmd() *cobra.Command {
	var status string
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List tickets",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			endpoint := "/v1/support/tickets"
			if status != "" {
				endpoint += "?status=" + status
			}

			resp, err := client.Get(endpoint)
			if err != nil {
				return err
			}

			var tickets []Ticket
			if err := json.Unmarshal(resp.Data, &tickets); err != nil {
				return fmt.Errorf("failed to parse tickets: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(tickets, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			if len(tickets) == 0 {
				fmt.Println("No tickets found")
				return nil
			}

			fmt.Printf("%-6s %-35s %-12s %-10s %-15s\n", "ID", "SUBJECT", "STATUS", "PRIORITY", "DEPARTMENT")
			fmt.Println(strings.Repeat("-", 85))
			for _, t := range tickets {
				fmt.Printf("%-6d %-35s %-12s %-10s %-15s\n",
					t.ID, truncate(t.Subject, 35), t.Status, t.Priority, t.Department)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "Filter by status (open/closed/pending)")
	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newTicketCreateCmd() *cobra.Command {
	var subject, message, department, priority string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new ticket",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()

			body := map[string]string{
				"subject":    subject,
				"message":    message,
				"department": department,
				"priority":   priority,
			}

			resp, err := client.Post("/v1/support/tickets", body)
			if err != nil {
				return err
			}

			var ticket Ticket
			if err := json.Unmarshal(resp.Data, &ticket); err != nil {
				return fmt.Errorf("failed to parse ticket: %w", err)
			}

			fmt.Printf("Ticket created successfully!\n")
			fmt.Printf("ID: %d\n", ticket.ID)
			fmt.Printf("Subject: %s\n", ticket.Subject)
			fmt.Printf("Status: %s\n", ticket.Status)

			return nil
		},
	}

	cmd.Flags().StringVar(&subject, "subject", "", "Ticket subject")
	cmd.Flags().StringVar(&message, "message", "", "Ticket message")
	cmd.Flags().StringVar(&department, "department", "support", "Department (support/billing/technical)")
	cmd.Flags().StringVar(&priority, "priority", "normal", "Priority (low/normal/high/urgent)")

	cmd.MarkFlagRequired("subject")
	cmd.MarkFlagRequired("message")

	return cmd
}

func newTicketGetCmd() *cobra.Command {
	var jsonOutput bool

	cmd := &cobra.Command{
		Use:   "get [ticket-id]",
		Short: "Get ticket details with replies",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/support/tickets/" + args[0])
			if err != nil {
				return err
			}

			var result struct {
				Ticket  Ticket        `json:"ticket"`
				Replies []TicketReply `json:"replies"`
			}
			if err := json.Unmarshal(resp.Data, &result); err != nil {
				return fmt.Errorf("failed to parse ticket: %w", err)
			}

			if jsonOutput {
				output, _ := json.MarshalIndent(result, "", "  ")
				fmt.Println(string(output))
				return nil
			}

			fmt.Printf("ID:         %d\n", result.Ticket.ID)
			fmt.Printf("Subject:    %s\n", result.Ticket.Subject)
			fmt.Printf("Status:     %s\n", result.Ticket.Status)
			fmt.Printf("Priority:   %s\n", result.Ticket.Priority)
			fmt.Printf("Department: %s\n", result.Ticket.Department)
			fmt.Printf("Created:    %s\n", result.Ticket.CreatedAt)
			fmt.Printf("Updated:    %s\n", result.Ticket.UpdatedAt)

			if len(result.Replies) > 0 {
				fmt.Println("\n--- Replies ---")
				for _, r := range result.Replies {
					authorType := "Customer"
					if r.IsStaff.Bool() {
						authorType = "Staff"
					}
					msg := r.Message
					if msg == "" {
						msg = r.Content
					}
					fmt.Printf("\n[%s] %s (%s):\n%s\n",
						r.CreatedAt, r.Author, authorType, msg)
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&jsonOutput, "json", false, "Output as JSON")

	return cmd
}

func newTicketReplyCmd() *cobra.Command {
	var message string

	cmd := &cobra.Command{
		Use:   "reply [ticket-id]",
		Short: "Reply to a ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/support/tickets/"+args[0]+"/replies", map[string]string{
				"message": message,
			})
			if err != nil {
				return err
			}

			fmt.Println("Reply sent successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&message, "message", "", "Reply message")
	cmd.MarkFlagRequired("message")

	return cmd
}

func newTicketCloseCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "close [ticket-id]",
		Short: "Close a ticket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			_, err := client.Post("/v1/support/tickets/"+args[0]+"/status", map[string]string{
				"status": "closed",
			})
			if err != nil {
				return err
			}

			fmt.Println("Ticket closed successfully")
			return nil
		},
	}
}

func newTicketDepartmentsCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "departments",
		Short: "List available departments",
		RunE: func(cmd *cobra.Command, args []string) error {
			client := api.NewClient()
			resp, err := client.Get("/v1/support/tickets/departments")
			if err != nil {
				return err
			}

			var departments []struct {
				ID   int    `json:"id"`
				Name string `json:"name"`
			}
			if err := json.Unmarshal(resp.Data, &departments); err != nil {
				return fmt.Errorf("failed to parse departments: %w", err)
			}

			fmt.Printf("%-6s %-20s\n", "ID", "NAME")
			fmt.Println(strings.Repeat("-", 30))
			for _, d := range departments {
				fmt.Printf("%-6d %-20s\n", d.ID, d.Name)
			}

			return nil
		},
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
