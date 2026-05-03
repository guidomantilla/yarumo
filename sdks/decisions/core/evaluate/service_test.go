package evaluate

import (
	"context"
	"errors"
	"testing"

	cexpressions "github.com/guidomantilla/yarumo/common/expressions"
	"github.com/guidomantilla/yarumo/compute/engine/bayesian/evidence"
	"github.com/guidomantilla/yarumo/compute/math/logic"

	"github.com/guidomantilla/yarumo/decisions/core/explain"
	"github.com/guidomantilla/yarumo/decisions/core/repository"
	"github.com/guidomantilla/yarumo/decisions/core/schema"
)

// --- test doubles ---

type testDomain struct {
	Value  string
	Amount float64
}

type testBinder struct{}

func (b testBinder) BindDeductive(_ testDomain) logic.Fact {
	return logic.Fact{"a": true, "b": true}
}

func (b testBinder) BindBayesian(_ testDomain) evidence.EvidenceBase {
	eb := evidence.NewEvidenceBase()
	eb.Observe("rain", "yes")

	return eb
}

func (b testBinder) BindFuzzy(_ testDomain) map[string]float64 {
	return map[string]float64{"temp": 75}
}

func (b testBinder) BindExpression(d testDomain) cexpressions.Context {
	return cexpressions.Context{
		"value":  d.Value,
		"amount": d.Amount,
	}
}

type testRepo struct {
	ruleSet *schema.RuleSet
	err     error
}

func (r *testRepo) Get(_ context.Context, _ string, _ string) (*schema.RuleSet, error) {
	return r.ruleSet, r.err
}

func (r *testRepo) List(_ context.Context) ([]schema.RuleSet, error) {
	return nil, nil
}

func (r *testRepo) Save(_ context.Context, _ *schema.RuleSet) error {
	return nil
}

func (r *testRepo) Delete(_ context.Context, _ string, _ string) error {
	return nil
}

type testAuditLog struct {
	entries []Entry
	err     error
}

func (a *testAuditLog) Record(_ context.Context, entry Entry) error {
	a.entries = append(a.entries, entry)

	return a.err
}

type failingExplainer struct {
	err error
}

func (e *failingExplainer) ExplainDeductive(_ context.Context, _ explain.DeductiveTrace) (string, error) {
	return "", e.err
}

func (e *failingExplainer) ExplainBayesian(_ context.Context, _ explain.BayesianTrace) (string, error) {
	return "", e.err
}

func (e *failingExplainer) ExplainFuzzy(_ context.Context, _ explain.FuzzyTrace) (string, error) {
	return "", e.err
}

func (e *failingExplainer) ExplainTable(_ context.Context, _ explain.TableTrace) (string, error) {
	return "", e.err
}

func (e *failingExplainer) ExplainScorecard(_ context.Context, _ explain.ScoreTrace) (string, error) {
	return "", e.err
}

func (e *failingExplainer) ExplainTree(_ context.Context, _ explain.TreeTrace) (string, error) {
	return "", e.err
}

// Verify interface compliance.
var _ explain.Explainer = (*failingExplainer)(nil)

// --- tests ---

func TestService_Execute_Deductive(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name:    "test",
				Version: "1",
				Deductive: &schema.DeductiveConfig{
					Rules: []schema.DeductiveRuleDef{
						{
							Name:       "r1",
							Condition:  "a and b",
							Conclusion: map[string]bool{"c": true},
						},
					},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo)
		result, err := svc.Execute(context.Background(), Request[testDomain]{
			Domain:         testDomain{Value: "test"},
			RuleSetName:    "test",
			RuleSetVersion: "1",
			Paradigm:       Deductive,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Deductive {
			t.Fatalf("expected deductive, got %s", result.Paradigm)
		}

		cVal, ok := result.Outcome.Facts["c"]
		if !ok || !cVal {
			t.Fatal("expected c=true in outcome")
		}

		if result.Explanation == "" {
			t.Fatal("expected non-empty explanation")
		}
	})

	t.Run("repo error", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{err: errors.New("not found")}
		svc := NewService[testDomain](testBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Deductive,
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrExecuteFailed) {
			t.Fatal("expected ErrExecuteFailed")
		}
	})

	t.Run("missing deductive config", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{Name: "test", Version: "1"},
		}
		svc := NewService[testDomain](testBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Deductive,
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatal("expected ErrMissingConfig")
		}
	})

	t.Run("unsupported paradigm", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{ruleSet: &schema.RuleSet{Name: "test"}}
		svc := NewService[testDomain](testBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Paradigm(99),
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrUnsupported) {
			t.Fatal("expected ErrUnsupported")
		}
	})
}

