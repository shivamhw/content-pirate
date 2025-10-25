package notifier

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/shivamhw/content-pirate/commons"
	"github.com/shivamhw/content-pirate/pkg/log"
)

type payload struct {
	File    string 
	ChatId  int64 
	ID      int64 
	Size    int64 
	Tokens  string 
}			

type BleveNotifier struct {
	Url string
	IdxName string
	client *http.Client
	fuse     bool
}

func NewBleveNotifier(url, idxName string) *BleveNotifier {
	log.Infof("using bleve notifier", "url", url, "idxName", idxName)
	return &BleveNotifier{
		Url:      url,
		IdxName:  idxName,
		client:   &http.Client{},
	}
}

func getTokens(fileName string) string {
	// This function should extract tokens from the fileName or item.
	// For simplicity, we return an empty string here.
	// Implement the actual logic as needed.
	fileName = strings.ReplaceAll(fileName, "_", " ")
	fileName = strings.ReplaceAll(fileName, "-", " ")
	fileName = strings.ReplaceAll(fileName, ".", " ")
	fileName = strings.ReplaceAll(fileName, ",", " ")
	fileName = strings.ReplaceAll(fileName, "@", "")
	fileName = strings.ReplaceAll(fileName, "#", "")
	fileName = strings.ReplaceAll(fileName, "(", " ")
	fileName = strings.ReplaceAll(fileName, ")", " ")
	fileName = strings.ReplaceAll(fileName, "[", " ")
	fileName = strings.ReplaceAll(fileName, "]", " ")
	
	return fileName
}

func (b *BleveNotifier) Notify(ctx context.Context, item *commons.Item, event NotifierEvent) error {
	// Construct the request to notify the Bleve index
	if b.fuse {
		return nil
	}
	url := fmt.Sprintf("%s/%s/%s", b.Url, b.IdxName, item.FileName)
	tokens := getTokens(item.FileName)
	id, _ := strconv.Atoi(item.DstId)
	chatId, _ := strconv.Atoi(item.Dst)
	payload := payload{
		File: item.FileName,
		ChatId: int64(chatId),
		ID: int64(id),
		Size: item.Size,
		Tokens: tokens,
	}
	payloadBt, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("failed to marshal payload: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(payloadBt))
	if err != nil {
		return err
	}

	// Set necessary headers and body if required
	req.Header.Set("Content-Type", "application/json")
	// Here you would typically marshal the item into JSON and set it as the body

	// Send the request
	resp, err := b.client.Do(req)
	if err != nil {
		b.fuse = true
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to notify Bleve: %s", resp.Status)
	}

	return nil
}