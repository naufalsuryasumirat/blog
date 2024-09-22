package utils

import (
    "crypto/sha1"
    "fmt"
    "strings"
)

var replacer *strings.Replacer = strings.NewReplacer(
    " ", "_",   ",", "",
    "/", "-",   "\\", "",
    "\"", "",   "'", "",
    "*", "",    "?", "",   
    "<", "",    ">", "",
    "(", "",    ")", "",
    "[", "",    "]", "",
    "{", "",    "}", "",
    ":", "",    ";", "",
    "|", "",    "~", "",
    "%", "",
)

func SanitizeFilename(filename string) string {
    return replacer.Replace(strings.ToLower(filename))
}

var hasher = sha1.New()

func HashUpload(datetime string, filename string) string {
    hasher.Write([]byte(datetime + "_" + filename))
    output := hasher.Sum(nil)
    hasher.Reset()
    
    return fmt.Sprintf("%x", output)
}

