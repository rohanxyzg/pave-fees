package fees

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestActivities_SaveFinalBillActivity_Unit(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)

	activities := NewActivities(mockRepo)

	billToSave := FinalBill{
		ID:          "bill-123",
		TotalAmount: 1000,
		Status:      BillStatusClosed,
	}
	mockRepo.EXPECT().
		UpdateBillStatus(gomock.Any(), "bill-123", BillStatusClosed, int64(1000)).
		Return(nil).
		Times(1)

	err := activities.SaveFinalBillActivity(context.Background(), billToSave)

	require.NoError(t, err)
}

func TestActivities_SaveFinalBillActivity_Unit_Error(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := NewMockRepositoryInterface(ctrl)
	activities := NewActivities(mockRepo)

	mockRepo.EXPECT().
		UpdateBillStatus(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		Return(errors.New("database is down")).
		Times(1)

	err := activities.SaveFinalBillActivity(context.Background(), FinalBill{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "database is down")
}
