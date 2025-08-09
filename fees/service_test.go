package fees

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBillService_CreateBill(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)
	mockTemporal := NewMockTemporalClientInterface(ctrl)
	service := NewBillService(mockRepo, mockTemporal)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateBillRequest{
			CustomerID: "customer-123",
			Currency:   USD,
		}

		mockRepo.EXPECT().
			CreateBill(ctx, gomock.Any()).
			DoAndReturn(func(ctx context.Context, bill *Bill) error {
				assert.Equal(t, req.CustomerID, bill.CustomerID)
				assert.Equal(t, req.Currency, bill.Currency)
				assert.Equal(t, BillStatusOpen, bill.Status)
				assert.Equal(t, int64(0), bill.TotalAmount)
				assert.NotEmpty(t, bill.ID)
				assert.True(t, time.Since(bill.CreatedAt) < time.Second)
				return nil
			})

		mockTemporal.EXPECT().
			ExecuteWorkflow(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, nil)

		response, err := service.CreateBill(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.NotEmpty(t, response.BillID)
		assert.Contains(t, response.BillID, req.CustomerID)
	})

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateBillRequest{
			CustomerID: "",
			Currency:   USD,
		}

		response, err := service.CreateBill(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateBillRequest{
			CustomerID: "customer-123",
			Currency:   USD,
		}

		mockRepo.EXPECT().
			CreateBill(ctx, gomock.Any()).
			Return(errors.New("database error"))

		response, err := service.CreateBill(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to create bill")
	})

	t.Run("TemporalError", func(t *testing.T) {
		ctx := context.Background()
		req := &CreateBillRequest{
			CustomerID: "customer-123",
			Currency:   USD,
		}

		mockRepo.EXPECT().
			CreateBill(ctx, gomock.Any()).
			Return(nil)

		mockTemporal.EXPECT().
			ExecuteWorkflow(ctx, gomock.Any(), gomock.Any(), gomock.Any()).
			Return(nil, errors.New("workflow error"))

		response, err := service.CreateBill(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to start bill workflow")
	})
}

func TestBillService_AddLineItem(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)
	mockTemporal := NewMockTemporalClientInterface(ctrl)
	service := NewBillService(mockRepo, mockTemporal)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"
		req := &AddLineItemRequest{
			Description: "Test item",
			Amount:      1000,
		}

		mockRepo.EXPECT().
			GetBillStatus(ctx, billID).
			Return(BillStatusOpen, nil)

		mockRepo.EXPECT().
			AddLineItem(ctx, billID, gomock.Any()).
			DoAndReturn(func(ctx context.Context, billID string, item *LineItem) error {
				assert.Equal(t, req.Description, item.Description)
				assert.Equal(t, req.Amount, item.Amount)
				assert.True(t, time.Since(item.Timestamp) < time.Second)
				return nil
			})

		mockTemporal.EXPECT().
			SignalWorkflow(ctx, billID, "", AddLineItemSignal, gomock.Any()).
			Return(nil)

		err := service.AddLineItem(ctx, billID, req)

		require.NoError(t, err)
	})

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"
		req := &AddLineItemRequest{
			Description: "",
			Amount:      1000,
		}

		err := service.AddLineItem(ctx, billID, req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("BillNotFound", func(t *testing.T) {
		ctx := context.Background()
		billID := "nonexistent-bill"
		req := &AddLineItemRequest{
			Description: "Test item",
			Amount:      1000,
		}

		mockRepo.EXPECT().
			GetBillStatus(ctx, billID).
			Return(BillStatusOpen, ErrBillNotFound)

		err := service.AddLineItem(ctx, billID, req)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBillNotFound)
	})

	t.Run("BillClosed", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"
		req := &AddLineItemRequest{
			Description: "Test item",
			Amount:      1000,
		}

		mockRepo.EXPECT().
			GetBillStatus(ctx, billID).
			Return(BillStatusClosed, nil)

		err := service.AddLineItem(ctx, billID, req)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBillAlreadyClosed)
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"
		req := &AddLineItemRequest{
			Description: "Test item",
			Amount:      1000,
		}

		mockRepo.EXPECT().
			GetBillStatus(ctx, billID).
			Return(BillStatusOpen, nil)

		mockRepo.EXPECT().
			AddLineItem(ctx, billID, gomock.Any()).
			Return(errors.New("database error"))

		err := service.AddLineItem(ctx, billID, req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save line item")
	})

	t.Run("SignalError", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"
		req := &AddLineItemRequest{
			Description: "Test item",
			Amount:      1000,
		}

		mockRepo.EXPECT().
			GetBillStatus(ctx, billID).
			Return(BillStatusOpen, nil)

		mockRepo.EXPECT().
			AddLineItem(ctx, billID, gomock.Any()).
			Return(nil)

		mockTemporal.EXPECT().
			SignalWorkflow(ctx, billID, "", AddLineItemSignal, gomock.Any()).
			Return(errors.New("signal error"))

		err := service.AddLineItem(ctx, billID, req)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to signal workflow")
	})
}

