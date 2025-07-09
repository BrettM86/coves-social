# Lexicon Test Data

This directory contains test data files for validating AT Protocol lexicon schemas.

## Naming Convention

Test files follow a specific naming pattern to distinguish between valid and invalid test cases:

- **Valid test files**: `{type}-valid.json` or `{type}-valid-{variant}.json`
  - Example: `profile-valid.json`, `post-valid-text.json`
  - These files should pass validation

- **Invalid test files**: `{type}-invalid-{reason}.json`
  - Example: `profile-invalid-missing-handle.json`, `post-invalid-enum-type.json`
  - These files should fail validation
  - Used to test that the validator correctly rejects malformed data

## Directory Structure

```
lexicon-test-data/
├── actor/
│   ├── profile-valid.json              # Valid actor profile
│   └── profile-invalid-missing-handle.json  # Missing required field
├── community/
│   └── profile-valid.json              # Valid community profile
├── interaction/
│   └── vote-valid.json                 # Valid vote record
├── moderation/
│   └── ban-valid.json                  # Valid ban record
└── post/
    ├── post-valid-text.json            # Valid text post
    └── post-invalid-enum-type.json     # Invalid postType value
```

## Running Tests

The validator automatically processes all files in this directory:
- Valid files are expected to pass validation
- Invalid files (containing `-invalid-` in the name) are expected to fail
- The validator reports if any files don't behave as expected

```bash
# Run full validation
go run cmd/validate-lexicon/main.go

# Run with verbose output to see each file
go run cmd/validate-lexicon/main.go -v
```

## Adding New Test Data

When adding new test data:

1. Create valid examples that showcase proper schema usage
2. Create invalid examples that test common validation errors:
   - Missing required fields
   - Invalid enum values
   - Wrong data types
   - Invalid formats (e.g., bad DIDs, malformed dates)

3. Name files according to the convention above
4. Run the validator to ensure your test files behave as expected