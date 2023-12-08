package chord

import "testing"

func TestBetween(t *testing.T) {
	start := ToBigInt("0")
	end := ToBigInt("10")
	elt := ToBigInt("5")
	if !between(start, elt, end, false) {
		t.Error("Expected true, got false")
	}

	elt = ToBigInt("10")
	if between(start, elt, end, false) {
		t.Error("Expected false, got true")
	}

	if !between(start, elt, end, true) {
		t.Error("Expected true, got false")
	}
}

