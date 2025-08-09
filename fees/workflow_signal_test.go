package fees

import (
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.temporal.io/sdk/testsuite"
)

func TestBillWorkflow_Signals(t *testing.T) {
	testSuite := &testsuite.WorkflowTestSuite{}

	t.Run("Success_With_Signals", func(t *testing.T) {
		env := testSuite.NewTestWorkflowEnvironment()

		activities := &Activities{}
		env.RegisterActivity(activities.CalculateTotalActivity)
		env.RegisterActivity(activities.SaveFinalBillActivity)

		expectedItems := []LineItem{
			{Description: "Item 1", Amount: 1000},
			{Description: "Item 2", Amount: 1500},
		}

		env.OnActivity("CalculateTotalActivity", mock.Anything, expectedItems).Return(int64(2500), nil)
		env.OnActivity("SaveFinalBillActivity", mock.Anything, mock.MatchedBy(func(bill FinalBill) bool {
			return bill.ID == "bill-123" && bill.TotalAmount == 2500 && bill.Status == BillStatusClosed
		})).Return(nil)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(AddLineItemSignal, LineItem{Description: "Item 1", Amount: 1000})
		}, time.Second*1)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(AddLineItemSignal, LineItem{Description: "Item 2", Amount: 1500})
		}, time.Second*2)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(CloseBillSignal, nil)
		}, time.Second*3)

		initialBill := Bill{
			ID:         "bill-123",
			CustomerID: "customer-456",
			Currency:   USD,
			Status:     BillStatusOpen,
		}

		env.ExecuteWorkflow(BillWorkflow, initialBill)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		env.AssertExpectations(t)
	})

	t.Run("Immediate_Close", func(t *testing.T) {
		env := testSuite.NewTestWorkflowEnvironment()

		activities := &Activities{}
		env.RegisterActivity(activities.CalculateTotalActivity)
		env.RegisterActivity(activities.SaveFinalBillActivity)

		var emptyItems []LineItem = nil
		env.OnActivity("CalculateTotalActivity", mock.Anything, emptyItems).Return(int64(0), nil)
		env.OnActivity("SaveFinalBillActivity", mock.Anything, mock.MatchedBy(func(bill FinalBill) bool {
			return bill.ID == "bill-empty" && bill.TotalAmount == 0 && bill.Status == BillStatusClosed
		})).Return(nil)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(CloseBillSignal, nil)
		}, time.Millisecond*100)

		initialBill := Bill{
			ID:         "bill-empty",
			CustomerID: "customer-empty",
			Currency:   GEL,
			Status:     BillStatusOpen,
		}

		env.ExecuteWorkflow(BillWorkflow, initialBill)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		env.AssertExpectations(t)
	})

	t.Run("Multiple_Items_Then_Close", func(t *testing.T) {
		env := testSuite.NewTestWorkflowEnvironment()

		activities := &Activities{}
		env.RegisterActivity(activities.CalculateTotalActivity)
		env.RegisterActivity(activities.SaveFinalBillActivity)

		expectedItems := []LineItem{
			{Description: "Item 1", Amount: 500},
			{Description: "Item 2", Amount: 1000},
			{Description: "Item 3", Amount: 750},
		}

		env.OnActivity("CalculateTotalActivity", mock.Anything, expectedItems).Return(int64(2250), nil)
		env.OnActivity("SaveFinalBillActivity", mock.Anything, mock.MatchedBy(func(bill FinalBill) bool {
			return bill.ID == "bill-multi" && bill.TotalAmount == 2250 && bill.Status == BillStatusClosed
		})).Return(nil)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(AddLineItemSignal, LineItem{Description: "Item 1", Amount: 500})
		}, time.Millisecond*100)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(AddLineItemSignal, LineItem{Description: "Item 2", Amount: 1000})
		}, time.Millisecond*200)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(AddLineItemSignal, LineItem{Description: "Item 3", Amount: 750})
		}, time.Millisecond*300)

		env.RegisterDelayedCallback(func() {
			env.SignalWorkflow(CloseBillSignal, nil)
		}, time.Millisecond*400)

		initialBill := Bill{
			ID:         "bill-multi",
			CustomerID: "customer-multi",
			Currency:   USD,
			Status:     BillStatusOpen,
		}

		env.ExecuteWorkflow(BillWorkflow, initialBill)

		require.True(t, env.IsWorkflowCompleted())
		require.NoError(t, env.GetWorkflowError())

		env.AssertExpectations(t)
	})
}
