package fees

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"encore.dev/storage/sqldb"
)

type Repository struct {
	db *sqldb.Database
}

func NewRepository(db *sqldb.Database) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateBill(ctx context.Context, bill *Bill) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO bills (id, customer_id, currency, status, total_amount, created_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, bill.ID, bill.CustomerID, bill.Currency, bill.Status, bill.TotalAmount, time.Now())
	
	if err != nil {
		return fmt.Errorf("failed to create bill: %w", err)
	}
	return nil
}

func (r *Repository) GetBillByID(ctx context.Context, billID string) (*Bill, error) {
	var bill Bill
	err := r.db.QueryRow(ctx, 
		"SELECT id, customer_id, currency, status, total_amount FROM bills WHERE id = $1", 
		billID,
	).Scan(&bill.ID, &bill.CustomerID, &bill.Currency, &bill.Status, &bill.TotalAmount)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, ErrBillNotFound
		}
		return nil, fmt.Errorf("failed to get bill: %w", err)
	}
	
	lineItems, err := r.GetLineItemsByBillID(ctx, billID)
	if err != nil {
		return nil, fmt.Errorf("failed to get line items for bill %s: %w", billID, err)
	}
	bill.LineItems = lineItems
	
	return &bill, nil
}

func (r *Repository) GetBillStatus(ctx context.Context, billID string) (BillStatus, error) {
	var status BillStatus
	err := r.db.QueryRow(ctx, "SELECT status FROM bills WHERE id = $1", billID).Scan(&status)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrBillNotFound
		}
		return "", fmt.Errorf("failed to get bill status: %w", err)
	}
	return status, nil
}

func (r *Repository) AddLineItem(ctx context.Context, billID string, item *LineItem) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO line_items (bill_id, description, amount, timestamp)
		VALUES ($1, $2, $3, $4)
	`, billID, item.Description, item.Amount, item.Timestamp)
	
	if err != nil {
		return fmt.Errorf("failed to add line item: %w", err)
	}
	return nil
}

func (r *Repository) GetLineItemsByBillID(ctx context.Context, billID string) ([]LineItem, error) {
	rows, err := r.db.Query(ctx, 
		"SELECT description, amount, timestamp FROM line_items WHERE bill_id = $1 ORDER BY timestamp ASC", 
		billID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query line items: %w", err)
	}
	defer rows.Close()
	
	var lineItems []LineItem
	for rows.Next() {
		var item LineItem
		if err := rows.Scan(&item.Description, &item.Amount, &item.Timestamp); err != nil {
			return nil, fmt.Errorf("failed to scan line item: %w", err)
		}
		lineItems = append(lineItems, item)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating line items: %w", err)
	}
	
	return lineItems, nil
}

func (r *Repository) UpdateBillStatus(ctx context.Context, billID string, status BillStatus, totalAmount int64) error {
	_, err := r.db.Exec(ctx, `
		UPDATE bills 
		SET status = $1, total_amount = $2 
		WHERE id = $3
	`, status, totalAmount, billID)
	
	if err != nil {
		return fmt.Errorf("failed to update bill status: %w", err)
	}
	
	var exists bool
	err = r.db.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM bills WHERE id = $1)", billID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to verify bill exists: %w", err)
	}
	
	if !exists {
		return ErrBillNotFound
	}
	
	return nil
}

func (r *Repository) listBills(ctx context.Context, customerID *string, status *BillStatus, limit, offset int, includeLineItems bool) ([]*Bill, error) {
	query := `SELECT id, customer_id, currency, status, total_amount FROM bills`
	var args []interface{}
	var conditions []string
	
	if customerID != nil {
		conditions = append(conditions, "customer_id = $1")
		args = append(args, *customerID)
	}
	
	if status != nil {
		placeholder := fmt.Sprintf("$%d", len(args)+1)
		conditions = append(conditions, fmt.Sprintf("status = %s", placeholder))
		args = append(args, *status)
	}
	
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)
	
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list bills: %w", err)
	}
	defer rows.Close()
	
	var bills []*Bill
	for rows.Next() {
		var bill Bill
		if err := rows.Scan(&bill.ID, &bill.CustomerID, &bill.Currency, &bill.Status, &bill.TotalAmount); err != nil {
			return nil, fmt.Errorf("failed to scan bill: %w", err)
		}
		
		if includeLineItems {
			lineItems, err := r.GetLineItemsByBillID(ctx, bill.ID)
			if err != nil {
				return nil, fmt.Errorf("failed to get line items for bill %s: %w", bill.ID, err)
			}
			bill.LineItems = lineItems
		} else {
			bill.TotalAmount = 0
		}
		
		bills = append(bills, &bill)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bills: %w", err)
	}
	
	return bills, nil
}

func (r *Repository) ListBillsByCustomer(ctx context.Context, customerID string, status *BillStatus, limit, offset int) ([]*Bill, error) {
	return r.listBills(ctx, &customerID, status, limit, offset, false)
}

func (r *Repository) ListAllBills(ctx context.Context, status *BillStatus, limit, offset int) ([]*Bill, error) {
	return r.listBills(ctx, nil, status, limit, offset, false)
}