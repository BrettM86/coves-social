# Forum Platform Domain Knowledge

This document serves as the central reference for domain-specific knowledge about our forum platform. It should be maintained and updated as the platform evolves.

## Overview

Our forum platform is a federated discussion system that combines traditional forum features with modern federation protocols (atproto and ActivityPub). The platform emphasizes user control, community autonomy, and cross-platform interoperability.

## Core Concepts

### Feed System

#### Home Feed
The home feed provides a personalized experience for each user based on their preferences and subscriptions.

**UI:**
- UI will be a tiktok style scrollable feed but left to right 

**Key Features:**
- **User-specific**: Each user sees a unique feed based on their subscriptions
- **Community subscription slider**: Users can adjust content visibility per community (1-5 scale)
    - 1 = Only show the best/most popular content from this community
    - 5 = Show all content from this community
- **Read state tracking**: Already read posts are automatically hidden from the feed
    - Requires `isRead` flag in the lexicon
    - Users can access read posts through a dedicated history tab
      - Store read history in PostgreSQL AppView, NOT in CAR files
      - This data never enters the user's repository
      - Only accessible through authenticated private endpoints

#### All Feed
A global feed showing content from across the entire platform.

**Key Features:**
- Displays all public content from all communities
- Respects read state (hides already-read posts)
- Provides discovery mechanism for new communities and content

#### Community Feed
A community feed showing content from a single community.

**Key Features:**
- Incorporates typical forum feed features such as hot, top (month, day, year), new

### Communities

Communities are the primary organizational unit for content and discussions.

#### Core Features
- **Creation**: Users can create new communities
- **Blocking**: Users can block communities from appearing in their feeds
- **Wiki**: Each community maintains its own wiki for documentation
- **Subscriptions**: Users can subscribe/unsubscribe to communities
- **NSFW Toggle**: Communities can be marked as NSFW with appropriate filtering

#### Reputation System
- **User reputation**: Gained through community interaction such as creating posts, comments, positive tags
- **Member access levels**: Reputation affects user privileges
- **Comment ordering**: Higher reputation may influence comment visibility
- **Voting**: Voting is more impactful for reputable community members

#### Rules System
Communities can enable specific rules to shape content and behavior:
- **Rule voting**: Community rules are enabled through democratic votes
- **Post type restrictions**: e.g., "Text Only" communities
- **Website blocklists**: Prevent specific domains from being shared
- **Geolocation restrictions**: Limit posting to users in specific locations

#### Bots and Plugins
- **RSS feed poster**: Automatically post content from RSS feeds
- **Game thread bots**: Automated creation of sports/gaming discussion threads
- Additional open source plugin system for community-specific automation

#### Moderation Models

##### Sortition-Based Moderation
A democratic approach to content moderation:
- **Tag-based removal**: Content hidden when enough users tag it as inappropriate, is flagged for tribunal review
- **Tribunal system**: Posts/accounts that have been tagged go through tribunal review
  - Minimum reputation required to serve


##### Traditional Moderation
- **Moderator hierarchy**: Standard moderator-based enforcement
- **Community override**: Users can still vote to remove moderators
- **Hybrid approach**: Even moderator-led communities incorporate user feedback

#### Flairs
- Support for post/user flairs (implementation details TBD)

### Federation

Our platform supports bidirectional federation with major protocols.

#### atproto (Bluesky) Integration
- **Post display**: Show Bluesky posts inline with native content
- **Backend implementation**: Posts stored as references to Bluesky records
- **Future roadmap**: Comment federation planned (posts-only for initial release)

#### ActivityPub Integration
Full two-way compatibility with ActivityPub networks:
- **Instance mapping**: All AP instances converted to "Hubs"
- **Community mapping**: AP communities mapped to Coves communities
- **User identity**: AP users assigned atproto DIDs
- **Action translation**: Bidirectional conversion between AP actions and Coves lexicon

### Posts

#### Post Types
- **Text**: Traditional text-based discussions
- **Video**: Video content with discussion threads
- **Image**: Image posts with comments
- **Article**: Long-form content/blog posts
- **Microblog**: Explicitly created for bsky posts

#### Post Features
- **Voting**: Upvote system
- **Tagging system**: Posts can be tagged (helpful, spam, hostile, etc.)
- **Share tracking**: Monitor post sharing/distribution
- **Comment threads**: Each post owns its comment thread
- **Federation indicators**: Display source platform (Lemmy, Bluesky, Mbin, etc.)

### Users

#### Identity and Authentication
- **Phone verification**: Required for account creation for Coves users (grants verified status)
- **Identity system**: Uses atproto DID (Decentralized Identifier) system
- **Federation**: Each user belongs to a federated platform (Bluesky, Lemmy, etc.)

#### Username System
- **Random generation**: When not provided, usernames follow patterns:
    - "Adjective Noun" (e.g., "BraveEagle")
    - "Adjective Adjective Noun" (e.g., "SmallQuietMouse")

#### User Features
- **Blocking**: Users can block other users
- **Offline sync**: Basic framework for offline functionality
- **Notification control**: Mute notifications from own posts
- **Save posts**: Bookmark posts for later reference
- **Post history**: View all previously seen posts

## Future Enhancements
- Comment federation for Bluesky
- Possibly Mastadon integration
- Extended plugin system for communities
- Edit/History for comments/posts/wiki
- Enhanced offline capabilities
- Incognito browsing
  - Disables browsing history
  - Read state not stored on device
- Community forking process (Theoretically built in via DID's)

---

*Last updated: [Date]*
*Version: 1.0*