func TestBillService_CloseBill(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)
	mockTemporal := NewMockTemporalClientInterface(ctrl)
	service := NewBillService(mockRepo, mockTemporal)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"

		bill := &Bill{
			ID:     billID,
			Status: BillStatusOpen,
		}

		mockRepo.EXPECT().
			GetBillByID(ctx, billID).
			Return(bill, nil)

		mockTemporal.EXPECT().
			SignalWorkflow(ctx, billID, "", CloseBillSignal, gomock.Any()).
			Return(nil)

		err := service.CloseBill(ctx, billID)

		require.NoError(t, err)
	})

	t.Run("BillNotFound", func(t *testing.T) {
		ctx := context.Background()
		billID := "nonexistent-bill"

		mockRepo.EXPECT().
			GetBillByID(ctx, billID).
			Return(nil, ErrBillNotFound)

		err := service.CloseBill(ctx, billID)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBillNotFound)
	})

	t.Run("BillAlreadyClosed", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"

		bill := &Bill{
			ID:     billID,
			Status: BillStatusClosed,
		}

		mockRepo.EXPECT().
			GetBillByID(ctx, billID).
			Return(bill, nil)

		err := service.CloseBill(ctx, billID)

		require.Error(t, err)
		assert.ErrorIs(t, err, ErrBillAlreadyClosed)
	})

	t.Run("SignalError", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"

		bill := &Bill{
			ID:     billID,
			Status: BillStatusOpen,
		}

		mockRepo.EXPECT().
			GetBillByID(ctx, billID).
			Return(bill, nil)

		mockTemporal.EXPECT().
			SignalWorkflow(ctx, billID, "", CloseBillSignal, gomock.Any()).
			Return(errors.New("signal error"))

		err := service.CloseBill(ctx, billID)

		require.Error(t, err)
		assert.Contains(t, err.Error(), "failed to signal bill to close")
	})
}

func TestBillService_GetBill(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)
	mockTemporal := NewMockTemporalClientInterface(ctrl)
	service := NewBillService(mockRepo, mockTemporal)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		billID := "bill-123"

		expectedBill := &Bill{
			ID:          billID,
			CustomerID:  "customer-456",
			Currency:    USD,
			Status:      BillStatusOpen,
			TotalAmount: 1500,
		}

		mockRepo.EXPECT().
			GetBillByID(ctx, billID).
			Return(expectedBill, nil)

		response, err := service.GetBill(ctx, billID)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, expectedBill, response.Bill)
	})

	t.Run("BillNotFound", func(t *testing.T) {
		ctx := context.Background()
		billID := "nonexistent-bill"

		mockRepo.EXPECT().
			GetBillByID(ctx, billID).
			Return(nil, ErrBillNotFound)

		response, err := service.GetBill(ctx, billID)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.ErrorIs(t, err, ErrBillNotFound)
	})
}

