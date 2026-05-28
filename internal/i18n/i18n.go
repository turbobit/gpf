package i18n

import (
	"embed"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

//go:embed *.json
var fs embed.FS

var defaultLocale = Locale()

// forcedLocale, when non-empty, overrides everything.
// Set via Force() from the --lang/-l CLI flag.
var forcedLocale string

// configPath is ~/.gpf/lang
var configPath = func() string {
	home, _ := os.UserHomeDir()
	if home == "" {
		home = "."
	}
	dir := filepath.Join(home, ".gpf")
	return filepath.Join(dir, "lang")
}()

// Translator provides access to translated strings.
type Translator struct {
	msgs map[string]string
}

// T returns the translated string for a key. Falls back to the key itself if not found.
func (t *Translator) T(key string) string {
	if v, ok := t.msgs[key]; ok {
		return v
	}
	return key
}

// Force overrides the auto-detected locale and persists it to disk.
// Called from the --lang/-l CLI flag. Has no effect if locale is empty.
func Force(locale string) {
	if locale != "" {
		forcedLocale = locale
		Save(locale)
	}
}

// Default returns a Translator using this priority:
// 1. --lang flag (Force)
// 2. Saved locale (~/.gpf/lang)
// 3. Auto-detected from env (LANG, etc.)
// 4. English fallback
func Default() *Translator {
	if forcedLocale != "" {
		return MustLoad(forcedLocale)
	}
	if saved := Loaded(); saved != "" {
		return MustLoad(saved)
	}
	return MustLoad(defaultLocale)
}

// For returns a Translator for a specific locale.
func For(locale string) *Translator {
	return MustLoad(locale)
}

// MustLoad loads and returns a Translator for the given locale.
// Returns English fallback if the locale is not available.
func MustLoad(locale string) *Translator {
	data, err := fs.ReadFile(locale + ".json")
	if err != nil {
		// Fallback to English
		data, err = fs.ReadFile("en.json")
		if err != nil {
			return &Translator{msgs: map[string]string{}}
		}
	}

	var msgs map[string]string
	if err := json.Unmarshal(data, &msgs); err != nil {
		return &Translator{msgs: map[string]string{}}
	}
	return &Translator{msgs: msgs}
}

// Save writes the locale to ~/.gpf/lang so it persists across sessions.
func Save(locale string) {
	dir := filepath.Dir(configPath)
	os.MkdirAll(dir, 0700)
	os.WriteFile(configPath, []byte(locale), 0600)
}

// Loaded reads the saved locale from ~/.gpf/lang.
// Returns empty string if no saved locale exists.
func Loaded() string {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}

// Locale detects the user's preferred locale from environment.
// Checks LANG, LANGUAGE, LC_ALL, LC_MESSAGES in order.
// Returns "en" as the default fallback.
func Locale() string {
	for _, env := range []string{"LANG", "LANGUAGE", "LC_ALL", "LC_MESSAGES"} {
		if v := os.Getenv(env); v != "" {
			return parseLocale(v)
		}
	}
	return "en"
}

// parseLocale extracts the short locale code from a full locale string.
// "ko_KR.UTF-8" → "ko", "en_US.utf8" → "en", "ja" → "ja"
func parseLocale(s string) string {
	// Remove encoding suffix (e.g., .UTF-8, .utf8)
	s = strings.SplitN(s, ".", 2)[0]

	// Remove variant (e.g., pt_BR_LATN → pt_BR)
	parts := strings.SplitN(s, "@", 2)
	s = parts[0]

	// Full locale like ko_KR → try ko first, then pt-BR
	parts = strings.SplitN(s, "_", 2)
	primary := strings.ToLower(parts[0])

	// Check if this locale file exists
	if _, err := fs.ReadFile(primary + ".json"); err == nil {
		return primary
	}

	// Also try the full code (e.g., pt-BR)
	full := strings.ToLower(strings.ReplaceAll(s, "_", "-"))
	if _, err := fs.ReadFile(full + ".json"); err == nil {
		return full
	}

	return primary
}
