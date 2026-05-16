package acceptance

import (
	"fmt"
	"math"
	"testing"

	"github.com/guidomantilla/yarumo/compute/math/fuzzy"
	"github.com/guidomantilla/yarumo/compute/math/stats"
)

// Section 1.6: fuzzy/ — Axioms (14 tests)

func TestAcceptance_TNorm_commutativity_exhaustive(t *testing.T) {
	t.Parallel()

	norms := []struct {
		name string
		fn   fuzzy.TNormFn
	}{
		{"Min", fuzzy.Min},
		{"Product", fuzzy.Product},
		{"Lukasiewicz", fuzzy.Lukasiewicz},
	}

	grid := fuzzyGrid(0.05)

	for _, norm := range norms {
		t.Run(norm.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range grid {
				for _, b := range grid {
					ab := norm.fn(fuzzy.Degree(a), fuzzy.Degree(b))
					ba := norm.fn(fuzzy.Degree(b), fuzzy.Degree(a))

					diff := math.Abs(float64(ab) - float64(ba))
					if diff > 1e-15 {
						t.Fatalf("%s: commutativity failed for a=%.2f b=%.2f: T(a,b)=%.15f T(b,a)=%.15f",
							norm.name, a, b, float64(ab), float64(ba))
					}
				}
			}
		})
	}
}

func TestAcceptance_TNorm_associativity_exhaustive(t *testing.T) {
	t.Parallel()

	norms := []struct {
		name string
		fn   fuzzy.TNormFn
	}{
		{"Min", fuzzy.Min},
		{"Product", fuzzy.Product},
		{"Lukasiewicz", fuzzy.Lukasiewicz},
	}

	grid := fuzzyGrid(0.05)

	for _, norm := range norms {
		t.Run(norm.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range grid {
				for _, b := range grid {
					for _, c := range grid {
						ab := norm.fn(fuzzy.Degree(a), fuzzy.Degree(b))
						lhs := norm.fn(ab, fuzzy.Degree(c))

						bc := norm.fn(fuzzy.Degree(b), fuzzy.Degree(c))
						rhs := norm.fn(fuzzy.Degree(a), bc)

						diff := math.Abs(float64(lhs) - float64(rhs))
						if diff > 1e-15 {
							t.Fatalf("%s: associativity failed for a=%.2f b=%.2f c=%.2f: T(T(a,b),c)=%.15f T(a,T(b,c))=%.15f",
								norm.name, a, b, c, float64(lhs), float64(rhs))
						}
					}
				}
			}
		})
	}
}

func TestAcceptance_TNorm_monotonicity_exhaustive(t *testing.T) {
	t.Parallel()

	norms := []struct {
		name string
		fn   fuzzy.TNormFn
	}{
		{"Min", fuzzy.Min},
		{"Product", fuzzy.Product},
		{"Lukasiewicz", fuzzy.Lukasiewicz},
	}

	grid := fuzzyGrid(0.05)

	for _, norm := range norms {
		t.Run(norm.name, func(t *testing.T) {
			t.Parallel()

			for i, a1 := range grid {
				for _, a2 := range grid[i:] {
					for _, b := range grid {
						v1 := norm.fn(fuzzy.Degree(a1), fuzzy.Degree(b))
						v2 := norm.fn(fuzzy.Degree(a2), fuzzy.Degree(b))

						if float64(v1) > float64(v2)+1e-15 {
							t.Fatalf("%s: monotonicity failed for a1=%.2f a2=%.2f b=%.2f: T(a1,b)=%.15f > T(a2,b)=%.15f",
								norm.name, a1, a2, b, float64(v1), float64(v2))
						}
					}
				}
			}
		})
	}
}

func TestAcceptance_TNorm_identity_exhaustive(t *testing.T) {
	t.Parallel()

	norms := []struct {
		name string
		fn   fuzzy.TNormFn
	}{
		{"Min", fuzzy.Min},
		{"Product", fuzzy.Product},
		{"Lukasiewicz", fuzzy.Lukasiewicz},
	}

	grid := fuzzyGrid(0.05)

	for _, norm := range norms {
		t.Run(norm.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range grid {
				result := norm.fn(fuzzy.Degree(a), 1.0)

				diff := math.Abs(float64(result) - a)
				if diff > 1e-15 {
					t.Fatalf("%s: identity failed for a=%.2f: T(a,1)=%.15f want %.15f",
						norm.name, a, float64(result), a)
				}
			}
		})
	}
}

