package typstpdfgenerator

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
)

type correlationIDContextKey struct{}

// WithCorrelationID returns a new context carrying the provided correlation ID.
//
// If correlationID is empty, ctx is returned unchanged.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	if correlationID == "" {
		return ctx
	}
	return context.WithValue(ctx, correlationIDContextKey{}, correlationID)
}

// CorrelationIDFromContext extracts a correlation ID previously set with WithCorrelationID.
// Returns an empty string if none is set.
func CorrelationIDFromContext(ctx context.Context) string {
	v, _ := ctx.Value(correlationIDContextKey{}).(string)
	return v
}

var (
	ErrNotGenerated   = errors.New("PDF not generated")
	ErrConnection     = errors.New("connection error")
	ErrInvalidAuth    = errors.New("auth key cannot be empty")
	ErrInvalidGateway = errors.New("FaaS gateway cannot be empty")
)

type NotGeneratedError struct {
	Message       string
	CorrelationID string
}

func (e *NotGeneratedError) Error() string {
	if e.CorrelationID != "" {
		return fmt.Sprintf("TypstPDF.NotGenerated '%s' (correlation_id=%s)", e.Message, e.CorrelationID)
	}
	return fmt.Sprintf("TypstPDF.NotGenerated '%s'", e.Message)
}

func (e *NotGeneratedError) Unwrap() error {
	return ErrNotGenerated
}

type ConnectionError struct {
	Message string
	Err     error
}

func (e *ConnectionError) Error() string {
	switch {
	case e.Message != "" && e.Err != nil:
		return fmt.Sprintf("connection error: %s: %v", e.Message, e.Err)
	case e.Message != "":
		return fmt.Sprintf("connection error: %s", e.Message)
	case e.Err != nil:
		return fmt.Sprintf("connection error: %v", e.Err)
	default:
		return "connection error"
	}
}

func (e *ConnectionError) Unwrap() error {
	if e.Err != nil {
		return e.Err
	}
	return ErrConnection
}

type HTTPError struct {
	StatusCode    int
	Status        string
	Body          string
	CorrelationID string
}

func (e *HTTPError) Error() string {
	if e == nil {
		return "http error"
	}
	if e.CorrelationID != "" && e.Body != "" {
		return fmt.Sprintf("HTTP %d: %s: %s (correlation_id=%s)", e.StatusCode, e.Status, e.Body, e.CorrelationID)
	}
	if e.CorrelationID != "" {
		return fmt.Sprintf("HTTP %d: %s (correlation_id=%s)", e.StatusCode, e.Status, e.CorrelationID)
	}
	if e.Body != "" {
		return fmt.Sprintf("HTTP %d: %s: %s", e.StatusCode, e.Status, e.Body)
	}
	return fmt.Sprintf("HTTP %d: %s", e.StatusCode, e.Status)
}

func (e *HTTPError) Unwrap() error {
	return ErrConnection
}

type MediaFile struct {
	Name string
	Data []byte
}

type ResponseInfo struct {
	Stdout        string
	Stderr        string
	CorrelationID string
}

type typstRequest struct {
	Content  string            `json:"content"`
	Template string            `json:"template"`
	Options  []string          `json:"options"`
	Media    map[string]string `json:"media"`
}

type typstResponse struct {
	Error   bool   `json:"error"`
	Message string `json:"message,omitempty"`
	PDF     string `json:"pdf,omitempty"`
	Stdout  string `json:"stdout,omitempty"`
	Stderr  string `json:"stderr,omitempty"`
}

type Client struct {
	authKey    string
	gateway    *url.URL
	httpClient *http.Client
}

func correlationIDFromResponse(resp *http.Response) string {
	if resp == nil {
		return ""
	}
	if v := strings.TrimSpace(resp.Header.Get("X-Correlation-ID")); v != "" {
		return v
	}
	if v := strings.TrimSpace(resp.Header.Get("X-Request-ID")); v != "" {
		return v
	}
	return ""
}

type Option func(*Client) error

func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) error {
		c.httpClient.Timeout = timeout
		return nil
	}
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Client) error {
		if client == nil {
			return fmt.Errorf("http client cannot be nil")
		}
		c.httpClient = client
		return nil
	}
}

func WithInsecureSkipVerify() Option {
	return func(c *Client) error {
		if c.httpClient.Transport == nil {
			// http.Client treats nil Transport as http.DefaultTransport.
			c.httpClient.Transport = http.DefaultTransport
		}

		transport, ok := c.httpClient.Transport.(*http.Transport)
		if !ok {
			return fmt.Errorf("cannot enable InsecureSkipVerify with non-*http.Transport (%T)", c.httpClient.Transport)
		}

		cloned := transport.Clone()
		if cloned.TLSClientConfig == nil {
			cloned.TLSClientConfig = &tls.Config{}
		}
		cloned.TLSClientConfig.InsecureSkipVerify = true
		c.httpClient.Transport = cloned
		return nil
	}
}

