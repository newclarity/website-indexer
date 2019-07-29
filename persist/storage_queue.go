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
func (me *Storage) LoadQueueItem() (i *Item, err error) {
	var stmt *sql.Stmt
	for range only.Once {
		stmt, err = me.dbh.Prepare(dml[SelectQueueItemDml])
		if err != nil {
			break
		}
		row := stmt.QueryRow()
		i = &Item{}
		var rh int64
		err = row.Scan(&rh)
		if err != nil {
			break
		}
		i.ResourceHash = Hash(rh)
	}
	return i, err
}
