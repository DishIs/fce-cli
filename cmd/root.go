package cmd

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/DishIs/fce-cli/internal/api"
	"github.com/DishIs/fce-cli/internal/auth"
	"github.com/DishIs/fce-cli/internal/config"
	"github.com/DishIs/fce-cli/internal/display"
	"github.com/DishIs/fce-cli/internal/ws"
	"github.com/spf13/cobra"
)

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
)

// ── Root command ──────────────────────────────────────────────────────────────

var rootCmd = &cobra.Command{
	Use:   "fce",
	Short: "FreeCustom.Email CLI — disposable inboxes, OTP extraction, real-time email",
	Long: `
  ◉ fce  —  FreeCustom.Email CLI

  Manage disposable inboxes, extract OTPs, and stream
  real-time email events from your terminal.

  Get started:
    fce login            Authenticate with your account
    fce watch random     Watch a random inbox for emails
    fce status           View your account and plan

  Docs: https://freecustom.email/api/docs
`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		display.Error(err.Error())
		os.Exit(1)
	}
}

func init() {
	rootCmd.AddCommand(
		loginCmd,
		logoutCmd,
		statusCmd,
		usageCmd,
		watchCmd,
		inboxCmd,
		messagesCmd,
		otpCmd,
		domainsCmd,
		versionCmd,
		devCmd,
		updateCmd,
		uninstallCmd,
	)
}

// ── fce uninstall ─────────────────────────────────────────────────────────────

var uninstallCmd = &cobra.Command{
	Use:   "uninstall",
	Short: "Remove all local configuration and credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Print("Are you sure you want to remove all local configuration and logout? (y/N): ")
		var confirm string
		fmt.Scanln(&confirm)
		if strings.ToLower(confirm) != "y" {
			display.Info("Aborted.")
			return nil
		}

		if err := config.Purge(); err != nil {
			return err
		}

		display.Success("Local configuration and credentials cleared.")
		
		// Platform-specific instructions
		display.Header("Next Steps")
		fmt.Println("To completely remove the fce binary, run the command for your platform:")
		fmt.Println()
		fmt.Println(display.TableBadge("Homebrew", "brew uninstall fce"))
		fmt.Println(display.TableBadge("Scoop",    "scoop uninstall fce"))
		fmt.Println(display.TableBadge("Choco",    "choco uninstall fce"))
		fmt.Println(display.TableBadge("NPM",      "npm uninstall -g fce-cli"))
		fmt.Println(display.TableBadge("Manual",   "sudo rm /usr/local/bin/fce"))
		fmt.Println()
		
		return nil
	},
}

// ── fce dev ───────────────────────────────────────────────────────────────────

var devCmd = &cobra.Command{
	Use:   "dev",
	Short: "Instantly register a dev inbox and start watching for emails",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		inbox := devInbox()
		display.Info(fmt.Sprintf("Temporary inbox: %s", inbox))

		if _, err := client.RegisterInbox(inbox); err != nil {
			return fmt.Errorf("failed to register inbox: %w", err)
		}

		display.Success("Watching for emails...")
		fmt.Println()

		return ws.Watch(client.GetAPIKey(), inbox)
	},
}

// ── fce update ────────────────────────────────────────────────────────────────

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update the CLI to the latest version",
	RunE: func(cmd *cobra.Command, args []string) error {
		display.Info("Checking for updates...")
		
		installCmd := "curl -fsSL https://raw.githubusercontent.com/DishIs/fce-cli/main/scripts/install.sh | sh"
		fmt.Printf("Running installer: %s\n", installCmd)
		
		// Execute the installation command
		var c *exec.Cmd
		if os.Getenv("OS") == "Windows_NT" {
			display.Warn("Auto-update on Windows is currently best handled via Scoop or Chocolatey.")
			fmt.Println("Try: scoop update fce  OR  choco upgrade fce")
			return nil
		} else {
			c = exec.Command("sh", "-c", installCmd)
		}
		
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		
		display.Success("Update complete!")
		return nil
	},
}

// ── fce version ───────────────────────────────────────────────────────────────

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version info",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("fce %s (%s) %s\n", Version, Commit, Date)
	},
}

