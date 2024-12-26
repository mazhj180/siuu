package test

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"golang.org/x/net/dns/dnsmessage"
	"net"
	"net/http"
	proxy2 "siu/tunnel/proxy"
	"strconv"
	"testing"
)

func TestTrojan(t *testing.T) {

	data := "GET / HTTP/1.1\nHost: www.bing.com\nUser-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/117.0.0.0 Safari/537.36\nAccept: text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7\nAccept-Encoding: gzip, deflate, br\nAccept-Language: en-US,en;q=0.9\nConnection: keep-alive"

	// 服务器地址和端口
	tp := &proxy2.TrojanProxy{
		Type:     proxy2.TROJAN,
		Name:     "xxxxxxx",
		Server:   "xxxxxx",
		Port:     0000,
		Protocol: proxy2.TCP,
		Password: "xcxasxxx",
		Sni:      "xxxxxx",
	}

	// 配置 TLS
	tlsConfig := &tls.Config{
		ServerName:         tp.Sni, // 设置 SNI
		InsecureSkipVerify: false,  // 验证证书
	}

	// 建立 TLS 连接
	conn, err := tls.Dial("tcp", net.JoinHostPort(tp.Server, strconv.Itoa(int(tp.Port))), tlsConfig)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()
	fmt.Println(data)
	hash := sha256.New224()
	hash.Write([]byte(tp.Password))
	pwd := hex.EncodeToString(hash.Sum(nil))
	fmt.Println(pwd)
	fmt.Println(len(pwd))

	fmt.Println(SHA224String(tp.Password))

}
func SHA224String(password string) string {
	hash := sha256.New224()
	hash.Write([]byte(password))
	val := hash.Sum(nil)
	str := ""
	for _, v := range val {
		str += fmt.Sprintf("%02x", v)
	}
	return str
}

func TestDoh(t *testing.T) {
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

func TestDns2(t *testing.T) {
	// 查询的域名
	domain := "bing.com."

	// 构建 DNS 查询消息
	var msg dnsmessage.Message
	msg.Header.RecursionDesired = true
	msg.Questions = []dnsmessage.Question{
		{
			Name:  dnsmessage.MustNewName(domain),
			Type:  dnsmessage.TypeA, // 查询 A 记录（IPv4 地址）
			Class: dnsmessage.ClassINET,
		},
	}

	// 编码为二进制
	rawQuery, err := msg.Pack()
	if err != nil {
		fmt.Printf("Failed to pack DNS message: %v\n", err)
		return
	}

	// Base64 编码查询
	base64Query := base64.RawURLEncoding.EncodeToString(rawQuery)

	// 构造 DoH 请求 URL
	dohURL := fmt.Sprintf("https://doh.pub/dns-query?dns=%s", base64Query)

	// 发送 HTTPS 请求
	resp, err := http.Get(dohURL)
	if err != nil {
		fmt.Printf("Failed to send DoH request: %v\n", err)
		return
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态
	if resp.StatusCode != http.StatusOK {
		fmt.Printf("DoH server returned status: %s\n", resp.Status)
		return
	}

	// 读取并解析 DNS 响应
	rawResponse := make([]byte, 512)
	n, err := resp.Body.Read(rawResponse)
	if err != nil {
		fmt.Printf("Failed to read response: %v\n", err)
		return
	}
	rawResponse = rawResponse[:n]

	// 打印原始响应数据（调试用）
	fmt.Printf("Raw DNS Response: %x\n", rawResponse)

	// 解码 DNS 响应
	var res dnsmessage.Message
	if err := res.Unpack(rawResponse); err != nil {
		fmt.Printf("Failed to unpack DNS response: %v\n", err)
		return
	}

	// 打印解析结果
	for _, answer := range res.Answers {
		if answer.Header.Type == dnsmessage.TypeA {
			ip := answer.Body.(*dnsmessage.AResource).A
			fmt.Printf("Domain: %s, IP: %v\n", domain, ip)
		}
	}
}

func TestTls(t *testing.T) {
	// 服务器地址和端口
	serverAddr := "xxxxx"
	serverName := "xxxxxx"

	// 配置 TLS
	tlsConfig := &tls.Config{
		ServerName:         serverName, // 设置 SNI
		InsecureSkipVerify: false,      // 验证证书
	}

	// 建立 TLS 连接
	conn, err := tls.Dial("tcp", serverAddr, tlsConfig)
	if err != nil {
		fmt.Printf("Failed to connect: %v\n", err)
		return
	}
	defer conn.Close()

	// 获取连接状态
	state := conn.ConnectionState()

	// 打印协议信息
	fmt.Printf("Protocol: %s\n", state.NegotiatedProtocol)
	fmt.Printf("Cipher Suite: %s\n", tls.CipherSuiteName(state.CipherSuite))
	fmt.Printf("Peer Certificates:\n")

	// 打印证书链
	for i, cert := range state.PeerCertificates {
		fmt.Printf(" Certificate %d:\n", i)
		printCertificate(cert)
	}

	// 检查证书验证结果
	if err := state.VerifiedChains; err != nil {
		fmt.Println("Certificate verified successfully")
	} else {
		fmt.Println("Failed to verify certificate")
	}
}

// 打印证书信息
func printCertificate(cert *x509.Certificate) {
	fmt.Printf("  Subject: %s\n", cert.Subject)
	fmt.Printf("  Issuer: %s\n", cert.Issuer)
	fmt.Printf("  Valid from: %s to %s\n", cert.NotBefore, cert.NotAfter)
	fmt.Printf("  DNS Names: %v\n", cert.DNSNames)

	// 打印公钥类型
	switch cert.PublicKeyAlgorithm {
	case x509.RSA:
		fmt.Println("  Public Key Type: RSA")
	case x509.ECDSA:
		fmt.Println("  Public Key Type: ECDSA")
	default:
		fmt.Println("  Public Key Type: Unknown")
	}

	// 打印原始证书（PEM 格式）
	block := &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: cert.Raw,
	}
	fmt.Printf("  PEM:\n%s\n", string(pem.EncodeToMemory(block)))
}
