package main

import (
	"context"
	"fmt"

	"github.com/guidomantilla/yarumo/internal/core"
	"github.com/guidomantilla/yarumo/pkg/boot"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/predicates"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic/propositional"
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

		type Product struct {
			Name     string
			Price    float64
			Category string
			InStock  bool
		}
		// Lista de productos
		products := []Product{
			{"Laptop", 1200.00, "Electronics", true},
			{"Smartphone", 800.00, "Electronics", false},
			{"Headphones", 150.00, "Accessories", true},
			{"Book", 20.00, "Books", true},
			{"Tablet", 300.00, "Electronics", true},
			{"Charger", 25.00, "Accessories", false},
		}

		// Predicados
		isElectronics := func(p Product) bool { return p.Category == "Electronics" }
		isAffordable := func(p Product) bool { return p.Price <= 500 }
		isInStock := func(p Product) bool { return p.InStock }

		// Productos electrónicos en stock
		electronicsInStock := predicates.Filter(products, predicates.And[Product](isElectronics, isInStock))
		fmt.Println("Electrónicos en stock:")
		for _, p := range electronicsInStock {
			fmt.Printf("- %s ($%.2f)\n", p.Name, p.Price)
		}

		// Productos asequibles
		affordableProducts := predicates.Filter(products, isAffordable)
		fmt.Println("\nProductos asequibles:")
		for _, p := range affordableProducts {
			fmt.Printf("- %s ($%.2f)\n", p.Name, p.Price)
		}

		// Productos que no están en stock
		outOfStockProducts := predicates.Filter(products, predicates.Not[Product](isInStock))
		fmt.Println("\nProductos fuera de stock:")
		for _, p := range outOfStockProducts {
			fmt.Printf("- %s ($%.2f)\n", p.Name, p.Price)

		}

		// Productos electrónicos asequibles o en stock
		affordableOrInStock := predicates.Filter(products, predicates.Or[Product](isAffordable, isInStock))
		fmt.Println("\nProductos asequibles o en stock:")
		for _, p := range affordableOrInStock {
			fmt.Printf("- %s ($%.2f)\n", p.Name, p.Price)
		}

		fmt.Println()
		fmt.Println()

		A := propositional.Var("A")
		B := propositional.Var("B")

		left := propositional.NotF{F: propositional.AndF{L: A, R: B}}
		right := propositional.OrF{L: propositional.NotF{F: A}, R: propositional.NotF{F: B}}

		formula := propositional.ImplF{L: left, R: right}

		env := map[string]bool{"A": true, "B": false}
		result := formula.Eval(env)

		fmt.Printf("Fórmula: %s\n", formula.String())
		fmt.Printf("Ambiente: %v => Resultado: %v\n", env, result)

		fmt.Println()
		fmt.Printf("Fórmula: %s\n", formula.String())
		table := propositional.TruthTable(formula)
		for _, row := range table {
			fmt.Println(row)
		}

		fmt.Println()
		// A ⇒ B ≡ ¬A ∨ B
		formula1 := propositional.ImplF{L: A, R: B}
		formula2 := propositional.OrF{L: propositional.NotF{F: A}, R: B}

		equiv := propositional.Equivalent(formula1, formula2)
		fmt.Printf("¿%s ≡ %s? %v\n", formula1.String(), formula2.String(), equiv)

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
