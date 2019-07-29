package persist

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/gocolly/colly"
	"github.com/gocolly/colly/queue"
	"github.com/sirupsen/logrus"
	"log"
	"net/url"
	"strings"
	"sync"
	"time"
	"website-indexer/config"
	"website-indexer/global"
	// This is imported to define "sqlite3" driver
	_ "github.com/mattn/go-sqlite3"
)

type Storager interface {
	Close() error
	QueueSize() (int, error)
	AddRequest([]byte) error
	GetRequest() ([]byte, error)
	LoadResource(Resource) (*Resource, error)
	UnmarshalResource([]byte) (*Resource, error)
}

var _ Storager = (*Storage)(nil)
var _ queue.Storage = (*Storage)(nil)

// Storage implements a SQLite3 storage backend for Colly
type Storage struct {
	Config   *config.Config
	Filename string       // Filename indicates the name of the sqlite file to use
	dbh      *sql.DB      // handle to the db
	mu       sync.RWMutex // Only used for cookie methods.
}

func (me *Storage) UnmarshalResource(b []byte) (res *Resource, err error) {
	for range only.Once {
		res = &Resource{}
		err = json.Unmarshal(b, &res)
		if err != nil {
			err = fmt.Errorf("unable to unmarshal resource '%s': %s", string(b), err)
			break
		}
		res, err = me.LoadResource(*res)
		if err != nil {
			err = fmt.Errorf("unable to load resource '%s': %s", string(b), err)
			break
		}
	}
	return res, err
}

// Init initializes the sqlite3 storage
func (me *Storage) Init() (err error) {
	for range only.Once {
		if me.dbh != nil {
			break
		}
		db, err := sql.Open("sqlite3", me.Filename)
		if err != nil {
			logrus.Fatalf("unable to open db file: %s", err.Error())
			break
		}

		err = db.Ping()
		if err != nil {
			logrus.Fatalf("unable to ping database: %s", err.Error())
		}
		me.dbh = db

		for _, s := range ddl {
			_, err = me.ExecSql(s)
			if err == nil {
				continue
			}
			logrus.Fatalf("unable to execute SQL '%s': %s", s, err.Error())
		}
	}
	return err
}

// AddRequest implements queue.Storage.AddRequest() function
func (me *Storage) AddRequest(b []byte) (err error) {
	for range only.Once {
		r := struct {
			colly.Request
			Url global.Url `json:"URL"`
		}{}
		err = json.Unmarshal(b, &r)
		if err != nil {
			logrus.Errorf("unable to unmarshal request '%s': %s", string(b), err)
		}
		_, _, err = me.AddResource(r.Url)
		if err != nil {
			logrus.Errorf("unable to add resource '%s': %s", r.Url, err)
			break
		}
		_, _, err = me.InsertQueueItem(Item{
			ResourceHash: NewHash(r.Url),
		})
		if err != nil {
			logrus.Errorf("unable to insert item '%s' into 'queue': %s", r.Url, err)
			break
		}
	}
	return err
}

type SqlResult int

const (
	RecordUnknown SqlResult = iota
	RecordAdded
	RecordExisted
	RecordAddFailed
)

func (me *Storage) AddResource(u global.Url) (resid SqlId, sr SqlResult, err error) {
	for range only.Once {
		var r *Resource
		r, err = me.LoadResourceByUrl(u)
		if err != nil {
			break
		}
		if r.Id != 0 {
			sr = RecordExisted
			break
		}
		r.url = u
		err = r.Init(nil)
		if err != nil {
			break
		}
		var res *Resource
		res, _, err = me.InsertResource(*r)
		if err != nil {
			sr = RecordAddFailed
			break
		}
		resid = res.Id
		sr = RecordAdded
	}
	return resid, sr, err
}

func (me *Storage) AddHost(u global.Url) (hid SqlId, sr SqlResult, err error) {
	for range only.Once {
		hid, err = me.LoadHostIdByUrl(u)
		if err != nil {
			break
		}
		if hid != 0 {
			sr = RecordExisted
			break
		}
		var h *Host
		h, _, err = me.InsertHost(*NewHost(u))
		if err != nil {
			sr = RecordAddFailed
			break
		}
		hid = h.Id
		sr = RecordAdded
	}
	return hid, sr, err
}

// GetRequest implements queue.Storage.GetRequest() function
func (me *Storage) GetRequest() (blob []byte, err error) {
	for range only.Once {
		me.mu.Lock()
		defer me.mu.Unlock()

		qi, err := me.LoadQueueItem()
		if err != nil {
			break
		}
		if qi.ResourceHash == 0 {
			break
		}
		r := &Resource{
			Hash: qi.ResourceHash,
		}
		r, err = me.LoadResourceProps(*r)
		if err != nil {
			break
		}
		blob, err = json.Marshal(r)
		if err != nil {
			return nil, err
		}
	}
	return blob, nil
}

