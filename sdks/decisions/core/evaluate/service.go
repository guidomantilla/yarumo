package evaluate

import (
	"context"
	"errors"
	"time"

	cassert "github.com/guidomantilla/yarumo/common/assert"
	cerrs "github.com/guidomantilla/yarumo/common/errs"
	cuids "github.com/guidomantilla/yarumo/common/uids"

	"github.com/guidomantilla/yarumo/decisions/core/repository"
)

// Service defines the interface for decision execution: bind → execute → explain → audit.
type Service[D any] interface {
	// Execute runs a decision for the given request.
	Execute(ctx context.Context, request Request[D]) (Result, error)
}

type service[D any] struct {
	deductiveBinder  DeductiveBinder[D]
	bayesianBinder   BayesianBinder[D]
	fuzzyBinder      FuzzyBinder[D]
	expressionBinder ExpressionBinder[D]
	repo             repository.Repository
	options          *Options
}

// NewService creates a new Service with the given binder, repository, and options.
// The binder must implement at least one of DeductiveBinder, BayesianBinder, FuzzyBinder,
// or ExpressionBinder. Use the combined Binder interface for convenience when all paradigms
// are needed.
func NewService[D any](binder any, repo repository.Repository, opts ...Option) Service[D] {
	cassert.NotNil(binder, "binder is nil")
	cassert.NotNil(repo, "repository is nil")

	svc := &service[D]{
		repo:    repo,
		options: NewOptions(opts...),
	}

	db, dbOk := binder.(DeductiveBinder[D])
	if dbOk {
		svc.deductiveBinder = db
	}

	bb, bbOk := binder.(BayesianBinder[D])
	if bbOk {
		svc.bayesianBinder = bb
	}

	fb, fbOk := binder.(FuzzyBinder[D])
	if fbOk {
		svc.fuzzyBinder = fb
	}

	eb, ebOk := binder.(ExpressionBinder[D])
	if ebOk {
		svc.expressionBinder = eb
	}

	cassert.True(dbOk || bbOk || fbOk || ebOk,
		"binder must implement at least one of DeductiveBinder, BayesianBinder, FuzzyBinder, or ExpressionBinder")

	return svc
}

// Execute runs a decision for the given request.
func (s *service[D]) Execute(ctx context.Context, request Request[D]) (Result, error) {
	cassert.NotNil(s, "service is nil")

	start := time.Now()

	ruleSet, err := s.repo.Get(ctx, request.RuleSetName, request.RuleSetVersion)
	if err != nil {
		return Result{}, ErrExecute(err)
	}

	result, err := s.dispatch(ctx, request, ruleSet)
	if err != nil {
		if errors.Is(err, ErrExplainFailed) {
			return Result{}, err
		}

		return Result{}, ErrExecute(err)
	}

	if s.options.auditLog != nil {
		id, idErr := cuids.UuidV7.Generate()
		if idErr != nil {
			return result, ErrAudit(idErr)
		}

		auditErr := s.options.auditLog.Record(ctx, Entry{
			ID:             id,
			Timestamp:      start,
			RuleSetName:    request.RuleSetName,
			RuleSetVersion: request.RuleSetVersion,
			Paradigm:       request.Paradigm.String(),
			Request:        request,
			Result:         result,
			Explanation:    result.Explanation,
			Duration:       time.Since(start),
		})
		if auditErr != nil {
			return result, ErrAudit(auditErr)
		}
	}

	return result, nil
}

func (s *service[D]) dispatch(ctx context.Context, request Request[D], ruleSet any) (Result, error) {
	switch request.Paradigm {
	case Deductive, Bayesian, Fuzzy:
		return s.dispatchInference(ctx, request, ruleSet)
	case Table, Scorecard, Tree:
		return s.dispatchModel(ctx, request, ruleSet)
	default:
		return Result{}, cerrs.Wrap(ErrUnsupported)
	}
}

func (s *service[D]) dispatchInference(ctx context.Context, request Request[D], ruleSet any) (Result, error) {
	input, query, err := s.bind(request)
	if err != nil {
		return Result{}, err
	}

	return dispatchParadigm(ctx, request.Paradigm, ruleSet, input, query, s.options.explainers())
}

func (s *service[D]) dispatchModel(ctx context.Context, request Request[D], ruleSet any) (Result, error) {
	if s.expressionBinder == nil {
		return Result{}, cerrs.Wrap(ErrNoBinder)
	}

	exprCtx := s.expressionBinder.BindExpression(request.Domain)

	return dispatchModelParadigm(ctx, request.Paradigm, ruleSet, exprCtx, s.options)
}

// bind converts domain data to the paradigm-specific input.
func (s *service[D]) bind(request Request[D]) (any, string, error) {
	switch request.Paradigm {
	case Deductive:
		if s.deductiveBinder == nil {
			return nil, "", cerrs.Wrap(ErrNoBinder)
		}

		return s.deductiveBinder.BindDeductive(request.Domain), "", nil

	case Bayesian:
		if s.bayesianBinder == nil {
			return nil, "", cerrs.Wrap(ErrNoBinder)
		}

		return s.bayesianBinder.BindBayesian(request.Domain), request.Query, nil

	case Fuzzy:
		if s.fuzzyBinder == nil {
			return nil, "", cerrs.Wrap(ErrNoBinder)
		}

		return s.fuzzyBinder.BindFuzzy(request.Domain), "", nil

	case Table, Scorecard, Tree:
		return nil, "", cerrs.Wrap(ErrUnsupported)
	default:
		return nil, "", cerrs.Wrap(ErrUnsupported)
	}
}
