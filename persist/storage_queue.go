package persist

import (
	"database/sql"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"time"
)

func (me *Storage) InsertQueueItem(itm Item) (qi *Item, sr sql.Result, err error) {
	for range only.Once {

		qi, err := me.LoadQueueItemByHash(itm.ResourceHash)
		if err != nil {
			break
		}
		if qi.Id != 0 {
			break
		}
		sr, err = me.ExecSql(dml[InsertQueueItemDml],
			int64(itm.ResourceHash),
			time.Now().Unix(),
		)
		if err != nil {
			err = fmt.Errorf("unable to insert queue item '%d': %s", itm.ResourceHash, err.Error())
			logrus.Error(err)
			break
		}
		qi = &itm
		var qiid int64
		qiid, err = sr.LastInsertId()
		if err != nil {
			err = fmt.Errorf("unable to access inserted ID for queue item '%d': %s", itm.ResourceHash, err.Error())
			logrus.Error(err)
			break
		}
		qi.Id = SqlId(qiid)
	}
	return qi, sr, err
}

func (me *Storage) LoadQueueItemByHash(h Hash) (qi *Item, err error) {
	for range only.Once {
		qi = &Item{}
		row := me.QueryRow(dml[SelectQueueItemByHashDml], int64(h))
		var id int64
		var h int64
		err = row.Scan(&id, &h)
		if err == nil {
			qi.Id = SqlId(id)
			qi.ResourceHash = Hash(h)
			break
		}
		if err == sql.ErrNoRows {
			err = nil
			break
		}
		err = fmt.Errorf("unable to load queue item for hash='%d': %s", h, err)
		logrus.Error(err)
		break
	}
	return qi, err
}

func (me *Storage) LoadQueueItem() (i *Item, err error) {
	var stmt *sql.Stmt
	for range only.Once {
		stmt, err = me.dbh.Prepare(dml[SelectQueueItemDml])
		if err != nil {
			break
		}
		row := stmt.QueryRow()
		i = &Item{}
		var id int64
		var rh int64
		err = row.Scan(&id, &rh)
		if err != nil {
			break
		}
		i.Id = SqlId(id)
		i.ResourceHash = Hash(rh)
	}
	return i, err
}

func (me *Storage) DeleteQueueItemsbyHash(h Hash) (sr sql.Result, err error) {
	sr, err = me.ExecSql(dml[DeleteQueueItemsByHashDml], int64(h))
	if err != nil {
		err = fmt.Errorf("unable to queue items for has='%d': %s", h, err.Error())
		logrus.Error(err)
	}
	return sr, err
}
