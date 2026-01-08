package codegen

import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/streamingfast/substreams-codegen/loop"
)

func Cmd(msg any) loop.Cmd {
	return func() loop.Msg {
		return msg
	}
}

func ZipFiles(files map[string][]byte) ([]byte, error) {
	tempDir, err := os.MkdirTemp(os.TempDir(), "zipper")
	if err != nil {
		return nil, fmt.Errorf("mkdir temp: %w", err)
	}

	if os.Getenv("GENERATOR_KEEP_FILES") != "true" {
		defer os.RemoveAll(tempDir)
	} else {
		fmt.Println("Keeping files in", tempDir)
	}

	zipFilepath := filepath.Join(tempDir, "source.zip")

	// write the content of the zip file here
	zipFile, err := os.Create(zipFilepath)
	if err != nil {
		return nil, fmt.Errorf("creating zip file: %w", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	for relativeFile, content := range files {
		fullFilepath := strings.ReplaceAll(relativeFile, "/", string(os.PathSeparator))

		fh := &zip.FileHeader{
			Name:   fullFilepath,
			Method: zip.Deflate,
		}
		if strings.HasSuffix(fullFilepath, ".sh") {
			fh.SetMode(0755)
		}
		// Create a writer for each file in the zip archive
		writer, err := zipWriter.CreateHeader(fh)
		if err != nil {
			return nil, fmt.Errorf("creating zip writer: %w", err)
		}

		// Write the file data to the zip archive
		_, err = writer.Write(content)
		if err != nil {
			return nil, fmt.Errorf("writing to zip: %w", err)
		}
	}

	// Close the zip archive
	err = zipWriter.Close()
	if err != nil {
		return nil, fmt.Errorf("closing zip: %w", err)
	}

	// Open the zip file to send the bytes
	zipFileB, err := os.ReadFile(zipFilepath)
	if err != nil {
		return nil, fmt.Errorf("opening zip file: %w", err)
	}

	return zipFileB, nil
}

// PrettifyJSON takes raw JSON bytes and returns prettified JSON with proper indentation.
// If the input is not valid JSON, it returns the original content unchanged.
func PrettifyJSON(rawJSON []byte) []byte {
	if len(rawJSON) == 0 {
		return rawJSON
	}

	var jsonData interface{}
	if err := json.Unmarshal(rawJSON, &jsonData); err != nil {
		// If it's not valid JSON, return the original content
		return rawJSON
	}

	prettified, err := json.MarshalIndent(jsonData, "", "  ")
	if err != nil {
		// If prettification fails, return the original content
		return rawJSON
	}

	return prettified
}

// errorPattern defines a pattern for matching and transforming error messages
type errorPattern struct {
	pattern *regexp.Regexp
	// transform takes the matched groups and returns the user-friendly message
	// The first element of matches is always the full match, subsequent elements are capture groups
	transform func(matches []string) string
}

var (
	// Compiled once at init time for performance
	wrappedFileErrorPattern = regexp.MustCompile(`^could not read file "([^"]+)":.+:\s*(.+)$`)
	fileNotFoundPattern     = regexp.MustCompile(`no such file or directory`)
	permissionDeniedPattern = regexp.MustCompile(`permission denied`)
	isDirectoryPattern      = regexp.MustCompile(`is a directory`)

	// errorPatterns is the list of error patterns to check, in order
	errorPatterns = []errorPattern{
		// Wrapped file errors with filename extraction
		{
			pattern: wrappedFileErrorPattern,
			transform: func(matches []string) string {
				filename := matches[1]
				innerErr := matches[2]

				// Check the inner error and provide context-specific message
				if fileNotFoundPattern.MatchString(innerErr) {
					return fmt.Sprintf("File not found: %s", filename)
				}
				if permissionDeniedPattern.MatchString(innerErr) {
					return fmt.Sprintf("Permission denied for file: %s", filename)
				}
				if isDirectoryPattern.MatchString(innerErr) {
					return fmt.Sprintf("Path is a directory, not a file: %s", filename)
				}
				// Unknown inner error, still show filename
				return fmt.Sprintf("%s (file: %s)", innerErr, filename)
			},
		},
	}

	// Simple string-based mappings (no regex needed)
	simpleErrorMappings = map[string]string{
		"no such file or directory":                             "File not found - please check the path and try again",
		"permission denied":                                      "Permission denied - please check file permissions",
		"is a directory":                                         "Path points to a directory, not a file - please provide a file path",
		"contract source code is not verified":                  "Contract source code is not verified on the block explorer - you'll need to provide the ABI manually",
		"invalid contract address or contract does not exist":   "Invalid contract address or contract does not exist at this address",
	}
)

// MapClientSideErrorToMessage converts technical errors from the CLI client into user-friendly messages.
//
// This function handles errors that originate from the CLI client (file reading, HTTP requests, etc.)
// which are transmitted as strings. It uses regex patterns to extract context (like filenames) and
// provides clear, actionable error messages.
//
// Note: While working with error strings is somewhat brittle (dependent on error message formats),
// it's acceptable for improving user experience with common error cases from client-side operations.
func MapClientSideErrorToMessage(err error) string {
	if err == nil {
		return ""
	}

	errStr := err.Error()

	// Try regex patterns first (they may extract additional context like filenames)
	for _, ep := range errorPatterns {
		if matches := ep.pattern.FindStringSubmatch(errStr); matches != nil {
			return ep.transform(matches)
		}
	}

	// Try simple substring mappings
	for errorSubstring, userMessage := range simpleErrorMappings {
		if strings.Contains(errStr, errorSubstring) {
			return userMessage
		}
	}

	// Return original error if no mapping found
	return errStr
}
