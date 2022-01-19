package compare

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"

	"github.com/phk13/dirsync/utils"
)

type fileResult struct {
	path string
	info os.FileInfo
}

func checkDir(dirChannel chan string, resultChannel chan fileResult, wg *sync.WaitGroup, exceptions []string, counter *int32){
	defer wg.Done()
	for dir := range dirChannel{
		files, err := ioutil.ReadDir(dir)
		if err != nil {
			fmt.Print(err)
			os.Exit(1)
		}
		for _, file := range files {
			if file.IsDir() {
				atomic.AddInt32(counter, 1)
				dirChannel <- filepath.Join(dir, file.Name())
			} else if !utils.HasExtension(exceptions, file.Name()) {
				resultChannel <- fileResult{filepath.Join(dir, file.Name()), file}
			}
		}
		atomic.AddInt32(counter, -1)
		if atomic.CompareAndSwapInt32(counter, 0, -1){
			close(dirChannel)
		}
	}
}

func WalkFiles(root1 string, root2 string, exceptions []string, threads int) ([]CompareResult, error) {
	
	var wg = sync.WaitGroup{}
	root1DirChan := make(chan string, 1000)
	root1ResultChan := make(chan fileResult, 1000)
	
	root2DirChan := make(chan string, 1000)
	root2ResultChan := make(chan fileResult, 1000)
	var counter1, counter2 int32 = 0, 0

	atomic.AddInt32(&counter1, 1)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go checkDir(root1DirChan, root1ResultChan, &wg, exceptions, &counter1)
	}
	root1DirChan <- root1

	atomic.AddInt32(&counter2, 1)
	for i := 0; i < threads; i++ {
		wg.Add(1)
		go checkDir(root2DirChan, root2ResultChan, &wg, exceptions, &counter2)
	}
	root2DirChan <- root2

	wg.Wait()
	close(root1ResultChan)
	close(root2ResultChan)
	
	var files []CompareResult

	for result := range(root1ResultChan) {
		if info2, err := os.Stat(root2 + result.path[len(root1):]); err == nil {
			files = append(files, CompareResult{result.path[len(root1):], result.info, info2})
		} else {
			files = append(files, CompareResult{result.path[len(root1):], result.info, nil})
		}
	}
	for result := range(root2ResultChan) {
		if _, err := os.Stat(root1 + result.path[len(root2):]); err != nil {
			files = append(files, CompareResult{result.path[len(root2):], nil, result.info})
		}
	}
	return files, nil
}
