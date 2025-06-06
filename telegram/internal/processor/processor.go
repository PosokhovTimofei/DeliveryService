package processor

import (
	"context"
	"encoding/json"

	"github.com/IBM/sarama"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/maksroxx/DeliveryService/telegram/internal/models"
	"github.com/maksroxx/DeliveryService/telegram/internal/repository"
	"github.com/sirupsen/logrus"
)

type NotificationProcessor struct {
	log  *logrus.Logger
	repo repository.Linker
	bot  *tgbotapi.BotAPI
}

func NewNotificationProcessor(log *logrus.Logger, repo repository.Linker, bot *tgbotapi.BotAPI) *NotificationProcessor {
	return &NotificationProcessor{
		log:  log,
		repo: repo,
		bot:  bot,
	}
}

func (p *NotificationProcessor) Setup(sarama.ConsumerGroupSession) error {
	p.log.Info("Notification processor setup")
	return nil
}

func (p *NotificationProcessor) Cleanup(sarama.ConsumerGroupSession) error {
	p.log.Info("Notification processor cleanup")
	return nil
}

func (p *NotificationProcessor) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		var notif models.Notification
		if err := json.Unmarshal(msg.Value, &notif); err != nil {
			p.log.WithError(err).Error("kafka error")
			continue
		}

		telegramID, err := p.repo.GetTelegramIDByUserID(context.Background(), notif.UserID)
		if err != nil {
			p.log.WithField("user_id", notif.UserID).Warn("telegram id not found")
			continue
		}

		message := tgbotapi.NewMessage(telegramID, notif.Message)
		if _, err := p.bot.Send(message); err != nil {
			p.log.WithError(err).WithField("telegram_id", telegramID).Error("telegramm delivery error")
		} else {
			p.log.WithField("telegram_id", telegramID).Info("success")
			session.MarkMessage(msg, "")
		}
	}
	return nil
}
