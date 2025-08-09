package fees

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCurrency_Validate(t *testing.T) {
	tests := []struct {
		name     string
		currency Currency
		wantErr  bool
	}{
		{"USD valid", USD, false},
		{"GEL valid", GEL, false},
		{"EUR invalid", "EUR", true},
		{"empty invalid", "", true},
		{"lowercase invalid", "usd", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.currency.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidCurrency)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCurrency_IsValid(t *testing.T) {
	tests := []struct {
		currency Currency
		want     bool
	}{
		{USD, true},
		{GEL, true},
		{"EUR", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.currency), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.currency.IsValid())
		})
	}
}

func TestBillStatus_IsValid(t *testing.T) {
	tests := []struct {
		status BillStatus
		want   bool
	}{
		{BillStatusOpen, true},
		{BillStatusClosed, true},
		{"INVALID", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(string(tt.status), func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestCreateBillRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateBillRequest
		wantErr error
	}{
		{
			name: "valid request USD",
			req: CreateBillRequest{
				CustomerID: "customer123",
				Currency:   USD,
			},
			wantErr: nil,
		},
		{
			name: "valid request GEL",
			req: CreateBillRequest{
				CustomerID: "customer123",
				Currency:   GEL,
			},
			wantErr: nil,
		},
		{
			name: "empty customer ID",
			req: CreateBillRequest{
				CustomerID: "",
				Currency:   USD,
			},
			wantErr: ErrEmptyCustomerID,
		},
		{
			name: "whitespace only customer ID",
			req: CreateBillRequest{
				CustomerID: "   ",
				Currency:   USD,
			},
			wantErr: ErrEmptyCustomerID,
		},
		{
			name: "invalid currency",
			req: CreateBillRequest{
				CustomerID: "customer123",
				Currency:   "EUR",
			},
			wantErr: ErrInvalidCurrency,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAddLineItemRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     AddLineItemRequest
		wantErr error
	}{
		{
			name: "valid request",
			req: AddLineItemRequest{
				Description: "Test item",
				Amount:      1000,
			},
			wantErr: nil,
		},
		{
			name: "empty description",
			req: AddLineItemRequest{
				Description: "",
				Amount:      1000,
			},
			wantErr: ErrEmptyDescription,
		},
		{
			name: "whitespace only description",
			req: AddLineItemRequest{
				Description: "   ",
				Amount:      1000,
			},
			wantErr: ErrEmptyDescription,
		},
		{
			name: "zero amount",
			req: AddLineItemRequest{
				Description: "Test item",
				Amount:      0,
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "negative amount",
			req: AddLineItemRequest{
				Description: "Test item",
				Amount:      -100,
			},
			wantErr: ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLineItem_Validate(t *testing.T) {
	tests := []struct {
		name    string
		item    LineItem
		wantErr error
	}{
		{
			name: "valid item",
			item: LineItem{
				Description: "Test item",
				Amount:      100,
				Timestamp:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "empty description",
			item: LineItem{
				Description: "",
				Amount:      100,
				Timestamp:   time.Now(),
			},
			wantErr: ErrEmptyDescription,
		},
		{
			name: "zero amount",
			item: LineItem{
				Description: "Test item",
				Amount:      0,
				Timestamp:   time.Now(),
			},
			wantErr: ErrInvalidAmount,
		},
		{
			name: "negative amount",
			item: LineItem{
				Description: "Test item",
				Amount:      -50,
				Timestamp:   time.Now(),
			},
			wantErr: ErrInvalidAmount,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.item.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBill_Validate(t *testing.T) {
	tests := []struct {
		name    string
		bill    Bill
		wantErr error
	}{
		{
			name: "valid bill",
			bill: Bill{
				ID:          "bill-123",
				CustomerID:  "customer123",
				Currency:    USD,
				Status:      BillStatusOpen,
				TotalAmount: 0,
				CreatedAt:   time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "empty customer ID",
			bill: Bill{
				ID:       "bill-123",
				Currency: USD,
				Status:   BillStatusOpen,
			},
			wantErr: ErrEmptyCustomerID,
		},
		{
			name: "invalid currency",
			bill: Bill{
				ID:         "bill-123",
				CustomerID: "customer123",
				Currency:   "EUR",
				Status:     BillStatusOpen,
			},
			wantErr: ErrInvalidCurrency,
		},
		{
			name: "invalid status",
			bill: Bill{
				ID:         "bill-123",
				CustomerID: "customer123",
				Currency:   USD,
				Status:     "INVALID",
			},
			wantErr: nil, // Will return a generic error, not a specific one
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.bill.Validate()
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else if tt.name == "invalid status" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), "invalid bill status")
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBill_CalculateTotal(t *testing.T) {
	tests := []struct {
		name     string
		bill     Bill
		expected int64
	}{
		{
			name:     "empty bill",
			bill:     Bill{LineItems: []LineItem{}},
			expected: 0,
		},
		{
			name: "single item",
			bill: Bill{
				LineItems: []LineItem{
					{Amount: 100},
				},
			},
			expected: 100,
		},
		{
			name: "multiple items",
			bill: Bill{
				LineItems: []LineItem{
					{Amount: 100},
					{Amount: 250},
					{Amount: 50},
				},
			},
			expected: 400,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			total := tt.bill.CalculateTotal()
			assert.Equal(t, tt.expected, total)
		})
	}
}

func TestBill_CanAddLineItem(t *testing.T) {
	tests := []struct {
		name     string
		status   BillStatus
		expected bool
	}{
		{"open bill", BillStatusOpen, true},
		{"closed bill", BillStatusClosed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bill := &Bill{Status: tt.status}
			assert.Equal(t, tt.expected, bill.CanAddLineItem())
		})
	}
}

func TestListBillsRequest_Validate(t *testing.T) {
	openStatus := BillStatusOpen
	closedStatus := BillStatusClosed
	invalidStatus := BillStatus("INVALID")

	tests := []struct {
		name           string
		req            ListBillsRequest
		wantErr        error
		expectedLimit  int
		expectedOffset int
	}{
		{
			name: "valid request with defaults",
			req: ListBillsRequest{
				CustomerID: "customer123",
			},
			wantErr:        nil,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "valid request with custom values",
			req: ListBillsRequest{
				CustomerID: "customer123",
				Status:     &openStatus,
				Limit:      50,
				Offset:     20,
			},
			wantErr:        nil,
			expectedLimit:  50,
			expectedOffset: 20,
		},
		{
			name: "limit too high gets capped",
			req: ListBillsRequest{
				CustomerID: "customer123",
				Limit:      200,
			},
			wantErr:        nil,
			expectedLimit:  100,
			expectedOffset: 0,
		},
		{
			name: "negative offset gets reset",
			req: ListBillsRequest{
				CustomerID: "customer123",
				Offset:     -10,
			},
			wantErr:        nil,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "zero limit gets default",
			req: ListBillsRequest{
				CustomerID: "customer123",
				Limit:      0,
			},
			wantErr:        nil,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "empty customer ID",
			req: ListBillsRequest{
				CustomerID: "",
			},
			wantErr: ErrEmptyCustomerID,
		},
		{
			name: "valid status filter",
			req: ListBillsRequest{
				CustomerID: "customer123",
				Status:     &closedStatus,
			},
			wantErr:        nil,
			expectedLimit:  10,
			expectedOffset: 0,
		},
		{
			name: "invalid status filter",
			req: ListBillsRequest{
				CustomerID: "customer123",
				Status:     &invalidStatus,
			},
			wantErr: fmt.Errorf("invalid status filter: INVALID"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				if tt.name == "invalid status filter" {
					assert.Contains(t, err.Error(), "invalid status filter")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLimit, tt.req.Limit)
				assert.Equal(t, tt.expectedOffset, tt.req.Offset)
			}
		})
	}
}

func TestListAllBillsRequest_Validate(t *testing.T) {
	tests := []struct {
		name           string
		req            ListAllBillsRequest
		wantErr        error
		expectedLimit  int
		expectedOffset int
	}{
		{
			name: "valid request with defaults",
			req: ListAllBillsRequest{
				Limit:  0,
				Offset: 0,
			},
			expectedLimit:  50,
			expectedOffset: 0,
		},
		{
			name: "valid request with status",
			req: ListAllBillsRequest{
				Status: &[]BillStatus{BillStatusOpen}[0],
				Limit:  25,
				Offset: 10,
			},
			expectedLimit:  25,
			expectedOffset: 10,
		},
		{
			name: "invalid status filter",
			req: ListAllBillsRequest{
				Status: &[]BillStatus{"INVALID"}[0],
				Limit:  50,
			},
			wantErr: errors.New("invalid status filter"),
		},
		{
			name: "limit too high",
			req: ListAllBillsRequest{
				Limit: 1001,
			},
			wantErr: errors.New("limit cannot exceed 1000"),
		},
		{
			name: "negative offset gets corrected",
			req: ListAllBillsRequest{
				Limit:  50,
				Offset: -5,
			},
			expectedLimit:  50,
			expectedOffset: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr != nil {
				assert.Error(t, err)
				if tt.name == "invalid status filter" {
					assert.Contains(t, err.Error(), "invalid status filter")
				} else if tt.name == "limit too high" {
					assert.Contains(t, err.Error(), "limit cannot exceed 1000")
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedLimit, tt.req.Limit)
				assert.Equal(t, tt.expectedOffset, tt.req.Offset)
			}
		})
	}
}
