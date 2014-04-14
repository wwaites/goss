package event

import (
	"errors"
	"hubs.net.uk/util"
        "labix.org/v2/mgo/bson"
	"time"
)

type Event map[string]interface{}

func New() (e Event, err error) {
	e = make(Event)
	e["uuid"], err = util.Uuid()
	if err != nil {
		return
	}
	e["timestamp"] = time.Now()
	return
}

func (e Event) Marshal() ([]byte, error) {
	return bson.Marshal(e)
}

func (e Event) Created() (t time.Time, err error) {
	t, ok := e["timestamp"].(time.Time)
	if !ok {
		err = errors.New("timestamp invalid or of wrong type")
		return
	}
	return
}

func Unmarshal(buf []byte) (e Event, err error) {
	err = bson.Unmarshal(buf, &e)
	return
}