func TestBillService_ListBills(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)
	mockTemporal := NewMockTemporalClientInterface(ctrl)
	service := NewBillService(mockRepo, mockTemporal)

	t.Run("Success", func(t *testing.T) {
		ctx := context.Background()
		req := &ListBillsRequest{
			CustomerID: "customer-123",
			Limit:      10,
			Offset:     0,
		}

		expectedBills := []*Bill{
			{ID: "bill-1", CustomerID: "customer-123", Status: BillStatusOpen, Currency: USD},
			{ID: "bill-2", CustomerID: "customer-123", Status: BillStatusClosed, Currency: USD},
		}

		expectedSummaries := []*BillSummary{
			{ID: "bill-1", CustomerID: "customer-123", Status: BillStatusOpen, Currency: USD},
			{ID: "bill-2", CustomerID: "customer-123", Status: BillStatusClosed, Currency: USD},
		}

		mockRepo.EXPECT().
			ListBillsByCustomer(ctx, req.CustomerID, req.Status, req.Limit, req.Offset).
			Return(expectedBills, nil)

		response, err := service.ListBills(ctx, req)

		require.NoError(t, err)
		require.NotNil(t, response)
		assert.Equal(t, expectedSummaries, response.Bills)
		assert.Equal(t, len(expectedBills), response.Total)
	})

	t.Run("ValidationError", func(t *testing.T) {
		ctx := context.Background()
		req := &ListBillsRequest{
			CustomerID: "",
		}

		response, err := service.ListBills(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("RepositoryError", func(t *testing.T) {
		ctx := context.Background()
		req := &ListBillsRequest{
			CustomerID: "customer-123",
			Limit:      10,
			Offset:     0,
		}

		mockRepo.EXPECT().
			ListBillsByCustomer(ctx, req.CustomerID, req.Status, req.Limit, req.Offset).
			Return(nil, errors.New("database error"))

		response, err := service.ListBills(ctx, req)

		require.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to list bills")
	})
}

func TestBillService_ListAllBills(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)
	mockTemporal := NewMockTemporalClientInterface(ctrl)
	service := NewBillService(mockRepo, mockTemporal)

	t.Run("success", func(t *testing.T) {
		ctx := context.Background()
		req := &ListAllBillsRequest{
			Status: nil,
			Limit:  50,
			Offset: 0,
		}

		expectedBills := []*Bill{
			{ID: "bill-1", CustomerID: "customer-1", Currency: USD, Status: BillStatusOpen},
			{ID: "bill-2", CustomerID: "customer-2", Currency: GEL, Status: BillStatusClosed},
		}

		expectedSummaries := []*BillSummary{
			{ID: "bill-1", CustomerID: "customer-1", Currency: USD, Status: BillStatusOpen},
			{ID: "bill-2", CustomerID: "customer-2", Currency: GEL, Status: BillStatusClosed},
		}

		mockRepo.EXPECT().
			ListAllBills(ctx, req.Status, req.Limit, req.Offset).
			Return(expectedBills, nil)

		response, err := service.ListAllBills(ctx, req)

		assert.NoError(t, err)
		assert.NotNil(t, response)
		assert.Equal(t, expectedSummaries, response.Bills)
		assert.Equal(t, len(expectedBills), response.Total)
	})

	t.Run("validation error", func(t *testing.T) {
		ctx := context.Background()
		req := &ListAllBillsRequest{
			Limit: 1001, // This exceeds the max limit and will cause validation to fail
		}

		response, err := service.ListAllBills(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "validation failed")
	})

	t.Run("repository error", func(t *testing.T) {
		ctx := context.Background()
		req := &ListAllBillsRequest{
			Status: nil,
			Limit:  50,
			Offset: 0,
		}

		mockRepo.EXPECT().
			ListAllBills(ctx, req.Status, req.Limit, req.Offset).
			Return(nil, errors.New("database error"))

		response, err := service.ListAllBills(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, response)
		assert.Contains(t, err.Error(), "failed to list all bills")
	})
}
