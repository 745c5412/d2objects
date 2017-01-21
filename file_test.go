package d2objects

import (
	"fmt"
	"os"
	"testing"
)

func openFixture(t *testing.T) *os.File {
	file, err := os.Open("./fixtures/Servers.d2o")
	if err != nil {
		t.Fatalf("open fixture failed: %v", err)
	}
	return file
}

func TestGetObjects(t *testing.T) {
	file := openFixture(t)
	defer func() { _ = file.Close() }()

	r := NewReader(file)
	f, err := ParseFile(r)
	if err != nil {
		t.Errorf("ParseFile: %v", err)
	}

	objs, err := f.GetObjects()
	if err != nil {
		t.Errorf("GetObjects: %v", err)
	}

	fmt.Printf("%#v\n", objs)
}
