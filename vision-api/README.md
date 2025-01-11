# Vision API Processor

A high-performance, concurrent image processing service that leverages Google Cloud Vision API for image labeling and analysis.

## Features

- Concurrent image processing with configurable worker pools
- Automatic image resizing and optimization
- Progress tracking with real-time statistics
- Multiple output formats (JSON, CSV, JSONL)
- Rate limiting and retry mechanisms
- Temporary file management
- Comprehensive error handling
- Modular and extensible architecture

## Prerequisites

- Go 1.21 or higher
- Google Cloud SDK
- ImageMagick (`convert` command-line tool)

### System Requirements

- Memory: At least 4GB RAM recommended
- Storage: Depends on image batch size
- CPU: Multi-core processor recommended for concurrent processing

## Installation

1. Clone the repository:
```bash
git clone https://github.com/your-username/vision-api.git
cd vision-api
```

2. Install dependencies:
```bash
go mod download
```

3. Install required system dependencies:

On macOS:
```bash
brew install imagemagick
```

On Ubuntu/Debian:
```bash
sudo apt-get update
sudo apt-get install imagemagick
```

4. Configure Google Cloud credentials:
```bash
gcloud auth application-default login
```

## Building

Build the binary:
```bash
# Basic build
go build -o vision-processor ./cmd/vision-processor

# Build with optimization
go build -o vision-processor -ldflags="-s -w" ./cmd/vision-processor

# Cross-compile for different platforms
GOOS=linux GOARCH=amd64 go build -o vision-processor-linux ./cmd/vision-processor
GOOS=windows GOARCH=amd64 go build -o vision-processor.exe ./cmd/vision-processor
```

## Configuration

Create a configuration file (`config.yaml`):

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  mode: "release"
  shutdown_timeout: 5

vision:
  max_retries: 3
  batch_size: 100
  pool_size: 8
  rate_limit: 1800
  timeout_seconds: 30

image:
  max_size_mb: 40
  max_width: 4096
  max_height: 4096
  quality: 85
  allowed_formats:
    - "jpeg"
    - "jpg"
    - "png"
    - "gif"
    - "bmp"

storage:
  output_dir: "./output"
  temp_dir: "./tmp"
```

## Usage Examples

### Basic Usage

Process a directory of images:
```bash
./vision-processor -input ./images -output ./results
```

### Advanced Usage

1. Process with custom concurrency and batch size:
```bash
./vision-processor \
  -input ./images \
  -output ./results \
  -config ./custom-config.yaml \
  -concurrency 16 \
  -debug
```

2. Programmatic Usage:

```go
package main

import (
    "context"
    "log"
    
    "github.com/your-username/vision-api/pkg/vision"
    "github.com/your-username/vision-api/internal/processor"
    "github.com/your-username/vision-api/internal/image"
)

func main() {
    // Initialize Vision client
    visionClient, err := vision.NewClient(
        vision.WithRateLimit(1800),
        vision.WithMaxRetries(3),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Initialize image handler
    imageHandler, err := image.NewHandler(
        image.WithMaxImageSize(40 * 1024 * 1024),
        image.WithMaxDimensions(4096, 4096),
    )
    if err != nil {
        log.Fatal(err)
    }

    // Create processor
    proc, err := processor.NewProcessor(
        processor.WithPoolSize(8),
        processor.WithBatchSize(100),
        processor.WithImageHandler(imageHandler),
        processor.WithVisionClient(visionClient),
        processor.WithOutputDir("./output"),
    )
    if err != nil {
        log.Fatal(err)
    }
    defer proc.Cleanup()

    // Process images
    ctx := context.Background()
    results, err := proc.ProcessBatch(ctx, inputs)
    if err != nil {
        log.Fatal(err)
    }

    // Handle results
    for _, result := range results {
        if result.Error != nil {
            log.Printf("Error processing %s: %v", result.Filename, result.Error)
            continue
        }
        log.Printf("Successfully processed %s: %d labels", result.Filename, len(result.Labels))
    }
}
```

### Handling Large Batches

For processing large batches of images:

```go
// Create a progress tracker
tracker := progress.NewTracker(totalFiles, os.Stdout)
processor.SetProgressTracker(tracker)
tracker.Start()
defer tracker.Finish()

// Process in smaller batches
batchSize := 1000
for i := 0; i < len(inputs); i += batchSize {
    end := i + batchSize
    if end > len(inputs) {
        end = len(inputs)
    }
    
    results, err := processor.ProcessBatch(ctx, inputs[i:end])
    if err != nil {
        log.Printf("Batch error: %v", err)
    }
}
```

## Error Handling

The service provides detailed error information:

```go
if err := processor.Process(ctx, input); err != nil {
    var processErr *processor.ProcessError
    if errors.As(err, &processErr) {
        log.Printf("Operation: %s", processErr.Op)
        log.Printf("File: %s", processErr.File)
        log.Printf("Details: %s", processErr.Details)
    }
}
```

## Development

### Running Tests

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific tests
go test -v ./internal/processor
```

### Code Quality

```bash
# Run linter
golangci-lint run

# Run code formatting
go fmt ./...
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Commit your changes
4. Push to the branch
5. Create a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.