// QueueSize implements queue.Storage.QueueSize() function
func (me *Storage) QueueSize() (int, error) {
	var count int
	statement, err := me.dbh.Prepare(dml[SelectQueueCountDml])
	if err != nil {
		return 0, err
	}
	row := statement.QueryRow()
	err = row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func (me *Storage) ExecSql(s global.Sql, args ...interface{}) (r sql.Result, err error) {
	stmt, err := me.dbh.Prepare(s)
	if err != nil {
		logrus.Fatalf("unable to execute SQL '%s'", s)
		return nil, err
	}
	return stmt.Exec(args...)
}

func (me *Storage) QueryRow(s global.Sql, args ...interface{}) *sql.Row {
	stmt, err := me.dbh.Prepare(s)
	if err != nil {
		logrus.Fatalf("unable to execute SQL '%s'", s)
		return nil
	}
	return stmt.QueryRow(args...)
}

func (me *Storage) Prepare(sql global.Sql) (stmt *sql.Stmt) {
	stmt, _ = me.dbh.Prepare(sql)
	return stmt
}

// Clear removes all entries from the storage
func (me *Storage) Clear() (err error) {
	me.mu.Lock()
	defer me.mu.Unlock()
	for _, t := range tables {
		err = me.clearTable(t)
		if err != nil {
			break
		}
	}
	return err
}

// Clear only the 'visited' table
func (me *Storage) ClearVisited() (err error) {
	me.mu.Lock()
	defer me.mu.Unlock()
	return me.clearTable("visited")
}

// Clears a single table
func (me *Storage) clearTable(t global.Tablename) (err error) {
	for range only.Once {
		_, err = me.Prepare("DROP TABLE " + t).Exec()
		if err != nil {
			err = fmt.Errorf("unable to drop SQL table '%s': %s", t, err.Error())
			logrus.Error(err.Error())
		}

	}
	return err
}

//Close the db
func (me *Storage) Close() error {
	err := me.dbh.Close()
	return err
}

// Visited implements colly/storage.Visited()
func (me *Storage) Visited(requestID uint64) error {
	_, _, err := me.InsertVisited(Visited{
		ResourceHash: Hash(requestID),
		Timestamp:    time.Now().Unix(),
		Headers:      "",
		Body:         "",
		Cookies:      "",
	})
	if err != nil {
		logrus.Errorf("unable to record visit for request ID '%d': %s", requestID, err)
	}
	return err
}

// IsVisited implements colly/storage.IsVisited()
func (me *Storage) IsVisited(requestID uint64) (yn bool, err error) {
	return !me.LoadShouldRevisitByHash(Hash(requestID)), nil
}

// SetCookies implements colly/storage..SetCookies()
func (me *Storage) SetCookies(u *url.URL, cookies string) {
	return
	// TODO Cookie methods currently have no way to return an error.

	// We need to use a write lock to prevent a race in the db:
	// if two callers set cookies in a very small window of time,
	// it is possible to drop the new cookies from one caller
	// ('last update wins' == best avoided).
	me.mu.Lock()
	defer me.mu.Unlock()

	statement, err := me.dbh.Prepare("INSERT INTO cookies (host, cookies) VALUES (?,?)")
	if err != nil {
		log.Printf("SetCookies() .Set error %me", err)
	}
	_, err = statement.Exec(u.Host, cookies)
	if err != nil {
		log.Printf("SetCookies() .Set error %me", err)
	}

}

// Cookies implements colly/storage.Cookies()
func (me *Storage) Cookies(u *url.URL) string {
	return ""
	// TODO Cookie methods currently have no way to return an error.
	var cookies string
	me.mu.RLock()

	//cookiesStr, err := me.Client.Get(me.getCookieID(u.Host)).Result()
	statement, err := me.dbh.Prepare("SELECT cookies FROM cookies where host = ?")
	if err != nil {
		log.Printf("Cookies() .Get error %s", err)
		return ""
	}
	row := statement.QueryRow(u.Host)

	err = row.Scan(&cookies)

	me.mu.RUnlock()

	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
		}

		log.Printf("Cookies() .Get error %s", err)
	}

	return cookies
}

//func (me *Storage) AddUrlPath(su *Resource) (hostid int64, sr SqlResult, err error) {
//	for range only.Once {
//		hash := NewHash(su.UrlPath)
//		hostId := ""
//		urlPath := ""
//		_, err = me.ExecSql("INSERT INTO urls (hash,host_id,url_path) VALUES (?,?,?)",
//			hash,
//			hostId,
//			urlPath,
//		)
//		if err != nil {
//			logrus.Errorf("unable to insert domain:port '%s:%s': %s", hn, p, err.Error())
//			break
//		}
//	}
//	return 0, RecordUnknown, nil
//}

// AddRequest implements queue.Storage.AddRequest() function
//func (me *Storage) AddUrl(u global.Resource) error {
//	me.mu.Lock()
//	defer me.mu.Unlock()
//	for range only.Once {
//		hostid,_,err := me.AddHost(u)
//		if err != nil {
//			break
//		}
//		up,err := getUrlPath(u)
//		upid,_,err := me.AddUrlPath(up)
//		if err != nil {
//			break
//		}
//
//	}
//
//
//	////return me.Client.RPush(me.getQueueID(), r).Err()
//	//statement, err := me.dbh.Prepare("INSERT INTO queue (data) VALUES (?)")
//	//if err != nil {
//	//	return err
//	//}
//	//_, err = statement.Exec(r)
//	//if err != nil {
//	//	return err
//	//}
//	return nil
//}
//func matchesCodes(err error, codes []int) (ok bool) {
//	for range only.Once {
//		var se sqlite3.Error
//		se, ok = err.(sqlite3.Error)
//		if !ok {
//			break
//		}
//		if len(codes) != 2 {
//			break
//		}
//		if se.Code != sqlite3.ErrNo(codes[0]) {
//			break
//		}
//		if se.ExtendedCode != sqlite3.ErrNoExtended(codes[1]) {
//			break
//		}
//	}
//	return ok
//}
