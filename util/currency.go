package util

const (
	ARS = "ARS"
	USD = "USD"
	EUR = "EUR"
)

func IsCurrencySuported(currency string) bool {
	switch currency {
	case ARS, USD, EUR:
		return true
	}

	return false
}
