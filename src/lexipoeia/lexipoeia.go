package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	flag.Usage = usage
	flag.Parse()

	if flag.NArg() < 1 {
		usage()
	}

	spec := LoadSpecification(flag.Arg(0))
	output := ""
	if flag.NArg() > 1 {
		output = flag.Arg(1)
	} else {
		output = flag.Arg(0) + ".words"
	}
	Generate(spec, output)
}

func usage() {
	fmt.Fprintf(os.Stderr, "usage: lexipoeia inputfile *outputfile\n")
	flag.PrintDefaults()
	os.Exit(2)
}
