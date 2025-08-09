package fees

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"pave-fees/fees/internal/temporal"
)

var (
	svc  *BillService
	once sync.Once
	err  error
)

func initService() (*BillService, error) {
	tc, err := temporal.NewClient(temporal.ClientOptions{
		Target:    "127.0.0.1:7233",
		Namespace: "default",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create temporal client: %w", err)
	}

	repo := NewRepository(getDB())
	activities := NewActivities(repo)

	tc.RegisterWorkflow(BillWorkflow)
	tc.RegisterActivity(activities.CalculateTotalActivity)
	tc.RegisterActivity(activities.SaveFinalBillActivity)

	if err := tc.StartWorker(); err != nil {
		return nil, fmt.Errorf("failed to start temporal worker: %w", err)
	}

	service := NewBillService(repo, tc)
	slog.Info("Fees service initialized successfully")

	return service, nil
}

func getService() (*BillService, error) {
	once.Do(func() {
		svc, err = initService()
		if err != nil {
			slog.Error("Failed to initialize fees service", "error", err)
		}
	})
	return svc, err
}


//encore:api public method=POST path=/bills
func CreateBill(ctx context.Context, req *CreateBillRequest) (*CreateBillResponse, error) {
	service, err := getService()
	if err != nil {
		return nil, fmt.Errorf("service initialization failed: %w", err)
	}
	return service.CreateBill(ctx, req)
}

//encore:api public method=POST path=/bills/:billID/items
func AddLineItem(ctx context.Context, billID string, req *AddLineItemRequest) error {
	service, err := getService()
	if err != nil {
		return fmt.Errorf("service initialization failed: %w", err)
	}
	return service.AddLineItem(ctx, billID, req)
}

//encore:api public method=POST path=/bills/:billID/close
func CloseBill(ctx context.Context, billID string) error {
	service, err := getService()
	if err != nil {
		return fmt.Errorf("service initialization failed: %w", err)
	}
	return service.CloseBill(ctx, billID)
}

//encore:api public method=GET path=/bills/:billID
func GetBill(ctx context.Context, billID string) (*GetBillResponse, error) {
	service, err := getService()
	if err != nil {
		return nil, fmt.Errorf("service initialization failed: %w", err)
	}
	return service.GetBill(ctx, billID)
}

type ListBillsParams struct {
	Status string `query:"status"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

//encore:api public method=GET path=/customers/:customerID/bills
func ListBills(ctx context.Context, customerID string, params ListBillsParams) (*ListBillsResponse, error) {
	req := &ListBillsRequest{
		CustomerID: customerID,
		Status:     nil,
		Limit:      10,
		Offset:     0,
	}
	
	// Convert string status to BillStatus if provided and not empty
	if params.Status != "" {
		status := BillStatus(params.Status)
		req.Status = &status
	}

	// Use provided limit if greater than 0, otherwise default to 10
	if params.Limit > 0 {
		req.Limit = params.Limit
	}

	// Use provided offset if greater than 0, otherwise default to 0
	if params.Offset > 0 {
		req.Offset = params.Offset
	}

	service, err := getService()
	if err != nil {
		return nil, fmt.Errorf("service initialization failed: %w", err)
	}
	return service.ListBills(ctx, req)
}

type ListAllBillsParams struct {
	Status string `query:"status"`
	Limit  int    `query:"limit"`
	Offset int    `query:"offset"`
}

//encore:api public method=GET path=/bills
func ListAllBills(ctx context.Context, params ListAllBillsParams) (*ListBillsResponse, error) {
	req := &ListAllBillsRequest{
		Status: nil,
		Limit:  50,
		Offset: 0,
	}
	
	// Convert string status to BillStatus if provided and not empty
	if params.Status != "" {
		status := BillStatus(params.Status)
		req.Status = &status
	}

	// Use provided limit if greater than 0, otherwise default to 50
	if params.Limit > 0 {
		req.Limit = params.Limit
	}

	// Use provided offset if greater than 0, otherwise default to 0
	if params.Offset > 0 {
		req.Offset = params.Offset
	}

	service, err := getService()
	if err != nil {
		return nil, fmt.Errorf("service initialization failed: %w", err)
	}
	return service.ListAllBills(ctx, req)
}
