package memorystorage

import (
	"context"
	"testing"

	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/models"
	"github.com/VictorLeskin/otus_final_project_of_golang_course/internal/storage"

	"github.com/stretchr/testify/assert"
)

func TestStorage_find(t *testing.T) {
	t0 := New()

	assert.Nil(t, t0.find("192.168.1.0/24", models.White))
	t1 := &models.IPList{
		Subnet:  "192.168.1.0/24",
		IsWhite: models.White,
	}
	t0.ipList = append(t0.ipList, t1)

	res := t0.find("192.168.1.0/24", models.White)
	assert.NotNil(t, res)
	assert.Equal(t, "white", res.IsWhite.String())
	assert.Equal(t, "192.168.1.0/24", res.Subnet)
	assert.Nil(t, t0.find("192.168.1.0/24", models.Black))
	assert.Nil(t, t0.find("192.168.1.0/21", models.White))

	t2 := &models.IPList{
		Subnet:  "192.168.1.1/24",
		IsWhite: models.White,
	}
	t0.ipList = append(t0.ipList, t2)

	t3 := &models.IPList{
		Subnet:  "192.168.1.2/24",
		IsWhite: models.Black,
	}
	t0.ipList = append(t0.ipList, t3)

	assert.NotNil(t, t0.find("192.168.1.2/24", models.Black))
}

func TestStorage_Add(t *testing.T) {
	t1 := &models.IPList{
		ID:      "id",
		Subnet:  "192.168.1.0/24",
		IsWhite: models.White,
	}

	t.Run("successful", func(t *testing.T) {
		t0 := New()
		err := t0.Add(context.Background(), t1)
		assert.NoError(t, err)

		savedEvent := t0.find("192.168.1.0/24", models.White)
		assert.Equal(t, "id", savedEvent.ID)
	})

	t.Run("invalid subnet", func(t *testing.T) {
		t0 := New()
		t2 := &models.IPList{
			Subnet:  "192.168.1.256/24",
			IsWhite: models.White,
		}

		err := t0.Add(context.Background(), t2)
		assert.Error(t, err)
		assert.ErrorIs(t, err, storage.ErrInvalidSubnetDetected)
	})

	t.Run("nothing to do: duplicate event", func(t *testing.T) {
		t0 := New()
		err := t0.Add(context.Background(), t1)
		assert.NoError(t, err)

		err = t0.Add(context.Background(), t1)
		assert.NoError(t, err)
	})

	t.Run("fail: context cancellation", func(t *testing.T) {
		t0 := New()

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := t0.Add(ctx, t1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)

		// Check what cancelling happened before adding
		exists := t0.find("192.168.1.0/24", models.White)
		assert.Nil(t, exists, "Event should NOT exist in storage")
	})
}
