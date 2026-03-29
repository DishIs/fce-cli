package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	// Output formatting flags globally accessible
	GlobalFormat string = "text"
	GlobalSilent bool   = false
)

func IsSilent() bool {
	return GlobalSilent || GlobalFormat != "text"
}

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	// Adaptive colors for both light and dark terminals
	colorDim    = lipgloss.AdaptiveColor{Light: "244", Dark: "240"}
	colorBright = lipgloss.AdaptiveColor{Light: "235", Dark: "255"}
	colorMuted  = lipgloss.AdaptiveColor{Light: "248", Dark: "245"}
	colorAccent = lipgloss.AdaptiveColor{Light: "232", Dark: "255"}
	colorBorder = lipgloss.AdaptiveColor{Light: "250", Dark: "238"}

	styleDim     = lipgloss.NewStyle().Foreground(colorDim)
	styleBright  = lipgloss.NewStyle().Foreground(colorBright).Bold(true)
	styleMuted   = lipgloss.NewStyle().Foreground(colorMuted)
	styleAccent  = lipgloss.NewStyle().Foreground(colorAccent).Bold(true)
	styleBorder  = lipgloss.NewStyle().Foreground(colorBorder)
	styleSuccess = lipgloss.NewStyle().Foreground(colorBright)
	styleError   = lipgloss.NewStyle().Foreground(colorBright).Bold(true)
	styleWarn    = lipgloss.NewStyle().Foreground(colorDim)

	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(colorBorder).
			Padding(0, 2)
)

// ── Logo ──────────────────────────────────────────────────────────────────────

// ASCII logo optimized for readability and Go raw string constraints
// Using @@ as a placeholder for backticks
const logoASCII = `
    ______              ______           __                       ______                _ __
   / ____/________ ___ / ____/_  _______/ /_____  ____ ___       / ____/___ ___  ____ _(_) /
  / /_  / ___/ _ \/ _ \ /   / / / / ___/ __/ __ \/ __ @@__ \    / __/ / __ @@__ \/ __ @@/ /
 / __/ / /  /  __/  __/ /___/ /_/ (__  ) /_/ /_/ / / / / / /   / /___/ / / / / / /_/ / / /  
/_/   /_/   \___/\___/\____/\__,_/____/\__/\____/_/ /_/ /_/( )/_____/_/ /_/ /_/\__,_/_/_/   
`

const tagline = `   FreeCustom.Email — disposable inbox API`

// PrintLogo prints the full logo + wordmark on first login
func PrintLogo() {
	if IsSilent() { return }
	fixedLogo := strings.ReplaceAll(logoASCII, "@@", "`")
	logo := styleBright.Render(fixedLogo)
	tag := styleMuted.Render(tagline)
	fmt.Println(logo)
	fmt.Println(tag)
	fmt.Println()
}

// PrintInlineLogo prints a compact single-line logo for command headers
func PrintInlineLogo() {
	if IsSilent() { return }
	icon := styleBright.Render("◉")
	name := styleBright.Render("fce")
	fmt.Printf("%s %s  ", icon, name)
}

// ── Section headers ───────────────────────────────────────────────────────────

func Header(title string) {
	if IsSilent() { return }
	bar  := styleDim.Render(strings.Repeat("─", 48))
	head := styleAccent.Render(title)
	fmt.Println()
	fmt.Println(bar)
	fmt.Printf("  %s\n", head)
	fmt.Println(bar)
}

// ── Status messages ───────────────────────────────────────────────────────────

func Success(msg string) {
	if IsSilent() { return }
	fmt.Printf("  %s  %s\n", styleBright.Render("✓"), styleSuccess.Render(msg))
}

func Error(msg string) {
	fmt.Printf("  %s  %s\n", styleBright.Render("✗"), styleError.Render(msg))
}

func Warn(msg string) {
	if IsSilent() { return }
	fmt.Printf("  %s  %s\n", styleBright.Render("!"), styleWarn.Render(msg))
}

func Info(msg string) {
	if IsSilent() { return }
	fmt.Printf("  %s  %s\n", styleDim.Render("·"), styleMuted.Render(msg))
}

func Step(n int, total int, msg string) {
	if IsSilent() { return }
	counter := styleDim.Render(fmt.Sprintf("[%d/%d]", n, total))
	fmt.Printf("  %s  %s\n", counter, styleMuted.Render(msg))
}

// ── Tables ────────────────────────────────────────────────────────────────────

type Row struct {
	Key   string
	Value string
}

func Table(rows []Row) {
	if IsSilent() { return }
	maxKey := 0
	for _, r := range rows {
		if len(r.Key) > maxKey {
			maxKey = len(r.Key)
		}
	}
	fmt.Println()
	for _, r := range rows {
		pad   := strings.Repeat(" ", maxKey-len(r.Key)+2)
		key   := styleDim.Render(r.Key)
		sep   := styleBorder.Render("·")
		value := styleBright.Render(r.Value)
		fmt.Printf("  %s%s%s  %s\n", key, pad, sep, value)
	}
	fmt.Println()
}

