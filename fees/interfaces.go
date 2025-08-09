package fees

import (
	"context"

	"go.temporal.io/sdk/client"
)

//go:generate mockgen -source=interfaces.go -destination=mock_interfaces_test.go -package=fees

type RepositoryInterface interface {
	CreateBill(ctx context.Context, bill *Bill) error
	GetBillByID(ctx context.Context, billID string) (*Bill, error)
	GetBillStatus(ctx context.Context, billID string) (BillStatus, error)
	AddLineItem(ctx context.Context, billID string, item *LineItem) error
	GetLineItemsByBillID(ctx context.Context, billID string) ([]LineItem, error)
	UpdateBillStatus(ctx context.Context, billID string, status BillStatus, totalAmount int64) error
	ListBillsByCustomer(ctx context.Context, customerID string, status *BillStatus, limit, offset int) ([]*Bill, error)
	ListAllBills(ctx context.Context, status *BillStatus, limit, offset int) ([]*Bill, error)
}

type TemporalClientInterface interface {
	ExecuteWorkflow(ctx context.Context, options client.StartWorkflowOptions, workflow interface{}, args ...interface{}) (client.WorkflowRun, error)
	SignalWorkflow(ctx context.Context, workflowID, runID, signalName string, arg interface{}) error
}