func New(authKey, faasGateway string, opts ...Option) (*Client, error) {
	if authKey == "" {
		return nil, ErrInvalidAuth
	}
	if faasGateway == "" {
		return nil, ErrInvalidGateway
	}

	gatewayURL, err := url.Parse(faasGateway)
	if err != nil {
		return nil, fmt.Errorf("invalid gateway URL: %w", err)
	}

	if gatewayURL.Scheme != "http" && gatewayURL.Scheme != "https" {
		return nil, &ConnectionError{Message: "invalid endpoint scheme: expected http or https"}
	}

	client := &Client{
		authKey: authKey,
		gateway: gatewayURL,
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
			Transport: &http.Transport{
				MaxIdleConns:        10,
				IdleConnTimeout:     30 * time.Second,
				DisableCompression:  false,
				TLSHandshakeTimeout: 10 * time.Second,
			},
		},
	}

	for _, opt := range opts {
		if err := opt(client); err != nil {
			return nil, err
		}
	}

	return client, nil
}

func (c *Client) convert(ctx context.Context, w io.Writer, content string, templateData []byte, options []string, media []MediaFile) (ResponseInfo, error) {
	correlationID := CorrelationIDFromContext(ctx)
	if correlationID == "" {
		correlationID = uuid.NewString()
	}
	info := ResponseInfo{CorrelationID: correlationID}

	mediaEncoded := make(map[string]string, len(media))
	for _, m := range media {
		mediaEncoded[m.Name] = base64.StdEncoding.EncodeToString(m.Data)
	}

	reqBody := typstRequest{
		Content:  content,
		Template: base64.StdEncoding.EncodeToString(templateData),
		Options:  options,
		Media:    mediaEncoded,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return info, &ConnectionError{Err: err}
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.gateway.String(), bytes.NewBuffer(jsonData))
	if err != nil {
		return info, &ConnectionError{Err: err}
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", c.authKey)
	req.Header.Set("X-Correlation-ID", correlationID)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return info, &ConnectionError{Err: err}
	}
	defer resp.Body.Close()

	if serverCorrelationID := correlationIDFromResponse(resp); serverCorrelationID != "" {
		correlationID = serverCorrelationID
		info.CorrelationID = serverCorrelationID
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return info, &ConnectionError{Err: err}
	}

	var response typstResponse
	if len(body) > 0 {
		_ = json.Unmarshal(body, &response)
	}
	info.Stdout = response.Stdout
	info.Stderr = response.Stderr

	if resp.StatusCode != http.StatusOK {
		msg := strings.TrimSpace(response.Message)
		if msg == "" {
			msg = strings.TrimSpace(string(body))
		}
		if msg != "" {
			if len(msg) > 1024 {
				msg = msg[:1024] + "..."
			}
			return info, &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, Body: msg, CorrelationID: correlationID}
		}
		return info, &HTTPError{StatusCode: resp.StatusCode, Status: resp.Status, CorrelationID: correlationID}
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return info, &ConnectionError{Err: err}
	}
	info.Stdout = response.Stdout
	info.Stderr = response.Stderr

	if response.Error {
		msg := response.Message
		if msg == "" {
			msg = "Unknown error"
		}
		return info, &NotGeneratedError{Message: msg, CorrelationID: correlationID}
	}

	if response.PDF == "" {
		return info, &NotGeneratedError{Message: "No PDF data in response", CorrelationID: correlationID}
	}

	pdfData, err := base64.StdEncoding.DecodeString(response.PDF)
	if err != nil {
		return info, fmt.Errorf("failed to decode PDF data: %w", err)
	}

	if _, err := w.Write(pdfData); err != nil {
		return info, fmt.Errorf("failed to write PDF data: %w", err)
	}

	return info, nil
}

func (c *Client) GeneratePDFFromFile(ctx context.Context, w io.Writer, content, templateFilePath string, options []string, media []MediaFile) (ResponseInfo, error) {
	templateData, err := os.ReadFile(templateFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ResponseInfo{}, fmt.Errorf("template file not found: %s", templateFilePath)
		}
		return ResponseInfo{}, fmt.Errorf("failed to read template file: %w", err)
	}

	return c.convert(ctx, w, content, templateData, options, media)
}

func (c *Client) GeneratePDFFromString(ctx context.Context, w io.Writer, content, templateString string, options []string, media []MediaFile) (ResponseInfo, error) {
	templateData := []byte(templateString)
	return c.convert(ctx, w, content, templateData, options, media)
}

func (c *Client) SavePDF(ctx context.Context, content, templateFilePath, outputPath string, options []string, media []MediaFile) (ResponseInfo, error) {
	templateData, err := os.ReadFile(templateFilePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return ResponseInfo{}, fmt.Errorf("template file not found: %s", templateFilePath)
		}
		return ResponseInfo{}, fmt.Errorf("failed to read template file: %w", err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return ResponseInfo{}, fmt.Errorf("failed to create output file: %w", err)
	}

	info, convErr := c.convert(ctx, file, content, templateData, options, media)
	closeErr := file.Close()
	if convErr != nil {
		_ = os.Remove(outputPath)
		return info, convErr
	}
	if closeErr != nil {
		_ = os.Remove(outputPath)
		return info, fmt.Errorf("failed to close output file: %w", closeErr)
	}

	return info, nil
}
