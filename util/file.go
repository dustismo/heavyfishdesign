package util

import (
	"os"
	"path/filepath"
	"strings"
)

// deletes all files of the given type from the requested directory
// this operates recursively, but does not remove the empty directories
func ClearDir(directory, extension string) error {
	files, err := FileList(directory, extension)
	if err != nil {
		return err
	}
	for _, fn := range files {
		err = os.Remove(fn)
		if err != nil {
			return err
		}
	}
	return nil
}

// recursively search for files of the given extension

// Filenames will have the passed in directory as a prefix:
// for instance:
// FileList("static", "hfd")
// static/hfd/shelf/four_sided_shelf_with_divider.hfd
// static/hfd/shelf/three_sided_shelf.hfd
// static/hfd/shelf/three_sided_shelf_with_divider.hfd
func FileList(directory string, extension string) ([]string, error) {
	filenames := []string{}
	err := filepath.Walk(directory,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(p, extension) {
				filenames = append(filenames, p)
			}
			return nil
		})
	return filenames, err
}

// Same as FileList, but strips the requested directory prefix.
// for instance:
// FileList("static", "hfd")
// hfd/shelf/four_sided_shelf_with_divider.hfd
// hfd/shelf/three_sided_shelf.hfd
// hfd/shelf/three_sided_shelf_with_divider.hfd
func RelFileList(directory string, extension string) ([]string, error) {
	filenames := []string{}
	err := filepath.Walk(directory,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && strings.HasSuffix(p, extension) {
				p = strings.TrimPrefix(p, directory)
				p = strings.TrimPrefix(p, string(os.PathSeparator))
				filenames = append(filenames, p)
			}
			return nil
		})
	return filenames, err
}

// recursively list the directories
func DirectoryList(directory string) ([]string, error) {
	directories := []string{}
	err := filepath.Walk(directory,
		func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() {
				directories = append(directories, p)
			}
			return nil
		})
	return directories, err
}
