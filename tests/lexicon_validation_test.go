package tests

import (
	"strings"
	"testing"

	lexicon "github.com/bluesky-social/indigo/atproto/lexicon"
)

func TestLexiconSchemaValidation(t *testing.T) {
	// Create a new catalog
	catalog := lexicon.NewBaseCatalog()

	// Load all schemas from the lexicon directory
	schemaPath := "../internal/atproto/lexicon"
	if err := catalog.LoadDirectory(schemaPath); err != nil {
		t.Fatalf("Failed to load lexicon schemas: %v", err)
	}

	// Test that we can resolve our key schemas
	expectedSchemas := []string{
		"social.coves.actor.profile",
		"social.coves.actor.subscription",
		"social.coves.actor.membership",
		"social.coves.community.profile",
		"social.coves.community.rules",
		"social.coves.community.wiki",
		"social.coves.post.text",
		"social.coves.post.image",
		"social.coves.post.video",
		"social.coves.post.article",
		"social.coves.richtext.facet",
		"social.coves.embed.image",
		"social.coves.embed.video",
		"social.coves.embed.external",
		"social.coves.embed.post",
		"social.coves.interaction.vote",
		"social.coves.interaction.tag",
		"social.coves.interaction.comment",
		"social.coves.interaction.share",
		"social.coves.moderation.vote",
		"social.coves.moderation.tribunalVote",
		"social.coves.moderation.ruleProposal",
	}

	for _, schemaID := range expectedSchemas {
		t.Run(schemaID, func(t *testing.T) {
			if _, err := catalog.Resolve(schemaID); err != nil {
				t.Errorf("Failed to resolve schema %s: %v", schemaID, err)
			}
		})
	}
}

func TestLexiconCrossReferences(t *testing.T) {
	// Create a new catalog
	catalog := lexicon.NewBaseCatalog()

	// Load all schemas
	if err := catalog.LoadDirectory("../internal/atproto/lexicon"); err != nil {
		t.Fatalf("Failed to load lexicon schemas: %v", err)
	}

	// Test specific cross-references that should work
	crossRefs := map[string]string{
		"social.coves.richtext.facet#byteSlice": "byteSlice definition in facet schema",
		"social.coves.actor.profile#geoLocation": "geoLocation definition in actor profile",
		"social.coves.community.rules#rule":      "rule definition in community rules",
	}

	for ref, description := range crossRefs {
		t.Run(ref, func(t *testing.T) {
			if _, err := catalog.Resolve(ref); err != nil {
				t.Errorf("Failed to resolve cross-reference %s (%s): %v", ref, description, err)
			}
		})
	}
}

func TestValidateRecord(t *testing.T) {
	// Create a new catalog
	catalog := lexicon.NewBaseCatalog()

	// Load all schemas
	if err := catalog.LoadDirectory("../internal/atproto/lexicon"); err != nil {
		t.Fatalf("Failed to load lexicon schemas: %v", err)
	}

	// Test cases for ValidateRecord
	tests := []struct {
		name        string
		recordType  string
		recordData  map[string]interface{}
		shouldFail  bool
		errorContains string
	}{
		{
			name:       "Valid actor profile",
			recordType: "social.coves.actor.profile",
			recordData: map[string]interface{}{
				"$type":       "social.coves.actor.profile",
				"handle":      "alice.example.com",
				"displayName": "Alice Johnson",
				"createdAt":   "2024-01-15T10:30:00Z",
			},
			shouldFail: false,
		},
		{
			name:       "Invalid actor profile - missing required field",
			recordType: "social.coves.actor.profile",
			recordData: map[string]interface{}{
				"$type":       "social.coves.actor.profile",
				"displayName": "Alice Johnson",
			},
			shouldFail:    true,
			errorContains: "required field missing: handle",
		},
		{
			name:       "Valid community profile",
			recordType: "social.coves.community.profile",
			recordData: map[string]interface{}{
				"$type":          "social.coves.community.profile",
				"name":           "programming",
				"displayName":    "Programming Community",
				"creator":        "did:plc:creator123",
				"moderationType": "moderator",
				"federatedFrom":  "coves",
				"createdAt":      "2023-12-01T08:00:00Z",
			},
			shouldFail: false,
		},
		{
			name:       "Valid post record",
			recordType: "social.coves.post.record",
			recordData: map[string]interface{}{
				"$type":           "social.coves.post.record",
				"community":       "did:plc:programming123",
				"postType":        "text",
				"title":           "Test Post",
				"text":            "This is a test post",
				"tags":            []string{"test", "golang"},
				"language":        "en",
				"contentWarnings": []string{},
				"createdAt":       "2025-01-09T14:30:00Z",
			},
			shouldFail: false,
		},
		{
			name:       "Invalid post record - invalid enum value",
			recordType: "social.coves.post.record",
			recordData: map[string]interface{}{
				"$type":           "social.coves.post.record",
				"community":       "did:plc:programming123",
				"postType":        "invalid-type",
				"title":           "Test Post",
				"text":            "This is a test post",
				"tags":            []string{"test"},
				"language":        "en",
				"contentWarnings": []string{},
				"createdAt":       "2025-01-09T14:30:00Z",
			},
			shouldFail:    true,
			errorContains: "string val not in required enum",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := lexicon.ValidateRecord(&catalog, tt.recordData, tt.recordType, lexicon.AllowLenientDatetime)
			
			if tt.shouldFail {
				if err == nil {
					t.Errorf("Expected validation to fail but it passed")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Expected error to contain '%s', got: %v", tt.errorContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("Expected validation to pass but got error: %v", err)
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && strings.Contains(s, substr))
}

func TestValidateRecordWithStrictMode(t *testing.T) {
	// Create a new catalog
	catalog := lexicon.NewBaseCatalog()

	// Load all schemas
	if err := catalog.LoadDirectory("../internal/atproto/lexicon"); err != nil {
		t.Fatalf("Failed to load lexicon schemas: %v", err)
	}

	// Test with strict validation flags
	recordData := map[string]interface{}{
		"$type":       "social.coves.actor.profile",
		"handle":      "alice.example.com",
		"displayName": "Alice Johnson",
		"createdAt":   "2024-01-15T10:30:00", // Missing timezone
	}

	// Should fail with strict validation
	err := lexicon.ValidateRecord(&catalog, recordData, "social.coves.actor.profile", lexicon.StrictRecursiveValidation)
	if err == nil {
		t.Error("Expected strict validation to fail on datetime without timezone")
	}

	// Should pass with lenient datetime validation
	err = lexicon.ValidateRecord(&catalog, recordData, "social.coves.actor.profile", lexicon.AllowLenientDatetime)
	if err != nil {
		t.Errorf("Expected lenient validation to pass, got error: %v", err)
	}
}
