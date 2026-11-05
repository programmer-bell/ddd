package memory

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/programmer-bell/ddd/aggregate"
)

func TestMemory_GetCustom(t *testing.T) {
	type testCase struct {
		name        string
		id          uuid.UUID
		expectedErr error
	}

	cust, err := aggregate.NewCustomer("bell")
	if err != nil {
		t.Fatal(err)
	}

	id := cust.GetID()

	repo := MemoryRepository{
		customers: map[uuid.UUID]aggregate.Customer{
			id: cust,
		},
	}
	testCases := []testCase{
		{
			name: "no customer by id",
			id:   uuid.MustParse("f47ac10b-58cc-0372-8567-0e3ht4qg6s3a"),
		}, {
			name:        "customer by id",
			id:          id,
			expectedErr: nil,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := repo.Get(tc.id)

			if !errors.Is(err, tc.expectedErr) {
				t.Errorf("expected error %v, got %v", tc.expectedErr, err)
			}
		})
	}
}