func TestAcceptance_TConorm_commutativity_exhaustive(t *testing.T) {
	t.Parallel()

	conorms := []struct {
		name string
		fn   fuzzy.TConormFn
	}{
		{"Max", fuzzy.Max},
		{"ProbabilisticSum", fuzzy.ProbabilisticSum},
		{"BoundedSum", fuzzy.BoundedSum},
	}

	grid := fuzzyGrid(0.05)

	for _, conorm := range conorms {
		t.Run(conorm.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range grid {
				for _, b := range grid {
					ab := conorm.fn(fuzzy.Degree(a), fuzzy.Degree(b))
					ba := conorm.fn(fuzzy.Degree(b), fuzzy.Degree(a))

					diff := math.Abs(float64(ab) - float64(ba))
					if diff > 1e-15 {
						t.Fatalf("%s: commutativity failed for a=%.2f b=%.2f: S(a,b)=%.15f S(b,a)=%.15f",
							conorm.name, a, b, float64(ab), float64(ba))
					}
				}
			}
		})
	}
}

func TestAcceptance_TConorm_associativity_exhaustive(t *testing.T) {
	t.Parallel()

	conorms := []struct {
		name string
		fn   fuzzy.TConormFn
	}{
		{"Max", fuzzy.Max},
		{"ProbabilisticSum", fuzzy.ProbabilisticSum},
		{"BoundedSum", fuzzy.BoundedSum},
	}

	grid := fuzzyGrid(0.05)

	for _, conorm := range conorms {
		t.Run(conorm.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range grid {
				for _, b := range grid {
					for _, c := range grid {
						ab := conorm.fn(fuzzy.Degree(a), fuzzy.Degree(b))
						lhs := conorm.fn(ab, fuzzy.Degree(c))

						bc := conorm.fn(fuzzy.Degree(b), fuzzy.Degree(c))
						rhs := conorm.fn(fuzzy.Degree(a), bc)

						diff := math.Abs(float64(lhs) - float64(rhs))
						if diff > 1e-15 {
							t.Fatalf("%s: associativity failed for a=%.2f b=%.2f c=%.2f: S(S(a,b),c)=%.15f S(a,S(b,c))=%.15f",
								conorm.name, a, b, c, float64(lhs), float64(rhs))
						}
					}
				}
			}
		})
	}
}

func TestAcceptance_TConorm_monotonicity_exhaustive(t *testing.T) {
	t.Parallel()

	conorms := []struct {
		name string
		fn   fuzzy.TConormFn
	}{
		{"Max", fuzzy.Max},
		{"ProbabilisticSum", fuzzy.ProbabilisticSum},
		{"BoundedSum", fuzzy.BoundedSum},
	}

	grid := fuzzyGrid(0.05)

	for _, conorm := range conorms {
		t.Run(conorm.name, func(t *testing.T) {
			t.Parallel()

			for i, a1 := range grid {
				for _, a2 := range grid[i:] {
					for _, b := range grid {
						v1 := conorm.fn(fuzzy.Degree(a1), fuzzy.Degree(b))
						v2 := conorm.fn(fuzzy.Degree(a2), fuzzy.Degree(b))

						if float64(v1) > float64(v2)+1e-15 {
							t.Fatalf("%s: monotonicity failed for a1=%.2f a2=%.2f b=%.2f: S(a1,b)=%.15f > S(a2,b)=%.15f",
								conorm.name, a1, a2, b, float64(v1), float64(v2))
						}
					}
				}
			}
		})
	}
}

