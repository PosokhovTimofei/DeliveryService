package service

import (
	"context"
	"fmt"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/kafka"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/sirupsen/logrus"
)

func StartAuction(
	pkg *models.Package,
	auctionSvc *AuctionService,
	producer *kafka.AuctionPublisher,
	repo repository.Packager,
	log *logrus.Logger,
) {
	go func() {
		timer := time.NewTimer(2 * time.Minute)
		defer timer.Stop()

		pkg.Status = "Auctioning"
		if err := repo.Update(context.Background(), pkg); err != nil {
			log.WithError(err).Error("Failed to update package when auction start")
			return
		}

		<-timer.C

		winner, err := auctionSvc.DetermineWinner(context.Background(), pkg.PackageID)
		if err != nil || winner == nil {
			log.Warnf("Auction finished with no winner for package %s", pkg.PackageID)

			pkg.Status = "Auction-failed"
			pkg.UpdatedAt = time.Now()
			if err := repo.Update(context.Background(), pkg); err != nil {
				log.WithError(err).Error("Failed to update package status to 'waiting'")
			}
			return
		}

		pkg.Status = "Finished"
		pkg.UserID = winner.UserID
		pkg.Cost = winner.Amount
		pkg.UpdatedAt = time.Now()

		if err := repo.Update(context.Background(), pkg); err != nil {
			log.WithError(err).Error("Failed to update package after auction win")
			return
		}

		result := &models.AuctionResult{
			PackageID:  pkg.PackageID,
			WinnerID:   winner.UserID,
			FinalPrice: winner.Amount,
			Currency:   pkg.Currency,
			FinishedAt: time.Now(),
		}

		if err := producer.PublishPayment(context.Background(), result); err != nil {
			log.WithError(err).Error("Failed to publish auction result")
		} else {
			log.WithField("package_id", pkg.PackageID).Info("Auction result published successfully")
		}

		notification := &models.Notification{
			UserID:  winner.UserID,
			Message: fmt.Sprintf("Поздравляем! Вы выиграли аукцион на пакет %s за %.2f %s", pkg.PackageID, winner.Amount, pkg.Currency),
		}

		if err := producer.PublishNotification(context.Background(), notification); err != nil {
			log.WithError(err).Error("Failed to send notification")
		} else {
			log.WithField("user_id", winner.UserID).Info("Notification sent to auction winner")
		}
	}()
}