func TestService_Execute_Bayesian(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name:    "bayes-test",
				Version: "1",
				Bayesian: &schema.BayesianConfig{
					Nodes: []schema.BayesianNodeDef{
						{
							Variable: "rain",
							Outcomes: []string{"yes", "no"},
							CPT: []schema.CPTRow{
								{Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}},
							},
						},
						{
							Variable: "wet",
							Parents:  []string{"rain"},
							Outcomes: []string{"yes", "no"},
							CPT: []schema.CPTRow{
								{Given: map[string]string{"rain": "yes"}, Probabilities: map[string]float64{"yes": 0.9, "no": 0.1}},
								{Given: map[string]string{"rain": "no"}, Probabilities: map[string]float64{"yes": 0.2, "no": 0.8}},
							},
						},
					},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo)
		result, err := svc.Execute(context.Background(), Request[testDomain]{
			Domain:         testDomain{Value: "test"},
			RuleSetName:    "bayes-test",
			RuleSetVersion: "1",
			Paradigm:       Bayesian,
			Query:          "wet",
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Bayesian {
			t.Fatalf("expected bayesian, got %s", result.Paradigm)
		}

		if len(result.Outcome.Distribution) == 0 {
			t.Fatal("expected non-empty distribution")
		}
	})

	t.Run("missing bayesian config", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{ruleSet: &schema.RuleSet{Name: "test"}}
		svc := NewService[testDomain](testBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Bayesian,
			Query:    "x",
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatal("expected ErrMissingConfig")
		}
	})
}

func TestService_Execute_Fuzzy(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name:    "fuzzy-test",
				Version: "1",
				Fuzzy: &schema.FuzzyConfig{
					InputVars: []schema.FuzzyVarDef{
						{
							Name: "temp",
							Min:  0,
							Max:  100,
							Terms: []schema.FuzzyTermDef{
								{Name: "cold", Type: "triangular", Params: []float64{0, 0, 50}},
								{Name: "hot", Type: "triangular", Params: []float64{50, 100, 100}},
							},
						},
					},
					OutputVars: []schema.FuzzyVarDef{
						{
							Name: "speed",
							Min:  0,
							Max:  100,
							Terms: []schema.FuzzyTermDef{
								{Name: "slow", Type: "triangular", Params: []float64{0, 0, 50}},
								{Name: "fast", Type: "triangular", Params: []float64{50, 100, 100}},
							},
						},
					},
					Rules: []schema.FuzzyRuleDef{
						{
							Name:       "r1",
							Conditions: []schema.FuzzyConditionDef{{Variable: "temp", Term: "hot"}},
							Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "fast"},
						},
						{
							Name:       "r2",
							Conditions: []schema.FuzzyConditionDef{{Variable: "temp", Term: "cold"}},
							Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "slow"},
						},
					},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo)
		result, err := svc.Execute(context.Background(), Request[testDomain]{
			Domain:         testDomain{Value: "test"},
			RuleSetName:    "fuzzy-test",
			RuleSetVersion: "1",
			Paradigm:       Fuzzy,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Fuzzy {
			t.Fatalf("expected fuzzy, got %s", result.Paradigm)
		}

		if len(result.Outcome.Outputs) == 0 {
			t.Fatal("expected non-empty outputs")
		}
	})

	t.Run("missing fuzzy config", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{ruleSet: &schema.RuleSet{Name: "test"}}
		svc := NewService[testDomain](testBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Fuzzy,
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatal("expected ErrMissingConfig")
		}
	})
}

