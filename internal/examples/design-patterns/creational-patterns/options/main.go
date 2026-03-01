package main

import (
	"fmt"
)

type Customer struct {
	id            string
	email         string
	firstName     *string
	lastName      *string
	loyaltyPoints int
}

type Option = func(c *Customer)

func WithName(firstName, lastName *string) Option {
	return func(c *Customer) {
		c.firstName = firstName
		c.lastName = lastName
	}
}

func WithLoyaltyPoints(loyaltyPoints int) Option {
	return func(c *Customer) {
		c.loyaltyPoints = loyaltyPoints
	}
}

func WithCreditCard() Option {
	return func(c *Customer) {
		//something
	}
}

func PremiumMember() []Option {
	return []Option{WithLoyaltyPoints(10_000), WithCreditCard()}
}

func NewCustomer(id, email string, opts ...Option) *Customer {
	c := &Customer{id: id, email: email, loyaltyPoints: 100}
	for _, opt := range opts {
		opt(c)
	}

	return c
}

func main() {
	id := "6fa49e0a"
	email := "jane@doe.com"
	firstName := "Jane"
	lastName := "Doe"

	c1 := NewCustomer(id, email)
	fmt.Println(c1)

	c2 := NewCustomer(id, email, WithName(&firstName, &lastName), WithLoyaltyPoints(200))
	fmt.Println(c2)

	c3 := NewCustomer(id, email, PremiumMember()...)
	fmt.Println(c3)
}
