package utils

import (
	"os"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfig_LoadConfig(t *testing.T) {
	cfg := &Config{}
	cfg.LoadConfig()

	assert.NotNil(t, cfg.Local.Locals)
	assert.NotEqual(t, cfg.Local.DefaultLocal, "")
}

func TestSearchInConf(t *testing.T) {
	f1 := "NoSuchFile.go"
	assert.Empty(t, SearchInConf(f1))

	f2 := "Checklist.xlsx"
	assert.NotEmpty(t, SearchInConf(f2))
	//t.Logf("Full path: %s", SearchInConf(f2))
}

func TestGetSplitGoPath(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skip for windows OS.")
	}

	originEnv := os.Getenv("GOPATH")

	// OSX Linux
	os.Setenv("GOPATH", "/Users/Tester/go:/etc/go:/Whatever/gogogo")
	assert.Equal(t, 3, len(GetSplitGoPath()))
	assert.Contains(t, GetSplitGoPath(), "/etc/go")

	/*
		// Windows
		os.Setenv("GOPATH", "c:\\go;d:\\go\\workspace;e:\\;f:\\users\\taihe\\gogogo")
		assert.Equal(t, 4, len(GetSplitGoPath()))
		assert.Contains(t, GetSplitGoPath(), "f:\\users\\taihe\\gogogo", )
	*/

	// OSX single
	os.Setenv("GOPATH", "/Users/Tester/go")
	assert.Equal(t, 1, len(GetSplitGoPath()))
	assert.Equal(t, "/Users/Tester/go", GetSplitGoPath()[0])

	/*
		// Windows single
		os.Setenv("GOPATH", "d:\\go")
		assert.Equal(t, 1, len(GetSplitGoPath()))
		assert.Equal(t, "d:\\go", GetSplitGoPath()[0])
	*/

	os.Setenv("GOPATH", originEnv)
}

func TestGetVCSRootPath(t *testing.T) {
	rootPath := GetVCSRootPath()
	// 确定结尾是root path
	assert.Equal(t, "vcs.taiyouxi.net", rootPath[len(rootPath)-len("vcs.taiyouxi.net"):])
}

func TestGetDataFilePath(t *testing.T) {

}

func TestConfig_LoadAllClientFileName(t *testing.T) {
	if !cfg.Dir.RunOnTeamCity {
		t.Skip("Skip for local running.")
	}
	cfg := &Config{}
	cfg.LoadConfig()

	assert.NotEmpty(t, cfg.AllClientFiles)
	/*
		for k, v := range cfg.AllClientFiles {
			t.Logf(":", k, v)
		}*/
}
