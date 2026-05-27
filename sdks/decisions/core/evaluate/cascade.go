package evaluate

import (
	"context"
	"fmt"
	"strings"

	cassert "github.com/guidomantilla/yarumo/core/common/assert"
	cerrs "github.com/guidomantilla/yarumo/core/common/errs"

	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// CascadeStage defines one stage in a cascade pipeline.
type CascadeStage struct {
	// Name identifies this stage.
	Name string
	// Paradigm selects the reasoning paradigm for this stage.
	Paradigm Paradigm
	// RuleSet is the pre-loaded ruleset for this stage.
	RuleSet *schema.RuleSet
	// Query is the target variable for Bayesian stages (ignored for other paradigms).
	Query string
}

// CascadePipeline defines the interface for chaining paradigm stages sequentially.
type CascadePipeline interface {
	// Execute runs the cascade pipeline with the given initial input.
	Execute(ctx context.Context, initialInput any) (CascadeResult, error)
}

type cascadePipeline struct {
	stages     []CascadeStage
	converters []StageConverter
	explainers explainerSet
}

// NewCascadePipeline creates a new cascade pipeline.
// Converters must have length len(stages) - 1.
func NewCascadePipeline(stages []CascadeStage, converters []StageConverter, opts ...Option) CascadePipeline {
	cassert.NotEmpty(stages, "stages are empty")
	cassert.True(len(converters) == len(stages)-1, "converters count must be len(stages)-1")

	options := NewOptions(opts...)

	return &cascadePipeline{
		stages:     stages,
		converters: converters,
		explainers: options.explainers(),
	}
}

// Execute runs the cascade pipeline with the given initial input.
// The initialInput type must match the first stage paradigm:
// logic.Fact for Deductive, evidence.EvidenceBase for Bayesian, map[string]float64 for Fuzzy.
func (p *cascadePipeline) Execute(ctx context.Context, initialInput any) (CascadeResult, error) {
	cassert.NotNil(p, "pipeline is nil")

	cascadeResult := CascadeResult{
		Stages: make([]StageResult, 0, len(p.stages)),
	}

	currentInput := initialInput

	for i, stage := range p.stages {
		result, err := p.executeStage(ctx, stage, currentInput)
		if err != nil {
			return CascadeResult{}, ErrCascade(cerrs.Wrap(ErrCascadeFailed, err))
		}

		cascadeResult.Stages = append(cascadeResult.Stages, StageResult{
			Name:   stage.Name,
			Result: result,
		})

		isLastStage := i == len(p.stages)-1
		if isLastStage {
			cascadeResult.Final = result

			break
		}

		currentInput, err = p.converters[i](result)
		if err != nil {
			return CascadeResult{}, ErrCascade(cerrs.Wrap(ErrCascadeFailed, err))
		}
	}

	explanations := make([]string, 0, len(cascadeResult.Stages))

	for _, sr := range cascadeResult.Stages {
		explanations = append(explanations, fmt.Sprintf("[%s] %s", sr.Name, sr.Result.Explanation))
	}

	cascadeResult.Explanation = strings.Join(explanations, " → ")

	return cascadeResult, nil
}

func (p *cascadePipeline) executeStage(ctx context.Context, stage CascadeStage, input any) (Result, error) {
	return dispatchParadigm(ctx, stage.Paradigm, stage.RuleSet, input, stage.Query, p.explainers)
}
