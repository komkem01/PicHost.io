package utils

import "strings"

func NextAlphaCode(last string) string {
	if last == "" {
		return "A"
	}
	// ถ้า last เป็นชุดเดียวกัน เช่น "A", "B", ..., "Z", "AA", "BB", ..., "ZZ", "AAA", ...
	// ให้เพิ่มความยาวอีก 1 ถ้า last เป็น Z, ZZ, ZZZ, ...
	upper := strings.ToUpper(last)
	n := len(upper)
	for i := 0; i < n; i++ {
		if upper[i] != 'Z' {
			break
		}
		if i == n-1 {
			// ถ้าเป็น Z, ZZ, ZZZ, ... ให้เพิ่มความยาวอีก 1
			return strings.Repeat("A", n+1)
		}
	}
	// ถ้าไม่ใช่ Z, ZZ, ... ให้เพิ่มตัวอักษรตัวแรก เช่น A->B, B->C, ..., Y->Z, AA->BB, BB->CC
	first := upper[0]
	next := first + 1
	return strings.Repeat(string(next), n)
}
