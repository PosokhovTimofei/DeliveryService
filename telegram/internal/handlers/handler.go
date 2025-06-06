package handlers

import (
	"context"
	"log"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maksroxx/DeliveryService/telegram/internal/service"
)

type Handler struct {
	serviceAuth    *service.AuthService
	servicePackage *service.PackageService
}

func NewHandler(serviceAuth *service.AuthService, servicePackage *service.PackageService) *Handler {
	return &Handler{
		serviceAuth:    serviceAuth,
		servicePackage: servicePackage,
	}
}

func (h *Handler) HandleUpdate(update tgbotapi.Update, bot *tgbotapi.BotAPI) {
	if update.Message == nil {
		return
	}

	chatID := update.Message.Chat.ID
	text := update.Message.Text

	switch {
	case text == "/start":
		h.sendStart(bot, chatID)

	case text == "/packages":
		msgText, err := h.servicePackage.GetUserPackages(context.Background(), update.Message.From.ID)
		if err != nil {
			log.Printf("❌ Ошибка получения посылок: %v", err)
			h.sendMarkdown(bot, chatID, buildErrorMessage("⚠️ Не удалось получить посылки. Возможно, вы не привязали аккаунт."))
		} else {
			h.sendMarkdown(bot, chatID, escapeMarkdown(msgText))
		}

	case strings.HasPrefix(text, "auth_"):
		err := h.serviceAuth.LinkTelegramAccount(context.Background(), text, update.Message.From.ID)
		if err != nil {
			log.Printf("❌ Ошибка привязки аккаунта: %v", err)
			h.sendMarkdown(bot, chatID, buildErrorMessage("❗ Не удалось привязать аккаунт. Возможно, код неверный или истёк."))
		} else {
			h.sendSuccessAuth(bot, chatID)
		}

	default:
		h.sendMarkdown(bot, chatID, buildErrorMessage("🚫 Неизвестная команда. Используйте `/start` или отправьте одноразовый код."))
	}
}

func (h *Handler) sendStart(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, buildStartMessage())
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = defaultKeyboard()
	bot.Send(msg)
}

func (h *Handler) sendSuccessAuth(bot *tgbotapi.BotAPI, chatID int64) {
	msg := tgbotapi.NewMessage(chatID, buildSuccessMessage("Аккаунт успешно привязан! Теперь вы можете использовать команды бота."))
	msg.ParseMode = "MarkdownV2"
	msg.ReplyMarkup = defaultKeyboard()
	bot.Send(msg)
}

func (h *Handler) sendMarkdown(bot *tgbotapi.BotAPI, chatID int64, msg string) {
	message := tgbotapi.NewMessage(chatID, msg)
	message.ParseMode = "MarkdownV2"
	_, err := bot.Send(message)
	if err != nil {
		log.Printf("❗ Ошибка отправки сообщения: %v", err)
	}
}
