package mysql

import (
	"ServerServing/config"
	"testing"
)

func TestInitMySQL(t *testing.T) {
	config.InitConfigWithFile("/Users/purchaser/go/src/ServerServing/config.yml", "dev")
	InitMySQL()
}
