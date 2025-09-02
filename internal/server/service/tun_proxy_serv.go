package service

import (
	"siuu/pkg/logger"
	"siuu/pkg/network"
	"siuu/pkg/network/dns"
	"siuu/pkg/network/vnic"
)

func TunTrafficHandler(nic vnic.VNIC, interceptor *dns.TUNDNSInterceptor, tcpManager *network.TCPManager, log *logger.Logger) {

	for {
		pock, err := vnic.ReadCompleteIPPacket(nic)
		if err != nil {
			log.Error("read complete ip packet error: %v", err)
			return
		}

		ipPacket, err := network.ParseIPPacket(pock)
		if err != nil {
			log.Error("parse ip packet error: %v", err)
			return
		}

		if interceptor.IsRunning() {
			if processedPacket, err := interceptor.ProcessPacket(ipPacket); err != dns.ErrNotDNSPacket {
				// only write back when it is dns and has response
				if err == nil && processedPacket != nil && len(processedPacket.Raw) > 0 {
					nic.Write(processedPacket.Raw)
					continue
				}
			}
		}

		if ipPacket.TCPHeader != nil {
			tcpManager.ProcessPacket(ipPacket)
			continue
		}

	}

}
