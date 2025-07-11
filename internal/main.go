package main

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/internal/core"
	"github.com/guidomantilla/yarumo/pkg/boot"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositions"
	"github.com/guidomantilla/yarumo/pkg/common/utils"
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

		xxx()
		yyy()

		return nil
	}, options...)
}

func xxx() {

	formula := UserRules[0].Formula
	for key, row := range propositions.Analyze(formula) {
		fmt.Println(fmt.Sprintf("%s: %+v", key, row)) //nolint:gosimple
	}
	fmt.Println()
	fmt.Println()
}

func yyy() {

	formula, predicate, _ := rules.Unwrap(UserRules)
	fmt.Println("Combined Formula:", fmt.Sprintf("%+v", formula))
	fmt.Println("Combined Predicate Result:", predicate(User)) // false

	fmt.Println()
	fmt.Println()

	result, _ := logic.EvaluateProposition(&User, formula, Predicates)
	fmt.Println(result)

	users := utils.FilterBy([]UserType{User}, utils.FilterFn[UserType](predicate))
	fmt.Println(users)

	fmt.Println()
	fmt.Println()

	results, _ := rules.EvaluateRules(&User, Predicates, UserRules)
	fmt.Println(results)
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
