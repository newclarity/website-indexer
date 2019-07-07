package persist

const (
	QueuedDir        = "queued"
	IndexedDir       = "indexed"
	ErroredDir       = "errored"
	JsonFileTemplate = "/{hash:2}/{hash}.json"
)

const (
	CannotExist Existence = iota
	CanExist
	MustExist
)
