package service

func ValidQuantity(n int) bool { return n >= 1 && n <= 99 }
func CanTransition(from, to, role string) bool {
	if role == "user" {
		return (from == "pending" && to == "cancelled") || (from == "shipped" && to == "completed")
	}
	if role != "admin" && role != "staff" {
		return false
	}
	return (from == "pending" && (to == "confirmed" || to == "cancelled")) || (from == "confirmed" && (to == "shipped" || to == "cancelled")) || (from == "shipped" && to == "completed")
}
