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
	t1 := models.IPList{
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
		t2 := models.IPList{
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

func TestStorage_Remove(t *testing.T) {
	t1 := models.IPList{
		Subnet:  "192.168.1.0/24",
		IsWhite: models.White,
	}

	t.Run("empty list", func(t *testing.T) {
		t0 := New()
		err := t0.Remove(context.Background(), t1)
		assert.NoError(t, err)
		assert.Equal(t, 0, len(t0.ipList))
	})

	t.Run("remove only subnet", func(t *testing.T) {
		t0 := New()
		t0.Add(context.Background(), t1)
		assert.Equal(t, 1, len(t0.ipList))

		t2 := models.IPList{
			Subnet:  "192.168.1.256/24",
			IsWhite: models.White,
		}

		err := t0.Remove(context.Background(), t2)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(t0.ipList))

		err = t0.Remove(context.Background(), t1)
		assert.Equal(t, 0, len(t0.ipList))
		assert.NoError(t, err)
	})

	t.Run("fail: context cancellation", func(t *testing.T) {
		t0 := New()
		t0.Add(context.Background(), t1)
		assert.Equal(t, 1, len(t0.ipList))

		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		err := t0.Remove(ctx, t1)
		assert.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled)
	})
}

func TestStorage_GetIpList(t *testing.T) {
	t0 := New()
	r1, err1 := t0.GetIpList(context.Background(), models.White)
	assert.Equal(t, 0, len(r1))
	assert.NoError(t, err1)

	t1 := &models.IPList{
		Subnet:  "192.168.1.0/24",
		IsWhite: models.White,
	}

	t2 := &models.IPList{
		Subnet:  "192.168.1.1/24",
		IsWhite: models.Black,
	}

	t3 := &models.IPList{
		Subnet:  "192.168.1.2/24",
		IsWhite: models.White,
	}

	t0.ipList = append(t0.ipList, t1)
	t0.ipList = append(t0.ipList, t2)
	t0.ipList = append(t0.ipList, t3)

	r2, err2 := t0.GetIpList(context.Background(), models.White)
	assert.Equal(t, 2, len(r2))
	assert.Equal(t, "192.168.1.0/24", r2[0].Subnet)
	assert.Equal(t, "192.168.1.2/24", r2[1].Subnet)
	assert.NoError(t, err2)

	r3, err3 := t0.GetIpList(context.Background(), models.Black)
	assert.Equal(t, 1, len(r3))
	assert.Equal(t, "192.168.1.1/24", r3[0].Subnet)
	assert.NoError(t, err3)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r2, err2 = t0.GetIpList(ctx, models.White)
	assert.Nil(t, r2)
	assert.ErrorIs(t, err2, context.Canceled)
}

