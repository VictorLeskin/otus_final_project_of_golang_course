package models

import (
	"net"
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

func TestIPList_AreSame(t *testing.T) {
	lhv := IPList{
		Subnet:  "ABCD",
		IsWhite: Black,
	}

	rhv := IPList{
		Subnet:  "ABCD",
		IsWhite: Black,
	}

	assert.True(t, lhv.AreSame(&rhv))

	lhv1 := lhv
	rhv1 := rhv

	lhv1.IsWhite = White
	assert.False(t, lhv1.AreSame(&rhv1))

	lhv2 := lhv
	rhv2 := rhv

	lhv2.Subnet = "ABCd"
	assert.False(t, lhv2.AreSame(&rhv2))
}

func TestIPList_AreSameS(t *testing.T) {
	lhv := IPList{
		Subnet:  "ABCD",
		IsWhite: Black,
	}

	assert.True(t, lhv.AreSameS("ABCD", Black))
	assert.False(t, lhv.AreSameS("ABCD", White))
	assert.False(t, lhv.AreSameS("ABCd", Black))
}

func TestIPList_Contains(t *testing.T) {
	ipList := &IPList{Subnet: "192.168.1.0/24"}

	assert.True(t, ipList.Contains(net.ParseIP("192.168.1.100")))
	assert.False(t, ipList.Contains(net.ParseIP("192.168.0.100")))
}

func TestIPList_ContainsAddress(t *testing.T) {
	ipList := &IPList{Subnet: "192.168.1.0/24"}

	// a valid IP in a subnet -> true, nil
	ok, err := ipList.ContainsAddress("192.168.1.100")
	assert.True(t, ok)
	assert.NoError(t, err)

	// a valid IP out of a subnet -> false, nil
	ok, err = ipList.ContainsAddress("192.168.0.100")
	assert.False(t, ok)
	assert.NoError(t, err)

	// invalid IP -> false, error
	ok, err = ipList.ContainsAddress("invalid")
	assert.False(t, ok)
	assert.Error(t, err)
}
