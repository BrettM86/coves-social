package repository

import (
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multihash"
)

var (
	// PlaceholderCID is used for empty repositories that have no content yet
	// This allows us to maintain consistency in the repository record
	// while the actual CAR data is created when records are added
	PlaceholderCID cid.Cid
)

func init() {
	// Initialize the placeholder CID once at startup
	emptyData := []byte("empty")
	mh, _ := multihash.Sum(emptyData, multihash.SHA2_256, -1)
	PlaceholderCID = cid.NewCidV1(cid.Raw, mh)
}
