package pinyin

import (
	"strings"

	"github.com/mozillazg/go-pinyin"
)

// JoinStrings 连接字符串（比 strings.Join() 更快）
func JoinStrings(parts []string) string {
	n := 0
	for _, p := range parts {
		n += len(p)
	}
	b := make([]byte, n)
	pos := 0
	for _, p := range parts {
		pos += copy(b[pos:], p)
	}
	return string(b)
}

// GetPinyin 获取字符串的拼音
func GetPinyin(s string) string {
	a := pinyin.NewArgs()
	pinyinList := pinyin.LazyPinyin(s, a)
	return JoinStrings(pinyinList)
}

// GetInitials 获取字符串的拼音首字母
func GetInitials(s string) string {
	a := pinyin.NewArgs()
	pinyinList := pinyin.LazyPinyin(s, a)
	initials := ""
	for _, py := range pinyinList {
		if len(py) > 0 {
			initials += string(py[0])
		}
	}
	return initials
}

// FuzzyMatch 模糊匹配（支持汉字、拼音、拼音首字母）
func FuzzyMatch(slice []string, keyword string) []string {
	var results []string
	keyword = strings.ToLower(keyword)

	for _, str := range slice {
		pinyinStr := GetPinyin(str)  // 获取完整拼音
		initials := GetInitials(str) // 获取拼音首字母

		// 直接匹配汉字、拼音或拼音首字母
		if strings.Contains(str, keyword) || strings.Contains(pinyinStr, keyword) || strings.Contains(initials, keyword) {
			results = append(results, str)
		}
	}

	return results
}
