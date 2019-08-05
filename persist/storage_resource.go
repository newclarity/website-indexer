package persist

import (
	"database/sql"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"website-indexer/global"
)

func (me *Storage) InsertResource(res Resource) (r *Resource, sr sql.Result, err error) {
	for range only.Once {
		if res.host == nil {
			res.host, err = me.LoadHostByResource(res)
			if err != nil {
				break
			}
		}
		if me.HasResource(res) {
			break
		}
		sr, err = me.ExecSql(dml[InsertResourceDml],
			int64(res.Hash),
			int64(res.host.Id),
			res.UrlPath,
		)
		//if matchesSqlite3ErrorCodes(err,sqlite3.ErrConstraint,sqlite3.ErrConstraintUnique) {
		//	break
		//}
		if err != nil {
			err = fmt.Errorf("unable to insert request '%s': %s", res.UrlPath, err.Error())
			logrus.Error(err)
			break
		}
		r = &res
		var rid int64
		rid, err = sr.LastInsertId()
		if err != nil {
			err = fmt.Errorf("unable to access inserted ID for request '%s': %s", res.url, err.Error())
			logrus.Error(err)
			break
		}
		r.Id = SqlId(rid)
	}
	return r, sr, err
}

func (me *Storage) LoadResource(res Resource) (r *Resource, err error) {
	for range only.Once {
		if res.Id != 0 {
			r, err = me.LoadResourceProps(res)
			break
		}
		var rid SqlId
		rid, err = me.LoadResourceId(res)
		if err != nil {
			break
		}
		res.Id = rid
		r = &res
	}
	return r, err
}

func (me *Storage) HasResource(res Resource) (ok bool) {
	q := dml[SelectResourceCountByIdDml]
	qr := me.QueryRow(q, res.Id)
	var cnt uint64
	err := qr.Scan(&cnt)
	return err == nil && cnt > 0
}

func (me *Storage) LoadResourceId(res Resource) (rid SqlId, err error) {
	var id int64
	for range only.Once {
		if res.Hash == 0 {
			logrus.Errorf("must specific Hash for resource when calling LoadResourceId()")
			break
		}
		q := dml[SelectResourceByHashDml]
		qr := me.QueryRow(q, int64(res.Hash))
		err = qr.Scan(&rid)
		if err != nil {
			switch err.Error() {
			case "sql: no rows in result set":
				err = nil
			default:
				logrus.Errorf("unable to select from 'resources' for url='%s': %s", res.url, err.Error())
				break
			}
		}
		if err != nil {
			break
		}
	}
	return SqlId(id), err
}

func (me *Storage) LoadResourceProps(res Resource) (r *Resource, err error) {
	for range only.Once {
		var qr *sql.Row
		switch {
		case res.Id != 0:
			qr = me.QueryRow(dml[SelectResourceByIdDml], int64(res.Id))
		case res.Hash != 0:
			qr = me.QueryRow(dml[SelectResourceByHashDml], int64(res.Hash))
		}
		if qr == nil {
			break
		}
		var rid int64
		var hid int64
		var hash int64
		tr := Resource{}
		err = qr.Scan(&rid, &hash, &hid, &tr.UrlPath)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				//err = nil
				break
			default:
				logrus.Errorf("unable to select for resource '%s': %s", res.String(), err.Error())
				break
			}
		}
		h, err := me.LoadHostById(SqlId(hid))
		if err != nil {
			break
		}
		tr.host = h
		tr.Id = SqlId(rid)
		tr.Hash = Hash(hash)
		var u global.Url
		u, err = tr.Url()
		if err != nil {
			break
		}
		tr.url = u
		r = &tr
	}
	return r, err
}

func (me *Storage) LoadResourceByUrl(u global.Url) (r *Resource, err error) {
	return me.LoadResourceByHash(NewHash(u))
}

func (me *Storage) LoadResourceByHash(hash Hash) (r *Resource, err error) {
	for range only.Once {
		qr := me.QueryRow(dml[SelectResourceByHashDml], int64(hash))

		var rid int64
		var h int64
		var hid int64
		r = &Resource{}
		err = qr.Scan(&rid, &h, &hid, &r.UrlPath)
		if err == sql.ErrNoRows {
			err = nil
			break
		}
		if err != nil {
			logrus.Errorf("unable to select from 'requests' for hash='%d': %s", hash, err)
			break
		}
		r.host, err = me.LoadHostById(SqlId(hid))
		if err != nil {
			logrus.Errorf("unable to load from 'hosts' for ID '%d': %s", hid, err)
			break
		}
		r.Id = SqlId(rid)
		r.Hash = Hash(h)
	}
	return r, err
}
