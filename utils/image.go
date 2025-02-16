package utils

import (
	"image"
	"image/jpeg"
	_ "image/png"
	"log"
	"os"
	"path/filepath"

	"github.com/nfnt/resize"
)

const targetHeight = 100
var optJpeg = jpeg.Options{ Quality: 60 }

func GenerateThumbnail(fpath string) {
    if reader, err := os.Open(fpath); err != nil {
    } else {
        defer reader.Close()
        img, _, err := image.Decode(reader)
        if err != nil {
            log.Println(err)
            return
        }

        resized := resize.Resize(0, targetHeight, img, resize.Bilinear)

        f, err := os.Create(fpath + ".thumbnail")
        if err != nil {
            log.Println(err)
            return
        }

        err = jpeg.Encode(f, resized, &optJpeg)
        if err != nil {
            log.Println(err)
        }
    }
}

func GenerateThumbnailDirectory(dirpath string) {
    entries, err := os.ReadDir(dirpath)
    if err != nil {
        log.Println(err)
    }

    for _, entry := range entries {
        ext := filepath.Ext(entry.Name())
        if ext != ".jpeg" && ext != ".jpg" && ext != ".png" {
            continue
        }

        GenerateThumbnail(filepath.Join(dirpath, entry.Name()))
        return
    }
}
