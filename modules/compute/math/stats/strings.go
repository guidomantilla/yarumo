package stats

import (
	"slices"
	"strconv"
	"strings"
)

// String returns a human-readable representation of a distribution.
func (d Distribution) String() string {
	if len(d) == 0 {
		return "{}"
	}

	keys := make([]Outcome, 0, len(d))

	for k := range d {
		keys = append(keys, k)
	}

	slices.Sort(keys)

	var b strings.Builder

	b.WriteString("{")

	for i, k := range keys {
		if i > 0 {
			b.WriteString(", ")
		}

		b.WriteString(string(k))
		b.WriteString("=")
		b.WriteString(strconv.FormatFloat(float64(d[k]), 'f', -1, 64))
	}

	b.WriteString("}")

	return b.String()
}

// String returns a string representation of a Normal distribution.
func (n Normal) String() string {
	return "Normal(μ=" + strconv.FormatFloat(n.Mu, 'g', -1, 64) +
		", σ=" + strconv.FormatFloat(n.Sigma, 'g', -1, 64) + ")"
}

// String returns a string representation of an Exponential distribution.
func (e Exponential) String() string {
	return "Exponential(λ=" + strconv.FormatFloat(e.Lambda, 'g', -1, 64) + ")"
}

// String returns a string representation of a Uniform distribution.
func (u Uniform) String() string {
	return "Uniform(" + strconv.FormatFloat(u.Min, 'g', -1, 64) +
		", " + strconv.FormatFloat(u.Max, 'g', -1, 64) + ")"
}

// String returns a string representation of a Beta distribution.
func (b Beta) String() string {
	return "Beta(α=" + strconv.FormatFloat(b.Alpha, 'g', -1, 64) +
		", β=" + strconv.FormatFloat(b.Bet, 'g', -1, 64) + ")"
}

// String returns a string representation of a Gamma distribution.
func (g Gamma) String() string {
	return "Gamma(α=" + strconv.FormatFloat(g.Alpha, 'g', -1, 64) +
		", β=" + strconv.FormatFloat(g.Bet, 'g', -1, 64) + ")"
}

// String returns a string representation of a ChiSquared distribution.
func (c ChiSquared) String() string {
	return "ChiSquared(k=" + strconv.FormatFloat(c.K, 'g', -1, 64) + ")"
}

// String returns a string representation of a StudentT distribution.
func (s StudentT) String() string {
	return "StudentT(ν=" + strconv.FormatFloat(s.Nu, 'g', -1, 64) + ")"
}

// String returns a string representation of a Lognormal distribution.
func (l Lognormal) String() string {
	return "Lognormal(μ=" + strconv.FormatFloat(l.Mu, 'g', -1, 64) +
		", σ=" + strconv.FormatFloat(l.Sigma, 'g', -1, 64) + ")"
}

// String returns a string representation of a Weibull distribution.
func (w Weibull) String() string {
	return "Weibull(k=" + strconv.FormatFloat(w.K, 'g', -1, 64) +
		", λ=" + strconv.FormatFloat(w.Lambda, 'g', -1, 64) + ")"
}

// String returns a string representation of an FDist distribution.
func (f FDist) String() string {
	return "F(d1=" + strconv.FormatFloat(f.D1, 'g', -1, 64) +
		", d2=" + strconv.FormatFloat(f.D2, 'g', -1, 64) + ")"
}

// String returns a string representation of a Poisson distribution.
func (p Poisson) String() string {
	return "Poisson(λ=" + strconv.FormatFloat(p.Lambda, 'g', -1, 64) + ")"
}

// String returns a string representation of a Binomial distribution.
func (b Binomial) String() string {
	return "Binomial(n=" + strconv.Itoa(b.N) +
		", p=" + strconv.FormatFloat(b.P, 'g', -1, 64) + ")"
}

// String returns a string representation of a Geometric distribution.
func (g Geometric) String() string {
	return "Geometric(p=" + strconv.FormatFloat(g.P, 'g', -1, 64) + ")"
}

// String returns a string representation of a Hypergeometric distribution.
func (h Hypergeometric) String() string {
	return "Hypergeometric(N=" + strconv.Itoa(h.N) +
		", K=" + strconv.Itoa(h.K) +
		", n=" + strconv.Itoa(h.Draws) + ")"
}

// String returns a string representation of a NegativeBinomial distribution.
func (nb NegativeBinomial) String() string {
	return "NegativeBinomial(r=" + strconv.Itoa(nb.R) +
		", p=" + strconv.FormatFloat(nb.P, 'g', -1, 64) + ")"
}

// String returns a string representation of a Gumbel distribution.
func (g Gumbel) String() string {
	return "Gumbel(μ=" + strconv.FormatFloat(g.Mu, 'g', -1, 64) +
		", β=" + strconv.FormatFloat(g.Beta, 'g', -1, 64) + ")"
}

// String returns a string representation of a Pareto distribution.
func (p Pareto) String() string {
	return "Pareto(xm=" + strconv.FormatFloat(p.Xm, 'g', -1, 64) +
		", α=" + strconv.FormatFloat(p.Alpha, 'g', -1, 64) + ")"
}
