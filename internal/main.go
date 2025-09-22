package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/guidomantilla/yarumo/internal/core"
	"github.com/guidomantilla/yarumo/pkg/boot"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
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

		process(Rules, "Colombian", "Adult")

		//xxx()
		//yyy()
		//zzz()
		//parser()

		return nil
	}, options...)
}

func xxx() { //nolint:unused

	result, err := Predicates.Evaluate(UserRules[0].Formula, &User)
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

func yyy() { //nolint:unused
	evaluator := logic.NewRuleSet(Predicates, UserRules)

	result, err := evaluator.Evaluate(&User)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error evaluating rules: %v", err)) //nolint:gosimple
		return
	}
	pretty, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error al formatear JSON:", err)
		return
	}

	fmt.Println(string(pretty))
}

func zzz() { //nolint:unused

	evaluator := logic.NewRuleSet(Predicates, UserInferableRules)

	result, err := evaluator.Evaluate(&UserInferable)
	if err != nil {
		fmt.Println(fmt.Sprintf("Error evaluating rules: %v", err)) //nolint:gosimple
		return
	}
	pretty, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		fmt.Println("Error al formatear JSON:", err)
		return
	}

	fmt.Println(string(pretty))
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
