package hosters

import "website-indexer/global"

type Object struct {
	global.Object
}

func NewObject(obj global.Object) *Object {
	o := Object{}
	o.Object = obj
	return &o
}

func (me *Object) AppendProps(propMap global.StringMap) {
	for n, v := range propMap {
		me.Object[n] = v
	}
}
