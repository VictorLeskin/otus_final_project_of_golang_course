package models

import (
	"errors"
	"net"
	"strings"
)

// IsValidIP проверяет, является ли строка валидным IP-адресом.
func IsValidIP(ip string) bool {
	return net.ParseIP(ip) != nil
}

// IsValidSubnet проверяет, является ли строка валидной подсетью в CIDR нотации.
func IsValidSubnet(subnet string) bool {
	if subnet == "" {
		return false
	}
	_, _, err := net.ParseCIDR(subnet)
	return err == nil
}

// NormalizeIP приводит IP к нормализованному виду (убирает ведущие нули и т.д.).
func NormalizeIP(ip string) string {
	parsed := net.ParseIP(ip)
	if parsed == nil {
		return ip
	}
	return parsed.String()
}

// IPType определяет тип входной строки.
type IPType int

const (
	IPTypeInvalid IPType = iota
	IPTypeSingle
	IPTypeSubnet
)

// DetectIPType определяет, что за строка пришла.
func DetectIPType(s string) IPType {
	if s == "" {
		return IPTypeInvalid
	}

	// Проверяем, содержит ли "/" (признак CIDR).
	if strings.Contains(s, "/") {
		if IsValidSubnet(s) {
			return IPTypeSubnet
		}
		return IPTypeInvalid
	}

	// Иначе проверяем как обычный IP.
	if IsValidIP(s) {
		return IPTypeSingle
	}

	return IPTypeInvalid
}

// ValidateIPOrSubnet универсальная проверка.
func ValidateIPOrSubnet(s string) error {
	if s == "" {
		return ErrEmptyString
	}

	switch DetectIPType(s) {
	case IPTypeSingle:
		return nil
	case IPTypeSubnet:
		return nil
	default:
		return ErrInvalidIPOrSubnet
	}
}

// Ошибки....
var (
	ErrEmptyString       = errors.New("subnet cannot be empty")
	ErrInvalidIPOrSubnet = errors.New("invalid IP address or subnet")
)
