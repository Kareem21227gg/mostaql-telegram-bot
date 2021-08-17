package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
)

//go server side webhock with telegram bot to check if there is any new messages in https://mostaql.com
type webhookReqBody struct {
	Message struct {
		Text string `json:"text"`
		Chat struct {
			ID int64 `json:"id"`
		} `json:"chat"`
	} `json:"message"`
}

func Handler(res http.ResponseWriter, req *http.Request) {

	body := &webhookReqBody{}
	if err := json.NewDecoder(req.Body).Decode(body); err != nil {
		fmt.Println("could not decode request body", err)
		return
	}

	fmt.Println("bot used by userID: ", body.Message.Chat.ID)
	messageState := "YES!"
	messagesNumber := getMessagesNumber()
	if messagesNumber == 0 {
		messageState = "no :)"
	} else if messagesNumber == -1 {
		messageState = "error :("
	}
	if err := sayPolo(body.Message.Chat.ID, messageState); err != nil {
		fmt.Println("error in sending reply:", err)
		return
	}

	fmt.Println("reply sent successfully to userID: ", body.Message.Chat.ID)
}
func getMessagesNumber() int {
	request, err := http.NewRequest("GET", "https://mostaql.com", http.NoBody)

	if err != nil {
		fmt.Println("error with create request:", err)
		return -1
	}
	// defer request.Body.Close()
	request.AddCookie(&http.Cookie{Domain: "mostaql.com", Name: "mostaqlweb", Value: "<YOUR_MOSTAQL_TOKEN>"})
	request.Header.Add("Host", "mostaql.com")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		fmt.Println("error with get respones:", err)
		return -1
	}
	defer response.Body.Close()
	mostaqlWebPage, err := ioutil.ReadAll(response.Body)
	if err != nil {
		fmt.Println("error with read reasponse body:", err)
		return -1
	}
	fileString := string(mostaqlWebPage)
	index := strings.Index(fileString, "<span class=\"text-alpha\">")
	if index == -1 {
		fmt.Println("wrong web page")
		return -1
	}
	fileString = fileString[index+25:]
	index = strings.Index(fileString, "<span class=\"text-alpha\">")
	if index == -1 {
		fmt.Println("wrong web page")
		return -1
	}
	n, err := strconv.Atoi(fileString[index+25 : index+26])
	if err != nil {
		fmt.Println("error with convert int to string:", err)
		return -1
	}
	return n
}

type sendMessageReqBody struct {
	ChatID int64                 `json:"chat_id"`
	Text   string                `json:"text"`
	Replay map[string][][]string `json:"reply_markup"`
}

func sayPolo(chatID int64, text string) error {

	mymap := make(map[string][][]string)
	mymap["keyboard"] = [][]string{{"messages?"}}
	reqBody := &sendMessageReqBody{
		ChatID: chatID,
		Text:   text,
		Replay: mymap,
	}

	reqBytes, err := json.Marshal(reqBody)
	if err != nil {
		return err
	}

	res, err := http.Post("https://api.telegram.org/bot<YOUR_BOT_TOKEN>/sendMessage", "application/json", bytes.NewBuffer(reqBytes))

	if err != nil {
		return err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		return errors.New("unexpected status" + res.Status)
	}
	return nil
}
func getPort() string {
	port := os.Getenv("PORT")
	if port == "" {
		port = "3030"
	}
	return ":" + port
}

func main() {
	http.ListenAndServe(getPort(), http.HandlerFunc(Handler))
}
