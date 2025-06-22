# Coves atProto Lexicon

This directory contains the atProto lexicon schemas for Coves, a federated forum platform built on the AT Protocol.

## Structure

### Actor (`social.coves.actor.*`)
- **profile**: User profile with verification status, federation info, location, moderation history, and violation tracking
- **subscription**: Records of communities a user wants to view content from  
- **membership**: Records of communities where user has reputation

### Community (`social.coves.community.*`)
- **profile**: Community information including moderation type (moderator/sortition/hybrid)
- **rules**: Community rules, restrictions, and moderation configuration
- **wiki**: Community wiki pages with markdown content

### Post (`social.coves.post.*`)
- **text**: Text posts with optional markdown formatting
- **image**: Image posts supporting up to 10 images
- **video**: Video posts with thumbnail support
- **article**: Link posts with preview metadata
- **microblog**: Short-form posts from federated platforms (Bluesky, Mastodon)

### Interaction (`social.coves.interaction.*`)
- **vote**: Upvotes on posts or comments
- **tag**: Tags applied to content (helpful, spam, hostile, etc.)
- **comment**: Comments supporting text, images, or stickers
- **share**: Sharing posts to other communities or platforms

### Moderation (`social.coves.moderation.*`)
- **vote**: Votes on moderation proposals
- **tribunalVote**: Jury decisions in sortition-based moderation
- **ruleProposal**: Proposals to change community rules

### Embed (`social.coves.embed.*`)
- **image**: Reusable image embed with alt text
- **video**: Video embed with metadata
- **external**: External link previews
- **post**: Embedded/quoted posts

### Federation (`social.coves.federation.*`)
- **post**: Reference to original federated post with platform info

### RichText (`social.coves.richtext.*`)
- **markup**: Text formatting (bold, italic, code, strikethrough, spoiler)
- **mention**: User and community mentions
- **link**: Links within text content

## Key Features

1. **Federation Support**: Posts and users track their origin platform (Bluesky, Lemmy, Mastodon, etc.) with dedicated microblog post type for federated content
2. **Dual Relationship Model**: Separate subscription (viewing) from membership (reputation)
3. **Flexible Moderation**: Support for moderator-based, sortition-based, or hybrid systems
4. **Rich Content**: Support for various post types with embedded media
5. **Community Governance**: Democratic rule changes and moderator removal
6. **Privacy-First**: Viewing activity is not recorded; only active participation
7. **Content Safety**: NSFW flags and content labels (nsfw, violence, spoilers) for communities and posts
8. **Accountability**: User profiles track moderation history and rule violations across communities

## Usage

These schemas define the data structures for Coves. They follow atProto conventions:
- Records are stored in user repositories (except community records)
- Communities have their own DIDs and repositories
- All timestamps use RFC 3339 format
- Text fields support Unicode with grapheme limits
- References use AT-URIs for cross-repository linking