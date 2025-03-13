package utils

func ReverseMap[Tx comparable, Ty comparable](m map[Tx]Ty) map[Ty]Tx {
	reversed := make(map[Ty]Tx)
	for key, value := range m {
		reversed[value] = key
	}
	return reversed
}