func TestService_Execute_Table(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name:    "test",
				Version: "1",
				Table: &schema.TableConfig{
					Rules: []schema.TableRuleDef{
						{Name: "r1", Conditions: []string{"amount > 100"}, Outputs: map[string]any{"approved": true}},
					},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo)
		result, err := svc.Execute(context.Background(), Request[testDomain]{
			Domain:         testDomain{Amount: 200},
			RuleSetName:    "test",
			RuleSetVersion: "1",
			Paradigm:       Table,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Table {
			t.Fatalf("expected Table paradigm, got %s", result.Paradigm)
		}

		if result.Outcome.Table == nil {
			t.Fatal("expected non-nil table outcome")
		}

		if result.Explanation == "" {
			t.Fatal("expected non-empty explanation")
		}
	})

	t.Run("missing table config", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{Name: "test", Version: "1"},
		}
		svc := NewService[testDomain](testBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Table,
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatal("expected ErrMissingConfig")
		}
	})
}

func TestService_Execute_Scorecard(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name:    "test",
				Version: "1",
				Scorecard: &schema.ScorecardConfig{
					BaseScore: 100,
					Attributes: []schema.ScorecardAttributeDef{
						{
							Name:   "amount",
							Weight: 1.0,
							Bins: []schema.ScorecardBinDef{
								{Condition: "amount > 100", Points: 50},
							},
						},
					},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo)
		result, err := svc.Execute(context.Background(), Request[testDomain]{
			Domain:         testDomain{Amount: 200},
			RuleSetName:    "test",
			RuleSetVersion: "1",
			Paradigm:       Scorecard,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Scorecard {
			t.Fatalf("expected Scorecard paradigm, got %s", result.Paradigm)
		}

		if result.Outcome.Score == nil {
			t.Fatal("expected non-nil score outcome")
		}
	})

	t.Run("missing scorecard config", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{Name: "test", Version: "1"},
		}
		svc := NewService[testDomain](testBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Scorecard,
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatal("expected ErrMissingConfig")
		}
	})
}

func TestService_Execute_Tree(t *testing.T) {
	t.Parallel()

	t.Run("success", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name:    "test",
				Version: "1",
				Tree: &schema.TreeConfig{
					Root: schema.TreeNodeDef{
						Condition: "amount > 100",
						True:      &schema.TreeNodeDef{Output: map[string]any{"risk": "low"}},
						False:     &schema.TreeNodeDef{Output: map[string]any{"risk": "high"}},
					},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo)
		result, err := svc.Execute(context.Background(), Request[testDomain]{
			Domain:         testDomain{Amount: 200},
			RuleSetName:    "test",
			RuleSetVersion: "1",
			Paradigm:       Tree,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result.Paradigm != Tree {
			t.Fatalf("expected Tree paradigm, got %s", result.Paradigm)
		}

		if result.Outcome.Tree == nil {
			t.Fatal("expected non-nil tree outcome")
		}
	})

	t.Run("missing tree config", func(t *testing.T) {
		t.Parallel()

		repo := &testRepo{
			ruleSet: &schema.RuleSet{Name: "test", Version: "1"},
		}
		svc := NewService[testDomain](testBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Tree,
		})

		if err == nil {
			t.Fatal("expected error")
		}

		if !errors.Is(err, ErrMissingConfig) {
			t.Fatal("expected ErrMissingConfig")
		}
	})
}

