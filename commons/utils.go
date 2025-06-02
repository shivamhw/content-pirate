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
func ReadFromJson(filePath string, v interface{}) error {
	if _, err := os.Stat(filePath); err != nil {
		return err
	}
	file, _ := os.Open(filePath)
	defer file.Close()
	data, err := io.ReadAll(file)
	if err != nil {
		log.Fatalf("%s", err)
		return err
	}
	if err = json.Unmarshal(data, v); err != nil {
		log.Fatalf("json unmarshell failed during %s read %s",filePath, err.Error())
	}
	return err
}

func IsImgLink(link string) bool {
	for _, suff := range IMG_SUFFIX {
		if strings.HasSuffix(link, suff) {
			return true
		}
	}
	return false
}
