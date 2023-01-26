/*
   Телеграмм бот таймер
*/

package main

import (
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func echo(w http.ResponseWriter, r *http.Request) {
	io.WriteString(w, "i`m work")
}

func main() {

	bot, err := tgbotapi.NewBotAPI(os.Getenv("TELEGRAM_APITOKEN"))
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true

	http.HandleFunc("/", echo)
	port := os.Getenv("PORT")
	go http.ListenAndServe(":"+port, nil)

	go func() {
		for {
			_, err = http.Get(fmt.Sprintf("https://%s/", os.Getenv("LINK_HEROKU")))
			if err != nil {
				log.Println("Нет соединения с сервером")
			}
			// Спим так как хероку рубит работу бота если запросы не приходят в течении 30 минут
			time.Sleep(time.Minute * 20)
		}
	}()

	// Start polling Telegram for updates.
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {

		mx := new(sync.Mutex)

		if update.Message == nil {
			continue
		}

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, "")

		if !update.Message.IsCommand() { // ignore any non-command Messages
			input, err := time.ParseDuration(update.Message.Text)
			if err != nil {
				msg.Text = "Введите время в формате 1h2m3s"
				if _, err := bot.Send(msg); err != nil {
					log.Println(err)
				}
				continue
			}

			go func() {
				var timer *time.Timer

				mx.Lock()
				timer = time.NewTimer(input)
				<-timer.C
				msg.Text = "Время вышло"
				if _, err := bot.Send(msg); err != nil {
					log.Println(err)
				}
				mx.Unlock()
			}()

			continue

		}

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "new":
			msg.Text = "Введите время в формате 1h2m3s"
			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}
		default:
			msg.Text = "Для создания таймера введите команду /new"
			if _, err := bot.Send(msg); err != nil {
				log.Println(err)
			}
		}
	}

}
