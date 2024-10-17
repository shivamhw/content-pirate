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

func GetCfgFromJson(filePath string, v interface{}) {
	file, _ := os.Open(filePath)
	defer file.Close()

	data, _ := io.ReadAll(file)
	if err := json.Unmarshal(data, v); err != nil {
		log.Fatal("json unmarshell failed during cfg read")
	}
}