// ── fce login ─────────────────────────────────────────────────────────────────

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with your FreeCustom.Email account",
	Long: `Opens your browser to complete authentication.
After logging in on the website, your API key is saved
securely in your OS keychain.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.Login()
	},
}

// ── fce logout ────────────────────────────────────────────────────────────────

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored credentials",
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.Logout()
	},
}

// ── fce status ────────────────────────────────────────────────────────────────

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show account info, plan, and inbox counts",
	Example: `  fce status`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		result, err := client.GetMe()
		if err != nil {
			return err
		}

		d, _ := result["data"].(map[string]interface{})
		if d == nil {
			d = result
		}

		plan        := strVal(d, "plan")
		planLabel   := strVal(d, "plan_label")
		price       := strVal(d, "price")
		credits     := fmt.Sprintf("%v", intVal(d, "credits"))
		apiInboxes  := fmt.Sprintf("%v", intVal(d, "api_inbox_count"))
		appInboxes  := fmt.Sprintf("%v", intVal(d, "app_inbox_count"))

		display.Header("Account")
		display.Table([]display.Row{
			{Key: "Plan",        Value: planLabel + "  " + display.PlanBadge(plan)},
			{Key: "Price",       Value: price},
			{Key: "Credits",     Value: credits + " remaining"},
			{Key: "API inboxes", Value: apiInboxes},
			{Key: "App inboxes", Value: appInboxes},
		})

		// Cache plan in config
		cfg, _ := config.LoadConfig()
		cfg.Plan = plan
		cfg.PlanLabel = planLabel
		_ = config.SaveConfig(cfg)

		return nil
	},
}

// ── fce usage ─────────────────────────────────────────────────────────────────

var usageCmd = &cobra.Command{
	Use:   "usage",
	Short: "Show request usage for the current billing period",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		result, err := client.GetUsage()
		if err != nil {
			return err
		}

		d, _ := result["data"].(map[string]interface{})
		if d == nil {
			d = result
		}

		used      := fmt.Sprintf("%v", d["requests_used"])
		limit     := fmt.Sprintf("%v", d["requests_limit"])
		remaining := fmt.Sprintf("%v", d["requests_remaining"])
		pct       := strVal(d, "percent_used")
		credits   := fmt.Sprintf("%v", d["credits_remaining"])
		resets    := strVal(d, "resets")

		display.Header("Usage")
		display.Table([]display.Row{
			{Key: "Requests used",      Value: used + " / " + limit},
			{Key: "Remaining",          Value: remaining},
			{Key: "Percent used",       Value: pct},
			{Key: "Credits remaining",  Value: credits},
			{Key: "Resets approx",      Value: resets},
		})

		return nil
	},
}

// ── fce watch ─────────────────────────────────────────────────────────────────

var watchCmd = &cobra.Command{
	Use:   "watch [inbox|random]",
	Short: "Stream emails in real time via WebSocket  [Startup plan+]",
	Long: `Connects to a WebSocket and streams incoming emails to your terminal.

  fce watch                        Watch all registered inboxes
  fce watch random                 Watch a new random inbox (auto-registers)
  fce watch test@ditmail.info      Watch a specific inbox