func TestService_Execute_WithAudit(t *testing.T) {
	t.Parallel()

	t.Run("audit success", func(t *testing.T) {
		t.Parallel()

		auditLog := &testAuditLog{}
		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name: "test",
				Deductive: &schema.DeductiveConfig{
					Rules: []schema.DeductiveRuleDef{
						{Name: "r1", Condition: "a", Conclusion: map[string]bool{"b": true}},
					},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo, WithAuditLog(auditLog))
		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Deductive,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(auditLog.entries) != 1 {
			t.Fatalf("expected 1 audit entry, got %d", len(auditLog.entries))
		}

		if auditLog.entries[0].ID == "" {
			t.Fatal("expected non-empty audit entry ID")
		}
	})

	t.Run("audit error", func(t *testing.T) {
		t.Parallel()

		auditLog := &testAuditLog{err: errors.New("audit failed")}
		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name: "test",
				Deductive: &schema.DeductiveConfig{
					Rules: []schema.DeductiveRuleDef{
						{Name: "r1", Condition: "a", Conclusion: map[string]bool{"b": true}},
					},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo, WithAuditLog(auditLog))
		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Deductive,
		})

		if err == nil {
			t.Fatal("expected audit error")
		}

		if !errors.Is(err, ErrAuditFailed) {
			t.Fatal("expected ErrAuditFailed")
		}
	})

	t.Run("model audit success", func(t *testing.T) {
		t.Parallel()

		auditLog := &testAuditLog{}
		repo := &testRepo{
			ruleSet: &schema.RuleSet{
				Name: "test",
				Table: &schema.TableConfig{
					HitPolicy: "collect",
					Rules:     []schema.TableRuleDef{},
				},
			},
		}

		svc := NewService[testDomain](testBinder{}, repo, WithAuditLog(auditLog))
		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Table,
		})

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(auditLog.entries) != 1 {
			t.Fatalf("expected 1 audit entry, got %d", len(auditLog.entries))
		}

		if auditLog.entries[0].Paradigm != "table" {
			t.Fatalf("expected paradigm=table, got %s", auditLog.entries[0].Paradigm)
		}
	})
}

func TestService_Execute_DeductiveParseError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name: "test",
			Deductive: &schema.DeductiveConfig{
				Rules: []schema.DeductiveRuleDef{
					{Name: "bad", Condition: "(((", Conclusion: map[string]bool{"x": true}},
				},
			},
		},
	}

	svc := NewService[testDomain](testBinder{}, repo)

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Deductive,
	})

	if err == nil {
		t.Fatal("expected error for parse failure")
	}
}

func TestService_Execute_BayesianBuildError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name: "test",
			Bayesian: &schema.BayesianConfig{
				Nodes: []schema.BayesianNodeDef{
					{
						Variable: "rain",
						Outcomes: []string{"yes", "no"},
						CPT:      []schema.CPTRow{{Probabilities: map[string]float64{"yes": 0.5, "no": 0.8}}},
					},
				},
			},
		},
	}

	svc := NewService[testDomain](testBinder{}, repo)

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Bayesian,
		Query:    "rain",
	})

	if err == nil {
		t.Fatal("expected error for network build failure")
	}
}

func TestService_Execute_FuzzyBuildError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name: "test",
			Fuzzy: &schema.FuzzyConfig{
				InputVars: []schema.FuzzyVarDef{
					{Name: "x", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "a", Type: "unknown", Params: []float64{0}}}},
				},
				OutputVars: []schema.FuzzyVarDef{},
				Rules:      []schema.FuzzyRuleDef{},
			},
		},
	}

	svc := NewService[testDomain](testBinder{}, repo)

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Fuzzy,
	})

	if err == nil {
		t.Fatal("expected error for fuzzy var build failure")
	}
}

func TestService_Execute_FuzzyOutputVarError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name: "test",
			Fuzzy: &schema.FuzzyConfig{
				InputVars: []schema.FuzzyVarDef{
					{Name: "x", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "a", Type: "triangular", Params: []float64{0, 0.5, 1}}}},
				},
				OutputVars: []schema.FuzzyVarDef{
					{Name: "y", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "b", Type: "unknown", Params: []float64{0}}}},
				},
				Rules: []schema.FuzzyRuleDef{},
			},
		},
	}

	svc := NewService[testDomain](testBinder{}, repo)

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Fuzzy,
	})

	if err == nil {
		t.Fatal("expected error for fuzzy output var build failure")
	}
}

func TestService_Execute_DeductiveExplainError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name: "test",
			Deductive: &schema.DeductiveConfig{
				Rules: []schema.DeductiveRuleDef{
					{Name: "r1", Condition: "a", Conclusion: map[string]bool{"b": true}},
				},
			},
		},
	}

	svc := NewService[testDomain](testBinder{}, repo, WithExplainer(&failingExplainer{err: errors.New("explain failed")}))

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Deductive,
	})

	if err == nil {
		t.Fatal("expected explain error")
	}

	if !errors.Is(err, ErrExplainFailed) {
		t.Fatal("expected ErrExplainFailed")
	}
}

