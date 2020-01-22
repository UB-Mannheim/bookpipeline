package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	"rescribe.xyz/bookpipeline"
	"rescribe.xyz/utils/pkg/hocr"
)

func main() {
	flag.Usage = func() {
		fmt.Fprintln(flag.CommandLine.Output(), "Usage: pagegraph file.hocr graph.png")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() != 2 {
		flag.Usage()
		return
	}

	wordconfs, err := hocr.GetWordConfs(flag.Arg(0))
	if err != nil {
		log.Fatal(err)
	}
	var confs []*bookpipeline.Conf
	for n, wc := range wordconfs {
		c := bookpipeline.Conf{
			Conf: wc,
			//Path: "fakepath",
			Path: fmt.Sprintf("word_%d", n),
		}
		confs = append(confs, &c)
	}

	// Structure to fit what bookpipeline.Graph needs
	// TODO: probably reorganise bookpipeline to just need []*Conf
	cconfs := make(map[string]*bookpipeline.Conf)
	for _, c := range confs {
		cconfs[c.Path] = c
	}

	fn := flag.Arg(1)
	f, err := os.Create(fn)
	if err != nil {
		log.Fatalln("Error creating file", fn, err)
	}
	defer f.Close()
	err = bookpipeline.GraphOpts(cconfs, filepath.Base(flag.Arg(0)), "Word number", false, f)
	if err != nil {
		log.Fatalln("Error creating graph", err)
	}
}
