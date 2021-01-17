package tgz_test

import (
	"bytes"
	"crypto/sha256"
	"os"
	"path/filepath"
	"testing"

	"github.com/zxdez/tgz"
)

var path = "testdata"
var target = "test.tgz"

func TestBytes(t *testing.T) {

	os.Mkdir(path, 0755)

	b := new(bytes.Buffer)
	b.WriteString("test file\nline1\nline2\n")

	h := sha256.New()
	w, err := os.Create(filepath.Join(path, target))
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if _, err := tgz.Bytes(b, nil, w, h); err != nil {
		t.Fatal(err)
	}

	t.Logf("target: %s\n hex: %x", target, h.Sum(nil))
}

func TestTarDir(t *testing.T) {

	TestBytes(t)
	defer os.Remove(filepath.Join(path, target))

	target := "test-file.tgz"
	defer os.Rename(target, filepath.Join(path, target))

	h := sha256.New()
	w, err := os.Create(target)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	if err := tgz.Tar(path, nil, w, h); err != nil {
		t.Fatal(err)
	}

	t.Logf("target: %s\n hex: %x", path, h.Sum(nil))

}

func TestTarFile(t *testing.T) {

	TestBytes(t)
	defer os.Remove(filepath.Join(path, target))

	target := "test-dir.tgz"
	defer os.Rename(target, filepath.Join(path, target))

	h := sha256.New()
	w, err := os.Create(target)
	if err != nil {
		t.Fatal(err)
	}
	defer w.Close()

	tgz.Tar(filepath.Join(path, target), nil, w, h)

	t.Logf("target: %s\n hex: %x", filepath.Join(path, target), h.Sum(nil))

}
