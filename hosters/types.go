package hosters

import "website-indexer/pages"

type IndexHoster interface {
	Initialize() error
	IndexPage(*pages.Page) bool
}
