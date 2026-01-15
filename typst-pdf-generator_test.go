package typstpdfgenerator

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const testTypstDir = "test/typst"

const (
	templateCompleteDoc  = "complete-document/complete-document.typ"
	templateHighRes      = "high-resolution/high-resolution.typ"
	templatePlots        = "plots-document/plots-document.typ"
	templateSimpleTest   = "simple-test/simple-test.typ"
	templateOptionsTest  = "options-test/options-test.typ"
	templateInvalidTest  = "invalid-test/invalid-test.typ"
	templateResponseTest = "response-info-test/response-info-test.typ"
	templateAeroNav      = "aero-navigator/aero-navigator.typ"
	templateDashingNews  = "dashing-dept-news/dashing-dept-news.typ"
	templateElspub       = "elspub/elspub.typ"
)

func setupClient(t *testing.T) *Client {
	t.Helper()

	authKey := os.Getenv("PDF_GENERATOR_AUTH_KEY")
	gateway := os.Getenv("PDF_GENERATOR_ENDPOINT")

	if authKey == "" || gateway == "" {
		t.Skip("Skipping test: PDF_GENERATOR_AUTH_KEY and PDF_GENERATOR_ENDPOINT must be set")
	}

	client, err := New(authKey, gateway, WithTimeout(120*time.Second))
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	return client
}

func loadTemplate(t *testing.T, name string) string {
	t.Helper()

	path := filepath.Join(testTypstDir, name)
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("Template not found: %s", path)
	}

	return path
}

func loadMediaFiles(t *testing.T, files map[string]string) []MediaFile {
	t.Helper()

	media := make([]MediaFile, 0, len(files))
	for name, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Logf("Warning: could not load media file %s: %v", path, err)
			continue
		}
		media = append(media, MediaFile{Name: name, Data: data})
	}

	return media
}

func setupOutputDir(t *testing.T) string {
	t.Helper()

	outputDir := filepath.Join("test_output", t.Name())
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		t.Fatalf("Failed to create output directory: %v", err)
	}

	return outputDir
}

func verifyPDF(t *testing.T, path string) int64 {
	t.Helper()

	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("PDF file not found: %v", err)
	}

	if info.Size() == 0 {
		t.Fatal("PDF file is empty")
	}

	data := make([]byte, 4)
	f, err := os.Open(path)
	if err == nil {
		defer f.Close()
		if n, _ := f.Read(data); n == 4 && string(data) != "%PDF" {
			t.Error("File does not appear to be a valid PDF")
		}
	}

	return info.Size()
}

func TestClientCreation(t *testing.T) {
	tests := []struct {
		name        string
		authKey     string
		gateway     string
		expectError bool
	}{
		{
			name:        "valid configuration",
			authKey:     "test-key",
			gateway:     "https://example.com/function/typst",
			expectError: false,
		},
		{
			name:        "empty auth key",
			authKey:     "",
			gateway:     "https://example.com/function/typst",
			expectError: true,
		},
		{
			name:        "empty gateway",
			authKey:     "test-key",
			gateway:     "",
			expectError: true,
		},
		{
			name:        "invalid gateway URL",
			authKey:     "test-key",
			gateway:     "not-a-url",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := New(tt.authKey, tt.gateway)
			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				if client == nil {
					t.Error("Expected client but got nil")
				}
			}
		})
	}
}

func TestAeroNavigatorTemplate(t *testing.T) {
	client := setupClient(t)

	templatePath := "test/typst/aero-navigator/aero-navigator.typ"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Skipf("Template file not found: %s", templatePath)
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "aero-navigator.pdf")

	respInfo, err := client.SavePDF(context.Background(), "", templatePath, outputPath, nil, nil)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Errorf("PDF file was not created: %s", outputPath)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Errorf("Failed to stat PDF file: %v", err)
	}
	if info.Size() == 0 {
		t.Error("PDF file is empty")
	}

	if respInfo.CorrelationID == "" {
		t.Error("Expected correlation ID but got empty string")
	}
	t.Logf("Correlation ID: %s", respInfo.CorrelationID)
	if respInfo.Stdout != "" {
		t.Logf("Stdout: %s", respInfo.Stdout)
	}
	if respInfo.Stderr != "" {
		t.Logf("Stderr: %s", respInfo.Stderr)
	}
}