func TestAcceptance_TConorm_identity_exhaustive(t *testing.T) {
	t.Parallel()

	conorms := []struct {
		name string
		fn   fuzzy.TConormFn
	}{
		{"Max", fuzzy.Max},
		{"ProbabilisticSum", fuzzy.ProbabilisticSum},
		{"BoundedSum", fuzzy.BoundedSum},
	}

	grid := fuzzyGrid(0.05)

	for _, conorm := range conorms {
		t.Run(conorm.name, func(t *testing.T) {
			t.Parallel()

			for _, a := range grid {
				result := conorm.fn(fuzzy.Degree(a), 0)

				diff := math.Abs(float64(result) - a)
				if diff > 1e-15 {
					t.Fatalf("%s: identity failed for a=%.2f: S(a,0)=%.15f want %.15f",
						conorm.name, a, float64(result), a)
				}
			}
		})
	}
}

func TestAcceptance_TNorm_Lukasiewicz_boundary(t *testing.T) {
	t.Parallel()

	t.Run("Luk_0.3_0.6_equals_0", func(t *testing.T) {
		t.Parallel()

		got := float64(fuzzy.Lukasiewicz(0.3, 0.6))
		assertFloat(t, "Luk(0.3,0.6)", got, 0, 1e-15)
	})

	t.Run("Luk_0.5_0.51_equals_0.01", func(t *testing.T) {
		t.Parallel()

		got := float64(fuzzy.Lukasiewicz(0.5, 0.51))
		assertFloat(t, "Luk(0.5,0.51)", got, 0.01, 1e-15)
	})

	t.Run("Luk_1_1_equals_1", func(t *testing.T) {
		t.Parallel()

		got := float64(fuzzy.Lukasiewicz(1, 1))
		assertFloat(t, "Luk(1,1)", got, 1, 1e-15)
	})
}

func TestAcceptance_TNorm_Product_epsilon(t *testing.T) {
	t.Parallel()

	t.Run("subnormal_non_negative", func(t *testing.T) {
		t.Parallel()

		got := float64(fuzzy.Product(fuzzy.Degree(1e-300), fuzzy.Degree(1e-300)))
		if got < 0 {
			t.Fatalf("Product(1e-300,1e-300) = %e, want >= 0", got)
		}
	})

	t.Run("small_product_accuracy", func(t *testing.T) {
		t.Parallel()

		got := float64(fuzzy.Product(fuzzy.Degree(1e-15), fuzzy.Degree(1e-15)))
		want := 1e-30
		assertFloat(t, "Product(1e-15,1e-15)", got, want, 1e-45)
	})
}

func TestAcceptance_Membership_exact_values(t *testing.T) {
	t.Parallel()

	t.Run("triangular_peak", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("Triangular(0,5,10) error: %v", err)
		}

		assertFloat(t, "Triangular(0,5,10) at x=5", float64(fn(5)), 1.0, 1e-15)
	})

	t.Run("triangular_left_zero", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("Triangular(0,5,10) error: %v", err)
		}

		assertFloat(t, "Triangular(0,5,10) at x=0", float64(fn(0)), 0.0, 1e-15)
	})

	t.Run("triangular_right_zero", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Triangular(0, 5, 10)
		if err != nil {
			t.Fatalf("Triangular(0,5,10) error: %v", err)
		}

		assertFloat(t, "Triangular(0,5,10) at x=10", float64(fn(10)), 0.0, 1e-15)
	})

	t.Run("trapezoidal_plateau", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Trapezoidal(0, 3, 7, 10)
		if err != nil {
			t.Fatalf("Trapezoidal(0,3,7,10) error: %v", err)
		}

		assertFloat(t, "Trapezoidal plateau at x=5", float64(fn(5)), 1.0, 1e-15)
	})

	t.Run("trapezoidal_left_zero", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Trapezoidal(0, 3, 7, 10)
		if err != nil {
			t.Fatalf("Trapezoidal(0,3,7,10) error: %v", err)
		}

		assertFloat(t, "Trapezoidal at x=0", float64(fn(0)), 0.0, 1e-15)
	})

	t.Run("gaussian_center", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Gaussian(5, 2)
		if err != nil {
			t.Fatalf("Gaussian(5,2) error: %v", err)
		}

		assertFloat(t, "Gaussian(5,2) at x=5", float64(fn(5)), 1.0, 1e-15)
	})

	t.Run("gaussian_symmetry", func(t *testing.T) {
		t.Parallel()

		fn, err := fuzzy.Gaussian(5, 2)
		if err != nil {
			t.Fatalf("Gaussian(5,2) error: %v", err)
		}

		left := float64(fn(3))
		right := float64(fn(7))

		diff := math.Abs(left - right)
		if diff > 1e-15 {
			t.Fatalf("Gaussian symmetry failed: f(3)=%.15f f(7)=%.15f diff=%e", left, right, diff)
		}
	})
}

