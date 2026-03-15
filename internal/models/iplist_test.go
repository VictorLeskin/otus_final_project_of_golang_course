package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestListType_String(t *testing.T) {
	{
		t0 := Black
		assert.Equal(t, "black", t0.String())
	}
	{
		t0 := White
		assert.Equal(t, "white", t0.String())
	}
}

func TestIPList_Validate(t *testing.T) {
	{
		t0 := IPList{
			Subnet: "",
		}
		assert.Equal(t, "subnet cannot be empty", t0.Validate().Error())
	}

	{
		t0 := IPList{
			Subnet: "195.208.65.151/25",
		}
		assert.Nil(t, t0.Validate())
	}

	{
		t0 := IPList{
			Subnet: "195.208.65.151",
		}
		assert.Nil(t, t0.Validate())
	}

	{
		t0 := IPList{
			Subnet: ".208.65.151",
		}
		assert.Equal(t, "invalid IP address or subnet", t0.Validate().Error())
	}
}
