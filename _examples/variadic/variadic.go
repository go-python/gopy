package variadic

func VariFunc(vargs ...int) int{
	total := 0
	for _, num := range vargs {
		total += num
	}
	return total
}

func NonVariFunc(arg1 int, arg2 []int, arg3 int) int{
	total := arg1
	for _, num := range arg2 {
		total += num
	}
	total += arg3

	return total
}
