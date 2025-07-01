package tests

import (
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
		"social.coves.richtext.markup",
		"social.coves.richtext.mention",
		"social.coves.richtext.link",
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
		"social.coves.richtext.markup#byteSlice": "byteSlice definition in markup schema",
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