Requires Startup plan or above. Emails arrive in under 200ms.`,
	Example: `  fce watch
  fce watch random
  fce watch mytest@ditube.info`,
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		// Plan gate
		me, err := client.GetMe()
		if err != nil {
			return err
		}
		d, _ := me["data"].(map[string]interface{})
		if d == nil {
			d = me
		}
		plan := strVal(d, "plan")
		if !api.HasPlan(plan, api.PlanStartup) {
			display.PlanGate("startup", "WebSocket watch")
			return nil
		}

		apiKey := client.GetAPIKey()
		mailbox := ""

		if len(args) > 0 {
			arg := args[0]
			if arg == "random" {
				// Generate a random inbox and register it
				mailbox = randomInbox()
				display.Info(fmt.Sprintf("Registering random inbox: %s", mailbox))
				if _, err := client.RegisterInbox(mailbox); err != nil {
					return fmt.Errorf("failed to register inbox: %w", err)
				}
				display.Success(fmt.Sprintf("Inbox ready: %s", mailbox))
				fmt.Println()
			} else {
				mailbox = arg
			}
		}

		return ws.Watch(apiKey, mailbox)
	},
}

// ── fce inbox ─────────────────────────────────────────────────────────────────

var inboxCmd = &cobra.Command{
	Use:   "inbox",
	Short: "Manage registered inboxes",
}

var inboxListCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "List all registered inboxes",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		inboxes, err := client.ListInboxes()
		if err != nil {
			return err
		}

		if len(inboxes) == 0 {
			display.Info("No inboxes registered.")
			display.Info("Add one with: fce inbox add <address>")
			return nil
		}

		display.Header(fmt.Sprintf("Inboxes  (%d)", len(inboxes)))
		items := make([]string, 0, len(inboxes))
		for _, item := range inboxes {
			if m, ok := item.(map[string]interface{}); ok {
				items = append(items, strVal(m, "inbox"))
			} else if s, ok := item.(string); ok {
				items = append(items, s)
			}
		}
		display.List(items)
		return nil
	},
}

var inboxAddCmd = &cobra.Command{
	Use:     "add <address>",
	Aliases: []string{"register"},
	Short:   "Register a new inbox",
	Example: `  fce inbox add mytest@ditmail.info
  fce inbox add random`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		inbox := args[0]
		if inbox == "random" {
			inbox = randomInbox()
			display.Info(fmt.Sprintf("Generated inbox: %s", inbox))
		}

		result, err := client.RegisterInbox(inbox)
		if err != nil {
			return err
		}

		registered := strVal(result, "inbox")
		if registered == "" {
			registered = inbox
		}
		display.Success(fmt.Sprintf("Registered: %s", registered))
		return nil
	},
}

var inboxRemoveCmd = &cobra.Command{
	Use:     "remove <address>",
	Aliases: []string{"rm", "delete", "unregister"},
	Short:   "Unregister an inbox",
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		if _, err := client.UnregisterInbox(args[0]); err != nil {
			return err
		}
		display.Success(fmt.Sprintf("Removed: %s", args[0]))
		return nil
	},
}

func init() {
	inboxCmd.AddCommand(inboxListCmd, inboxAddCmd, inboxRemoveCmd)
}

// ── fce messages ──────────────────────────────────────────────────────────────

var messagesCmd = &cobra.Command{
	Use:     "messages <inbox> [id]",
	Aliases: []string{"msgs", "mail"},
	Short:   "List messages in an inbox or view a specific message",
	Example: `  fce messages mytest@ditmail.info
  fce messages mytest@ditmail.info u7hPpV5sA`,
	Args:    cobra.MinimumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		inbox := args[0]

		// View specific message
		if len(args) > 1 {
			id := args[1]
			msg, err := client.GetMessage(inbox, id)
			if err != nil {
				return err
			}
			display.MessageContent(msg)
			return nil
		}

		// List messages
		msgs, err := client.ListMessages(inbox)
		if err != nil {
			return err
		}

		if len(msgs) == 0 {
			display.Info("No messages in this inbox.")
			return nil
		}

		display.Header(fmt.Sprintf("Messages in %s  (%d)", inbox, len(msgs)))
		fmt.Println()

		for _, item := range msgs {
			m, ok := item.(map[string]interface{})
			if !ok {
				continue
			}
			id      := strVal(m, "id")
			from    := strVal(m, "from")
			subject := strVal(m, "subject")
			date    := strVal(m, "date")
			otp     := strVal(m, "otp")

			display.Table([]display.Row{
				{Key: "ID",      Value: id},
				{Key: "From",    Value: from},
				{Key: "Subject", Value: subject},
				{Key: "Date",    Value: date},
				{Key: "OTP",     Value: func() string {
					if otp == "" || otp == "__DETECTED__" {
						return "—"
					}
					return otp
				}()},
			})
			display.Divider()
		}
		fmt.Println()
		return nil
	},
}

// ── fce otp ───────────────────────────────────────────────────────────────────

var otpCmd = &cobra.Command{
	Use:   "otp <inbox>",
	Short: "Get the latest OTP from an inbox  [Growth plan+]",
	Long: `Fetches the most recent one-time password from an inbox.
OTP is extracted automatically — no regex needed.

