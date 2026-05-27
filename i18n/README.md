# gpf Internationalization

This directory contains translation files for gpf's user-facing strings.

## File Format

- **Format**: JSON, UTF-8 encoded.
- **Keys**: Must match the English (`en.json`) keys exactly. Never rename or remove keys.
- **Values**: Translated text for the corresponding key.

## Adding a New Language

1. **Copy** `en.json` to a new file named `<locale>.json` (e.g., `ja.json`, `fr.json`, `de.json`).
2. **Translate** all string values while keeping keys unchanged.
3. **Save** as UTF-8 encoded JSON.
4. **Test** by running gpf with the new locale to verify all strings display correctly.
5. **Open a Pull Request** with your new translation file.

### Locale Naming Conventions

Use standard BCP 47 locale codes:

| Example | File        |
|---------|-------------|
| English | `en.json`   |
| Korean  | `ko.json`   |
| Japanese| `ja.json`   |
| French  | `fr.json`   |
| Brazilian Portuguese | `pt-BR.json` |
| German  | `de.json`   |

## Contributing

We welcome community translations. To contribute:

- Fork the repository.
- Add or update your `i18n/<locale>.json` file.
- Ensure the JSON is valid (use a JSON linter).
- Submit a Pull Request with a description of the language and any context-specific notes.

If you notice missing or incorrect strings in an existing translation, feel free to open a PR with the fix.
