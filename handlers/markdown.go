package handlers

import (
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"

	renderer "github.com/naufalsuryasumirat/blog/renderer"
	t "github.com/naufalsuryasumirat/blog/templates"
	util "github.com/naufalsuryasumirat/blog/utils"
)

type FormMD struct {
	Document *multipart.FileHeader   `form:"document" binding:"required"`
	Images   []*multipart.FileHeader `form:"images" binding:"omitempty"`
    Category string                  `form:"category" binding:"omitempty"`
    Blurb    string                  `form:"blurb" binding:"required"`
    Password string                  `form:"password" binding:"required"`
}

// can change only blurb
type PatchFormMD struct {
	Document *multipart.FileHeader   `form:"document" binding:"omitempty"`
    Blurb    string                  `form:"blurb" binding:"omitempty"`
    Password string                  `form:"password" binding:"required"`
}

var knownDocumentExtensions = map[string]bool{
	"md": true,
}

var knownImageExtensions = map[string]bool{
	"png":  true,
	"jpg":  true,
	"jpeg": true,
}

// should be in database, but ok
var knownCategories = map[string]bool{
    "tech": true,
    "ent": true,
}

func PostUploadMarkdown(c *gin.Context) {
	var form FormMD
	if err := c.ShouldBind(&form); err != nil {
		log.Println(err.Error())
		c.String(http.StatusBadRequest, "Malformed request, make sure all forms are filled")
        c.Abort()
        return
	}

    if form.Password != os.Getenv("PASSWORD") {
        c.String(http.StatusForbidden, "You're not authorized to post articles here, ding-dong!")
        c.Abort()
        return
    }

	timestamp, timestampFormatted := util.GetTimestamp()

	document, _ := form.Document.Open()
	mdFile, _ := io.ReadAll(document)
	filenameFull := form.Document.Filename

    category := form.Category
    if !knownCategories[category] {
        category = "tech" // defaults to tech category
    }

	filename, ext, found := strings.Cut(filenameFull, ".")
	if !found || !knownDocumentExtensions[ext] {
		c.String(http.StatusBadRequest, "Unknown document extension")
        c.Abort()
		return
	}

	for _, image := range form.Images {
		_, ext, found := strings.Cut(image.Filename, ".")
		if !found || !knownImageExtensions[ext] {
			c.String(http.StatusBadRequest, "Unknown image extension")
            c.Abort()
			return
		}
	}

	hashOutput := util.HashUpload(timestampFormatted, filename)

	dirpath := filepath.Join(util.StorageDir, hashOutput)
	os.MkdirAll(dirpath, 0755)

	insertResult, err := util.GetDB().Exec(
        `INSERT INTO entries (dirpath, doc)
            VALUES (?, ?);`,
        hashOutput,
        timestamp)
	if err != nil {
		log.Println(err.Error())
        c.String(http.StatusInternalServerError, "Couldn't insert to entries")
        c.Abort()
        return
	}

	lastEntryID, err := insertResult.LastInsertId()
	if err != nil {
		log.Println(err.Error())
        c.String(http.StatusInternalServerError, "Couldn't get inserted entry")
        c.Abort()
        return
	}

    // insert article entry
    _ , err = util.GetDB().Exec(
        `INSERT INTO articles (entry_id, title, blurb, category)
            VALUES (?, ?, ?, ?)`,
        lastEntryID,
        filename,
        form.Blurb,
        category)
    if err != nil {
        log.Println(err.Error())
        c.String(http.StatusInternalServerError, "Couldn't insert to article")
        c.Abort()
        return
    }

	var insertImagesQuery string = `
        INSERT INTO images (entry_id, fname) VALUES
    `
	values := []interface{}{}
	for _, image := range form.Images {
		imagename := util.SanitizeFilename(image.Filename)
		savepath := filepath.Join(dirpath, imagename)
		if err := c.SaveUploadedFile(image, savepath); err != nil {
			log.Println(err.Error())
            // not informing failed to save image
		}
		insertImagesQuery += "(?, ?),"
		values = append(values, lastEntryID, imagename)
	}

	if len(form.Images) > 0 {
		statement, err := util.GetDB().Prepare(
			insertImagesQuery[0 : len(insertImagesQuery)-1],
		)
		if err != nil {
			log.Println(err.Error())
            // not informing failed to prepare query
		}

		if _, err = statement.Exec(values...); err != nil {
			log.Println(err.Error())
            // not informing failed to exec query
		}
	}

	c.SaveUploadedFile(form.Document, filepath.Join(dirpath, filenameFull))
	generatedHTML := renderer.MdToHTML(mdFile, &hashOutput, util.SanitizeFilename)
	os.WriteFile(
		filepath.Join(dirpath, filename+".html"),
		generatedHTML,
		0755,
	)

	c.Data(http.StatusOK, "text/html; charset=utf-8", generatedHTML)
}

