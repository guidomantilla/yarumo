// Package main demonstrates the imperative API of common/validation: the
// caller assembles leaves into the validation flow their handler needs, and
// aggregates the resulting domain error.
package main

import (
	"errors"
	"fmt"

	cerrs "github.com/guidomantilla/yarumo/common/errs"
	extuids "github.com/guidomantilla/yarumo/extension/common/uids"
	cvalidation "github.com/guidomantilla/yarumo/common/validation"
)

// Pokemon is a sample domain type validated by this example.
type Pokemon struct {
	ID    string
	Name  string
	Email string
	Level int
	Tags  []string
}

func main() {
	good := Pokemon{
		ID:    "550e8400-e29b-41d4-a716-446655440000",
		Name:  "Pikachu",
		Email: "ash@kanto.com",
		Level: 35,
		Tags:  []string{"electric", "starter"},
	}

	bad := Pokemon{
		ID:    "not-a-uuid",
		Name:  "",
		Email: "not-an-email",
		Level: 200,
		Tags:  []string{},
	}

	demoLeaves(good)
	demoLeaves(bad)
	demoReflection(good)
}

// demoLeaves validates a Pokemon using the leaf catalogue and aggregates
// every violation into a single domain error.
func demoLeaves(p Pokemon) {
	fmt.Println("=== Validating", p.Name, "===")

	var causes []error

	err := cvalidation.IsUID(p.ID, extuids.IsUUID)
	if err != nil {
		causes = append(causes, err)
	}

	err = cvalidation.IsRequired(p.Name)
	if err != nil {
		causes = append(causes, err)
	}

	err = cvalidation.MaxLen(p.Name, 50)
	if err != nil {
		causes = append(causes, err)
	}

	err = cvalidation.IsEmail(p.Email)
	if err != nil {
		causes = append(causes, err)
	}

	err = cvalidation.InRange(p.Level, 1, 100)
	if err != nil {
		causes = append(causes, err)
	}

	err = cvalidation.NonEmpty(p.Tags)
	if err != nil {
		causes = append(causes, err)
	}

	if len(causes) == 0 {
		fmt.Println("  OK")

		return
	}

	domain := cvalidation.ErrValidation(causes...)
	for _, info := range cerrs.AsErrorInfo(domain) {
		fmt.Printf("  type=%s\n", info.Type)

		for _, msg := range info.Messages {
			fmt.Printf("    - %s\n", msg)
		}
	}
}

// demoReflection shows how the reflection helpers walk into a struct by path,
// the same way modules/validation/ uses them to address fields from config.
func demoReflection(p Pokemon) {
	fmt.Println("=== Reflection ===")

	v, err := cvalidation.GetField(p, "Email")
	if err != nil {
		fmt.Println("  Email lookup failed:", err)
	} else {
		fmt.Printf("  Email = %v\n", v)
	}

	v, err = cvalidation.GetField(p, "Tags[0]")
	if err != nil {
		fmt.Println("  Tags[0] lookup failed:", err)
	} else {
		fmt.Printf("  Tags[0] = %v\n", v)
	}

	_, err = cvalidation.GetField(p, "Missing")
	if errors.Is(err, cvalidation.ErrPathNotFound) {
		fmt.Println("  Missing -> ErrPathNotFound as expected")
	}
}
