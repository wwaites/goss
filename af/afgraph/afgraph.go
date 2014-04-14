package main

import (
	"flag"
	"fmt"
	"hubs.net.uk/event"
	"hubs.net.uk/log"
	"github.com/ziutek/rrd"
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
	"os"
	"path"
	"time"
)

var mongo string
var output string
var overwrite bool
func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage: %s [flags]

`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&mongo, "db", "localhost:27017", "Mongo Database")
	flag.StringVar(&output, "o", ".", "Output Directory")
	flag.BoolVar(&overwrite, "x", false, "Overwrite")
}

func main() {
	flag.Parse()

	s, err := mgo.Dial(mongo)
	if err != nil {
		log.Err.Fatal(err)
	}
	defer s.Close()

	var hosts []string

	collection := s.DB("af24").C("event")
	err = collection.Find(nil).Distinct("host", &hosts)
	if err != nil {
		log.Err.Fatal(err)
	}

	for _, host := range hosts {
		filename := path.Join(output, host + ".rrd")
		log.Info.Printf("processing %s", host)

		var q bson.M
		info, err := rrd.Info(filename)
		if err != nil {
			q = bson.M{"type": "aflog", "host": host}
		} else {
			lasti, ok := info["last_update"]
			if !ok {
				log.Err.Fatal("rrd file exists, but no last update")
			}
			last, ok := lasti.(uint)
			if !ok {
				log.Err.Fatalf("last update is of wrong type: %#v", lasti)
			}
			utime := time.Unix(int64(last), 0)
			q = bson.M{"type": "aflog", "host": host, "timestamp": bson.M{ "$gt": utime }}
		}

		it := collection.Find(q).Sort("timestamp").Iter()
		if it == nil {
			log.Err.Fatal("iterator is nil!")
		}

		var creator *rrd.Creator
		var updater *rrd.Updater
		var e event.Event
		for it.Next(&e)  {
			if creator == nil {
				start := e["timestamp"].(time.Time)
				creator = rrd.NewCreator(filename, start, 10)
				creator.DS("rx0", "GAUGE", 60, -120, 0)
				creator.DS("rx1", "GAUGE", 60, -120, 0)
				creator.DS("cap", "GAUGE", 60, 0, 1e9)
				creator.DS("pow", "GAUGE", 60, -50, 50)
				creator.RRA("AVERAGE", 0, 1, 6*60*24*365)
				err = creator.Create(overwrite)
				if err != nil && !os.IsExist(err) {
					log.Err.Fatal(err)
				}
			}
			if updater == nil {
				updater = rrd.NewUpdater(filename)
			}
			timestamp := e["timestamp"].(time.Time)

			locali, ok := e["local"]
			if !ok {
				log.Warn.Printf("[%v] no local section", e["uuid"])
				continue
			}
			local, ok := locali.(event.Event)
			if !ok {
				log.Warn.Printf("[%v] local section has wrong type: %#v", e["uuid"], locali)
				continue
			}

			rx0i, ok := local["rxpower0"]
			if !ok {
				log.Warn.Printf("[%v] local rxpower0 absent", e["uuid"])
				continue
			}
			rx0, ok := rx0i.(int64)
			if !ok {
				log.Warn.Printf("[%v] local rxpower0 has wrong type: %#v", e["uuid"], rx0i)
				continue
			}

			rx1i, ok := local["rxpower1"]
			if !ok {
				log.Warn.Printf("[%v] local rxpower1 absent", e["uuid"])
				continue
			}
			rx1, ok := rx1i.(int64)
			if !ok {
				log.Warn.Printf("[%v] local rxpower1 has wrong type: %#v", e["uuid"], rx1i)
				continue
			}

			rxcapi, ok := local["rxcapacity"]
			if !ok {
				log.Warn.Printf("[%v] local rxcapacity absent", e["uuid"])
				continue
			}
			rxcap, ok := rxcapi.(int64)
			if !ok {
				log.Warn.Printf("[%v] local rxcapacity has wrong type: %#v", e["uuid"], rxcapi)
				continue
			}

			remotei, ok := e["remote"]
			if !ok {
				log.Warn.Printf("[%v] no remote section",e["uuid"])
				continue
			}
			remote, ok := remotei.(event.Event)
			if !ok {
				log.Warn.Printf("[%v] remote section has wrong type: %#v", e["uuid"], remotei)
				continue
			}

			poweri, ok := remote["rpowerout"]
			if !ok {
				log.Warn.Printf("[%v] remote rpowerout is absent", e["uuid"])
				continue
			}
			power, ok := poweri.(int64)
			if !ok {
				log.Warn.Printf("[%v] remote rpowerout has wrong type: %v", e["uuid"], poweri)
				continue
			}
			
			err = updater.Update(timestamp.Unix(), rx0, rx1, rxcap, power)
			if err != nil {
				log.Err.Print(err)
				continue
			}
		}
	}
}