Requires Growth plan or above.`,
	Example: `  fce otp mytest@ditmail.info`,
	Args:    cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		// Plan gate
		me, err := client.GetMe()
		if err != nil {
			return err
		}
		d, _ := me["data"].(map[string]interface{})
		if d == nil {
			d = me
		}
		plan := strVal(d, "plan")
		if !api.HasPlan(plan, api.PlanGrowth) {
			display.PlanGate("growth", "OTP extraction")
			return nil
		}

		result, err := client.GetOTP(args[0])
		if err != nil {
			return err
		}

		otp  := strVal(result, "otp")
		link := strVal(result, "verification_link")
		from := strVal(result, "from")
		subj := strVal(result, "subject")
		ts   := fmt.Sprintf("%v", result["timestamp"])

		if otp == "" || otp == "null" {
			display.Info("No OTP found in recent messages.")
			return nil
		}

		display.Header("OTP")
		rows := []display.Row{
			{Key: "OTP",  Value: otp},
			{Key: "From", Value: from},
			{Key: "Subj", Value: subj},
			{Key: "Time", Value: ts},
		}
		if link != "" && link != "null" {
			rows = append(rows, display.Row{Key: "Link", Value: link})
		}
		display.Table(rows)
		return nil
	},
}

// ── fce domains ───────────────────────────────────────────────────────────────

var domainsCmd = &cobra.Command{
	Use:   "domains",
	Short: "List available domains on your plan",
	RunE: func(cmd *cobra.Command, args []string) error {
		client, err := requireAuth()
		if err != nil {
			return err
		}

		domains, err := client.ListDomains()
		if err != nil {
			return err
		}

		if len(domains) == 0 {
			display.Info("No domains available.")
			return nil
		}

		display.Header(fmt.Sprintf("Available Domains  (%d)", len(domains)))
		items := make([]string, 0, len(domains))
		for _, item := range domains {
			if m, ok := item.(map[string]interface{}); ok {
				domain := strVal(m, "domain")
				tier   := strVal(m, "tier")
				suffix := ""
				if tier == "pro" {
					suffix = "  " + display.PlanBadge("pro")
				}
				items = append(items, domain+suffix)
			}
		}
		display.List(items)
		return nil
	},
}

// ── Helpers ───────────────────────────────────────────────────────────────────

func requireAuth() (*api.Client, error) {
	client, err := api.New()
	if err != nil {
		display.Error("Not logged in.")
		display.Info("Run: fce login")
		return nil, fmt.Errorf("not authenticated")
	}
	return client, nil
}

func strVal(m map[string]interface{}, key string) string {
	v, _ := m[key].(string)
	return v
}

func intVal(m map[string]interface{}, key string) int {
	switch v := m[key].(type) {
	case float64:
		return int(v)
	case int:
		return v
	}
	return 0
}

// randomInbox generates a random inbox address using a free-tier domain
var randomDomains = []string{
	"ditube.info", "ditmail.info", "ditapi.info",
	"ditgame.info", "ditplay.info", "ditcloud.info",
}

var adjectives = []string{
	"swift", "clear", "quiet", "bright", "calm",
	"sharp", "bold", "cool", "crisp", "light",
}

var nouns = []string{
	"fox", "hawk", "mint", "wave", "peak",
	"pine", "vale", "reef", "beam", "dusk",
}

func randomInbox() string {
	rng    := rand.New(rand.NewSource(time.Now().UnixNano()))
	adj    := adjectives[rng.Intn(len(adjectives))]
	noun   := nouns[rng.Intn(len(nouns))]
	n      := rng.Intn(9000) + 1000
	domain := randomDomains[rng.Intn(len(randomDomains))]
	return strings.ToLower(fmt.Sprintf("%s%s%d@%s", adj, noun, n, domain))
}

func devInbox() string {
	rng    := rand.New(rand.NewSource(time.Now().UnixNano()))
	chars  := "abcdefghijklmnopqrstuvwxyz0123456789"
	suffix := make([]byte, 4)
	for i := range suffix {
		suffix[i] = chars[rng.Intn(len(chars))]
	}
	domain := randomDomains[rng.Intn(len(randomDomains))]
	return fmt.Sprintf("dev-%s@%s", string(suffix), domain)
}
