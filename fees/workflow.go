package fees

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
)

const (
	AddLineItemSignal = "ADD_LINE_ITEM"
	CloseBillSignal   = "CLOSE_BILL"
)

func BillWorkflow(ctx workflow.Context, initialBill Bill) error {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting bill workflow", "bill_id", initialBill.ID)

	isClosed := false
	var lineItems []LineItem

	addLineItemChan := workflow.GetSignalChannel(ctx, AddLineItemSignal)
	closeBillChan := workflow.GetSignalChannel(ctx, CloseBillSignal)

	for !isClosed {
		selector := workflow.NewSelector(ctx)

		selector.AddReceive(addLineItemChan, func(c workflow.ReceiveChannel, more bool) {
			var item LineItem
			c.Receive(ctx, &item)
			logger.Info("Received line item", "description", item.Description, "amount", item.Amount)
			lineItems = append(lineItems, item)
		})

		selector.AddReceive(closeBillChan, func(c workflow.ReceiveChannel, more bool) {
			c.Receive(ctx, nil)
			logger.Info("Received close bill signal", "total_line_items", len(lineItems))
			isClosed = true
		})

		selector.Select(ctx)
	}

	ao := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy: &temporal.RetryPolicy{
			MaximumAttempts:        3,
			BackoffCoefficient:     2.0,
			InitialInterval:        time.Second,
			MaximumInterval:        30 * time.Second,
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var total int64
	err := workflow.ExecuteActivity(ctx, "CalculateTotalActivity", lineItems).Get(ctx, &total)
	if err != nil {
		logger.Error("Failed to calculate total", "error", err)
		return fmt.Errorf("failed to calculate total: %w", err)
	}

	finalBill := FinalBill{
		ID:          initialBill.ID,
		TotalAmount: total,
		Status:      BillStatusClosed,
	}

	err = workflow.ExecuteActivity(ctx, "SaveFinalBillActivity", finalBill).Get(ctx, nil)
	if err != nil {
		logger.Error("Failed to save final bill", "error", err)
		return fmt.Errorf("failed to save final bill: %w", err)
	}

	logger.Info("Bill workflow completed successfully", 
		"bill_id", initialBill.ID, 
		"total_amount", total, 
		"line_items_count", len(lineItems))

	return nil
}