func GetUploadMarkdown(c *gin.Context) {
    cmp := t.Add(false, "")
    c.Status(http.StatusOK)
    t.Layout(cmp, "post").Render(c.Request.Context(), c.Writer)
}

func PatchArticle(c *gin.Context) {
    art := c.Param("art")

    if art == "" || art == "/" {
        NotFound(c)
        c.Abort()
        return
    }

    _, found := GetLatestArticle(art)
    if !found {
        NotFound(c)
        c.Abort()
        return
    }

	var form PatchFormMD 
	if err := c.ShouldBind(&form); err != nil {
		log.Println(err.Error())
		c.String(http.StatusBadRequest, "Malformed request, make sure password is filled")
        c.Abort()
        return
	}

    if form.Password != os.Getenv("PASSWORD") {
        c.String(http.StatusForbidden, "You're not authorized to patch articles here, ding-dong!")
        c.Abort()
        return
    }

    if form.Document == nil && len(form.Blurb) == 0 {
        c.String(http.StatusBadRequest, "Malformed request, make sure either blurb of markdown is filled")
        c.Abort()
        return
    }

    if form.Document != nil {
        document, _ := form.Document.Open()
        mdFile, _ := io.ReadAll(document)
        filenameFull := form.Document.Filename

        // to work properly, image names must be the same
        _, ext, found := strings.Cut(filenameFull, ".")
        if !found || !knownDocumentExtensions[ext] {
            c.String(http.StatusBadRequest, "Unknown document extension")
            c.Abort()
            return
        }

        currentFiles := GetLatestFiles(art)
        idxMd := slices.IndexFunc(currentFiles, func(s string) bool {
            return filepath.Ext(s) == ".md"
        })
        if idxMd == -1 {
            c.String(http.StatusInternalServerError, "Couldn't find .md article to patch")
            c.Abort()
            return 
        }

        BackupFiles(art, currentFiles)
        // article title won't change
        dirpath := filepath.Join(util.StorageDir, art)
        oldMdFile := currentFiles[idxMd]
        filename := strings.TrimSuffix(oldMdFile, filepath.Ext(oldMdFile))
        c.SaveUploadedFile(form.Document, filepath.Join(dirpath, oldMdFile))
        generatedHTML := renderer.MdToHTML(mdFile, &art, util.SanitizeFilename)
        os.WriteFile(
            filepath.Join(dirpath, filename + ".html"),
            generatedHTML,
            0755,
        )
    }

    if form.Blurb != "" {
        _, err := util.GetDB().Exec(
            `UPDATE articles
                SET blurb = ?
                WHERE entry_id IN (
                    SELECT t1.entry_id FROM articles t1 JOIN entries t2 ON t1.entry_id = t2.id
                        WHERE t2.dirpath = ?
                );
            `,
            form.Blurb,
            art,
        )

        if err != nil {
            c.String(http.StatusInternalServerError, "Failed to patch blurb for article")
            c.Abort()
            return
        }
    }

    c.String(http.StatusOK, "article successfully patched")
}

func GetEditArticle(c *gin.Context) {
    art := c.Param("art")

    if art == "" || art == "/" {
        NotFound(c)
        c.Abort()
        return
    }

    _, found := GetLatestArticle(art)
    if !found {
        NotFound(c)
        c.Abort()
        return
    }

    cmp := t.Add(true, art)
    c.Status(http.StatusOK)
    t.Layout(cmp, "post").Render(c.Request.Context(), c.Writer)
}
