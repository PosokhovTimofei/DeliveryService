package bot

import (
	"log"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maksroxx/DeliveryService/telegram/internal/handlers"
)

type Bot struct {
	API *tgbotapi.BotAPI
}

func NewBot(token string) *Bot {
	api, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalf("Failed to create bot: %v", err)
	}

	api.Debug = false
	log.Printf("Authorized on account %s", api.Self.UserName)

	return &Bot{API: api}
}

func (b *Bot) Run(handler *handlers.Handler) {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.API.GetUpdatesChan(u)

	log.Println("Telegram сервис запущен...")
	for update := range updates {
		if update.Message != nil {
			go handler.HandleUpdate(update, b.API)
		}
	}
}
