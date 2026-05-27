package command

import "strings"

// Action represents a gpf command mode.
type Action int

const (
	InteractiveConfig Action = iota
	ShowPorts
	CreateTunnel
	ShowTunnels
	StopTunnel
	StopAllTunnels
)

// Options holds parsed CLI options.
type Options struct {
	Action  Action
	Value   string
	Locale  string // empty = auto-detect from LANG
}

// Which parses CLI args and returns the action, value, and options.
func Which(args []string) Options {
	// Strip --lang / -l flag from args
	locale, remaining := extractLocale(args)

	if len(remaining) == 0 {
		return Options{Action: InteractiveConfig, Locale: locale}
	}

	// "gpf -" or "gpf - keyword"
	if remaining[0] == "-" {
		keyword := strings.Join(remaining[1:], " ")
		return Options{Action: InteractiveConfig, Value: keyword, Locale: locale}
	}

	// Subcommands
	switch remaining[0] {
	case "ports":
		return Options{Action: ShowPorts, Value: strings.Join(remaining[1:], " "), Locale: locale}
	case "forward":
		return Options{Action: CreateTunnel, Value: strings.Join(remaining[1:], " "), Locale: locale}
	case "tunnels":
		return Options{Action: ShowTunnels, Locale: locale}
	case "stop":
		if len(remaining) < 2 {
			return Options{Action: InteractiveConfig, Locale: locale}
		}
		return Options{Action: StopTunnel, Value: remaining[1], Locale: locale}
	case "stop-all":
		return Options{Action: StopAllTunnels, Locale: locale}
	case "version", "-v", "--version":
		return Options{Action: InteractiveConfig, Value: remaining[0], Locale: locale}
	default:
		// "gpf mac" → search ~/.ssh/config for "mac"
		return Options{Action: InteractiveConfig, Value: strings.Join(remaining, " "), Locale: locale}
	}
}

// extractLocale parses --lang <locale> or -l <locale> from args.
// Returns the locale and the remaining args with the flag removed.
func extractLocale(args []string) (string, []string) {
	var locale string
	var remaining []string

	for i := 0; i < len(args); i++ {
		if args[i] == "--lang" || args[i] == "-l" {
			if i+1 < len(args) {
				i++
				locale = strings.ToLower(args[i])
				// Normalize: ko_KR → ko, pt_BR → pt-BR
				parts := strings.SplitN(locale, "_", 2)
				locale = parts[0]
				if len(parts) == 2 {
					locale = parts[0] + "-" + strings.ToUpper(parts[1])
				}
			}
			continue
		}
		// Also handle --lang=ko or -l=ko
		if strings.HasPrefix(args[i], "--lang=") {
			locale = strings.ToLower(strings.TrimPrefix(args[i], "--lang="))
			parts := strings.SplitN(locale, "_", 2)
			locale = parts[0]
			if len(parts) == 2 {
				locale = parts[0] + "-" + strings.ToUpper(parts[1])
			}
			continue
		}
		if strings.HasPrefix(args[i], "-l=") {
			locale = strings.ToLower(strings.TrimPrefix(args[i], "-l="))
			parts := strings.SplitN(locale, "_", 2)
			locale = parts[0]
			if len(parts) == 2 {
				locale = parts[0] + "-" + strings.ToUpper(parts[1])
			}
			continue
		}
		remaining = append(remaining, args[i])
	}

	return locale, remaining
}