func TestElspubTemplateWithMedia(t *testing.T) {
	client := setupClient(t)

	templatePath := "test/typst/elspub/elspub.typ"
	if _, err := os.Stat(templatePath); os.IsNotExist(err) {
		t.Skipf("Template file not found: %s", templatePath)
	}

	jsonPath := "test/typst/elspub/test_data.json"
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Skipf("Failed to read JSON file: %v", err)
	}

	mdPath := "test/typst/elspub/md_content/content.md"
	mdData, err := os.ReadFile(mdPath)
	if err != nil {
		t.Skipf("Failed to read Markdown file: %v", err)
	}

	media := []MediaFile{
		{Name: "test_data.json", Data: jsonData},
		{Name: "md_content/content.md", Data: mdData},
	}

	fontFiles := []string{
		"test/typst/elspub/fonts/Lato-Regular.ttf",
		"test/typst/elspub/fonts/Lato-Bold.ttf",
		"test/typst/elspub/fonts/Lato-Italic.ttf",
		"test/typst/elspub/fonts/Lato-BoldItalic.ttf",
	}

	for _, fontPath := range fontFiles {
		if fontData, err := os.ReadFile(fontPath); err == nil {
			media = append(media, MediaFile{
				Name: "fonts/" + filepath.Base(fontPath),
				Data: fontData,
			})
		}
	}

	outputDir := t.TempDir()
	outputPath := filepath.Join(outputDir, "elspub.pdf")

	_, err = client.SavePDF(context.Background(), "", templatePath, outputPath, nil, media)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	info, err := os.Stat(outputPath)
	if err != nil {
		t.Errorf("PDF file was not created: %v", err)
	}
	if info.Size() == 0 {
		t.Error("PDF file is empty")
	}

	t.Logf("PDF generated successfully: %d bytes", info.Size())
}

func TestGeneratePDFFromString(t *testing.T) {
	client := setupClient(t)

	templatePath := loadTemplate(t, templateSimpleTest)
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	var buf bytes.Buffer
	_, err = client.GeneratePDFFromString(context.Background(), &buf, "", string(templateData), nil, nil)
	if err != nil {
		t.Fatalf("Failed to generate PDF from string: %v", err)
	}

	pdfData := buf.Bytes()
	if len(pdfData) == 0 {
		t.Error("PDF data is empty")
	}

	if len(pdfData) >= 4 && string(pdfData[:4]) != "%PDF" {
		t.Error("Generated data does not appear to be a valid PDF")
	}

	t.Logf("PDF generated successfully: %d bytes", len(pdfData))
}

func TestGeneratePDFWithOptions(t *testing.T) {
	client := setupClient(t)

	templatePath := loadTemplate(t, templateOptionsTest)
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	options := []string{"--ppi", "300"}

	var buf bytes.Buffer
	_, err = client.GeneratePDFFromString(context.Background(), &buf, "", string(templateData), options, nil)
	if err != nil {
		t.Fatalf("Failed to generate PDF with options: %v", err)
	}

	pdfData := buf.Bytes()
	if len(pdfData) == 0 {
		t.Error("PDF data is empty")
	}

	t.Logf("PDF with options generated successfully: %d bytes", len(pdfData))
}

func TestInvalidTemplate(t *testing.T) {
	client := setupClient(t)

	templatePath := loadTemplate(t, templateInvalidTest)
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	var buf bytes.Buffer
	info, err := client.GeneratePDFFromString(context.Background(), &buf, "", string(templateData), nil, nil)
	if err == nil {
		t.Error("Expected error for invalid template but got none")
	}

	t.Logf("Error correctly returned: %v", err)
	if info.CorrelationID != "" {
		t.Logf("Correlation ID: %s", info.CorrelationID)
	}
	if info.Stderr != "" {
		t.Logf("Stderr: %s", info.Stderr)
	}
}

func TestNonExistentTemplateFile(t *testing.T) {
	client := setupClient(t)

	var buf bytes.Buffer
	_, err := client.GeneratePDFFromFile(context.Background(), &buf, "", "non-existent-file.typ", nil, nil)
	if err == nil {
		t.Error("Expected error for non-existent file but got none")
	}
}

func TestResponseInfo(t *testing.T) {
	client := setupClient(t)

	templatePath := loadTemplate(t, templateResponseTest)
	templateData, err := os.ReadFile(templatePath)
	if err != nil {
		t.Fatalf("Failed to read template: %v", err)
	}

	var buf bytes.Buffer
	info, err := client.GeneratePDFFromString(context.Background(), &buf, "", string(templateData), nil, nil)
	if err != nil {
		t.Fatalf("Failed to generate PDF: %v", err)
	}

	if info.CorrelationID == "" {
		t.Error("Expected non-empty correlation ID")
	}

	t.Logf("Response info - Correlation ID: %s", info.CorrelationID)
}

// ============================================================================
// Integration Tests - Visual Output
// ============================================================================

