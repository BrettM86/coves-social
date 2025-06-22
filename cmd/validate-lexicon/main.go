package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	lexicon "github.com/bluesky-social/indigo/atproto/lexicon"
)

func main() {
	var (
		schemaPath = flag.String("path", "internal/atproto/lexicon", "Path to lexicon schemas directory")
		verbose    = flag.Bool("v", false, "Verbose output")
	)
	flag.Parse()

	if *verbose {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	// Check if path exists
	if _, err := os.Stat(*schemaPath); os.IsNotExist(err) {
		log.Fatalf("Schema path does not exist: %s", *schemaPath)
	}

	// Create a new catalog
	catalog := lexicon.NewBaseCatalog()

	// Load all schemas from the directory
	fmt.Printf("Loading schemas from: %s\n", *schemaPath)
	if err := loadSchemasWithDebug(&catalog, *schemaPath, *verbose); err != nil {
		log.Fatalf("Failed to load schemas: %v", err)
	}

	fmt.Printf("✅ Successfully loaded schemas from %s\n", *schemaPath)

	// Validate schema structure by trying to resolve some known schemas
	if err := validateSchemaStructure(&catalog, *schemaPath, *verbose); err != nil {
		log.Fatalf("Schema validation failed: %v", err)
	}

	fmt.Println("✅ All schemas validated successfully!")
}

// validateSchemaStructure performs additional validation checks
func validateSchemaStructure(catalog *lexicon.BaseCatalog, schemaPath string, verbose bool) error {
	var validationErrors []string
	var schemaFiles []string

	// Collect all JSON schema files
	err := filepath.Walk(schemaPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .json files
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			schemaFiles = append(schemaFiles, path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking schema directory: %w", err)
	}

	if verbose {
		fmt.Printf("\nFound %d schema files to validate:\n", len(schemaFiles))
		for _, file := range schemaFiles {
			fmt.Printf("  - %s\n", file)
		}
	}

	// Try to resolve some of our expected schemas
	expectedSchemas := []string{
		"social.coves.actor.profile",
		"social.coves.community.profile",
		"social.coves.post.text",
		"social.coves.richtext.markup",
	}

	if verbose {
		fmt.Println("\nValidating key schemas:")
	}

	for _, schemaID := range expectedSchemas {
		if _, err := catalog.Resolve(schemaID); err != nil {
			validationErrors = append(validationErrors, fmt.Sprintf("Failed to resolve schema %s: %v", schemaID, err))
		} else if verbose {
			fmt.Printf("  ✅ %s\n", schemaID)
		}
	}

	if len(validationErrors) > 0 {
		fmt.Println("❌ Schema validation errors found:")
		for _, errMsg := range validationErrors {
			fmt.Printf("  %s\n", errMsg)
		}
		return fmt.Errorf("found %d validation errors", len(validationErrors))
	}

	return nil
}

// loadSchemasWithDebug loads schemas one by one to identify problematic files
func loadSchemasWithDebug(catalog *lexicon.BaseCatalog, schemaPath string, verbose bool) error {
	var schemaFiles []string

	// Collect all JSON schema files
	err := filepath.Walk(schemaPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Only process .json files
		if !info.IsDir() && filepath.Ext(path) == ".json" {
			schemaFiles = append(schemaFiles, path)
		}
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking schema directory: %w", err)
	}

	// Try to load schemas one by one
	for _, schemaFile := range schemaFiles {
		if verbose {
			fmt.Printf("  Loading: %s\n", schemaFile)
		}

		// Create a temporary catalog for this file
		tempCatalog := lexicon.NewBaseCatalog()
		if err := tempCatalog.LoadDirectory(filepath.Dir(schemaFile)); err != nil {
			return fmt.Errorf("failed to load schema file %s: %w", schemaFile, err)
		}
	}

	// If all individual files loaded OK, try loading the whole directory
	return catalog.LoadDirectory(schemaPath)
}
