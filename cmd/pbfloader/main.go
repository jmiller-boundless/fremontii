package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"time"

	"github.com/jmiller-boundless/fremontii"

	"github.com/qedus/osmpbf"
)

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

		var cache map[int64]*directededge.SimpleNode
		cache = make(map[int64]*directededge.SimpleNode)

		var sws []*directededge.SimpleWay

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
					//sn := &SimpleNode{}
					sn := &directededge.SimpleNode{}
					sn.Id = uint64(v.ID)
					sn.Lat = float32(v.Lat)
					sn.Lon = float32(v.Lon)
					if isBarrier(v) {
						sn.Countflag = 1 | 0x20
					} else {
						sn.Countflag = 1
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
		//f.Close()
		elapsed := time.Since(start)
		start2 := time.Now()
		log.Printf("Node storage took: ", elapsed)

		//for _, value := range cache {
		//	if value.Countflag != 1 {
		//		PrettyPrint(value)
		//	}
		//}

		g, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer g.Close()
		q := osmpbf.NewDecoder(g)

		// use more memory from the start, it is faster
		q.SetBufferSize(osmpbf.MaxBlobSize)

		// start decoding with several goroutines, it is faster
		err = q.Start(runtime.GOMAXPROCS(-1))
		if err != nil {
			log.Fatal(err)
		}

		for {
			if v, err := q.Decode(); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			} else {
				switch v := v.(type) {
				case *osmpbf.Node:
					nc++
				case *osmpbf.Way:
					//d := &directededge.DirectedEdge{}
					d := &directededge.SimpleWay{}
					d.Nodes = make([]*directededge.SimpleNode, len(v.NodeIDs))
					for key, nodeid := range v.NodeIDs {
						if val, ok := cache[nodeid]; ok {
							d.Nodes[key] = val
						}
					}
					processWayTags(v, d)
					sws = append(sws, d)
					wc++
				case *osmpbf.Relation:
					// Process Relation v.
					rc++
				default:
					log.Fatalf("unknown type %T\n", v)
				}
			}
		}
		//g.Close()
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

func processWayTags(w *osmpbf.Way, sw *directededge.SimpleWay) {
	if val, ok := w.Tags["oneway"]; ok {
		sw.Oneway = val
	}
	if val, ok := w.Tags["highway"]; ok {
		sw.Highway = val
	}
	if val, ok := w.Tags["junction"]; ok {
		sw.Junction = val
	}
	if val, ok := w.Tags["access"]; ok {
		sw.Access = val
	}
	if val, ok := w.Tags["motor_vehicle"]; ok {
		sw.MotorVehicle = val
	}
	if val, ok := w.Tags["service"]; ok {
		sw.Service = val
	}
	if val, ok := w.Tags["area"]; ok {
		sw.Area = val
	}
}
