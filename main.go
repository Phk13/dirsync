package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/mattn/go-tty"
	"github.com/schollz/progressbar/v3"
)

func setupCloseHandler() {
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		fmt.Println("\n- Cancelled.")
		os.Exit(0)
	}()
}

func hasExtension(s []string, str string) bool {
	for _, v := range s {
		if len(str) > len(v) {
			if v == str[len(str)-len(v):] {
				return true
			}
		}
	}
	return false
}

func walkFiles(root1 string, root2 string, exceptions []string) ([]result, error) {
	var files []result
	errWalk := filepath.Walk(root1, func(path string, info os.FileInfo, err error) error { // HL
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || hasExtension(exceptions, path) {
			return nil
		}
		if info2, err := os.Stat(root2 + path[len(root1):]); err == nil {
			files = append(files, result{path[len(root1):], info, info2})
		} else {
			files = append(files, result{path[len(root1):], info, nil})
		}
		return nil
	})
	if errWalk != nil {
		return nil, errWalk
	}
	errWalk = filepath.Walk(root2, func(path string, info os.FileInfo, err error) error { // HL
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || hasExtension(exceptions, path) {
			return nil
		}
		if _, err := os.Stat(root1 + path[len(root2):]); err != nil {
			files = append(files, result{path[len(root2):], nil, info})
		}
		return nil
	})
	if errWalk != nil {
		return nil, errWalk
	}
	return files, nil
}

type result struct {
	path string
	f1   os.FileInfo
	f2   os.FileInfo
}

func compare(files []result) []result {
	var diff []result
	for _, file := range files {
		if file.f1 == nil || file.f2 == nil || file.f1.Size() != file.f2.Size() {
			diff = append(diff, file)
		}
	}
	return diff
}

func askAction(def bool) bool {
	answer := "?"
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()
	for answer != "y" && answer != "n" && answer != "\n" {
		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		answer = string(r)
		answer = strings.ToLower(answer)
	}
	fmt.Printf("%s\n", answer)
	if answer == "y" || (answer == "" && def == true) {
		return true
	}
	return false
}

func askActionSides(def bool) (bool, string) {
	answer := "?"
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()
	for answer != "<" && answer != ">" && answer != "n" {
		r, err := tty.ReadRune()
		if err != nil {
			log.Fatal(err)
		}
		answer = string(r)
		answer = strings.ToLower(answer)
	}
	fmt.Printf("%s\n", answer)
	if answer == ">" || answer == "<" {
		return true, answer
	}
	return false, ""
}

func getActions(root1 string, root2 string, files []result) ([][]string, error) {
	var copyMoves [][]string
	for _, file := range files {
		if file.f2 == nil {
			fmt.Printf("New file on %s%s - Copy? (Y/n)): ", root1, file.path)
			if askAction(true) {
				copyMoves = append(copyMoves, []string{root1 + file.path, root2 + file.path})
			}
		} else if file.f1 == nil {
			fmt.Printf("New file on %s%s - Copy? (Y/n)): ", root2, file.path)
			if askAction(true) {
				copyMoves = append(copyMoves, []string{root2 + file.path, root1 + file.path})
			}
		} else {
			if file.f1.ModTime().After(file.f2.ModTime()) {
				fmt.Printf("File %s%s is newer - Copy? (y/N)): ", root1, file.path)
				if askAction(false) {
					copyMoves = append(copyMoves, []string{root1 + file.path, root2 + file.path})
				}
			} else if file.f1.ModTime().Before(file.f2.ModTime()) {
				fmt.Printf("File %s%s is newer - Copy? (y/N)): ", root2, file.path)
				if askAction(false) {
					copyMoves = append(copyMoves, []string{root2 + file.path, root1 + file.path})
				}
			} else {
				fmt.Println("Modification time matches but size is different. User request")
				fmt.Printf("File %s has different sizes but same modification time.\n", file.path)
				fmt.Printf("< - %s%s - Size: %d - ModTime: %v\n", root1, file.path, file.f1.Size(), file.f1.ModTime())
				fmt.Printf("> - %s%s - Size: %d - ModTime: %v\n", root2, file.path, file.f2.Size(), file.f2.ModTime())
				fmt.Printf("Copy? (</>/n)): ")
				if ok, side := askActionSides(false); ok {
					if side == "<" {
						copyMoves = append(copyMoves, []string{root1 + file.path, root2 + file.path})
					} else if side == ">" {
						copyMoves = append(copyMoves, []string{root2 + file.path, root1 + file.path})
					}
				}
			}
		}
	}
	return copyMoves, nil
}

func create(p string) (*os.File, error) {
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		return nil, err
	}
	return os.Create(p)
}

func copy(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	bar := progressbar.NewOptions(int(sourceFileStat.Size()),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetDescription("Copying..."),
		progressbar.OptionFullWidth(),
		progressbar.OptionSetTheme(progressbar.Theme{
			Saucer:        "=",
			SaucerHead:    ">",
			SaucerPadding: " ",
			BarStart:      "[",
			BarEnd:        "]",
		}))
	_, err = io.Copy(io.MultiWriter(destination, bar), source)
	return err
}

func syncFiles(actions [][]string) error {
	for i, m := range actions {
		fmt.Printf("\n(%d/%d) \t- Copying %s \t\t-> \t%s\n", i+1, len(actions), m[0], m[1])
		err := copy(m[0], m[1])
		if err != nil {
			return err
		}
	}
	return nil
}

type exceptList []string

func (i *exceptList) String() string {
	return "exceptions"
}

func (i *exceptList) Set(value string) error {
	*i = append(*i, value)
	return nil
}

var exceptions exceptList

func main() {
	setupCloseHandler()
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage of %s [options] directory1 directory2\n", os.Args[0])
		flag.PrintDefaults()
	}
	flag.Var(&exceptions, "x", "define a extension to ignore (can be done multiple times)")
	origin := flag.String("d1", "", "directory #1 to sync.")
	destination := flag.String("d2", "", "directory #2 to sync.")
	flag.Parse()
	fmt.Println("-----------------------------")
	fmt.Println("--------Dir Sync v1.0--------")
	fmt.Println()
	if *origin == "" || *destination == "" {
		fmt.Println("Usage: godirsync [options] -d1 pathToDir1 -d2 pathToDir2")
		return
	}

	root1 := filepath.Clean(*origin)
	root2 := filepath.Clean(*destination)
	fmt.Println("-----------------------------")
	fmt.Println("-------Checking files--------")
	fmt.Println()
	files, err := walkFiles(root1, root2, exceptions)
	if err != nil {
		fmt.Println(err)
		return
	}
	diff := compare(files)
	fmt.Printf("%d / %d different files.\n", len(diff), len(files))

	fmt.Println("-----------------------------")
	fmt.Println("--Preparing synchronization--")
	fmt.Println()
	copyMoves, err := getActions(root1, root2, diff)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println("-----------------------------")
	fmt.Println("--------Synchronizing--------")
	err = syncFiles(copyMoves)
	if err != nil {
		fmt.Println(err)
		return
	}
	fmt.Println()
	fmt.Println("-----------------------------")
	fmt.Println("----------Finished-----------")
	fmt.Println("-----------------------------")
}
