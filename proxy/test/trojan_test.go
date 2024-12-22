package test

import (
	"context"
	"fmt"
	"net"
	"testing"
)

func TestTrojan(t *testing.T) {
	// 自定义 DNS 服务器地址
	dnsServer := "8.8.8.8:53" // Google Public DNS

	// 创建自定义 Resolver
	resolver := &net.Resolver{
		PreferGo: true, // 使用 Go 的 DNS 解析逻辑
		Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
			// 替换为自定义 DNS 服务器
			d := net.Dialer{}
			return d.DialContext(ctx, network, dnsServer)
		},
	}

	// 使用自定义 Resolver 解析域名
	domain := "bing.com"
	ips, err := resolver.LookupIP(context.Background(), "ip", domain)
	if err != nil {
		fmt.Println("Failed to resolve domain:", err)
		return
	}

	// 打印解析结果
	fmt.Printf("IP addresses for %s:\n", domain)
	for _, ip := range ips {
		fmt.Println(ip.String())
	}
}

func TestDns(t *testing.T) {
	ip, err := net.LookupIP("bing.com")
	if err != nil {
		return
	}
	fmt.Println(ip)
}
