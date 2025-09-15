package service

import (
	"context"
	"fmt"
	"time"

	"github.com/maksroxx/DeliveryService/auction/internal/kafka"
	"github.com/maksroxx/DeliveryService/auction/internal/metrics"
	"github.com/maksroxx/DeliveryService/auction/internal/models"
	"github.com/maksroxx/DeliveryService/auction/internal/repository"
	"github.com/sirupsen/logrus"
)

func StartAuction(
	pkg *models.Package,
	auctionSvc *AuctionService,
	producer kafka.AucPublisher,
	repo repository.Packager,
	log *logrus.Logger,
	auctionDuration time.Duration,
) {
	go func() {
		metrics.AuctionStartedTotal.Inc()

		timer := time.NewTimer(auctionDuration)
		defer timer.Stop()

		pkg.Status = "Auctioning"
		if err := repo.Update(context.Background(), pkg); err != nil {
			log.WithError(err).Error("Failed to update package when auction start")
			return
		}

		<-timer.C
		metrics.AuctionFinishedTotal.Inc()

		winner, err := auctionSvc.DetermineWinner(context.Background(), pkg.PackageID)
		if err != nil || winner == nil {
			metrics.AuctionFinishedWithoutWinner.Inc()

			log.Warnf("Auction finished with no winner for package %s", pkg.PackageID)
			pkg.Status = "Auction-failed"
			pkg.UpdatedAt = time.Now()
			_ = repo.Update(context.Background(), pkg)
			return
		}

		metrics.AuctionFinishedWithWinner.Inc()

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
		_ = producer.PublishPayment(context.Background(), result)

		notification := &models.Notification{
			UserID:  winner.UserID,
			Message: fmt.Sprintf("Поздравляем! Вы выиграли аукцион на пакет %s за %.2f %s", pkg.PackageID, winner.Amount, pkg.Currency),
		}
		_ = producer.PublishNotification(context.Background(), notification)
	}()
}
