package acquirer

var (
	cards3dsRequiredSuccess = []string{
		"4000000000003220",
		"4000000000003063",
	}

	cards3dsRequiredFailure = []string{
		"4000008400001280",
		"4000000000003097",
	}

	cardsNo3dsSuccess = []string{
		"4242424242424242",
		"5555555555554444",
		"4000000000007726",
		"4000000000005126",
	}

	cardsRefunded = []string{
		"4000000000007726",
		"4000000000005126",
	}
)

func is3dSecureRequired(cardNumber string) bool {
	for _, card := range cards3dsRequiredSuccess {
		if card == cardNumber {
			return true
		}
	}
	for _, card := range cards3dsRequiredFailure {
		if card == cardNumber {
			return true
		}
	}
	return false
}

func isSuccess(cardNumber string) bool {
	for _, card := range cards3dsRequiredSuccess {
		if card == cardNumber {
			return true
		}
	}
	for _, card := range cardsNo3dsSuccess {
		if card == cardNumber {
			return true
		}
	}
	return false
}

func shouldRefund(cardNumber string) bool {
	for _, card := range cardsRefunded {
		if card == cardNumber {
			return true
		}
	}
	return false
}
