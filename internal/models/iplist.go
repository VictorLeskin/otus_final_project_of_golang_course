// internal/models/ip_list.go
package models

import (
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
	ID        int64
	Subnet    string   // "192.168.1.0/24"
	IsWhite   ListType // true=white, false=black
	CreatedAt time.Time
}

// Validate проверяет корректность записи
func (il *IPList) Validate() error {
	return ValidateIPOrSubnet(il.Subnet)
}
