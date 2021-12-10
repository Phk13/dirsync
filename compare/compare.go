package compare

import "os"

/* CompareResult stores a file path and its FileInfo in both root directories, or nil if it doesn't exist in either.*/
type CompareResult struct {
	Path string
	F1   os.FileInfo
	F2   os.FileInfo
}

/* Compare takes a slice of CompareResult and returns all the files that are not equal in both root directories.*/
func Compare(files []CompareResult) []CompareResult {
	var diff []CompareResult
	for _, file := range files {
		if file.F1 == nil || file.F2 == nil || file.F1.Size() != file.F2.Size() {
			diff = append(diff, file)
		}
	}
	return diff
}