package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var (
	DIR = flag.String("dir", "", "directory to scan")
)

func directoryListing(dirname string) ([]string, error) {
	res := []string{}
	err := filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		res = append(res, path)
		return nil
	})
	return res, err
}

func FindDuplicates(dirname string) (map[string][]string, error) {
	listing, err := directoryListing(*DIR)
	if err != nil {
		return nil, err
	}

	pathByName := map[string][]string{}
	for _, path := range listing {
		basename := filepath.Base(path)
		v, ok := pathByName[basename]
		if !ok {
			pathByName[basename] = []string{path}
		} else {
			pathByName[basename] = append(v, path)
		}
	}
	duplicates := map[string][]string{}
	for basename, paths := range pathByName {
		if len(paths) > 1 {
			duplicates[basename] = paths
		}
	}
	return duplicates, nil
}

func main() {
	flag.Parse()
	if DIR == nil || *DIR == "" {
		flag.Usage()
		return
	}
	duplicates, err := FindDuplicates(*DIR)
	if err != nil {
		log.Fatal(err.Error())
		return
	}
	for basename, paths := range duplicates {
		fmt.Printf("%s:\n", basename)
		for _, path := range paths {
			fmt.Println("    *", path)
		}
	}
}