func TestAcceptance_Defuzzification_output_in_range(t *testing.T) {
	t.Parallel()

	fn, err := fuzzy.Triangular(0, 5, 10)
	if err != nil {
		t.Fatalf("Triangular(0,5,10) error: %v", err)
	}

	samples := 100
	xs := make([]float64, samples+1)
	ys := make([]fuzzy.Degree, samples+1)

	for i := range samples + 1 {
		x := float64(i) * 10.0 / float64(samples)
		xs[i] = x
		ys[i] = fn(x)
	}

	methods := []struct {
		name string
		fn   fuzzy.DefuzzifyFn
	}{
		{"Centroid", fuzzy.Centroid},
		{"Bisector", fuzzy.Bisector},
		{"MeanOfMax", fuzzy.MeanOfMax},
		{"LargestOfMax", fuzzy.LargestOfMax},
		{"SmallestOfMax", fuzzy.SmallestOfMax},
	}

	for _, method := range methods {
		t.Run(method.name, func(t *testing.T) {
			t.Parallel()

			result := method.fn(xs, ys)

			if result < 0 || result > 10 {
				t.Fatalf("%s: result %.6f outside range [0,10]", method.name, result)
			}
		})
	}
}

// Section 1.7: stats/ — Distributions (12 tests)

func TestAcceptance_Stats_known_answer_Normal(t *testing.T) {
	t.Parallel()

	dist, err := stats.NewNormal(0, 1)
	if err != nil {
		t.Fatalf("NewNormal(0,1) error: %v", err)
	}

	t.Run("PDF_at_0", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Normal(0,1).PDF(0)", dist.PDF(0), 0.398942280401433, floatTolerance)
	})

	t.Run("CDF_at_0", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Normal(0,1).CDF(0)", dist.CDF(0), 0.5, floatTolerance)
	})

	t.Run("CDF_at_1.96", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Normal(0,1).CDF(1.96)", dist.CDF(1.96), 0.975002104852, floatTolerance)
	})

	t.Run("CDF_at_-1.96", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Normal(0,1).CDF(-1.96)", dist.CDF(-1.96), 0.024997895148, floatTolerance)
	})

	t.Run("Mean", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Normal(0,1).Mean()", dist.Mean(), 0, floatTolerance)
	})

	t.Run("Variance", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Normal(0,1).Variance()", dist.Variance(), 1, floatTolerance)
	})
}

func TestAcceptance_Stats_known_answer_Exponential(t *testing.T) {
	t.Parallel()

	dist, err := stats.NewExponential(2)
	if err != nil {
		t.Fatalf("NewExponential(2) error: %v", err)
	}

	t.Run("PDF_at_0", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Exponential(2).PDF(0)", dist.PDF(0), 2.0, floatTolerance)
	})

	t.Run("CDF_at_1", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Exponential(2).CDF(1)", dist.CDF(1), 0.864664716763387, floatTolerance)
	})

	t.Run("Mean", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Exponential(2).Mean()", dist.Mean(), 0.5, floatTolerance)
	})

	t.Run("Variance", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Exponential(2).Variance()", dist.Variance(), 0.25, floatTolerance)
	})
}

func TestAcceptance_Stats_known_answer_Beta(t *testing.T) {
	t.Parallel()

	dist, err := stats.NewBeta(2, 5)
	if err != nil {
		t.Fatalf("NewBeta(2,5) error: %v", err)
	}

	t.Run("Mean", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Beta(2,5).Mean()", dist.Mean(), 2.0/7.0, floatTolerance)
	})

	t.Run("Variance", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Beta(2,5).Variance()", dist.Variance(), 10.0/392.0, floatTolerance)
	})
}

