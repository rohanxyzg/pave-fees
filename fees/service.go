package fees

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"pave-fees/fees/internal/temporal"

	"go.temporal.io/sdk/client"
)

type BillService struct {
	repo     RepositoryInterface
	temporal TemporalClientInterface
}

func NewBillService(repo RepositoryInterface, temporalClient TemporalClientInterface) *BillService {
	return &BillService{
		repo:     repo,
		temporal: temporalClient,
	}
}

func (s *BillService) CreateBill(ctx context.Context, req *CreateBillRequest) (*CreateBillResponse, error) {
	if err := req.Validate(); err != nil {
		slog.Error("invalid create bill request", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	billID := fmt.Sprintf("bill-%s-%d", req.CustomerID, time.Now().UnixNano())

	bill := &Bill{
		ID:          billID,
		CustomerID:  req.CustomerID,
		Currency:    req.Currency,
		Status:      BillStatusOpen,
		TotalAmount: 0,
		CreatedAt:   time.Now(),
		LineItems:   make([]LineItem, 0),
	}

	if err := s.repo.CreateBill(ctx, bill); err != nil {
		slog.Error("failed to create bill in repository", "bill_id", billID, "error", err)
		return nil, fmt.Errorf("failed to create bill: %w", err)
	}

	workflowOptions := client.StartWorkflowOptions{
		ID:        billID,
		TaskQueue: temporal.TaskQueue,
	}

	_, err := s.temporal.ExecuteWorkflow(ctx, workflowOptions, BillWorkflow, *bill)
	if err != nil {
		slog.Error("failed to start bill workflow", "bill_id", billID, "error", err)
		return nil, fmt.Errorf("failed to start bill workflow: %w", err)
	}

	slog.Info("bill created successfully", "bill_id", billID, "customer_id", req.CustomerID)
	return &CreateBillResponse{BillID: billID}, nil
}

func (s *BillService) AddLineItem(ctx context.Context, billID string, req *AddLineItemRequest) error {
	if err := req.Validate(); err != nil {
		slog.Error("invalid add line item request", "bill_id", billID, "error", err)
		return fmt.Errorf("validation failed: %w", err)
	}

	status, err := s.repo.GetBillStatus(ctx, billID)
	if err != nil {
		slog.Error("failed to get bill status", "bill_id", billID, "error", err)
		return err
	}

	if status == BillStatusClosed {
		slog.Warn("attempted to add line item to closed bill", "bill_id", billID)
		return ErrBillAlreadyClosed
	}

	item := &LineItem{
		Description: req.Description,
		Amount:      req.Amount,
		Timestamp:   time.Now(),
	}

	if err := s.repo.AddLineItem(ctx, billID, item); err != nil {
		slog.Error("failed to add line item to repository", "bill_id", billID, "error", err)
		return fmt.Errorf("failed to save line item: %w", err)
	}

	if err := s.temporal.SignalWorkflow(ctx, billID, "", AddLineItemSignal, *item); err != nil {
		slog.Warn("failed to signal workflow for new line item", "bill_id", billID, "error", err)
		return fmt.Errorf("failed to signal workflow: %w", err)
	}

	slog.Info("line item added successfully", "bill_id", billID, "description", req.Description, "amount", req.Amount)
	return nil
}

func (s *BillService) CloseBill(ctx context.Context, billID string) error {
	bill, err := s.repo.GetBillByID(ctx, billID)
	if err != nil {
		slog.Error("failed to get bill for closing", "bill_id", billID, "error", err)
		return err
	}

	if bill.Status == BillStatusClosed {
		slog.Warn("attempted to close already closed bill", "bill_id", billID)
		return ErrBillAlreadyClosed
	}

	if err := s.temporal.SignalWorkflow(ctx, billID, "", CloseBillSignal, struct{}{}); err != nil {
		slog.Error("failed to signal bill to close", "bill_id", billID, "error", err)
		return fmt.Errorf("failed to signal bill to close: %w", err)
	}

	slog.Info("bill close signal sent successfully", "bill_id", billID)
	return nil
}

func (s *BillService) GetBill(ctx context.Context, billID string) (*GetBillResponse, error) {
	bill, err := s.repo.GetBillByID(ctx, billID)
	if err != nil {
		slog.Error("failed to get bill", "bill_id", billID, "error", err)
		return nil, err
	}

	slog.Debug("bill retrieved successfully", "bill_id", billID, "status", bill.Status)
	return &GetBillResponse{Bill: bill}, nil
}

func (s *BillService) ListBills(ctx context.Context, req *ListBillsRequest) (*ListBillsResponse, error) {
	if err := req.Validate(); err != nil {
		slog.Error("invalid list bills request", "error", err)
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	bills, err := s.repo.ListBillsByCustomer(ctx, req.CustomerID, req.Status, req.Limit, req.Offset)
	if err != nil {
		slog.Error("failed to list bills", "customer_id", req.CustomerID, "error", err)
		return nil, fmt.Errorf("failed to list bills: %w", err)
	}

	slog.Debug("bills listed successfully", "customer_id", req.CustomerID, "count", len(bills))
	return &ListBillsResponse{Bills: bills, Total: len(bills)}, nil
}
