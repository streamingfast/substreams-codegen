package codegen

import (
	"embed"
	"io/fs"
	"text/template"

	"github.com/bmatcuk/doublestar/v4"
)

//go:embed common-templates/*
var commonTemplatesFS embed.FS

var commonTemplates *template.Template

func init() {
	var err error
	commonTemplates, err = parseCommonTemplates()
	if err != nil {
		panic(err)
	}
}

func MarkdownEscape(s string) string {
	return "```\n" + s + "\n```\n"
}

// NetworkToEndpoint maps network names to their corresponding StreamingFast endpoints
func NetworkToEndpoint(network string) string {
	endpoints := map[string]string{
		"mainnet":   "mainnet.eth.streamingfast.io:443",
		"bnb":       "bnb.streamingfast.io:443",
		"bsc":       "bnb.streamingfast.io:443",
		"polygon":   "polygon.streamingfast.io:443",
		"arbitrum":  "arbitrum.streamingfast.io:443",
		"optimism":  "optimism.streamingfast.io:443",
		"base":      "base.streamingfast.io:443",
		"avalanche": "avalanche.streamingfast.io:443",
		"fantom":    "fantom.streamingfast.io:443",
		"amoy":      "polygon.streamingfast.io:443", // Polygon testnet uses same endpoint
		"holesky":   "holesky.eth.streamingfast.io:443",
		"sepolia":   "sepolia.eth.streamingfast.io:443",
	}
	
	if endpoint, exists := endpoints[network]; exists {
		return endpoint
	}
	
	// Fallback to a generic pattern if not found
	return network + ".streamingfast.io:443"
}

func parseCommonTemplates() (*template.Template, error) {
	t := template.New("").Funcs(templateFuncs)
	filenames, err := doublestar.Glob(commonTemplatesFS, "**/*.gotmpl")
	if err != nil {
		return nil, err
	}

	for _, filename := range filenames {
		b, err := fs.ReadFile(commonTemplatesFS, filename)
		if err != nil {
			return nil, err
		}
		_, err = t.New(filename).Parse(string(b))
		if err != nil {
			return nil, err
		}
	}
	return t, nil
}
