package fees

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This test works perfectly once the global init() panic is removed.
// It tests the activity's logic in complete isolation.
func TestActivities_SaveFinalBillActivity_Unit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	// 1. Create a mock of the repository dependency.
	mockRepo := NewMockRepositoryInterface(ctrl)

	// 2. Inject the mock into the Activities struct.
	activities := NewActivities(mockRepo)

	// 3. Define the expected call on the mock.
	// We expect UpdateBillStatus to be called once with these exact arguments.
	billToSave := FinalBill{
		ID:          "bill-123",
		TotalAmount: 1000,
		Status:      BillStatusClosed,
	}
	mockRepo.EXPECT().
		UpdateBillStatus(gomock.Any(), "bill-123", BillStatusClosed, int64(1000)).
		Return(nil). // Tell the mock to return no error.
		Times(1)     // Expect it to be called exactly once.

	// 4. Run the activity function.
	err := activities.SaveFinalBillActivity(context.Background(), billToSave)

	// 5. Assert that the result is what you expect.
	require.NoError(t, err)
}

// Test for the error case
func TestActivities_SaveFinalBillActivity_Unit_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)
	activities := NewActivities(mockRepo)

	// Define that the mock should return an error
	mockRepo.EXPECT().
		UpdateBillStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("database is down")).
		Times(1)

	err := activities.SaveFinalBillActivity(context.Background(), FinalBill{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database is down")
}
