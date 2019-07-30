package persist

import (
	"database/sql"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"time"
	"website-indexer/global"
)

func (me *Storage) InsertVisited(vis Visited) (v *Visited, sr sql.Result, err error) {
	for range only.Once {
		sr, err = me.ExecSql(dml[InsertVisitedDml],
			int64(vis.ResourceHash),
			vis.Timestamp,
			vis.Headers,
			vis.Body,
			vis.Cookies,
		)
		if err != nil {
			err = fmt.Errorf("unable to insert visited '%d': %s", vis.ResourceHash, err)
			logrus.Error(err)
			break
		}
		v = &vis
		var vid int64
		vid, err = sr.LastInsertId()
		if err != nil {
			err = fmt.Errorf("unable to access inserted ID for visited '%s': %s", vis.ResourceHash, err)
			logrus.Error(err)
			break
		}
		v.Id = SqlId(vid)
		_, err = me.DeleteQueueItemsbyHash(vis.ResourceHash)
		if err != nil {
			err = fmt.Errorf("unable to remove queued item for hash='%d' from queue: %s", vis.ResourceHash, err)
			logrus.Error(err)
			break
		}
	}
	return v, sr, err
}

func (me *Storage) LoadVisitedCountByUrl(u global.Url) (cnt int64, err error) {
	cnt, _, err = me.loadVisitedStatsByHash(NewHash(u))
	if err != nil {
		err = fmt.Errorf("unable to count 'visited' for url='%s': %s", u, err)
		logrus.Error(err)
	}
	return
}

func (me *Storage) LoadVisitedCountByHash(h Hash) (cnt int64, err error) {
	cnt, _, err = me.loadVisitedStatsByHash(h)
	if err != nil {
		err = fmt.Errorf("unable to count 'visited' for hash='%s': %s", h, err)
		logrus.Error(err)
	}
	return
}

func (me *Storage) LoadShouldRevisitByHash(h Hash) bool {
	revisit := true
	for range only.Once {
		cnt, ts, err := me.loadVisitedStatsByHash(h)
		if err != nil {
			// Assume we should revisit?
			break
		}
		if cnt == 0 {
			break
		}
		r := time.Unix(ts, 0).Add(me.Config.RevisitDuration)
		if time.Now().After(r) {
			break
		}
		revisit = false
	}
	return revisit
}

func (me *Storage) loadVisitedStatsByHash(h Hash) (cnt int64, ts int64, err error) {
	r := me.QueryRow(dml[SelectVisitedStatsByHashDml], int64(h))
	err = r.Scan(&cnt, &ts)
	if err != nil {
		err = fmt.Errorf("unable to count 'visited' for hash='%d': %s", h, err)
		logrus.Error(err)
	}
	return cnt, ts, err
}
