package display

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	Title   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	Success = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	Error   = lipgloss.NewStyle().Foreground(lipgloss.Color("1"))
	Warn    = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	Dim     = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	Logo    = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
)

func Banner() string {
	return Title.Render("Archcore") + Dim.Render(" — Shared architectural memory for AI coding agents")
}

func WelcomeBanner() string {
	logoLines := []string{
		"╔══════╗",
		"║      ║",
		"║  ╔═══╣",
		"║  ║ ╔═╣",
		"╚══╩═╩═╝",
	}
	logo := Logo.Render(strings.Join(logoLines, "\n"))

	textLines := []string{
		Title.Render("Archcore — Shared architectural memory for AI coding agents"),
		Dim.Render("Context engineering for your codebase"),
		Dim.Render("https://archcore.ai"),
	}
	text := strings.Join(textLines, "\n")

	return lipgloss.JoinHorizontal(lipgloss.Center, logo, "   ", text)
}

func CheckLine(msg string) string {
	return Success.Render("  ✓ ") + msg
}

func FailLine(msg string) string {
	return Error.Render("  ✗ ") + msg
}

func WarnLine(msg string) string {
	return Warn.Render("  ! ") + msg
}

func HintLine(msg string) string {
	return Dim.Render("    → ") + Dim.Render(msg)
}

func KeyValue(key, value string) string {
	return fmt.Sprintf("  %s %s", Dim.Render(key+":"), value)
}

func HookConnectedLine(version string, docCount int) string {
	return fmt.Sprintf("Archcore %s · MCP connected · %d docs", version, docCount)
}
