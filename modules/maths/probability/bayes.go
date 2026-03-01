package probability

// Bayes computes P(H|E) using Bayes' theorem: P(H|E) = P(E|H)*P(H) / P(E).
func Bayes(prior, likelihood, evidence Prob) Prob {
	if evidence == 0 {
		return 0
	}

	return Prob(float64(likelihood) * float64(prior) / float64(evidence))
}

// TotalProbability computes P(E) = Sum P(E|Hi)*P(Hi).
// priors and likelihoods must have the same length.
func TotalProbability(priors, likelihoods []Prob) Prob {
	var sum float64

	for i := range priors {
		sum += float64(priors[i]) * float64(likelihoods[i])
	}

	return Prob(sum)
}

// ChainRule computes the joint probability: P(A,B,...) = P(A)*P(B|A)*P(C|A,B)*...
func ChainRule(conditionals ...Prob) Prob {
	result := 1.0

	for _, p := range conditionals {
		result *= float64(p)
	}

	return Prob(result)
}

// Independent computes P(A intersection B) for independent events: P(A)*P(B).
func Independent(a, b Prob) Prob {
	return Prob(float64(a) * float64(b))
}
