package main

import (
	"fmt"
)

type Card struct {
	amount   float64
	currency string
	address  string
}

func NewCard() *Card {
	return &Card{}
}

func (c *Card) Charge(amount float64) *Card {
	c.amount = amount
	return c
}

func (c *Card) WithCurrency(currency string) *Card {
	c.currency = currency
	return c
}

func (c *Card) WithAddress(address string) *Card {
	c.address = address
	return c
}

func (c *Card) Execute() *Card {
	fmt.Printf("Dear Customer, \n%s %v is Debited from your account \n", c.currency, c.amount)
	return c
}
