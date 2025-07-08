# Rich Text Facets Documentation

## Overview

Rich text facets provide a way to annotate ranges of text with formatting, mentions, links, and other features in the Coves platform. This implementation follows the AT Protocol standards while extending them with additional formatting options.

## UTF-8 Byte Counting

**IMPORTANT**: All byte indices in facets use UTF-8 byte positions, not character positions or UTF-16 code units.

### Why UTF-8 Bytes?

The AT Protocol uses UTF-8 byte counting to ensure consistent text indexing across all platforms and programming languages. This is crucial because:

1. **Character counting varies** - What counts as one "character" differs between Unicode grapheme clusters, code points, and visual characters
2. **UTF-16 inconsistencies** - JavaScript uses UTF-16 internally, but other languages don't
3. **Network efficiency** - AT Protocol data is transmitted as UTF-8

### Calculating Byte Positions

```go
text := "Hello 👋 @alice!"
// Finding byte position of "@alice"
prefix := "Hello 👋 "
byteStart := len([]byte(prefix))  // 9 bytes (not 8 characters!)
byteEnd := byteStart + len([]byte("@alice"))  // 9 + 6 = 15
```

### Common Pitfalls

1. **Emoji can be multiple bytes**:
   - "👋" = 4 bytes
   - "👨‍👩‍👧‍👧" = 25 bytes (family emoji with zero-width joiners)

2. **Non-ASCII text**:
   - "café" = 5 bytes (é is 2 bytes)
   - "Привет" = 12 bytes (each Cyrillic letter is 2 bytes)

## Facet Structure

Each facet consists of:
- **index**: Byte range in the text
- **features**: Array of features applied to this range

```json
{
  "index": {
    "byteStart": 5,
    "byteEnd": 11
  },
  "features": [
    {
      "$type": "social.coves.richtext.facet#mention",
      "did": "did:plc:example123",
      "handle": "alice.bsky.social"
    }
  ]
}
```

## Supported Feature Types

### 1. Mention (`social.coves.richtext.facet#mention`)
For @mentions of users or !mentions of communities.

```json
{
  "$type": "social.coves.richtext.facet#mention",
  "did": "did:plc:example123",
  "handle": "alice.bsky.social"  // Optional, for display
}
```

### 2. Link (`social.coves.richtext.facet#link`)
For hyperlinks in text.

```json
{
  "$type": "social.coves.richtext.facet#link",
  "uri": "https://example.com"
}
```

### 3. Bold (`social.coves.richtext.facet#bold`)
For **bold** text formatting.

```json
{
  "$type": "social.coves.richtext.facet#bold"
}
```

### 4. Italic (`social.coves.richtext.facet#italic`)
For *italic* text formatting.

```json
{
  "$type": "social.coves.richtext.facet#italic"
}
```

### 5. Strikethrough (`social.coves.richtext.facet#strikethrough`)
For ~~strikethrough~~ text formatting.

```json
{
  "$type": "social.coves.richtext.facet#strikethrough"
}
```

### 6. Spoiler (`social.coves.richtext.facet#spoiler`)
For hidden/spoiler text that requires user interaction to reveal.

```json
{
  "$type": "social.coves.richtext.facet#spoiler",
  "reason": "Movie spoiler"  // Optional
}
```

## Examples

### Complete Post with Facets

```json
{
  "text": "Check out **this** amazing post by @alice about ~secret stuff~!",
  "facets": [
    {
      "index": {"byteStart": 10, "byteEnd": 18},
      "features": [{"$type": "social.coves.richtext.facet#bold"}]
    },
    {
      "index": {"byteStart": 36, "byteEnd": 42},
      "features": [{
        "$type": "social.coves.richtext.facet#mention",
        "did": "did:plc:alice123",
        "handle": "alice.coves.social"
      }]
    },
    {
      "index": {"byteStart": 49, "byteEnd": 62},
      "features": [{
        "$type": "social.coves.richtext.facet#spoiler",
        "reason": "Plot details"
      }]
    }
  ]
}
```

### Multiple Features on Same Range

Text can have multiple formatting features:

```json
{
  "text": "This is ***really*** important!",
  "facets": [
    {
      "index": {"byteStart": 8, "byteEnd": 20},
      "features": [
        {"$type": "social.coves.richtext.facet#bold"},
        {"$type": "social.coves.richtext.facet#italic"}
      ]
    }
  ]
}
```

## Best Practices

1. **Validate byte ranges**: Ensure byteEnd > byteStart and both are within text bounds
2. **Sort facets**: Order facets by byteStart for easier processing
3. **Handle overlaps**: Multiple facets can overlap - render them in a sensible order
4. **Validate features**: Each feature must have a valid `$type` field
5. **UTF-8 safety**: Always calculate bytes using UTF-8 encoding, not string length

## Integration with Bluesky

When federating content from Bluesky:
- Bluesky uses `app.bsky.richtext.facet` with similar structure
- Convert their facet types to Coves equivalents
- Preserve byte indices (they also use UTF-8)

## Client Implementation Notes

For web clients:
```javascript
// Converting JavaScript string index to UTF-8 bytes
const textEncoder = new TextEncoder();
const bytes = textEncoder.encode(text.substring(0, charIndex));
const byteIndex = bytes.length;
```

For Go implementations:
```go
// Already UTF-8 native
byteIndex := len(text[:runeIndex])
```

## Validation

Always validate:
1. Byte indices are non-negative integers
2. ByteEnd > byteStart
3. Byte ranges don't exceed text length
4. Each feature has required fields
5. `$type` values are recognized