func TestStorage_GetAll(t *testing.T) {
	t0 := New()
	r1, err1 := t0.GetAll(context.Background())
	assert.Equal(t, 0, len(r1))
	assert.NoError(t, err1)

	t1 := &models.IPList{
		Subnet:  "192.168.1.0/24",
		IsWhite: models.White,
	}

	t2 := &models.IPList{
		Subnet:  "192.168.1.1/24",
		IsWhite: models.Black,
	}

	t3 := &models.IPList{
		Subnet:  "192.168.1.2/24",
		IsWhite: models.White,
	}

	t0.ipList = append(t0.ipList, t1)
	t0.ipList = append(t0.ipList, t2)
	t0.ipList = append(t0.ipList, t3)

	r2, err2 := t0.GetAll(context.Background())
	assert.Equal(t, 3, len(r2))
	assert.Equal(t, "192.168.1.0/24", r2[0].Subnet)
	assert.Equal(t, "192.168.1.1/24", r2[1].Subnet)
	assert.Equal(t, "192.168.1.2/24", r2[2].Subnet)
	assert.NoError(t, err2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	r2, err2 = t0.GetAll(ctx)
	assert.Nil(t, r2)
	assert.ErrorIs(t, err2, context.Canceled)
}

func TestStorage_Clear(t *testing.T) {
	t0 := New()
	err1 := t0.Clear(context.Background(), models.Black)
	assert.NoError(t, err1)

	t1 := &models.IPList{
		Subnet:  "192.168.1.0/24",
		IsWhite: models.White,
	}

	t2 := &models.IPList{
		Subnet:  "192.168.1.1/24",
		IsWhite: models.Black,
	}

	t3 := &models.IPList{
		Subnet:  "192.168.1.2/24",
		IsWhite: models.White,
	}

	t0.ipList = append(t0.ipList, t1)
	t0.ipList = append(t0.ipList, t2)
	t0.ipList = append(t0.ipList, t3)

	err2 := t0.Clear(context.Background(), models.Black)
	r2, _ := t0.GetAll(context.Background())

	assert.Equal(t, 2, len(r2))
	assert.Equal(t, "192.168.1.0/24", r2[0].Subnet)
	assert.Equal(t, "192.168.1.2/24", r2[1].Subnet)
	assert.NoError(t, err2)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err2 = t0.Clear(ctx, models.White)
	r2, _ = t0.GetAll(context.Background())
	assert.ErrorIs(t, err2, context.Canceled)
	assert.Equal(t, 2, len(r2))

	err2 = t0.Clear(context.Background(), models.White)
	r2, _ = t0.GetAll(context.Background())
	assert.Nil(t, err2)
	assert.Equal(t, 0, len(r2))
}

func TestStorage_ClearAll(t *testing.T) {
	{
		t0 := New()
		err1 := t0.Clear(context.Background(), models.Black)
		assert.NoError(t, err1)

		t1 := &models.IPList{
			Subnet:  "192.168.1.0/24",
			IsWhite: models.White,
		}

		t2 := &models.IPList{
			Subnet:  "192.168.1.1/24",
			IsWhite: models.Black,
		}

		t3 := &models.IPList{
			Subnet:  "192.168.1.2/24",
			IsWhite: models.White,
		}

		t0.ipList = append(t0.ipList, t1)
		t0.ipList = append(t0.ipList, t2)
		t0.ipList = append(t0.ipList, t3)

		err2 := t0.ClearAll(context.Background())
		assert.NoError(t, err2)
		assert.Equal(t, 0, len(t0.ipList))
	}

	{
		t0 := New()
		err1 := t0.Clear(context.Background(), models.Black)
		assert.NoError(t, err1)

		t1 := &models.IPList{
			Subnet:  "192.168.1.0/24",
			IsWhite: models.White,
		}

		t2 := &models.IPList{
			Subnet:  "192.168.1.1/24",
			IsWhite: models.Black,
		}

		t3 := &models.IPList{
			Subnet:  "192.168.1.2/24",
			IsWhite: models.White,
		}

		t0.ipList = append(t0.ipList, t1)
		t0.ipList = append(t0.ipList, t2)
		t0.ipList = append(t0.ipList, t3)

		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err2 := t0.ClearAll(ctx)
		assert.ErrorIs(t, err2, context.Canceled)
		assert.Equal(t, 3, len(t0.ipList))
	}
}
func TestStorage_Contains(t *testing.T) {
	t0 := New()
	t1 := &models.IPList{
		Subnet:  "191.168.1.0/24",
		IsWhite: models.White,
	}

	t2 := &models.IPList{
		Subnet:  "192.168.1.1/24",
		IsWhite: models.Black,
	}

	t3 := &models.IPList{
		Subnet:  "193.168.1.2/24",
		IsWhite: models.White,
	}

	t0.ipList = append(t0.ipList, t1)
	t0.ipList = append(t0.ipList, t2)
	t0.ipList = append(t0.ipList, t3)

	res0, err0 := t0.Contains(context.Background(), models.White, "191.168.1.111")
	assert.True(t, res0)
	assert.NoError(t, err0)

	res1, err1 := t0.Contains(context.Background(), models.Black, "192.168.1.111")
	assert.True(t, res1)
	assert.NoError(t, err1)

	res2, err2 := t0.Contains(context.Background(), models.White, "193.168.1.111")
	assert.True(t, res2)
	assert.NoError(t, err2)

	resF, errF := t0.Contains(context.Background(), models.White, "194.168.1.111")
	assert.False(t, resF)
	assert.NoError(t, errF)

	resI, errI := t0.Contains(context.Background(), models.White, "invalid")
	assert.False(t, resI)
	assert.ErrorIs(t, errI, storage.ErrInvalidAddressDetected)

	res0, err0 = t0.Contains(context.Background(), models.White, "191.168.1.111")
	assert.True(t, res0)
	assert.NoError(t, err0)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res0, err0 = t0.Contains(ctx, models.White, "191.168.1.111")
	assert.False(t, res0)
	assert.ErrorIs(t, err0, context.Canceled)
}

func TestStorage_count(t *testing.T) {
	t0 := New()
	t1 := &models.IPList{
		Subnet:  "191.168.1.0/24",
		IsWhite: models.White,
	}

	t2 := &models.IPList{
		Subnet:  "192.168.1.1/24",
		IsWhite: models.Black,
	}

	t3 := &models.IPList{
		Subnet:  "193.168.1.2/24",
		IsWhite: models.White,
	}

	assert.Equal(t, 0, t0.count(models.White))
	assert.Equal(t, 0, t0.count(models.Black))

	t0.ipList = append(t0.ipList, t1)
	t0.ipList = append(t0.ipList, t2)
	t0.ipList = append(t0.ipList, t3)

	assert.Equal(t, 2, t0.count(models.White))
	assert.Equal(t, 1, t0.count(models.Black))
}

func TestStorage_IsIPAuthorized(t *testing.T) {
	t0 := New()
	t1 := &models.IPList{
		Subnet:  "191.168.1.0/32",
		IsWhite: models.White,
	}

	t2 := &models.IPList{
		Subnet:  "192.168.1.1/32",
		IsWhite: models.Black,
	}

	t3 := &models.IPList{
		Subnet:  "193.168.1.2/32",
		IsWhite: models.White,
	}

	res, err := t0.IsIPAuthorized(context.Background(), "200.200.200.200")
	assert.True(t, res)
	assert.NoError(t, err)

	t0.ipList = append(t0.ipList, t2)

	// in the black list
	res0, err0 := t0.IsIPAuthorized(context.Background(), "192.168.1.1")
	assert.False(t, res0)
	assert.NoError(t, err0)

	res1, err1 := t0.IsIPAuthorized(context.Background(), "200.200.200.200")
	assert.True(t, res1)
	assert.NoError(t, err1)

	t0.ipList = append(t0.ipList, t1)
	t0.ipList = append(t0.ipList, t3)

	res2, err2 := t0.IsIPAuthorized(context.Background(), "192.168.1.1")
	assert.False(t, res2)
	assert.NoError(t, err2)

	res3, err3 := t0.IsIPAuthorized(context.Background(), "200.200.200.200")
	assert.False(t, res3)
	assert.NoError(t, err3)

	res4, err4 := t0.IsIPAuthorized(context.Background(), "193.168.1.2")
	assert.True(t, res4)
	assert.NoError(t, err4)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	res5, err5 := t0.IsIPAuthorized(ctx, "193.168.1.2")
	assert.False(t, res5)
	assert.ErrorIs(t, err5, context.Canceled)
}
