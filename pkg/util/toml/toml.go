package toml

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

func LoadTomlFromFile(filePath string, value any) error {
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return err
	}

	dir := filepath.Dir(filePath)
	filename := filepath.Base(filePath)
	name := strings.TrimSuffix(filename, filepath.Ext(filename))

	v := viper.New()
	v.SetConfigName(name)
	v.SetConfigType("toml")
	v.AddConfigPath(dir)

	if err := v.ReadInConfig(); err != nil {
		return err
	}

	return v.Unmarshal(value)
}
