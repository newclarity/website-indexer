package persist

const (
	QueuedDir        = "queued"
	IndexedDir       = "indexed"
	ErroredDir       = "errored"
	JsonFileTemplate = "/{hash:2}/{hash}.json"
	SqliteDbFilename = "crawl.db"
)

const (
	CannotExist Existence = iota
	CanExist
	MustExist
)