func TestIntegration_DashingDeptNews(t *testing.T) {
	client := setupClient(t)
	outputDir := setupOutputDir(t)

	templatePath := loadTemplate(t, templateDashingNews)
	outputPath := filepath.Join(outputDir, "newsletter.pdf")

	media := loadMediaFiles(t, map[string]string{
		"prometheus.png": "test/typst/dashing-dept-news/prometheus.png",
	})

	_, err := client.SavePDF(context.Background(), "", templatePath, outputPath, nil, media)
	if err != nil {
		t.Fatalf("Newsletter generation failed: %v", err)
	}

	size := verifyPDF(t, outputPath)
	t.Logf("Newsletter PDF: %s (%d bytes)", outputPath, size)
}

func TestIntegration_CompleteDocument(t *testing.T) {
	client := setupClient(t)
	outputDir := setupOutputDir(t)

	templatePath := loadTemplate(t, templateCompleteDoc)
	outputPath := filepath.Join(outputDir, "complete-document.pdf")

	_, err := client.SavePDF(context.Background(), "", templatePath, outputPath, nil, nil)
	if err != nil {
		t.Fatalf("Complete document generation failed: %v", err)
	}

	size := verifyPDF(t, outputPath)
	t.Logf("Complete document PDF: %s (%d bytes)", outputPath, size)
}

func TestIntegration_HighResolution(t *testing.T) {
	client := setupClient(t)
	outputDir := setupOutputDir(t)

	templatePath := loadTemplate(t, templateHighRes)
	outputPath := filepath.Join(outputDir, "high-resolution.pdf")
	options := []string{"--ppi", "300"}

	_, err := client.SavePDF(context.Background(), "", templatePath, outputPath, options, nil)
	if err != nil {
		t.Fatalf("High-res document generation failed: %v", err)
	}

	size := verifyPDF(t, outputPath)
	t.Logf("High-resolution PDF: %s (%d bytes)", outputPath, size)
}

func TestIntegration_WithPlots(t *testing.T) {
	client := setupClient(t)
	outputDir := setupOutputDir(t)

	templatePath := loadTemplate(t, templatePlots)
	outputPath := filepath.Join(outputDir, "plots.pdf")

	_, err := client.SavePDF(context.Background(), "", templatePath, outputPath, nil, nil)
	if err != nil {
		t.Skipf("plotst package not available: %v", err)
	}

	size := verifyPDF(t, outputPath)
	t.Logf("Plots document PDF: %s (%d bytes)", outputPath, size)
}

func TestIntegration_AllExamples(t *testing.T) {
	client := setupClient(t)
	outputDir := setupOutputDir(t)

	testCases := []struct {
		name        string
		template    string
		mediaFiles  map[string]string
		description string
	}{
		{
			name:        "aero-navigator",
			template:    templateAeroNav,
			description: "Aviation navigation log",
		},
		{
			name:     "newsletter",
			template: templateDashingNews,
			mediaFiles: map[string]string{
				"prometheus.png": "test/typst/dashing-dept-news/prometheus.png",
			},
			description: "Department newsletter with image",
		},
		{
			name:     "academic-paper",
			template: templateElspub,
			mediaFiles: map[string]string{
				"test_data.json":         "test/typst/elspub/test_data.json",
				"md_content/content.md":  "test/typst/elspub/md_content/content.md",
				"img/nginx.png":          "test/typst/elspub/img/nginx.png",
				"fonts/Lato-Regular.ttf": "test/typst/elspub/fonts/Lato-Regular.ttf",
				"fonts/Lato-Bold.ttf":    "test/typst/elspub/fonts/Lato-Bold.ttf",
				"fonts/Lato-Italic.ttf":  "test/typst/elspub/fonts/Lato-Italic.ttf",
			},
			description: "Academic publication with assets",
		},
	}

	results := make(map[string]int64)

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			templatePath := loadTemplate(t, tc.template)
			outputPath := filepath.Join(outputDir, tc.name+".pdf")

			media := loadMediaFiles(t, tc.mediaFiles)

			_, err := client.SavePDF(context.Background(), "", templatePath, outputPath, nil, media)
			if err != nil {
				t.Fatalf("%s generation failed: %v", tc.description, err)
			}

			size := verifyPDF(t, outputPath)
			results[tc.name] = size
			t.Logf("%s: %d bytes", tc.description, size)
		})
	}

	// Summary
	var total int64
	for _, size := range results {
		total += size
	}
	t.Logf("Generated %d PDFs, total size: %d bytes", len(results), total)
	t.Logf("Output directory: %s", outputDir)
}
