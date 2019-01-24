package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"git.rescribe.xyz/testingtools/parse"
	"git.rescribe.xyz/testingtools/parse/prob"
)

func main() {
	b := parse.BucketSpecs{
		// minimum confidence, name
		{ 0, "bad" },
		{ 0.95, "95to98" },
		{ 0.98, "98plus" },
	}

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: bucket-lines-prob [-d dir] prob1 [prob2] [...]\n")
		fmt.Fprintf(os.Stderr, "Copies image-text line pairs into different directories according\n")
		fmt.Fprintf(os.Stderr, "to the average character probability for the line.\n")
		fmt.Fprintf(os.Stderr, "This uses the .prob files generated by ocropy-rpred's --probabilities\n")
		fmt.Fprintf(os.Stderr, "option, which it assumes will be in the same directory as the line's\n")
		fmt.Fprintf(os.Stderr, "image and text files.\n")
		flag.PrintDefaults()
	}
	dir := flag.String("d", "buckets", "Directory to store the buckets")
	flag.Parse()
	if flag.NArg() < 1 {
		flag.Usage()
		os.Exit(1)
	}

	lines := make(parse.LineDetails, 0)

	for _, f := range flag.Args() {
		newlines, err := prob.GetLineDetails(f)
		if err != nil {
			log.Fatal(err)
		}

                for _, l := range newlines {
                        lines = append(lines, l)
                }
	}

	stats, err := parse.BucketUp(lines, b, *dir)
	if err != nil {
		log.Fatal(err)
	}

	parse.PrintBucketStats(os.Stdout, stats)
}