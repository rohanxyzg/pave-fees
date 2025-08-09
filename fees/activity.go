package fees

import (
	"context"
	"fmt"
	"log/slog"
)

type Activities struct {
	repo RepositoryInterface
}

func NewActivities(repo RepositoryInterface) *Activities {
	return &Activities{repo: repo}
}

func (a *Activities) CalculateTotalActivity(_ context.Context, items []LineItem) (int64, error) {
	var total int64
	for _, item := range items {
		total += item.Amount
	}

	slog.Debug("calculated total for bill", "line_items_count", len(items), "total", total)
	return total, nil
}

type FinalBill struct {
	ID          string
	TotalAmount int64
	Status      BillStatus
}

func (a *Activities) SaveFinalBillActivity(ctx context.Context, bill FinalBill) error {
	err := a.repo.UpdateBillStatus(ctx, bill.ID, bill.Status, bill.TotalAmount)
	if err != nil {
		slog.Error("failed to save final bill", "bill_id", bill.ID, "error", err)
		return fmt.Errorf("failed to save final bill: %w", err)
	}

	slog.Info("bill finalized successfully", "bill_id", bill.ID, "total_amount", bill.TotalAmount)
	return nil
}
