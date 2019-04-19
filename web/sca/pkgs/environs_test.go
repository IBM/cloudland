package pkgs

import "testing"

func TestLoadEnvirons(t *testing.T) {
	environs := LoadEnvirons("cladmin")
	for k, v := range environs {
		t.Log(k, v)
	}
}
