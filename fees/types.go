package fees

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrBillNotFound      = errors.New("bill not found")
	ErrBillAlreadyClosed = errors.New("bill is already closed")
	ErrInvalidCurrency   = errors.New("invalid currency")
	ErrInvalidAmount     = errors.New("amount must be positive")
	ErrEmptyDescription  = errors.New("description cannot be empty")
	ErrEmptyCustomerID   = errors.New("customer ID cannot be empty")
	ErrInvalidBillID     = errors.New("invalid bill ID format")
)

type Currency string

const (
	USD Currency = "USD"
	GEL Currency = "GEL"
)

func (c Currency) IsValid() bool {
	return c == USD || c == GEL
}

func (c Currency) Validate() error {
	if !c.IsValid() {
		return fmt.Errorf("%w: %s. Supported currencies: USD, GEL", ErrInvalidCurrency, c)
	}
	return nil
}

type BillStatus string

const (
	BillStatusOpen   BillStatus = "OPEN"
	BillStatusClosed BillStatus = "CLOSED"
)

func (bs BillStatus) IsValid() bool {
	return bs == BillStatusOpen || bs == BillStatusClosed
}

type LineItem struct {
	Description string    `json:"description"`
	Amount      int64     `json:"amount"`
	Timestamp   time.Time `json:"timestamp"`
}

func (li *LineItem) Validate() error {
	if strings.TrimSpace(li.Description) == "" {
		return ErrEmptyDescription
	}
	if li.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

type Bill struct {
	ID          string     `json:"id"`
	CustomerID  string     `json:"customerId"`
	Currency    Currency   `json:"currency"`
	Status      BillStatus `json:"status"`
	LineItems   []LineItem `json:"lineItems"`
	TotalAmount int64      `json:"totalAmount"`
	CreatedAt   time.Time  `json:"createdAt"`
}

type BillSummary struct {
	ID         string     `json:"id"`
	CustomerID string     `json:"customerId"`
	Currency   Currency   `json:"currency"`
	Status     BillStatus `json:"status"`
	CreatedAt  time.Time  `json:"createdAt"`
}

func (b *Bill) Validate() error {
	if strings.TrimSpace(b.CustomerID) == "" {
		return ErrEmptyCustomerID
	}
	if err := b.Currency.Validate(); err != nil {
		return err
	}
	if !b.Status.IsValid() {
		return fmt.Errorf("invalid bill status: %s", b.Status)
	}
	return nil
}

func (b *Bill) CalculateTotal() int64 {
	var total int64
	for _, item := range b.LineItems {
		total += item.Amount
	}
	return total
}

func (b *Bill) CanAddLineItem() bool {
	return b.Status == BillStatusOpen
}

type CreateBillRequest struct {
	CustomerID string   `json:"customerId"`
	Currency   Currency `json:"currency"`
}

func (r *CreateBillRequest) Validate() error {
	if strings.TrimSpace(r.CustomerID) == "" {
		return ErrEmptyCustomerID
	}
	return r.Currency.Validate()
}

type CreateBillResponse struct {
	BillID string `json:"billId"`
}

type AddLineItemRequest struct {
	Description string `json:"description"`
	Amount      int64  `json:"amount"`
}

func (r *AddLineItemRequest) Validate() error {
	if strings.TrimSpace(r.Description) == "" {
		return ErrEmptyDescription
	}
	if r.Amount <= 0 {
		return ErrInvalidAmount
	}
	return nil
}

type GetBillResponse struct {
	Bill *Bill `json:"bill"`
}

type ListBillsRequest struct {
	CustomerID string      `json:"customerId"`
	Status     *BillStatus `json:"status,omitempty"`
	Limit      int         `json:"limit,omitempty"`
	Offset     int         `json:"offset,omitempty"`
}

func (r *ListBillsRequest) Validate() error {
	if strings.TrimSpace(r.CustomerID) == "" {
		return ErrEmptyCustomerID
	}
	if r.Status != nil && !r.Status.IsValid() {
		return fmt.Errorf("invalid status filter: %s", *r.Status)
	}
	if r.Limit <= 0 {
		r.Limit = 10
	}
	if r.Limit > 100 {
		r.Limit = 100
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
	return nil
}

type ListBillsResponse struct {
	Bills []*BillSummary `json:"bills"`
	Total int           `json:"total"`
}

type ListAllBillsRequest struct {
	Status *BillStatus `json:"status,omitempty"`
	Limit  int         `json:"limit,omitempty"`
	Offset int         `json:"offset,omitempty"`
}

func (r *ListAllBillsRequest) Validate() error {
	if r.Status != nil && !r.Status.IsValid() {
		return fmt.Errorf("invalid status filter: %s", *r.Status)
	}
	if r.Limit <= 0 {
		r.Limit = 50
	}
	if r.Offset < 0 {
		r.Offset = 0
	}
	if r.Limit > 1000 {
		return errors.New("limit cannot exceed 1000")
	}
	return nil
}
