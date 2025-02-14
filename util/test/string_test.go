package test

import (
	"fmt"
	"math/rand"
	"siuu/util"
	"testing"
)

func TestFuzzy(t *testing.T) {
	data := []string{"香港", "上海", "北京", "广州", "深圳"}
	testCases := []string{"xia", "x", "xiang", "shang", "bei", "gz", "bj", "北"}

	for _, keyword := range testCases {
		matches := util.FuzzyMatch(data, keyword)
		fmt.Printf("搜索 \"%s\": %v\n", keyword, matches)
	}
}

// 生成 10 万条数据
func generateLargeDataset(n int) []string {
	data := []string{"北京", "上海", "广州", "深圳", "杭州", "武汉", "成都", "南京", "重庆", "西安"}
	result := make([]string, n)
	for i := 0; i < n; i++ {
		result[i] = data[rand.Intn(len(data))]
	}
	return result
}

// 测试 fuzzyMatch 在大数据集上的性能
func BenchmarkFuzzyMatchLargeDataset(b *testing.B) {
	data := generateLargeDataset(100000)
	keyword := "shang" // 匹配 "上海"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = util.FuzzyMatch(data, keyword)
	}
}