func TestAcceptance_Stats_known_answer_StudentT(t *testing.T) {
	t.Parallel()

	dist, err := stats.NewStudentT(5)
	if err != nil {
		t.Fatalf("NewStudentT(5) error: %v", err)
	}

	t.Run("CDF_at_0", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "StudentT(5).CDF(0)", dist.CDF(0), 0.5, floatTolerance)
	})

	t.Run("Mean", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "StudentT(5).Mean()", dist.Mean(), 0, floatTolerance)
	})

	t.Run("Variance", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "StudentT(5).Variance()", dist.Variance(), 5.0/3.0, floatTolerance)
	})
}

func TestAcceptance_Stats_known_answer_Poisson(t *testing.T) {
	t.Parallel()

	dist, err := stats.NewPoisson(3)
	if err != nil {
		t.Fatalf("NewPoisson(3) error: %v", err)
	}

	t.Run("PMF_at_0", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Poisson(3).PMF(0)", dist.PMF(0), 0.049787068368, floatTolerance)
	})

	t.Run("PMF_at_3", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Poisson(3).PMF(3)", dist.PMF(3), 0.224041807660, floatTolerance)
	})

	t.Run("Mean", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Poisson(3).Mean()", dist.Mean(), 3.0, floatTolerance)
	})

	t.Run("Variance", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Poisson(3).Variance()", dist.Variance(), 3.0, floatTolerance)
	})
}

func TestAcceptance_Stats_known_answer_Binomial(t *testing.T) {
	t.Parallel()

	dist, err := stats.NewBinomial(10, 0.3)
	if err != nil {
		t.Fatalf("NewBinomial(10,0.3) error: %v", err)
	}

	t.Run("PMF_at_3", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Binomial(10,0.3).PMF(3)", dist.PMF(3), 0.266827932, floatTolerance)
	})

	t.Run("Mean", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Binomial(10,0.3).Mean()", dist.Mean(), 3.0, floatTolerance)
	})

	t.Run("Variance", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Binomial(10,0.3).Variance()", dist.Variance(), 2.1, floatTolerance)
	})
}

func TestAcceptance_Stats_adversarial_Welford(t *testing.T) {
	t.Parallel()

	t.Run("catastrophic_cancellation", func(t *testing.T) {
		t.Parallel()

		var rs stats.RunningStats
		rs.Push(1e8 + 1)
		rs.Push(1e8 + 2)
		rs.Push(1e8 + 3)

		assertFloat(t, "Welford mean", rs.Mean(), 1e8+2, floatTolerance)
		assertFloat(t, "Welford population variance", rs.Variance(), 2.0/3.0, floatTolerance)
	})

	t.Run("identical_data", func(t *testing.T) {
		t.Parallel()

		var rs stats.RunningStats
		rs.Push(7)
		rs.Push(7)
		rs.Push(7)
		rs.Push(7)

		assertFloat(t, "identical variance", rs.Variance(), 0.0, floatTolerance)
	})

	t.Run("single_datum", func(t *testing.T) {
		t.Parallel()

		var rs stats.RunningStats
		rs.Push(42)

		assertFloat(t, "single mean", rs.Mean(), 42.0, floatTolerance)
		assertFloat(t, "single variance", rs.Variance(), 0.0, floatTolerance)
	})
}

func TestAcceptance_Stats_adversarial_WindowedStats(t *testing.T) {
	t.Parallel()

	windowSize := 10
	totalValues := 100

	ws := stats.NewWindowedStats(windowSize)
	if ws == nil {
		t.Fatalf("NewWindowedStats(%d) returned nil", windowSize)
	}

	values := make([]float64, totalValues)
	for i := range totalValues {
		values[i] = float64(i*i) + 0.5
	}

	for i, v := range values {
		ws.Push(v)

		if i < windowSize-1 {
			continue
		}

		// Capture actual stats before the parallel subtest runs, because
		// subsequent Push calls will mutate the WindowedStats.
		start := i - windowSize + 1
		window := values[start : i+1]

		var sum float64
		for _, val := range window {
			sum += val
		}

		wantMean := sum / float64(windowSize)

		var varSum float64

		for _, val := range window {
			d := val - wantMean
			varSum += d * d
		}

		wantVariance := varSum / float64(windowSize)

		gotMean := ws.Mean()
		gotVariance := ws.Variance()

		step := i
		t.Run(fmt.Sprintf("step_%d", step), func(t *testing.T) {
			t.Parallel()

			assertFloat(t, fmt.Sprintf("mean at step %d", step), gotMean, wantMean, floatTolerance)
			assertFloat(t, fmt.Sprintf("variance at step %d", step), gotVariance, wantVariance, 1e-6)
		})
	}
}

