package api

import (
	"os"
	"testing"

	"github.com/quanttide/qtcloud-connect-provider-example/internal/store"
)

// testSetup 创建基于临时目录的 file store，供 api 测试使用。
// 源实现位于 qtadmin provider 的 human_test.go，此处为自包含而随附复制。
func testSetup(t *testing.T) (store.Store, func()) {
	t.Helper()
	dir, err := os.MkdirTemp("", "api-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	s, err := store.New(store.Config{Driver: "file", Path: dir})
	if err != nil {
		os.RemoveAll(dir)
		t.Fatalf("failed to create store: %v", err)
	}
	return s, func() {
		s.Close()
		os.RemoveAll(dir)
	}
}