func TestService_Execute_BayesianExplainError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name: "test",
			Bayesian: &schema.BayesianConfig{
				Nodes: []schema.BayesianNodeDef{
					{
						Variable: "rain",
						Outcomes: []string{"yes", "no"},
						CPT:      []schema.CPTRow{{Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}}},
					},
				},
			},
		},
	}

	svc := NewService[testDomain](testBinder{}, repo, WithExplainer(&failingExplainer{err: errors.New("explain failed")}))

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Bayesian,
		Query:    "rain",
	})

	if err == nil {
		t.Fatal("expected explain error")
	}

	if !errors.Is(err, ErrExplainFailed) {
		t.Fatal("expected ErrExplainFailed")
	}
}

func TestService_Execute_FuzzyExplainError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name: "test",
			Fuzzy: &schema.FuzzyConfig{
				InputVars: []schema.FuzzyVarDef{
					{Name: "temp", Min: 0, Max: 100, Terms: []schema.FuzzyTermDef{
						{Name: "cold", Type: "triangular", Params: []float64{0, 0, 50}},
						{Name: "hot", Type: "triangular", Params: []float64{50, 100, 100}},
					}},
				},
				OutputVars: []schema.FuzzyVarDef{
					{Name: "speed", Min: 0, Max: 100, Terms: []schema.FuzzyTermDef{
						{Name: "slow", Type: "triangular", Params: []float64{0, 0, 50}},
						{Name: "fast", Type: "triangular", Params: []float64{50, 100, 100}},
					}},
				},
				Rules: []schema.FuzzyRuleDef{
					{Name: "r1", Conditions: []schema.FuzzyConditionDef{{Variable: "temp", Term: "hot"}}, Consequent: schema.FuzzyConsequentDef{Variable: "speed", Term: "fast"}},
				},
			},
		},
	}

	svc := NewService[testDomain](testBinder{}, repo, WithExplainer(&failingExplainer{err: errors.New("explain failed")}))

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Fuzzy,
	})

	if err == nil {
		t.Fatal("expected explain error")
	}

	if !errors.Is(err, ErrExplainFailed) {
		t.Fatal("expected ErrExplainFailed")
	}
}

// deductiveOnlyBinder only implements DeductiveBinder.
type deductiveOnlyBinder struct{}

func (b deductiveOnlyBinder) BindDeductive(_ testDomain) logic.Fact {
	return logic.Fact{"a": true}
}

// fuzzyOnlyBinder only implements FuzzyBinder.
type fuzzyOnlyBinder struct{}

func (b fuzzyOnlyBinder) BindFuzzy(_ testDomain) map[string]float64 {
	return map[string]float64{"x": 0.5}
}

// expressionOnlyBinder only implements ExpressionBinder.
type expressionOnlyBinder struct{}

func (b expressionOnlyBinder) BindExpression(d testDomain) cexpressions.Context {
	return cexpressions.Context{"amount": d.Amount}
}

