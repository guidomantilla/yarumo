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

		// (¬(A ∧ B) ⇒ (¬A ∨ ¬B))
		formula := A.And(B).Not().Implies(A.Not().Or(B.Not()))

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
		formula1 := A.Implies(B)
		formula2 := A.Not().Or(B)

		equiv := propositional.Equivalent(formula1, formula2)
		fmt.Printf("¿%s ≡ %s? %v\n", formula1.String(), formula2.String(), equiv)

		fmt.Println()
		formula = A.Implies(B).Not() // ¬(A ⇒ B)
		fmt.Println("Original:", formula.String())

		nnf := propositional.ToNNF(formula)
		fmt.Println("NNF:     ", nnf.String())
		propositional.PrintTruthTable(nnf)

		cnf := propositional.ToCNF(formula)
		fmt.Println("CNF:     ", cnf.String())
		propositional.PrintTruthTable(cnf)

		fmt.Println()
		fmt.Println()

		C := propositional.V("C")

		formula = A.Implies(B.Or(C)) // A ⇒ (B ∨ C)
		fmt.Println("Original:", formula.String())

		cnf = propositional.ToCNF(formula)
		fmt.Println("CNF:     ", cnf.String())

		propositional.PrintTruthTable(cnf)

		fmt.Println()
		fmt.Println()

		fmt.Println()
		fmt.Println()

		formula = propositional.V("P").Implies(propositional.V("Q")).And(propositional.V("P")).And(propositional.V("Q").Not())
		fmt.Println("Formula:", formula.String())
		propositional.ResolutionTrace(formula)

		fmt.Println()
		fmt.Println()

		formula = propositional.V("A").Implies(propositional.V("B")).And(propositional.V("A")).And(propositional.V("B").Not())
		fmt.Println("Formula:", formula.String())
		fmt.Println("¿Contradictoria?", propositional.IsContradiction(formula)) // true
		fmt.Println("¿Satisfactible?", propositional.IsSatisfiable(formula))    // false

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
