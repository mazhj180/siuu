package test

import (
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"net"
	"siuu/util"
	"testing"
)

func TestXdb(t *testing.T) {
	file, _ := xdb.LoadContentFromFile(util.ProjectRootPath + "/conf/ip2region.xdb")
	s, _ := xdb.NewWithBuffer(file)
	str, err := s.SearchByStr("204.79.197.200")
	if err != nil {
		return
	}
	fmt.Println(str)
}

func TestXdbVector(t *testing.T) {
	path := util.ProjectRootPath + "/conf/ip2region.xdb"
	file, _ := xdb.LoadVectorIndexFromFile(path)
	s, _ := xdb.NewWithVectorIndex(path, file)
	str, err := s.SearchByStr("204.79.197.200")
	if err != nil {
		return
	}
	fmt.Println(str)
}

func TestLookupIp(t *testing.T) {
	domain := "www.cn.bing.com"
	ips, err := net.LookupIP(domain)
	if err != nil {
		t.Fatalf("parse fail %s: %s\n", domain, err)
		return
	}
	for _, ip := range ips {
		fmt.Println("ip:", ip.String())
	}
}

func BenchmarkXdbVector(b *testing.B) {
	path := util.ProjectRootPath + "/conf/ip2region.xdb"
	file, _ := xdb.LoadVectorIndexFromFile(path)
	s, _ := xdb.NewWithVectorIndex(path, file)
	str, err := s.SearchByStr("204.79.197.200")
	if err != nil {
		return
	}
	fmt.Println(str)
}
