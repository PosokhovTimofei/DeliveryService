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
	ctx context.Context,
	pkg *models.Package,
	auctionSvc *AuctionService,
	producer *kafka.AuctionPublisher,
	repo repository.Packager,
	log *logrus.Logger,
) {
	timer := time.NewTimer(5 * time.Minute)

	go func() {
		select {
		case <-ctx.Done():
			log.Info("Auction cancelled before timeout")
			return

		case <-timer.C:
			winner, err := auctionSvc.DetermineWinner(ctx, pkg.PackageID)
			if err != nil || winner == nil {
				log.Warnf("Auction finished with no winner for package %s", pkg.PackageID)

				pkg.Status = "Auction-failed"
				pkg.UpdatedAt = time.Now()
				if err := repo.Update(ctx, pkg); err != nil {
					log.WithError(err).Error("Failed to update package status to 'auction-failed'")
				}
				return
			}

			pkg.Status = "Finished"
			pkg.UserID = winner.UserID
			pkg.Cost = winner.Amount
			pkg.UpdatedAt = time.Now()

			if err := repo.Update(ctx, pkg); err != nil {
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

			if err := producer.PublishPayment(ctx, result); err != nil {
				log.WithError(err).Error("Failed to publish auction result")
			} else {
				log.WithField("package_id", pkg.PackageID).Info("Auction result published successfully")
			}

			//
			notification := &models.Notification{
				UserID:  winner.UserID,
				Message: fmt.Sprintf("Поздравляем! Вы выиграли аукцион на пакет %s за %.2f %s", pkg.PackageID, winner.Amount, pkg.Currency),
			}

			if err := producer.PublishNotification(ctx, notification); err != nil {
				log.WithError(err).Error("Failed to send notification")
			} else {
				log.WithField("user_id", winner.UserID).Info("Notification sent to auction winner")
			}
		}
	}()
}
