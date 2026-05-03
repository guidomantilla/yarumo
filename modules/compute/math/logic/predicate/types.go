// Package predicate provides bounded quantifiers (FORALL/EXISTS) over finite
// collections with propositional predicates.
package predicate

import "github.com/guidomantilla/yarumo/compute/math/logic"

// Collection is a finite set of fact assignments to quantify over.
type Collection []logic.Fact
