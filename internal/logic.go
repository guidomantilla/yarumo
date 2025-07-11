package main

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
	"github.com/guidomantilla/yarumo/pkg/rules"
)

type UserType struct {
	Name          string
	Age           int
	Active        bool
	Email         string
	Country       string
	IsAdmin       bool
	Has2FA        bool
	TermsAccepted bool
}

var (
	User          = UserType{Name: "Ana", Age: 22, Active: true, Email: "", Country: "CO", IsAdmin: true, Has2FA: false, TermsAccepted: false}
	Adult         = propositions.Var("Adult")
	Active        = propositions.Var("Active")
	Colombian     = propositions.Var("Colombian")
	EmailValid    = propositions.Var("EmailValid")
	Has2FA        = propositions.Var("Has2FA")
	TermsAccepted = propositions.Var("TermsAccepted")
	Admin         = propositions.Var("Admin")
	Predicates    = map[propositions.Var]predicates.Predicate[UserType]{
		Adult:         func(u UserType) bool { return u.Age >= 18 },
		Active:        func(u UserType) bool { return u.Active },
		Colombian:     func(u UserType) bool { return u.Country == "CO" },
		EmailValid:    func(u UserType) bool { return u.Email != "" },
		Has2FA:        func(u UserType) bool { return u.Has2FA },
		TermsAccepted: func(u UserType) bool { return u.TermsAccepted },
		Admin:         func(u UserType) bool { return u.IsAdmin },
	}
	UserRules = []rules.Rule[UserType]{
		{
			Label:     "R1 - Colombian adults must be active",
			Formula:   Colombian.And(Adult).Implies(Active),
			Predicate: Predicates[Colombian].And(Predicates[Adult].Implies(Predicates[Active])),
		},
		{
			Label:     "R2 - All users must accept terms to be active",
			Formula:   Active.Implies(TermsAccepted),
			Predicate: Predicates[Active].Implies(Predicates[TermsAccepted]),
		},
		{
			Label:     "R3 - Admins must have 2FA",
			Formula:   Admin.Implies(Has2FA),
			Predicate: Predicates[Admin].Implies(Predicates[Has2FA]),
		},
		{
			Label:     "R4 - All users must have email",
			Formula:   Active.Implies(EmailValid),
			Predicate: Predicates[Active].Implies(Predicates[EmailValid]),
		},
	}
)
