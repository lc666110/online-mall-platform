package service

import "testing"

func TestOrderTransitions(t *testing.T) {
	cases := []struct {
		from, to, role string
		ok             bool
	}{
		{"pending", "confirmed", "admin", true}, {"confirmed", "shipped", "staff", true},
		{"shipped", "completed", "user", true}, {"pending", "cancelled", "user", true},
		{"completed", "cancelled", "admin", false}, {"pending", "shipped", "admin", false},
		{"pending", "confirmed", "user", false}, {"shipped", "cancelled", "user", false},
	}
	for _, c := range cases {
		if CanTransition(c.from, c.to, c.role) != c.ok {
			t.Errorf("%+v", c)
		}
	}
}
func TestCheckoutValidation(t *testing.T) {
	for _, n := range []int{-1, 0, 1000} {
		if ValidQuantity(n) {
			t.Fatalf("accepted %d", n)
		}
	}
	if !ValidQuantity(1) || !ValidQuantity(99) {
		t.Fatal("valid quantities rejected")
	}
}
