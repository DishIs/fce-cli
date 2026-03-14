package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ── Styles ────────────────────────────────────────────────────────────────────

var (
	// Neutral palette — works on both light and dark terminals
	styleDim     = lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	styleBright  = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	styleMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
	styleAccent  = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	styleBorder  = lipgloss.NewStyle().Foreground(lipgloss.Color("238"))
	styleSuccess = lipgloss.NewStyle().Foreground(lipgloss.Color("255"))
	styleError   = lipgloss.NewStyle().Foreground(lipgloss.Color("255")).Bold(true)
	styleWarn    = lipgloss.NewStyle().Foreground(lipgloss.Color("250"))

	// Box style for info panels
	panelStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("238")).
			Padding(0, 2)

	// Plan badge
	planStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("242")).
			Padding(0, 1).
			Foreground(lipgloss.Color("252"))
)

// ── Logo ──────────────────────────────────────────────────────────────────────
// ASCII recreation of the FreeCustom.Email envelope-in-circle logo

const logoASCII = `
    ╭─────────────────────╮
    │  ╭───────────────╮  │
    │  │ ╲           ╱ │  │
    │  │   ╲       ╱   │  │
    │  │     ╲   ╱     │  │
    │  │      ╲╱       │  │
    │  │               │  │
    │  ╰───────────────╯  │
    ╰─────────────────────╯`

const wordmark = `  FreeCustom.Email`
const tagline  = `  disposable inbox API`

// PrintLogo prints the full logo + wordmark on first login
func PrintLogo() {
	logo := styleBright.Render(logoASCII)
	wm   := styleBright.Render(wordmark)
	tag  := styleMuted.Render(tagline)
	fmt.Println(logo)
	fmt.Println(wm)
	fmt.Println(tag)
	fmt.Println()
}

// PrintInlineLogo prints a compact single-line logo for command headers
func PrintInlineLogo() {
	icon := styleBright.Render("◉")
	name := styleBright.Render("fce")
	fmt.Printf("%s %s  ", icon, name)
}

// ── Section headers ───────────────────────────────────────────────────────────

func Header(title string) {
	bar  := styleDim.Render(strings.Repeat("─", 48))
	head := styleAccent.Render(title)
	fmt.Println()
	fmt.Println(bar)
	fmt.Printf("  %s\n", head)
	fmt.Println(bar)
}

// ── Status messages ───────────────────────────────────────────────────────────

func Success(msg string) {
	fmt.Printf("  %s  %s\n", styleBright.Render("✓"), styleSuccess.Render(msg))
}

func Error(msg string) {
	fmt.Printf("  %s  %s\n", styleBright.Render("✗"), styleError.Render(msg))
}

func Warn(msg string) {
	fmt.Printf("  %s  %s\n", styleBright.Render("!"), styleWarn.Render(msg))
}

func Info(msg string) {
	fmt.Printf("  %s  %s\n", styleDim.Render("·"), styleMuted.Render(msg))
}

func Step(n int, total int, msg string) {
	counter := styleDim.Render(fmt.Sprintf("[%d/%d]", n, total))
	fmt.Printf("  %s  %s\n", counter, styleMuted.Render(msg))
}

// ── Tables ────────────────────────────────────────────────────────────────────

type Row struct {
	Key   string
	Value string
}

func Table(rows []Row) {
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
	return planStyle.Render(strings.ToUpper(plan))
}

// ── Live event (for watch) ────────────────────────────────────────────────────

func EmailEvent(from, subject, otp, link string, timestamp string) {
	div   := styleDim.Render(strings.Repeat("─", 52))
	fmt.Println()
	fmt.Println(div)
	fmt.Printf("  %s  %s\n", styleDim.Render("FROM"), styleBright.Render(from))
	fmt.Printf("  %s  %s\n", styleDim.Render("SUBJ"), styleMuted.Render(subject))
	fmt.Printf("  %s  %s\n", styleDim.Render("TIME"), styleDim.Render(timestamp))
	if otp != "" && otp != "__DETECTED__" && otp != "null" {
		otpVal := styleAccent.Render(otp)
		fmt.Printf("  %s  %s\n", styleBright.Render("OTP "), otpVal)
	}
	if link != "" && link != "null" {
		fmt.Printf("  %s  %s\n", styleDim.Render("LINK"), styleDim.Render(link))
	}
	fmt.Println(div)
	fmt.Println()
}

// ── Waiting spinner (simple) ──────────────────────────────────────────────────

func Waiting(msg string) {
	fmt.Printf("  %s  %s\n", styleDim.Render("◌"), styleMuted.Render(msg))
}

// ── Plan gate error ───────────────────────────────────────────────────────────

func PlanGate(requiredPlan string, feature string) {
	fmt.Println()
	Error(fmt.Sprintf("%s requires %s plan or above.", feature, strings.Title(requiredPlan)))
	Info("Upgrade at: https://freecustom.email/api/pricing")
	fmt.Println()
}

// ── Divider ───────────────────────────────────────────────────────────────────

func Divider() {
	fmt.Println(styleDim.Render("  " + strings.Repeat("─", 48)))
}
