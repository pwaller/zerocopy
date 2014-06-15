// This work is subject to the CC0 1.0 Universal (CC0 1.0) Public Domain Dedication
// license. Its contents can be found at:
// http://creativecommons.org/publicdomain/zero/1.0/

package zerocopy

import (
	"bytes"
	"log"
	"os"
	"testing"
)

func TestZeroCopyBuffer(t *testing.T) {
	buf := []byte("Hello, world")
	r := bytes.NewReader(buf)
	zr, err := NewReader(r)
	if err != nil {
		t.Fatal(err)
	}
	result, err := zr.Read(6)
	if string(result) != "Hello," {
		t.Fatalf("string(result) != \"Hello,\": %q", string(result))
	}
	result, err = zr.Read(6)
	if string(result) != " world" {
		t.Fatalf("string(result) != \" world\": %q", string(result))
	}
}

// Try reading 100 bytes of /etc/passwd
func TestZeroCopyFile(t *testing.T) {
	fd, err := os.Open("/etc/passwd")
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	zr, err := NewReader(fd)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()

	const N = 100

	result, err := zr.Read(N)
	if err != nil {
		log.Fatalln("err != nil: ", err)
	}
	if len(result) != N {
		log.Fatalln("len(result) !=", N, "it's", len(result))
	}
}