func TestService_Execute_NoBinder(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name:      "test",
			Deductive: &schema.DeductiveConfig{Rules: []schema.DeductiveRuleDef{{Name: "r1", Condition: "a", Conclusion: map[string]bool{"b": true}}}},
			Bayesian: &schema.BayesianConfig{
				Nodes: []schema.BayesianNodeDef{{Variable: "rain", Outcomes: []string{"yes", "no"}, CPT: []schema.CPTRow{{Probabilities: map[string]float64{"yes": 0.3, "no": 0.7}}}}},
			},
			Fuzzy: &schema.FuzzyConfig{
				InputVars:  []schema.FuzzyVarDef{{Name: "x", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "a", Type: "triangular", Params: []float64{0, 0.5, 1}}}}},
				OutputVars: []schema.FuzzyVarDef{{Name: "y", Min: 0, Max: 1, Terms: []schema.FuzzyTermDef{{Name: "b", Type: "triangular", Params: []float64{0, 0.5, 1}}}}},
				Rules:      []schema.FuzzyRuleDef{{Name: "r1", Conditions: []schema.FuzzyConditionDef{{Variable: "x", Term: "a"}}, Consequent: schema.FuzzyConsequentDef{Variable: "y", Term: "b"}}},
			},
		},
	}

	t.Run("no deductive binder", func(t *testing.T) {
		t.Parallel()

		svc := NewService[testDomain](fuzzyOnlyBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Deductive,
		})

		if err == nil {
			t.Fatal("expected error for missing deductive binder")
		}

		if !errors.Is(err, ErrNoBinder) {
			t.Fatal("expected ErrNoBinder")
		}
	})

	t.Run("no bayesian binder", func(t *testing.T) {
		t.Parallel()

		svc := NewService[testDomain](deductiveOnlyBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Bayesian,
			Query:    "rain",
		})

		if err == nil {
			t.Fatal("expected error for missing bayesian binder")
		}

		if !errors.Is(err, ErrNoBinder) {
			t.Fatal("expected ErrNoBinder")
		}
	})

	t.Run("no fuzzy binder", func(t *testing.T) {
		t.Parallel()

		svc := NewService[testDomain](deductiveOnlyBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Fuzzy,
		})

		if err == nil {
			t.Fatal("expected error for missing fuzzy binder")
		}

		if !errors.Is(err, ErrNoBinder) {
			t.Fatal("expected ErrNoBinder")
		}
	})

	t.Run("no expression binder for table", func(t *testing.T) {
		t.Parallel()

		svc := NewService[testDomain](deductiveOnlyBinder{}, repo)

		_, err := svc.Execute(context.Background(), Request[testDomain]{
			Paradigm: Table,
		})

		if err == nil {
			t.Fatal("expected error for missing expression binder")
		}

		if !errors.Is(err, ErrNoBinder) {
			t.Fatal("expected ErrNoBinder")
		}
	})
}

func TestService_Execute_ModelExplainError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{
		ruleSet: &schema.RuleSet{
			Name: "test",
			Table: &schema.TableConfig{
				Rules: []schema.TableRuleDef{
					{Name: "r1", Conditions: []string{"amount > 0"}, Outputs: map[string]any{"val": 1}},
				},
			},
		},
	}

	failExplainer := &failingExplainer{err: errors.New("explain failed")}
	svc := NewService[testDomain](testBinder{}, repo, WithTableExplainer(failExplainer))

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Domain:   testDomain{Amount: 100},
		Paradigm: Table,
	})

	if err == nil {
		t.Fatal("expected explain error")
	}

	if !errors.Is(err, ErrExplainFailed) {
		t.Fatal("expected ErrExplainFailed")
	}
}

func TestService_Execute_ModelRepoError(t *testing.T) {
	t.Parallel()

	repo := &testRepo{err: errors.New("not found")}
	svc := NewService[testDomain](testBinder{}, repo)

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Table,
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrExecuteFailed) {
		t.Fatal("expected ErrExecuteFailed")
	}
}

func TestService_Execute_ModelUnsupported(t *testing.T) {
	t.Parallel()

	repo := &testRepo{ruleSet: &schema.RuleSet{Name: "test"}}
	svc := NewService[testDomain](expressionOnlyBinder{}, repo)

	_, err := svc.Execute(context.Background(), Request[testDomain]{
		Paradigm: Paradigm(99),
	})

	if err == nil {
		t.Fatal("expected error")
	}

	if !errors.Is(err, ErrUnsupported) {
		t.Fatal("expected ErrUnsupported")
	}
}

// Verify Binder[D] interface compliance.
var _ Binder[testDomain] = testBinder{}

// Verify individual binder interface compliance.
var _ DeductiveBinder[testDomain] = deductiveOnlyBinder{}

var _ FuzzyBinder[testDomain] = fuzzyOnlyBinder{}

var _ ExpressionBinder[testDomain] = expressionOnlyBinder{}

// Verify Repository interface compliance.
var _ repository.Repository = (*testRepo)(nil)

// Verify AuditLog interface compliance.
var _ Log = (*testAuditLog)(nil)
