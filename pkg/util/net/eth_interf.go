package net

import (
	"fmt"
	"net"
)

func GetEthInterface() (net.Interface, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return net.Interface{}, err
	}

	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp != 0 && iface.Flags&net.FlagLoopback == 0 {
			return iface, nil
		}
	}

	return net.Interface{}, fmt.Errorf("no eth interface found")
}
