package handlers

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/a-h/templ"
	u "github.com/naufalsuryasumirat/blog/utils"
)

const backupDirPrefix = "rev"

// returns article html, boolean found
func GetLatestArticle(path string) (templ.Component, bool) {
    none := func() (templ.Component, bool) {
        return nil, false
    }

	entry, found := u.GetLatestEntry(path)
	if !found {
        return none()
	}

	dirpath := filepath.Join(u.StorageDir, entry.Dirpath)
	d, err := os.Stat(dirpath)
	if err != nil {
		if err != fs.ErrNotExist {
			log.Println(err)
		}
        return none()
	}

	if !d.IsDir() {
		log.Println("hashed name not a directory")
        return none()
	}

	files, err := os.ReadDir(dirpath)
	chk(err)

	foundHtml := false
	var fHtml string
	foundMd := false
	var fMd string
	for _, file := range files {
		if file.IsDir() {
			continue
		}

		if foundHtml && foundMd {
			break
		}

		ext := filepath.Ext(file.Name())
		if ext == ".html" {
			foundHtml = true
			fHtml = file.Name()
		} else if ext == ".md" {
			foundMd = true
			fMd = file.Name()
		}
	}

    if !foundHtml && !foundMd {
        return none()
    }

    if foundHtml {
        fname := filepath.Join(dirpath, fHtml)
        buf, err := os.ReadFile(fname)
        if err != nil {
            log.Println(err)
            return none()
        }

        return templ.Raw(string(buf)), true
    }

    // found md only, TODO: generate html here
    log.Println(fMd)
    return none()
}

func GetArticleAudio(path string) (string, bool) {
    none := func() (string, bool) {
        return "", false
    }

	dirpath := filepath.Join(u.StorageDir, path)
	d, err := os.Stat(dirpath)
	if err != nil {
		if err != fs.ErrNotExist {
			log.Println(err)
		}
        return none()
	}

	if !d.IsDir() {
		log.Println("hashed name not a directory")
        return none()
	}

	files, err := os.ReadDir(dirpath)
	chk(err)

    found := false
    var res string
    for _, file := range files {
        ext := filepath.Ext(file.Name())
        if ext == ".mp3" {
            res = file.Name()
            found = true
            break
        }
    }

    return res, found
}

// Gets latest file in the dirpath (root, prevs are in rev incremented subdir)
// Assumes path exist, no os stat checking
func GetLatestFiles(path string) []string {
    var res []string

    dir := filepath.Join(u.StorageDir, path)
    files, err := os.ReadDir(dir)
    chk(err)

    for _, file := range files {
        if file.IsDir() {
            continue
        }

        ext := filepath.Ext(file.Name())
        if ext == ".html" || ext == ".md" {
            res = append(res, file.Name())
        }
    }

    return res
}

// Backup files into rev{%d,inc} subdirectory in the given path.
// Assumes path exist, no os stat checking.
func BackupFiles(path string, files []string) {
    dir := filepath.Join(u.StorageDir, path)

    // increments rev dir
    entries, err := os.ReadDir(dir)
    chk(err)

    if len(entries) <= 0 {
        return
    }

    curRev := 0
    for i := len(entries)-1; i >= 0; i-- {
        entry := entries[i]
        if !entry.IsDir() {
            continue
        }

        rev := strings.TrimPrefix(entry.Name(), backupDirPrefix)
        if rev == entry.Name() {
            continue
        }

        revInt, err := strconv.Atoi(rev)
        chk(err)

        curRev = revInt
        break
    }

    curRev += 1
    // create new bakdir
    bakDir := filepath.Join(dir, fmt.Sprintf("%s%d", backupDirPrefix, curRev))
    os.Mkdir(bakDir, 0755)

    // move old files
    for _, f := range files {
        oldPath := filepath.Join(dir, f)
        newPath := filepath.Join(bakDir, f)
        if err := os.Rename(oldPath, newPath); err != nil {
            log.Println(err)
            continue
        }
    }
}
