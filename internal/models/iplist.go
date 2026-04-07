// internal/models/ip_list.go
package models

import (
	"fmt"
	"net"
	"time"
)

// ListType как булев тип: false = black, true = white.
type ListType bool

const (
	Black ListType = false
	White ListType = true
)

// String для отладки (необязательно).
func (lt ListType) String() string {
	if lt {
		return "white"
	}
	return "black"
}

// IPList представляет запись в списке il-адресов.
type IPList struct {
	ID        int64
	Subnet    string   // "192.168.1.0/24" ...
	IsWhite   ListType // true=white, false=black ...
	CreatedAt time.Time
}

// Validate проверяет корректность записи.
func (il *IPList) Validate() error {
	return ValidateIPOrSubnet(il.Subnet)
}

// AreSame проверяет что записи ииеет такие subnet и тип.
func (il *IPList) AreSameS(subnet string, isWhite ListType) bool {
	return (il.IsWhite == isWhite) && (il.Subnet == subnet)
}

// AreSame проверяет что записи совпадают.
func (il *IPList) AreSame(rhv *IPList) bool {
	return il.AreSameS(rhv.Subnet, rhv.IsWhite)
}

func (il *IPList) Contains(ipAddr net.IP) bool {
	_, subnet, _ := net.ParseCIDR(il.Subnet)

	return subnet.Contains(ipAddr)
}

func (il *IPList) ContainsAddress(address string) (bool, error) {
	ipAddr := net.ParseIP(address)
	if ipAddr == nil {
		return false, fmt.Errorf("wrong IP address")
	}

	return il.Contains(ipAddr), nil
}
