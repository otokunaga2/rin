package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"time"
)

func InitDb(db *sql.DB) {
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS user_utterances(user_id varchar(64),utterance varchar(10000), recorded_at timestamp)"); err != nil {
		log.Fatal(err)
		return
	}
}
func GetDBConnection() (*sql.DB, error) {
	dbUrl := os.Getenv("DATABASE_URL")
	db, err := sql.Open("postgres", dbUrl)
	return db, err
}
func InsertDB(user_id string, text string, current time.Time) error {
	db, err := GetDBConnection()
	defer db.Close()
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
		return err
	}
	_, err2 := db.Exec("INSERT INTO user_utterances(user_id, utterance, recorded_at) values($1,$2, $3)", user_id, text, current)
	if err2 != nil {
		log.Fatal(err2)
	}
	return nil

}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!\n")
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently system is running \n")
}

type BotMessageType int

const (
	ActiveListen BotMessageType = iota // ActiveListen == 0
)

var MESSAGE_LIST = [2]string{"そうだったんですね", "うん、うん"}

func GenerateMessage() string {
	rand.Seed(time.Now().UnixNano())
	selected := rand.Intn(2)
	fmt.Printf("Selected index %d \n", selected)
	reply := MESSAGE_LIST[selected]
	return reply
}
func SendMessageWithStrategy(c BotMessageType, userId string, bot *linebot.Client) {
	reply := GenerateMessage()
	switch c {
	case ActiveListen:
		_, err := bot.PushMessage(userId, linebot.NewTextMessage(reply)).Do()
		if err != nil {
			log.Fatalf("Fail to send message to %s", userId)
		}
	}
}
func lineHandler(w http.ResponseWriter, r *http.Request) {

	bot, err := linebot.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	events, err := bot.ParseRequest(r)
	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				id := message.ID
				log.Print("Logging id :", id)
				log.Printf("Received From USER ID: %s, Text: %s \n", event.Source.UserID, event.Message)
				err2 := InsertDB(event.Source.UserID, message.Text, time.Now())
				if err2 != nil {
					log.Fatalf("Fail when insertion data %s", err2)
				}
				SendMessageWithStrategy(ActiveListen, string(event.Source.UserID), bot)
			case *linebot.StickerMessage:
				replyMessage := fmt.Sprintf(
					"sticker id is %s, stickerResourceType is %s", message.StickerID, message.StickerResourceType)
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(replyMessage)).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func main() {
	result := GenerateMessage()
	fmt.Printf("Generated response %s \n", result)
	db, err := GetDBConnection()
	if err != nil {
		//log.Error("Fail to get db connection")
		log.Fatalf("Fail when insertion data %s", err)
	}
	InitDb(db)
	serverPort, _ := strconv.Atoi(os.Args[1])
	fmt.Printf("Starting server at Port %d", serverPort)
	http.HandleFunc("/", handler)
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/callback", lineHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", serverPort), nil)
}
