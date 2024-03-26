package services

import (
	"net/netip"
)

var _ IPChecker = (*IPCheckService)(nil)

// IPCheckService service is used to check if the IP is in trusted subnet.
type IPCheckService struct {
	prefix netip.Prefix
}

// NewIPChecker creates instance of the IP checker.
func NewIPChecker(prefix netip.Prefix) *IPCheckService {
	return &IPCheckService{
		prefix: prefix,
	}
}

// IsTrusted checks if the IP is in trusted subnet.
// It returns true if IP is in trusted subnet and false otherwise.
func (i *IPCheckService) IsTrusted(ipAddr string) (bool, error) {
	a, err := netip.ParseAddr(ipAddr)
	if err != nil {
		return false, err
	}

	return i.prefix.Contains(a), nil
}
