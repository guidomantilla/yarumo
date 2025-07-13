package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/guidomantilla/yarumo/internal/core"
	"github.com/guidomantilla/yarumo/pkg/boot"
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

		//xxx()
		yyy()
		//zzz()
		//parser()

		return nil
	}, options...)
}

func parser() { //nolint:unused
	f, _ := rules.ParseFormula("isAdult THEN isActive")
	fmt.Println("Parsed Formula:", fmt.Sprintf("%+v", f))

	f, _ = rules.ParseFormula("has2FA IFF isAdmin")
	fmt.Println("Parsed Formula:", fmt.Sprintf("%+v", f))

	f1, _ := rules.ParseFormula("isAdmin THEN has2FA")
	f2, _ := rules.ParseFormula("NOT(has2FA) THEN NOT(isAdmin)")
	fmt.Println(propositions.Equivalent(f1, f2)) //

	exp := "(NOT isAdult AND isColombian) OR (isAdmin THEN (has2FA AND isActive)) IFF (TRUE OR (FALSE AND hasEmail))"
	f3, err := rules.ParseFormula(exp)
	fmt.Println(fmt.Sprintf("Parsed Formula: %+v, Error: %+v", f3, err)) //nolint:gosimple

	exp = "((NOT isAdmin OR isActive) AND (hasEmail AND (isColombian IFF isAdult))) THEN ((termsAccepted OR has2FA) AND NOT FALSE)"
	f4, err := rules.ParseFormula(exp)
	fmt.Println(fmt.Sprintf("Parsed Formula: %+v, Error: %+v", f4, err)) //nolint:gosimple
}

func xxx() { //nolint:unused

	fmt.Println()
	fmt.Println()
}

func yyy() { //nolint:unused

	result, err := Predicates.Evaluate(UserRules[0].Formula, User)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error evaluating proposition: %v", err)) //nolint:gosimple
		return
	}

	pretty, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error al formatear JSON:", err)
		return
	}

	fmt.Println(string(pretty))
	fmt.Println()
	fmt.Println()

}

func zzz() { //nolint:unused

	results, _ := rules.EvaluateRules(&User, Predicates, UserRules)
	for _, result := range results {
		fmt.Println(fmt.Sprintf("Formula: %+v", result.Rule.Formula)) //nolint:gosimple
		fmt.Println(fmt.Sprintf("Input: %+v", result.Input))          //nolint:gosimple
		fmt.Println(fmt.Sprintf("Satisfied: %t", result.Satisfied))   //nolint:gosimple
		fmt.Println(fmt.Sprintf("Violated: %t", result.Violated))     //nolint:gosimple
		fmt.Println(fmt.Sprintf("Explanation: %+v", result.EvalTree)) //nolint:gosimple
		fmt.Println()
		fmt.Println()

	}
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
