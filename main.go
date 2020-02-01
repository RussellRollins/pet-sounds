package main

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"
)

const (
	defaultFileName = "pets.hcl"
)

func main() {
	if err := inner(); err != nil {
		fmt.Printf("pet-sounds error: %s\n", err.Error())
		os.Exit(1)
	}
}

func inner() error {
	var inputFile string
	flag.StringVar(&inputFile, "file", defaultFileName, "the file to read pet configuration from")
	flag.StringVar(&inputFile, "f", defaultFileName, "the file to read pet configuration from (shorthand)")
	flag.Parse()

	// There is a random function for the HCL configuration.
	rand.Seed(time.Now().Unix())

	pets, err := ReadConfig(inputFile)
	if err != nil {
		return err
	}

	for _, p := range pets {
		p.Say()
		p.Act()
	}

	return nil
}
