package codegen

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

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
