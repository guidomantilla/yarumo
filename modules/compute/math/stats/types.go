// Package stats provides descriptive statistics and probability primitives.
package stats

// Var is a random variable name.
type Var string

// Outcome is a possible value of a random variable.
type Outcome string

// Prob is a probability value in [0,1].
type Prob float64

// Distribution maps outcomes to their probabilities.
type Distribution map[Outcome]Prob

// Assignment maps variables to observed outcomes.
type Assignment map[Var]Outcome

// ContinuousDist defines the interface for continuous probability distributions.
type ContinuousDist interface {
	// PDF returns the probability density function value at x.
	PDF(x float64) float64
	// CDF returns the cumulative distribution function value at x.
	CDF(x float64) float64
	// Mean returns the expected value.
	Mean() float64
	// Variance returns the variance.
	Variance() float64
	// Quantile returns the inverse CDF at probability p in [0,1].
	Quantile(p float64) float64
}

// Normal is the normal (Gaussian) distribution.
type Normal struct {
	Mu    float64
	Sigma float64
}

// Exponential is the exponential distribution with rate Lambda.
type Exponential struct {
	Lambda float64
}

// Uniform is the continuous uniform distribution over [Min, Max].
type Uniform struct {
	Min float64
	Max float64
}

// Beta is the Beta distribution with shape parameters Alpha and Beta.
type Beta struct {
	Alpha float64
	Bet   float64
}

// Gamma is the Gamma distribution with shape Alpha and rate Beta.
type Gamma struct {
	Alpha float64
	Bet   float64
}

// ChiSquared is the chi-squared distribution with K degrees of freedom.
type ChiSquared struct {
	K float64
}

// StudentT is the Student's t-distribution with Nu degrees of freedom.
type StudentT struct {
	Nu float64
}

// Lognormal is the log-normal distribution with parameters Mu and Sigma.
type Lognormal struct {
	Mu    float64
	Sigma float64
}

// Weibull is the Weibull distribution with shape K and scale Lambda.
type Weibull struct {
	K      float64
	Lambda float64
}

// FDist is the F-distribution with D1 and D2 degrees of freedom.
type FDist struct {
	D1 float64
	D2 float64
}

// Poisson is the Poisson distribution with rate Lambda.
type Poisson struct {
	Lambda float64
}

// Binomial is the binomial distribution with N trials and success probability P.
type Binomial struct {
	N int
	P float64
}

// Geometric is the geometric distribution with success probability P.
// Models the number of failures before the first success.
type Geometric struct {
	P float64
}

// Hypergeometric is the hypergeometric distribution.
// N is the population size, K is the number of success states, Draws is the number of draws.
type Hypergeometric struct {
	N     int
	K     int
	Draws int
}

// NegativeBinomial is the negative binomial distribution.
// R is the number of successes required, P is the success probability.
type NegativeBinomial struct {
	R int
	P float64
}

// Gumbel is the Gumbel (type I extreme value) distribution.
// Mu is the location parameter and Beta is the scale parameter.
type Gumbel struct {
	Mu   float64
	Beta float64
}

// Pareto is the Pareto (type I) distribution.
// Xm is the minimum value (scale) and Alpha is the shape parameter.
type Pareto struct {
	Xm    float64
	Alpha float64
}

// DiscreteDist defines the interface for discrete probability distributions.
type DiscreteDist interface {
	// PMF returns the probability mass function value at k.
	PMF(k int) float64
	// CDFDiscrete returns the cumulative distribution function value at k.
	CDFDiscrete(k int) float64
	// Mean returns the expected value.
	Mean() float64
	// Variance returns the variance.
	Variance() float64
}

// Type compliance.
var (
	_ ContinuousDist = Normal{}
	_ ContinuousDist = Exponential{}
	_ ContinuousDist = Uniform{}
	_ ContinuousDist = Beta{}
	_ ContinuousDist = Gamma{}
	_ ContinuousDist = ChiSquared{}
	_ ContinuousDist = StudentT{}
	_ ContinuousDist = Lognormal{}
	_ ContinuousDist = Weibull{}
	_ ContinuousDist = FDist{}
	_ ContinuousDist = Gumbel{}
	_ ContinuousDist = Pareto{}
	_ DiscreteDist   = Poisson{}
	_ DiscreteDist   = Binomial{}
	_ DiscreteDist   = Geometric{}
	_ DiscreteDist   = Hypergeometric{}
	_ DiscreteDist   = NegativeBinomial{}
)
