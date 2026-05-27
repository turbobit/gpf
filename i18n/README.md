# gpf Internationalization

This directory contains translation files for gpf's user-facing strings.

## How Language is Detected

gpf automatically detects the UI language from these environment variables (in order):

1. `LANG` (e.g., `ko_KR.UTF-8` → `ko.json`)
2. `LANGUAGE`
3. `LC_ALL`
4. `LC_MESSAGES`

The locale string is parsed to extract the primary language code (e.g., `ko_KR.UTF-8` → `ko`, `pt_BR.UTF-8` → `pt-BR`). If the detected locale file does not exist, English (`en.json`) is used as the fallback.

### Override Language

```bash
# Temporarily switch to Korean
LANG=ko_KR.UTF-8 gpf

# Temporarily switch to English
LANG=en_US.UTF-8 gpf
```

## File Format

- **Format**: JSON, UTF-8 encoded.
- **Keys**: Must match the English (`en.json`) keys exactly. Never rename or remove keys.
- **Values**: Translated text for the corresponding key.

## Adding a New Language

1. **Copy** `en.json` to a new file named `<locale>.json` (e.g., `ja.json`, `fr.json`, `de.json`).
2. **Translate** all string values while keeping keys unchanged.
3. **Save** as UTF-8 encoded JSON.
4. **Test** by running `LANG=<locale>.UTF-8 gpf` to verify all strings display correctly.
5. **Open a Pull Request** with your new translation file.

### Locale Naming Conventions

Use standard BCP 47 locale codes:

| Example | File | Locale |
|---------|------|--------|
| English | `en.json` | `en`, `en_US.UTF-8` |
| Korean | `ko.json` | `ko`, `ko_KR.UTF-8` |
| Japanese | `ja.json` | `ja`, `ja_JP.UTF-8` |
| French | `fr.json` | `fr`, `fr_FR.UTF-8` |
| Brazilian Portuguese | `pt-BR.json` | `pt_BR`, `pt_BR.UTF-8` |
| German | `de.json` | `de`, `de_DE.UTF-8` |

## Contributing

We welcome community translations. To contribute:

- Fork the repository.
- Add or update your `i18n/<locale>.json` file.
- Ensure the JSON is valid (use a JSON linter).
- Submit a Pull Request with a description of the language and any context-specific notes.

If you notice missing or incorrect strings in an existing translation, feel free to open a PR with the fix.
