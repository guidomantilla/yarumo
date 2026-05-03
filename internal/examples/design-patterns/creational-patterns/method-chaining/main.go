package main

func main() {

	card := NewCard().Charge(10.4).
		WithCurrency("USD").
		WithAddress("Gulshan Karachi").
		Execute()

	// you can write it as below as well in a single line
	card.Charge(10.4).WithCurrency("PKR").WithAddress("Gulshan Karachi").Execute()
}
