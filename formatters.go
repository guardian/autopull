package main

import "fmt"

type ByteSizePrefix string

var ByteSizePrefixes = [...]ByteSizePrefix{"b", "kiB", "MiB", "GiB", "TiB"}

/**
reduce the given byte count to the most appropriate multiplier
*/
func FormatByteSize(number int64, divisor int) string {
	if divisor > len(ByteSizePrefixes) {
		return fmt.Sprintf("%d %s", number, ByteSizePrefixes[divisor-1])
	}

	if number < 1024 {
		return fmt.Sprintf("%d %s", number, ByteSizePrefixes[divisor])
	}
	return FormatByteSize(number/1024, divisor+1)
}