func TestAcceptance_Stats_adversarial_Bayes(t *testing.T) {
	t.Parallel()

	prior := stats.Prob(0.001)
	sensitivity := stats.Prob(0.99)
	falsePositive := stats.Prob(0.05)

	pEvidenceGivenNotH := falsePositive
	pNotH := stats.Prob(1 - float64(prior))

	pEvidence, err := stats.TotalProbability(
		[]stats.Prob{prior, pNotH},
		[]stats.Prob{sensitivity, pEvidenceGivenNotH},
	)
	if err != nil {
		t.Fatalf("TotalProbability error: %v", err)
	}

	posterior, err := stats.Bayes(prior, sensitivity, pEvidence)
	if err != nil {
		t.Fatalf("Bayes error: %v", err)
	}

	wantPosterior := (0.99 * 0.001) / (0.99*0.001 + 0.05*0.999)

	t.Run("posterior_low_despite_positive_test", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "Bayes posterior", float64(posterior), wantPosterior, floatTolerance)
	})

	t.Run("posterior_below_prior_intuition", func(t *testing.T) {
		t.Parallel()

		if float64(posterior) > 0.5 {
			t.Fatalf("base rate fallacy: posterior %.6f should be << 0.5 for rare condition", float64(posterior))
		}
	})

	t.Run("evidence_total_probability", func(t *testing.T) {
		t.Parallel()

		wantEvidence := 0.99*0.001 + 0.05*0.999
		assertFloat(t, "P(E)", float64(pEvidence), wantEvidence, floatTolerance)
	})
}

func TestAcceptance_Stats_known_answer_ChiSquared(t *testing.T) {
	t.Parallel()

	dist, err := stats.NewChiSquared(3)
	if err != nil {
		t.Fatalf("NewChiSquared(3) error: %v", err)
	}

	t.Run("PDF_at_1", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "ChiSquared(3).PDF(1)", dist.PDF(1), 0.24197, 1e-4)
	})

	t.Run("CDF_at_7.815", func(t *testing.T) {
		t.Parallel()
		// The 95th-percentile critical value for df=3 is ~7.8147; at x=7.815
		// the CDF slightly exceeds 0.95.
		assertFloat(t, "ChiSquared(3).CDF(7.815)", dist.CDF(7.815), 0.9611, 1e-3)
	})

	t.Run("Mean", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "ChiSquared(3).Mean()", dist.Mean(), 3.0, floatTolerance)
	})

	t.Run("Variance", func(t *testing.T) {
		t.Parallel()
		assertFloat(t, "ChiSquared(3).Variance()", dist.Variance(), 6.0, floatTolerance)
	})
}

func TestAcceptance_Stats_hypothesis_TTest(t *testing.T) {
	t.Parallel()

	t.Run("non_rejection", func(t *testing.T) {
		t.Parallel()

		sample := []float64{2.1, 2.3, 1.9, 2.0, 2.2}

		_, pValue, err := stats.TTest(sample, 2.0)
		if err != nil {
			t.Fatalf("TTest error: %v", err)
		}

		if pValue < 0.05 {
			t.Fatalf("non-rejection: p-value %.6f < 0.05, should not reject H0: mu=2.0", pValue)
		}
	})

	t.Run("rejection", func(t *testing.T) {
		t.Parallel()

		sample := []float64{10.1, 10.3, 9.9, 10.2, 10.4, 10.0, 10.5}

		_, pValue, err := stats.TTest(sample, 9.0)
		if err != nil {
			t.Fatalf("TTest error: %v", err)
		}

		if pValue >= 0.05 {
			t.Fatalf("rejection: p-value %.6f >= 0.05, should reject H0: mu=9.0", pValue)
		}
	})
}
