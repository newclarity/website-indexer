package persist

import (
	"database/sql"
	"fmt"
	"github.com/gearboxworks/go-status/only"
	"github.com/sirupsen/logrus"
	"website-indexer/global"
)

func (me *Storage) InsertHost(host Host) (h *Host, sr sql.Result, err error) {
	for range only.Once {
		if host.Domain == "" {
			err = host.Init()
			if err != nil {
				break
			}
		}
		sr, err = me.ExecSql(dml[InsertHostDml],
			host.Scheme,
			host.Domain,
			host.Port,
		)
		if err != nil {
			err = fmt.Errorf("unable to insert host '%s': %s", host.url, err.Error())
			logrus.Error(err)
			break
		}
		h = &host
		var hid int64
		hid, err = sr.LastInsertId()
		if err != nil {
			err = fmt.Errorf("unable to access inserted ID for host '%s': %s", host.url, err.Error())
			logrus.Error(err)
			break
		}
		h.Id = SqlId(hid)

	}
	return h, sr, err
}

func (me *Storage) LoadHostByUrl(u global.Url) (h *Host, err error) {
	for range only.Once {
		var hid SqlId
		hid, err = me.LoadHostIdByUrl(u)
		if err != nil {
			break
		}
		h, err = me.LoadHostById(hid)
		if err != nil {
			break
		}
	}
	return h, err
}

func (me *Storage) LoadHostIdByUrl(u global.Url) (hostid SqlId, err error) {
	for range only.Once {
		u, err = getRootUrl(u)
		if err != nil {
			break
		}
		r := me.QueryRow(dml[SelectHostByUrlDml], u+"%")
		var hid int64
		err = r.Scan(&hid)
		if err == nil {
			hostid = SqlId(hid)
			break
		}
		if err == sql.ErrNoRows {
			err = nil
			break
		}
		err = fmt.Errorf("unable to select from 'hosts' for url='%s': %s", u, err.Error())
		logrus.Error(err)
	}
	return hostid, err
}

func (me *Storage) LoadHostById(hid SqlId) (h *Host, err error) {
	for range only.Once {
		h, err = me.LoadHostProps(Host{
			Id: hid,
		})
		if err != nil {
			err = fmt.Errorf("unable to load host by ID '%d': %s", hid, err)
			logrus.Error(err)
			break
		}
	}
	return h, err
}

func (me *Storage) LoadHost(host Host) (h *Host, err error) {
	for range only.Once {
		h = &Host{}
		if !host.Initialized() {
			err = host.Init()
		}
		if err != nil {
			break
		}
		if host.Id != 0 {
			h, err = me.LoadHostProps(host)
		} else {
			var hid SqlId
			hid, err = me.LoadHostId(host)
			host.Id = hid
		}
		if err != nil {
			break
		}
		h = &host
	}
	return h, err
}

func (me *Storage) LoadHostId(host Host) (hid SqlId, err error) {
	var id int64
	for range only.Once {
		q := dml[SelectHostBySDPDml]
		r := me.QueryRow(q, host.Scheme, host.Domain, host.Port)
		err = r.Scan(&id)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				err = nil
			default:
				err = fmt.Errorf("unable to select from 'hosts' for url='%s': %s", host.url, err)
				logrus.Error(err)
				break
			}
		}
		if err != nil {
			break
		}
	}
	return SqlId(id), err
}

func (me *Storage) LoadHostProps(host Host) (h *Host, err error) {
	for range only.Once {
		h = &Host{}
		r := me.QueryRow(dml[SelectHostByIdDml], host.Id)
		var hid int64
		var s global.Protocol
		var d global.Domain
		var p global.Port
		err = r.Scan(&hid, &s, &d, &p)
		if err != nil {
			switch err {
			case sql.ErrNoRows:
				err = nil
			default:
				err = fmt.Errorf("unable to select from 'hosts' for Id='%d': %s", host.Id, err)
				logrus.Error(err)
				break
			}
		}
		if err != nil {
			break
		}
		h.Id = SqlId(hid)
		h.Domain = d
		h.Scheme = s
		h.Port = p
		h.url = h.Url()
	}
	return h, err
}

func (me *Storage) LoadHostByResource(res Resource) (h *Host, err error) {
	var hid SqlId
	for range only.Once {
		if res.host != nil {
			hid = res.Host().Id
			break
		}
		var u global.Url
		u, err = res.Url()
		if err != nil {
			break
		}
		hid, _, err = me.AddHost(u)
		if err != nil {
			break
		}
	}
	if err == nil {
		h, err = me.LoadHostById(hid)
	}
	return h, err
}
