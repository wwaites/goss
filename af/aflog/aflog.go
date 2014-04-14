package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"hubs.net.uk/oss/event"
	"hubs.net.uk/oss/log"
	"hubs.net.uk/oss/rcmd"
	"labix.org/v2/mgo"
	"os"
	"regexp"
	"strconv"
	"time"
)

var mongo string

func init() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, `
Usage: %s [flags]

`, os.Args[0])
		flag.PrintDefaults()
	}
	flag.StringVar(&mongo, "db", "localhost:27017", "Mongo Database")
}

func NewWriter(m *mgo.Session) (chan event.Event) {
	in := make(chan event.Event)
	collection := m.DB("af24").C("event")
	go func() {
		for {
			select {
			case e, ok := <- in:
				if !ok {
					break 
				}
				err := collection.Insert(&e)
				if err != nil {
					panic(err)
				}
			}
		}
	}()
	return in
}

var sectionRe *regexp.Regexp
var kvpRe *regexp.Regexp

func init() {
	sectionRe = regexp.MustCompile("^ +[*]+ ([^*]+): [*]+$")
	kvpRe = regexp.MustCompile("^ +([^ ]+): ([^ ]+)$")
}

func coerce(s string) interface{} {
	intv, err := strconv.ParseInt(s, 10, 32)
	if err == nil {
		return intv
	}
	floatv, err := strconv.ParseFloat(s, 32)
	if err == nil {
		return floatv
	}
	return s
}

func ParseAF(data []byte) (e event.Event, err error) {
	e, err = event.New() // make(map[string]interface{})
	if err != nil {
		return
	}
	var section string
	var sdata map[string]interface{}
	for _, line := range bytes.Split(data, []byte("\n")) {
		if len(line) > 0 && line[len(line)-1] == '\r' {
			line = line[:len(line)-1]
		}
		if len(line) == 0 {
			continue
		}
		m := sectionRe.FindSubmatch(line)
		if m != nil {
//			log.Info.Print(string(m[1]))
			switch string(m[1]) {
			case "Local Data": 
				section = "local"
				sdata = make(map[string]interface{})
				e[section] = sdata 
				break
			case "Remote Data":
				section = "remote"
				sdata = make(map[string]interface{})
				e[section] = sdata
				break
			case "Local Config":
				section = "config"
				sdata = make(map[string]interface{})
				e[section] = sdata
				break
			case "link Data":
				section = "link"
				sdata = make(map[string]interface{})
				e[section] = sdata
				break
			default:
				err = errors.New("unknown section: " + string(m[1]))
				return
			}
			continue
		}
		m = kvpRe.FindSubmatch(line)
		if m != nil {
			key, val := string(m[1]), string(m[2])
//			log.Info.Printf("%s = %s", key, val)
			if sdata == nil {
				err = errors.New("kvp before section")
				return
			}
			sdata[key] = coerce(val)
			continue
		}
		err = errors.New("unknown input: " + string(line))
		return
	}
	return
}

const testData = `
       ******* Local Data: *******
                   status: slave-operational
                 rxpower0: -67
                 rxpower1: -67
               rxcapacity: 771989760
                txmodrate: 6x
                 gpspulse: detected
                   dpstat: 1000Mbps-Full
                    miles: 3.860
                     feet: 20380
                    rssi0: 41
                    rssi1: 41
                    temp0: 17
                    temp1: 20
      ******* Remote Data: ********
                rrxpower0: -68
                rrxpower1: -67
               txcapacity: 761975040
               rtxmodrate: 6x
                rpowerout: 33
     ******* Local Config: ********
                 powerout: 33
                   rxgain: high
              txfrequency: 24.2GHz
              rxfrequency: 24.1GHz
                   duplex: full
               modcontrol: automatic
                    speed: 6x
                      gps: on
                 gpspulse: detected
                 linkname: TEGOLA24........................
                      key: 0000:0000:0000:0000:0000:0000:0000:0000
        ******* link Data: ********
                 baseline: -101
                     fade: 1
`

const interval = 10e9
func LogAF(host string, writer chan event.Event) {
	var session rcmd.Session
	var err error

	for {
		if session == nil {
			session, err = rcmd.NewExpSession("admin", host, "^AF\\.v[0-9.]+# $")
			if err != nil {
				log.Err.Printf("%s %s", host, err)
				time.Sleep(60e9)
				continue
			}
		}
		output, err := session.Exec("af af")
		if err != nil {
			log.Err.Printf("%s: %s", host, err)
			session.Close()
			session = nil
			continue
		}
		data, err := ParseAF([]byte(output))
		if err != nil {
			log.Err.Printf("%s: %s", host, err)
			time.Sleep(interval)
			continue
		}
		data["host"] = host
		data["type"] = "aflog"

		log.Debug.Printf("%v", data)
		writer <- data
		time.Sleep(interval)
	}
}

func main() {
	s, err := mgo.Dial(mongo)
	if err != nil {
		log.Err.Fatal(err)
	}

	writer := NewWriter(s)

	go LogAF("af24-mah.west", writer)
	go LogAF("af24-cor.west", writer)
	for {
		time.Sleep(60e9)
	}
	close(writer)
}
