package conf

import (
	"testing"
)

func TestUnmarshal(t *testing.T) {
	Init("./test.conf")
	type tMap struct {
		TMap map[string]string `json:"t_map:key"`
	}

	m := tMap{}
	err := Unmarshal(&m)
	t.Log(err)
	t.Log(m)
}
