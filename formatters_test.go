package main

import "testing"

func TestFormatByteSize(t *testing.T) {
	bytesString := FormatByteSize(1011, 0)
	if bytesString != "1011 b" {
		t.Errorf("FormatByteSize should have returned 1011 b but got %s", bytesString)
	}

	kbString := FormatByteSize(25600, 0)
	if kbString != "25 kiB" {
		t.Errorf("FormatByteSize should have returned 25 kb but got %s", bytesString)
	}

	mbString := FormatByteSize(26214400, 0)
	if mbString != "25 MiB" {
		t.Errorf("FormatByteSize should have returned 25 MiB but got %s", mbString)
	}
}
