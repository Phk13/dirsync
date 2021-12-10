package filesync

import (
	"fmt"

	"github.com/phk13/dirsync/compare"
	"github.com/phk13/dirsync/utils"
)

func GetActions(root1 string, root2 string, files []compare.CompareResult) ([][]string, error) {
	var copyMoves [][]string
	for _, file := range files {
		if file.F2 == nil {
			fmt.Printf("New file on %s%s - Copy? (Y/n)): ", root1, file.Path)
			if utils.AskAction(true) {
				copyMoves = append(copyMoves, []string{root1 + file.Path, root2 + file.Path})
			}
		} else if file.F1 == nil {
			fmt.Printf("New file on %s%s - Copy? (Y/n)): ", root2, file.Path)
			if utils.AskAction(true) {
				copyMoves = append(copyMoves, []string{root2 + file.Path, root1 + file.Path})
			}
		} else {
			if file.F1.ModTime().After(file.F2.ModTime()) {
				fmt.Printf("File %s%s is newer - Copy? (y/N)): ", root1, file.Path)
				if utils.AskAction(false) {
					copyMoves = append(copyMoves, []string{root1 + file.Path, root2 + file.Path})
				}
			} else if file.F1.ModTime().Before(file.F2.ModTime()) {
				fmt.Printf("File %s%s is newer - Copy? (y/N)): ", root2, file.Path)
				if utils.AskAction(false) {
					copyMoves = append(copyMoves, []string{root2 + file.Path, root1 + file.Path})
				}
			} else {
				fmt.Println("Modification time matches but size is different. User request")
				fmt.Printf("File %s has different sizes but same modification time.\n", file.Path)
				fmt.Printf("< - %s%s - Size: %d - ModTime: %v\n", root1, file.Path, file.F1.Size(), file.F1.ModTime())
				fmt.Printf("> - %s%s - Size: %d - ModTime: %v\n", root2, file.Path, file.F2.Size(), file.F2.ModTime())
				fmt.Printf("Copy? (</>/n)): ")
				if ok, side := utils.AskActionSides(false); ok {
					if side == "<" {
						copyMoves = append(copyMoves, []string{root1 + file.Path, root2 + file.Path})
					} else if side == ">" {
						copyMoves = append(copyMoves, []string{root2 + file.Path, root1 + file.Path})
					}
				}
			}
		}
	}
	return copyMoves, nil
}

