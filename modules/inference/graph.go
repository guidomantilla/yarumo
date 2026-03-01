package inference

import (
	_ "github.com/guidomantilla/yarumo/inference/bayesian/engine"
	_ "github.com/guidomantilla/yarumo/inference/bayesian/evidence"
	_ "github.com/guidomantilla/yarumo/inference/bayesian/explain"
	_ "github.com/guidomantilla/yarumo/inference/bayesian/network"
	_ "github.com/guidomantilla/yarumo/inference/classical/engine"
	_ "github.com/guidomantilla/yarumo/inference/classical/explain"
	_ "github.com/guidomantilla/yarumo/inference/classical/facts"
	_ "github.com/guidomantilla/yarumo/inference/classical/rules"
	_ "github.com/guidomantilla/yarumo/inference/fuzzy/engine"
	_ "github.com/guidomantilla/yarumo/inference/fuzzy/explain"
	_ "github.com/guidomantilla/yarumo/inference/fuzzy/rules"
	_ "github.com/guidomantilla/yarumo/inference/fuzzy/variable"
)
