# typst-pdf-generator-go
Go client for generating PDFs with typst-pdf-generator.

## Usage

All generation methods accept a `context.Context` and return a `ResponseInfo` containing `CorrelationID`, `Stdout`, and `Stderr`.
The correlation ID can be set by the caller via `WithCorrelationID`, and the client will prefer a correlation/request ID returned by the server when present.
checkout the [examples](./examples) folder for more usage examples.

```go
package main

import (
	"context"
	"log"
	"time"

	typstpdfgenerator "github.com/4Sigma/typst-pdf-generator-go"
)

func main() {
	client, err := typstpdfgenerator.New(
		"YOUR_AUTH_KEY",
		"https://YOUR_FAAS_GATEWAY/function/typst",
		typstpdfgenerator.WithTimeout(120*time.Second),
	)
	if err != nil {
		log.Fatal(err)
	}

	ctx := typstpdfgenerator.WithCorrelationID(context.Background(), "my-correlation-id")

	info, err := client.SavePDF(
		ctx,
		"",               // content
		"template.typ",   // template file
		"output.pdf",     // output file
		nil,              // options
		nil,              // media
	)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("generated PDF, correlation_id=%s", info.CorrelationID)
}
```
