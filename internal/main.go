package main

import (
	"context"
	"fmt"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
	"github.com/guidomantilla/yarumo/pkg/rules"

	"github.com/guidomantilla/yarumo/internal/core"
	"github.com/guidomantilla/yarumo/pkg/boot"
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

		xxx()

		return nil
	}, options...)
}
func xxx() {

	formula := UserRules[0].Formula
	for key, row := range propositions.Analyze(formula) {
		fmt.Println(fmt.Sprintf("%s: %+v", key, row))
	}
	fmt.Println()
	fmt.Println()
}

func yyy() {

	formula, predicate := UserRules[0].Formula, UserRules[0].Predicate
	for _, rule := range UserRules[1:] {
		formula, predicate = formula.And(rule.Formula), predicate.And(rule.Predicate)
	}
	fmt.Println("Combined Formula:", fmt.Sprintf("%+v", formula))
	fmt.Println("Combined Predicate Result:", predicate(User)) // false

	fmt.Println()
	fmt.Println()

	eval := logic.CompileProposition(formula, Predicates)
	fmt.Println(eval(User)) // true

	fmt.Println()
	fmt.Println()

	results := rules.EvaluateRules(Predicates, UserRules, User)
	rules.PrintRuleEvaluation(results)
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
