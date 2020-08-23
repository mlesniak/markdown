package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
)

// TODO Error handling and configurability
func sendMessage(text string) {
	botID := os.Getenv("BOTID")
	chatID, err := strconv.Atoi(os.Getenv("CHATID"))
	if err != nil {
		panic(err)
	}

	body := struct {
		ChatID int    `json:"chat_id"`
		Text   string `json:"text"`
	}{
		ChatID: chatID,
		Text:   text,
	}
	bs, _ := json.Marshal(&body)
	buf := bytes.NewReader(bs)
	// println(string(bs))

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", botID)
	http.Post(url, "application/json", buf)
	// if err != nil {
	// 	panic(err)
	// }
	// all, err := ioutil.ReadAll(post.Body)
	// if err != nil {
	// 	panic(err)
	// }
	// println(string(all))
}
