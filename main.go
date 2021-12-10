package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/phk13/dirsync/compare"
	"github.com/phk13/dirsync/filesync"
	"github.com/phk13/dirsync/utils"
)


func main() {
	version := "2.0"
	var exceptions utils.ExceptList
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] directory1 directory2\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Var(&exceptions, "x", "define a extension to ignore (can be done multiple times).")
	origin := flag.String("d1", "", "directory #1 to sync.")
	destination := flag.String("d2", "", "directory #2 to sync.")
	threads := flag.Int("t", 4, "Number of threads to scan each directory (total threads would be double of this number).")
	flag.Parse()
	fmt.Println("-----------------------------")
	fmt.Printf("--------Dir Sync v%s--------\n\n", version)
	
	if *origin == "" || *destination == "" {
		flag.Usage()
		return
	}
	start := time.Now()
	root1 := filepath.Clean(*origin)
	root2 := filepath.Clean(*destination)
	fmt.Println("-----------------------------")
	fmt.Println("-------Checking files--------")
	fmt.Println()
	files, err := compare.WalkFiles(root1, root2, exceptions, *threads)
	fmt.Printf("WalkFiles -> %s\n", time.Since(start))
	start = time.Now()
	if err != nil {
		fmt.Println(err)
		return
	}
	diff := compare.Compare(files)
	fmt.Printf("Compare -> %s\n", time.Since(start))
	fmt.Printf("%d / %d different files.\n", len(diff), len(files))

	fmt.Println("-----------------------------")
	fmt.Println("--Preparing synchronization--")
	fmt.Println()
	copyMoves, err := filesync.GetActions(root1, root2, diff)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("-----------------------------")
	fmt.Println("--------Synchronizing--------")
	err = filesync.SyncFiles(copyMoves)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
	fmt.Println("-----------------------------")
	fmt.Println("----------Finished-----------")
	fmt.Println("-----------------------------")
}
