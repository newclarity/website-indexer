package persist

import "website-indexer/global"

var tables = []global.Tablename{
	"hosts",
	"resources",
	"queue",
	"visited",
}
var ddl = []global.Sql{
	"CREATE TABLE IF NOT EXISTS hosts (id INTEGER PRIMARY KEY, scheme TEXT, domain TEXT, port INTEGER)",
	"CREATE UNIQUE INDEX IF NOT EXISTS idx_hosts ON hosts (scheme,domain,port)",

	"CREATE TABLE IF NOT EXISTS resources (id INTEGER PRIMARY KEY, hash INTEGER, host_id INTEGER, urlpath TEXT)",
	"CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_hash ON resources (hash)",
	"CREATE UNIQUE INDEX IF NOT EXISTS idx_resources_url ON resources (host_id,urlpath)",

	"CREATE TABLE IF NOT EXISTS queue (id INTEGER PRIMARY KEY, resource_hash INTEGER, timestamp INTEGER)",
	"CREATE UNIQUE INDEX IF NOT EXISTS idx_queue_resource_hash ON queue (resource_hash,timestamp)",

	"CREATE TABLE IF NOT EXISTS visited (id INTEGER PRIMARY KEY, resource_hash INTEGER, timestamp INTEGER, headers TEXT, body BLOB, cookies BLOB)",
	"CREATE UNIQUE INDEX IF NOT EXISTS idx_visited ON visited (resource_hash,timestamp)",
}

const (
	SelectQueueCountDml     = "select-queue-count"
	SelectQueueItemDml      = "select-queue-item"
	SelectResourceDml       = "select-resource"
	SelectResourceByIdDml   = "select-resource-by-id"
	SelectResourceByHashDml = "select-resource-by-hash"
	SelectHostByUrlDml      = "select-host-by-url"
	SelectHostByIdDml       = "select-host-by-id"
	SelectHostBySDPDml      = "select-host-by-sdp"

	SelectVisitedStatsByHashDml = "select-visited-count-by-hash"

	InsertHostDml      = "insert-host"
	InsertResourceDml  = "insert-resource"
	InsertVisitedDml   = "insert-visited"
	InsertQueueItemDml = "insert-queue-item"

	DeleteQueueItemDml = "delete-queue-item"
)

var dml = map[global.Name]global.Sql{
	SelectQueueCountDml:     "SELECT COUNT(*) FROM queue",
	SelectQueueItemDml:      "SELECT CASE WHEN count(*)=0 THEN 0 ELSE min(resource_hash) END FROM queue",
	SelectResourceDml:       "SELECT id,hash,host_id,urlpath FROM resources",
	SelectResourceByIdDml:   "SELECT id,hash,host_id,urlpath FROM resources WHERE id = ?",
	SelectResourceByHashDml: "SELECT id,hash,host_id,urlpath FROM resources WHERE hash = ?",
	SelectHostByIdDml:       "SELECT id,scheme,domain,port FROM hosts WHERE id = ?",
	SelectHostByUrlDml:      "SELECT id FROM hosts WHERE scheme || '://' || domain || CASE WHEN port IN (0,80) THEN '' ELSE ':'||CAST(port AS text) END || '/' LIKE ?",
	SelectHostBySDPDml:      "SELECT id FROM hosts WHERE scheme=? AND domain=? AND port=?",

	SelectVisitedStatsByHashDml: "SELECT COUNT(*),MAX(timestamp) FROM visited WHERE resource_hash = ?",

	InsertHostDml:      "INSERT INTO hosts (scheme,domain,port) VALUES (?,?,?)",
	InsertResourceDml:  "INSERT INTO resources (hash,host_id,urlpath) VALUES (?,?,?)",
	InsertVisitedDml:   "INSERT INTO visited (resource_hash,timestamp,headers,body,cookies) VALUES(?,?,?,?,?)",
	InsertQueueItemDml: "INSERT INTO queue (resource_hash,timestamp) VALUES (?,?)",

	DeleteQueueItemDml: "DELETE FROM queue WHERE id = ?",
}
