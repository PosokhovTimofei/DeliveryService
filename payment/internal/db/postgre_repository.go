package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgconn"
	"github.com/jackc/pgx/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/maksroxx/DeliveryService/payment/internal/models"
)

type PostgresPaymenter struct {
	db *pgxpool.Pool
}

func NewPostgresPaymenter(db *pgxpool.Pool) *PostgresPaymenter {
	return &PostgresPaymenter{db: db}
}

func (p *PostgresPaymenter) CreatePayment(ctx context.Context, payment models.Payment) error {
	if p.db == nil {
		return fmt.Errorf("PostgreSQL connection is nil")
	}

	query := `INSERT INTO payments (user_id, package_id, cost, currency, status) 
			  VALUES ($1, $2, $3, $4, $5)`

	_, err := p.db.Exec(ctx, query, payment.UserID, payment.PackageID, payment.Cost, payment.Currency, payment.Status)
	if err != nil {
		if pgErr, ok := err.(*pgconn.PgError); ok {
			if pgErr.Code == "23505" {
				return fmt.Errorf("payment already exists")
			}
		}
		return fmt.Errorf("failed to create payment: %w", err)
	}

	return nil
}

func (p *PostgresPaymenter) UpdatePayment(ctx context.Context, update models.Payment) (*models.Payment, error) {
	query := `UPDATE payments 
			  SET status = $1 
			  WHERE user_id = $2 AND package_id = $3 AND status != 'PAID'
			  RETURNING user_id, package_id, cost, currency, status`

	row := p.db.QueryRow(ctx, query, update.Status, update.UserID, update.PackageID)

	var payment models.Payment
	err := row.Scan(&payment.UserID, &payment.PackageID, &payment.Cost, &payment.Currency, &payment.Status)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("payment already confirmed or not found")
		}
		return nil, fmt.Errorf("failed to update payment: %w", err)
	}

	return &payment, nil
}
