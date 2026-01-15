# Typst Test Templates

This directory contains test templates organized by use case. Each template is in its own directory with all required assets (images, fonts, data files, etc.).

## Structure

```
test/typst/
├── aero-navigator/          # Aviation navigation log template
├── complete-document/       # Multi-page document with various features
├── dashing-dept-news/       # Newsletter template with images
├── elspub/                  # Academic publication with JSON and Markdown
├── high-resolution/         # High-res output test (--ppi 300)
├── invalid-test/            # Invalid syntax for error testing
├── options-test/            # Options parameter testing
├── plots-document/          # Data visualization with plotst
├── response-info-test/      # Response metadata testing
└── simple-test/             # Basic template for quick tests
```

## Template Organization

Each template directory contains:
- `<name>.typ` - The main Typst template file
- Additional assets as needed (images, fonts, JSON, etc.)

## Usage in Tests

Templates are loaded using the `loadTemplate()` helper:

```go
templatePath := loadTemplate(t, templateSimpleTest)
```

Media files are loaded with `loadMediaFiles()`:

```go
media := loadMediaFiles(t, map[string]string{
    "image.png": "test/typst/example/image.png",
})
```

## Adding New Templates

1. Create a new directory: `test/typst/my-template/`
2. Add the template file: `my-template.typ`
3. Add any required assets (images, fonts, etc.)
4. Add a constant in `typst-pdf-generator_test.go`:
   ```go
   templateMyTemplate = "my-template/my-template.typ"
   ```
5. Write your test using the helper functions
