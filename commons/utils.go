package commons

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"strings"
)

func GetMIME(media string) string {
	mime := strings.Split(media, "/")
	if len(mime) == 1 {
		return "jpg"
	}
	return mime[1]
}

func GetExtFromLink(link string) string {
	return strings.Split(link, ".")[len(strings.Split(link, "."))-1]
}
func ReadFromJson(filePath string, v interface{}) {
	file, _ := os.Open(filePath)
	defer file.Close()
	data, _ := io.ReadAll(file)
	if err := json.Unmarshal(data, v); err != nil {
		log.Fatal("json unmarshell failed during cfg read")
	}
}

func IsImgLink(link string) bool {
	for _, suff := range IMG_SUFFIX {
		if strings.HasSuffix(link, suff) {
			return true
		}
	}
	return false
}
