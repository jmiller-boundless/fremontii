package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/qedus/osmpbf"
)

type SimpleWay struct {
	id           int64
	nodeids      []int64
	oneway       string
	highway      string
	junction     string
	access       string
	motorVehicle string
	service      string
	area         string
}

type SimpleNode struct {
	id        int64
	lat       float64
	lon       float64
	countflag byte
}

func main() {
	argsWithoutProg := os.Args[1:]
	if len(argsWithoutProg) == 1 {
		path := argsWithoutProg[0]
		fmt.Println("Path:" + path)

		f, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer f.Close()

		var cache map[int64]*SimpleNode
		cache = make(map[int64]*SimpleNode)

		d := osmpbf.NewDecoder(f)

		// use more memory from the start, it is faster
		d.SetBufferSize(osmpbf.MaxBlobSize)
		start := time.Now()
		// start decoding with several goroutines, it is faster
		err = d.Start(runtime.GOMAXPROCS(-1))
		if err != nil {
			log.Fatal(err)
		}

		var nc, wc, rc uint64
		for {
			if v, err := d.Decode(); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			} else {
				switch v := v.(type) {
				case *osmpbf.Node:
					sn := &SimpleNode{}
					sn.id = v.ID
					sn.lat = v.Lat
					sn.lon = v.Lon
					if isBarrier(v) {
						sn.countflag = 1 | 0x20
					} else {
						sn.countflag = 1
					}
					cache[v.ID] = sn
					nc++
				case *osmpbf.Way:
					//d := &directededge.DirectedEdge{}
					wc++
				case *osmpbf.Relation:
					// Process Relation v.
					rc++
				default:
					log.Fatalf("unknown type %T\n", v)
				}
			}
		}
		elapsed := time.Since(start)
		start2 := time.Now()
		log.Printf("Node storage took: ", elapsed)

		elapsed2 := time.Since(start2)
		log.Printf("Way processing took: ", elapsed2)
		fmt.Printf("Nodes: %d, Ways: %d, Relations: %d\n", nc, wc, rc)

		if err != nil {
			fmt.Println(err)
			return
		}
	} else {
		fmt.Println("Importing the OSM file requires the path to the pbf file")
	}
}

func PrettyPrint(v interface{}) (err error) {
	b, err := json.MarshalIndent(v, "", "  ")
	if err == nil {
		fmt.Println(string(b))
	}
	return
}

func serializeNode(v interface{}) (bytes []byte, err error) {
	return json.Marshal(v)

}

func isBarrier(n *osmpbf.Node) bool {
	out := false
	//("barrier") && (value.equals("gate") || value.equals("bollard") || value.equals("lift_gate"))) {

	if val, ok := n.Tags["barrier"]; ok {
		barriers := map[string]bool{
			"gate":      true,
			"bollard":   true,
			"lift_gate": true,
		}
		if barriers[val] {
			out = true
		}
	}

	return out
}
