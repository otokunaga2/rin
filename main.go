package main

import (
	"database/sql"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"
	"strconv"
)

func InitDb(db *sql.DB) {
	if _, err := db.Exec("CREATE TABLE IF NOT EXISTS user_utterances(utterance varchar(10000), recorded_at timestamp)"); err != nil {
		log.Fatal(err)
		return
	}
}
func handler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello world!\n")
}

func healthzHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Currently system is running \n")
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
				log.Print("Received From USER ID: ", event.Source.UserID)
				replyMessage := linebot.NewTextMessage(message.Text)
				if _, err = bot.ReplyMessage(event.ReplyToken, replyMessage).Do(); err != nil {
					log.Print(err)
				}
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
	dbUrl := os.Getenv("DATABASE_URL")
	fmt.Printf("DATABSE URL is %s", dbUrl)
	db, err := sql.Open("postgres", dbUrl)
	if err != nil {
		log.Fatalf("Error opening database: %q", err)
	}
	InitDb(db)
	port, _ := strconv.Atoi(os.Args[1])
	fmt.Printf("Starting server at Port %d", port)
	http.HandleFunc("/", handler)
	http.HandleFunc("/healthz", healthzHandler)
	http.HandleFunc("/callback", lineHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}
