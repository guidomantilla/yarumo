package main

import (
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
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
	Adult         = propositions.Var("Adult")
	Active        = propositions.Var("Active")
	Colombian     = propositions.Var("Colombian")
	EmailValid    = propositions.Var("EmailValid")
	Has2FA        = propositions.Var("Has2FA")
	TermsAccepted = propositions.Var("TermsAccepted")
	Admin         = propositions.Var("Admin")
	CanLogin      = propositions.Var("CanLogin")
	Predicates    = logic.PredicatesRegistry[UserType]{
		Adult:         func(u UserType) bool { return u.Age >= 18 },
		Active:        func(u UserType) bool { return u.Active },
		Colombian:     func(u UserType) bool { return u.Country == "CO" },
		EmailValid:    func(u UserType) bool { return u.Email != "" },
		Has2FA:        func(u UserType) bool { return u.Has2FA },
		TermsAccepted: func(u UserType) bool { return u.TermsAccepted },
		Admin:         func(u UserType) bool { return u.IsAdmin },
		CanLogin:      func(u UserType) bool { return false },
	}
	UserRules = []logic.Rule[UserType]{
		{
			Label:   "R1 - Colombian adults must be active",
			Formula: Colombian.And(Adult).Implies(Active),
		},
		{
			Label:   "R2 - All users must accept terms to be active",
			Formula: Admin.Implies(TermsAccepted),
		},
		{
			Label:   "R3 - Admins must have 2FA",
			Formula: Admin.Implies(Has2FA),
		},
		{
			Label:   "R4 - All users must have email",
			Formula: Active.Implies(EmailValid),
		},
	}
	UserInferableRules = []logic.Rule[UserType]{
		{
			Label:       "R1 - Colombian adults will be active",
			Formula:     Colombian.And(Adult),
			Consequence: &Active,
		},
		{
			Label:       "R2 - All active admins must accept terms to be active",
			Formula:     Active.And(Admin.Implies(TermsAccepted)),
			Consequence: &TermsAccepted,
		},
		{
			Label:       "R3 - All active admins must have 2FA",
			Formula:     Active.And(Admin.Implies(Has2FA)),
			Consequence: &Has2FA,
		},
		{
			Label:       "R4 - All active users must have email, then they can login",
			Formula:     Active.And(EmailValid.Implies(CanLogin)),
			Consequence: &CanLogin,
		},
	}
)
