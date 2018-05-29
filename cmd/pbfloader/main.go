package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strconv"
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

		//cache, _ := bigcache.NewBigCache(bigcache.DefaultConfig(0 * time.Minute))
		var cache map[int64]*osmpbf.Node
		cache = make(map[int64]*osmpbf.Node)

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
					//if val, ok := v.Tags["name"]; ok {
					//fmt.Println("name: " + val)
					//}
					//jval, error := json.Marshal(v)
					//if error == nil {
					//cache.Set(strconv.FormatInt(v.ID, 10), jval)
					cache[v.ID] = v
					//wbin, _ := cache.Get(strconv.FormatInt(v.ID, 10))
					//fmt.Println(string(wbin))
					//}
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
		g, err := os.Open(path)
		if err != nil {
			log.Fatal(err)
		}
		defer g.Close()
		d = osmpbf.NewDecoder(g)
		// start decoding with several goroutines, it is faster
		err = d.Start(runtime.GOMAXPROCS(-1))
		if err != nil {
			log.Fatal(err)
		}

		for {
			if v, err := d.Decode(); err == io.EOF {
				break
			} else if err != nil {
				log.Fatal(err)
			} else {
				switch v := v.(type) {
				case *osmpbf.Node:
					nc++
				case *osmpbf.Way:
					d := &directededge.DirectedEdge{}
					endnodeid := v.NodeIDs[0]
					//wbin, _ := cache.Get(strconv.FormatInt(endnodeid, 10))
					if wbin, ok := cache[endnodeid]; ok {
						//var node osmpbf.Node
						//e := json.Unmarshal(wbin, &node)
						//if e != nil {
						//	d.Endnode = uint64(node.ID)
						d.Endnode = uint64(wbin.ID)
						//	fmt.Println(strconv.FormatUint(d.Endnode, 10))
						//} else {
						//	log.Fatal(e)
						//}
						fmt.Println(strconv.FormatUint(d.Endnode, 10))
					}
					wc++
				case *osmpbf.Relation:
					// Process Relation v.
					rc++
				default:
					log.Fatalf("unknown type %T\n", v)
				}
			}
		}
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
