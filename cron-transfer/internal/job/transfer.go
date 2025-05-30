package job

import (
	"context"
	"time"

	pb "github.com/maksroxx/DeliveryService/proto/database"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type TransferJob struct {
	client pb.PackageServiceClient
	log    *logrus.Entry
}

func NewTransferJob(conn *grpc.ClientConn, logger *logrus.Logger) *TransferJob {
	return &TransferJob{
		client: pb.NewPackageServiceClient(conn),
		log:    logger.WithField("module", "transfer-job"),
	}
}

func (t *TransferJob) Run() {
	t.log.Info("Starting expired package transfer")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err := t.client.TransferExpiredPackages(ctx, &pb.Empty{})
	if err != nil {
		t.log.WithError(err).Error("Transfer failed")
		return
	}

	t.log.Info("Transfer completed successfully")
}