// ── List ──────────────────────────────────────────────────────────────────────

func List(items []string) {
	if IsSilent() { return }
	fmt.Println()
	for i, item := range items {
		n   := styleDim.Render(fmt.Sprintf("%02d", i+1))
		dot := styleBorder.Render("·")
		fmt.Printf("  %s %s %s\n", n, dot, styleBright.Render(item))
	}
	fmt.Println()
}

// ── Plan badge ────────────────────────────────────────────────────────────────

func PlanBadge(plan string) string {
	p := strings.ToLower(plan)
	if p == "free" || p == "developer" {
		return styleDim.Render("[" + strings.ToUpper(plan) + "]")
	}

	// High visibility solid badge for paid plans
	return lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}).
		Background(lipgloss.AdaptiveColor{Light: "235", Dark: "255"}).
		Padding(0, 1).
		Bold(true).
		Render(strings.ToUpper(plan))
}

// ── Live event (for watch) ────────────────────────────────────────────────────

func EmailEvent(id, from, subject, otp, link string, timestamp string) {
	div   := styleDim.Render(strings.Repeat("─", 52))
	fmt.Println()
	fmt.Println(div)
	fmt.Printf("  %s  %s\n", styleDim.Render("ID  "), styleDim.Render(id))
	fmt.Printf("  %s  %s\n", styleDim.Render("FROM"), styleBright.Render(from))
	fmt.Printf("  %s  %s\n", styleDim.Render("SUBJ"), styleMuted.Render(subject))
	fmt.Printf("  %s  %s\n", styleDim.Render("TIME"), styleDim.Render(timestamp))
	if otp != "" {
		otpVal := styleAccent.Render(otp)
		fmt.Printf("  %s  %s\n", styleBright.Render("OTP "), otpVal)
	}
	if link != "" {
		button := lipgloss.NewStyle().
			Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}).
			Background(lipgloss.AdaptiveColor{Light: "235", Dark: "255"}).
			Padding(0, 1).
			Bold(true).
			Render("OPEN EMAIL")
		fmt.Printf("\n  %s  %s\n", button, styleDim.Render(link))
	}
	fmt.Println(div)
	fmt.Println()
}

func MessageContent(data map[string]interface{}) {
	from    := fmt.Sprintf("%v", data["from"])
	subject := fmt.Sprintf("%v", data["subject"])
	date    := fmt.Sprintf("%v", data["date"])
	text    := fmt.Sprintf("%v", data["text"])
	body    := fmt.Sprintf("%v", data["body"])
	html    := fmt.Sprintf("%v", data["html"])

	Header("Message Details")
	Table([]Row{
		{Key: "From",    Value: from},
		{Key: "Subject", Value: subject},
		{Key: "Date",    Value: date},
	})

	if text != "" && text != "<nil>" {
		fmt.Println("\n" + styleDim.Render("── Text Content ──────────────────────────────────────────────"))
		fmt.Println(text)
	} else if body != "" && body != "<nil>" {
		fmt.Println("\n" + styleDim.Render("── Body ──────────────────────────────────────────────────────"))
		fmt.Println(body)
	} else if html != "" && html != "<nil>" {
		fmt.Println("\n" + styleDim.Render("── HTML (Source) ─────────────────────────────────────────────"))
		fmt.Println(html)
	}
	fmt.Println()
}

// ── Waiting spinner (simple) ──────────────────────────────────────────────────

func Waiting(msg string) {
	if IsSilent() { return }
	fmt.Printf("  %s  %s\n", styleDim.Render("◌"), styleMuted.Render(msg))
}

// ── Plan gate error ───────────────────────────────────────────────────────────

func PlanGate(requiredPlan string, feature string) {
	fmt.Println()
	Error(fmt.Sprintf("%s requires %s plan or above.", feature, strings.Title(requiredPlan)))
	Info("Upgrade at: https://www.freecustom.email/api/pricing")
	fmt.Println()
}

// ── Divider ───────────────────────────────────────────────────────────────────

func Divider() {
	if IsSilent() { return }
	fmt.Println(styleDim.Render("  " + strings.Repeat("─", 48)))
}

// TableBadge returns a formatted string with a badge and value
func TableBadge(label, value string) string {
	badge := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "255", Dark: "0"}).
		Background(lipgloss.AdaptiveColor{Light: "244", Dark: "240"}).
		Padding(0, 1).
		Bold(true).
		Width(10).
		Align(lipgloss.Center).
		Render(strings.ToUpper(label))

	return fmt.Sprintf("  %s  %s", badge, styleBright.Render(value))
}
