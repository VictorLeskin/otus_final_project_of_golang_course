// internal/models/ip_list.go
package models

import (
	"fmt"
	"net"
	"time"
)

// ListType как булев тип: false = black, true = white
type ListType bool

const (
	Black ListType = false
	White ListType = true
)

// String для отладки (необязательно)
func (lt ListType) String() string {
	if lt {
		return "white"
	}
	return "black"
}

// IPList представляет запись в списке IP-адресов
type IPList struct {
	ID        string
	Subnet    string   // "192.168.1.0/24"
	IsWhite   ListType // true=white, false=black
	CreatedAt time.Time
}

// Validate проверяет корректность записи
func (il *IPList) Validate() error {
	return ValidateIPOrSubnet(il.Subnet)
}

// AreSame проверяет что записи ииеет такие subnet и тип
func (lhv *IPList) AreSameS(subnet string, isWhite ListType) bool {
	return (lhv.IsWhite == isWhite) && (lhv.Subnet == subnet)
}

// AreSame проверяет что записи совпадают
func (lhv *IPList) AreSame(rhv *IPList) bool {
	return lhv.AreSameS(rhv.Subnet, rhv.IsWhite)
}

func (ip *IPList) Contains(ipAddr net.IP) bool {
	_, subnet, _ := net.ParseCIDR(ip.Subnet)

	return subnet.Contains(ipAddr)
}

func (ip *IPList) ContainsAddress(address string) (bool, error) {
	ipAddr := net.ParseIP(address)
	if ipAddr == nil {
		return false, fmt.Errorf("wrong ip address")
	}

	return ip.Contains(ipAddr), nil
}
