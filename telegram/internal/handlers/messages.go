package handlers

import (
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func buildStartMessage() string {
	return escapeMarkdown("👋 *Привет!* Отправь мне одноразовый код для привязки аккаунта, начинающийся с `auth_XXXX`.\n\nИли воспользуйся кнопкой ниже ⬇️")
}

func buildSuccessMessage(text string) string {
	return escapeMarkdown("✅ " + text)
}

func buildErrorMessage(text string) string {
	return escapeMarkdown("❌ " + text)
}

func escapeMarkdown(msg string) string {
	replacer := strings.NewReplacer(
		"_", "\\_",
		"*", "\\*",
		"[", "\\[",
		"]", "\\]",
		"(", "\\(",
		")", "\\)",
		"~", "\\~",
		"`", "\\`",
		">", "\\>",
		"#", "\\#",
		"+", "\\+",
		"-", "\\-",
		"=", "\\=",
		"|", "\\|",
		"{", "\\{",
		"}", "\\}",
		".", "\\.",
		"!", "\\!",
	)
	return replacer.Replace(msg)
}

func defaultKeyboard() tgbotapi.ReplyKeyboardMarkup {
	return tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("/packages"),
		),
	)
}
