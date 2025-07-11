package main

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/internal/core"
	"github.com/guidomantilla/yarumo/pkg/boot"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
	"github.com/guidomantilla/yarumo/pkg/rules"
	"github.com/guidomantilla/yarumo/pkg/servers"
)

func main() {

	name, version := "yarumo-app", "1.0.0"
	ctx, options := context.Background(), GetOptions()
	boot.Run[core.Config](ctx, name, version, func(ctx context.Context, app servers.Application) error {
		wctx, err := boot.Context[core.Config]()
		if err != nil {
			return fmt.Errorf("error getting context: %w", err)
		}

		fmt.Println("Configuration:", fmt.Sprintf("%+v", wctx.Config))
		fmt.Println()
		fmt.Println()

		type User struct {
			Name          string
			Age           int
			Active        bool
			Email         string
			Country       string
			IsAdmin       bool
			Has2FA        bool
			TermsAccepted bool
		}

		adult := propositions.Var("Adult")
		active := propositions.Var("Active")
		colombian := propositions.Var("Colombian")
		emailValid := propositions.Var("EmailValid")
		has2FA := propositions.Var("Has2FA")
		termsAccepted := propositions.Var("TermsAccepted")
		admin := propositions.Var("Admin")

		predicatex := map[propositions.Var]predicates.Predicate[User]{
			adult:         func(u User) bool { return u.Age >= 18 },
			active:        func(u User) bool { return u.Active },
			colombian:     func(u User) bool { return u.Country == "CO" },
			emailValid:    func(u User) bool { return u.Email != "" },
			has2FA:        func(u User) bool { return u.Has2FA },
			termsAccepted: func(u User) bool { return u.TermsAccepted },
			admin:         func(u User) bool { return u.IsAdmin },
		}

		userRules := []rules.Rule[User]{
			{
				Label:     "R1 - Colombian adults must be active",
				Formula:   colombian.And(adult).Implies(active),
				Predicate: predicatex[colombian].And(predicatex[adult].Implies(predicatex[active])),
			},
			{
				Label:     "R2 - All users must accept terms to be active",
				Formula:   active.Implies(termsAccepted),
				Predicate: predicatex[active].Implies(predicatex[termsAccepted]),
			},
			{
				Label:     "R3 - Admins must have 2FA",
				Formula:   admin.Implies(has2FA),
				Predicate: predicatex[admin].Implies(predicatex[has2FA]),
			},
			{
				Label:     "R4 - All users must have email",
				Formula:   active.Implies(emailValid),
				Predicate: predicatex[active].Implies(predicatex[emailValid]),
			},
		}

		user := User{
			Name:          "Ana",
			Age:           22,
			Active:        true,
			Email:         "",
			Country:       "CO",
			IsAdmin:       true,
			Has2FA:        false,
			TermsAccepted: false,
		}

		formula, predicate := userRules[0].Formula, userRules[0].Predicate
		for _, rule := range userRules[1:] {
			formula, predicate = formula.And(rule.Formula), predicate.And(rule.Predicate)
		}
		fmt.Println("Combined Formula:", fmt.Sprintf("%+v", formula))
		fmt.Println("Combined Predicate Result:", predicate(user)) // false

		fmt.Println()
		fmt.Println()

		eval := logic.CompileProposition(formula, predicatex)
		fmt.Println(eval(user)) // true

		fmt.Println()
		fmt.Println()

		predicatex = map[propositions.Var]predicates.Predicate[User]{
			adult:         func(u User) bool { return u.Age >= 18 },
			active:        func(u User) bool { return u.Active },
			colombian:     func(u User) bool { return u.Country == "CO" },
			emailValid:    func(u User) bool { return u.Email != "" },
			has2FA:        func(u User) bool { return u.Has2FA },
			termsAccepted: func(u User) bool { return u.TermsAccepted },
			admin:         func(u User) bool { return u.IsAdmin },
		}

		userRules = []rules.Rule[User]{
			{
				Label:     "R1 - Colombian adults must be active",
				Formula:   colombian.And(adult).Implies(active),
				Predicate: predicatex[colombian].And(predicatex[adult].Implies(predicatex[active])),
			},
			{
				Label:     "R2 - All users must accept terms to be active",
				Formula:   active.Implies(termsAccepted),
				Predicate: predicatex[active].Implies(predicatex[termsAccepted]),
			},
			{
				Label:     "R3 - Admins must have 2FA",
				Formula:   admin.Implies(has2FA),
				Predicate: predicatex[admin].Implies(predicatex[has2FA]),
			},
			{
				Label:     "R4 - All users must have email",
				Formula:   active.Implies(emailValid),
				Predicate: predicatex[active].Implies(predicatex[emailValid]),
			},
		}

		results := rules.EvaluateRules(predicatex, userRules, user)
		rules.PrintRuleEvaluation(results)

		return nil
	}, options...)
}

/*
	timeoutCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()

		rest := comm.NewRESTClient("https://fakerestapi.azurewebsites.net", comm.WithHTTPClient(wctx.HttpClient))
		resp, err := rest.Call(timeoutCtx, http.MethodGet, "/api/v1/Activities/1", nil)
		if err != nil {
			return fmt.Errorf("error making request: %w", err)
		}

		if pointer.IsSlice(resp.Data) {
			sliceMaps, err := comm.ToSliceOfMapsOfAny(resp.Data)
			if err != nil {
				return fmt.Errorf("error converting response data to map: %w", err)
			}
			fmt.Println(fmt.Sprintf("Response status: %+v", sliceMaps)) //nolint:gosimple
		}
		if pointer.IsMap(resp.Data) {
			maps, err := comm.ToMapOfAny(resp.Data)
			if err != nil {
				return fmt.Errorf("error converting response data to map: %w", err)
			}
			fmt.Println(fmt.Sprintf("Response status: %+v", maps)) //nolint:gosimple
		}
*/
