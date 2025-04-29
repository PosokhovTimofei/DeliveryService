package db_test

import (
	"context"
	"testing"

	"github.com/maksroxx/DeliveryService/payment/internal/db"
	"github.com/maksroxx/DeliveryService/payment/internal/models"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func TestCreatePayment_Success(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("successful insert", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())
		repo := db.NewPaymentMongoRepository(mt.DB, "payments")
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		err := repo.CreatePayment(context.Background(), models.Payment{
			UserID:    "user123",
			PackageID: "pkg456",
			Cost:      100.0,
			Currency:  "USD",
		})

		assert.NoError(t, err)
	})
}

func TestCreatePayment_Duplicate(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("duplicate key", func(mt *mtest.T) {
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
			mtest.CreateWriteErrorsResponse(mtest.WriteError{
				Code:    11000,
				Message: "duplicate key error",
			}),
		)

		repo := db.NewPaymentMongoRepository(mt.DB, "payments")

		err := repo.CreatePayment(context.Background(), models.Payment{
			UserID:    "user123",
			PackageID: "pkg456",
			Cost:      100,
			Currency:  "USD",
		})

		assert.EqualError(t, err, "payment has already exists")
	})

	mt.Run("other insert error", func(mt *mtest.T) {
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
			mtest.CreateWriteErrorsResponse(mtest.WriteError{
				Code:    999,
				Message: "some other error",
			}),
		)

		repo := db.NewPaymentMongoRepository(mt.DB, "payments")

		err := repo.CreatePayment(context.Background(), models.Payment{
			UserID:    "user123",
			PackageID: "pkg456",
			Cost:      100,
			Currency:  "USD",
		})

		assert.Error(t, err)
		assert.NotEqual(t, "payment has already exists", err.Error())
	})
}

func TestUpdatePayment(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))

	mt.Run("unexpected error", func(mt *mtest.T) {
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    101,
				Message: "random mongo error",
			}),
		)

		repo := db.NewPaymentMongoRepository(mt.DB, "payments")

		_, err := repo.UpdatePayment(context.Background(), models.Payment{
			UserID:    "user123",
			PackageID: "pkg456",
			Status:    models.PaymentStatusPaid,
		})

		assert.Error(t, err)
		assert.NotEqual(t, "payment already confirmed", err.Error())
	})

	mt.Run("payment already confirmed", func(mt *mtest.T) {
		mt.AddMockResponses(
			mtest.CreateSuccessResponse(),
			mtest.CreateCommandErrorResponse(mtest.CommandError{
				Code:    101,
				Message: "payment already confirmed",
			}),
		)

		repo := db.NewPaymentMongoRepository(mt.DB, "payments")

		_, err := repo.UpdatePayment(context.Background(), models.Payment{
			UserID:    "user123",
			PackageID: "pkg456",
			Status:    models.PaymentStatusPaid,
		})

		assert.Error(t, err)
		assert.Equal(t, "payment already confirmed", err.Error())
	})
}
