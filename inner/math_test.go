package inner

import "testing"

func TestSum(t *testing.T) {
	expected := 5
	result := Sum(3, 2)

	if result != expected {
		t.Error("Expected", expected, "Got", result)
	}
}
