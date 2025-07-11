package propositional

type Formula interface {
	String() string
	Eval(env map[string]bool) bool
	Vars() []string
}

type Var string

type NotF struct{ F Formula }

type AndF struct{ L, R Formula }

type OrF struct{ L, R Formula }

type ImplF struct{ L, R Formula }

type IffF struct{ L, R Formula }

func (v Var) String() string { return string(v) }

func (v Var) Eval(env map[string]bool) bool { return env[string(v)] }

func (v Var) Vars() []string { return []string{string(v)} }

func (f NotF) String() string { return "¬" + f.F.String() }

func (f NotF) Eval(env map[string]bool) bool { return !f.F.Eval(env) }

func (f NotF) Vars() []string { return f.F.Vars() }

func (f AndF) String() string { return "(" + f.L.String() + " ∧ " + f.R.String() + ")" }

func (f AndF) Eval(env map[string]bool) bool { return f.L.Eval(env) && f.R.Eval(env) }

func (f AndF) Vars() []string { return union(f.L.Vars(), f.R.Vars()) }

func (f OrF) String() string { return "(" + f.L.String() + " ∨ " + f.R.String() + ")" }

func (f OrF) Eval(env map[string]bool) bool { return f.L.Eval(env) || f.R.Eval(env) }

func (f OrF) Vars() []string { return union(f.L.Vars(), f.R.Vars()) }

func (f ImplF) String() string { return "(" + f.L.String() + " ⇒ " + f.R.String() + ")" }

func (f ImplF) Eval(env map[string]bool) bool { return !f.L.Eval(env) || f.R.Eval(env) }

func (f ImplF) Vars() []string { return union(f.L.Vars(), f.R.Vars()) }

func (f IffF) String() string { return "(" + f.L.String() + " ⇔ " + f.R.String() + ")" }

func (f IffF) Eval(env map[string]bool) bool { return f.L.Eval(env) == f.R.Eval(env) }

func (f IffF) Vars() []string { return union(f.L.Vars(), f.R.Vars()) }

func union(a, b []string) []string {
	set := make(map[string]struct{})
	for _, x := range a {
		set[x] = struct{}{}
	}
	for _, x := range b {
		set[x] = struct{}{}
	}
	var out []string
	for x := range set {
		out = append(out, x)
	}
	return out
}

func TruthTable(f Formula) []map[string]bool {
	vars := f.Vars()
	n := len(vars)
	var rows []map[string]bool
	for i := 0; i < 1<<n; i++ {
		row := make(map[string]bool)
		for j, v := range vars {
			row[v] = (i>>j)&1 == 1
		}
		row["result"] = f.Eval(row)
		rows = append(rows, row)
	}
	return rows
}

func Equivalent(a, b Formula) bool {
	ttA := TruthTable(a)
	ttB := TruthTable(b)
	if len(ttA) != len(ttB) {
		return false
	}
	for i := range ttA {
		if ttA[i]["result"] != ttB[i]["result"] {
			return false
		}
	}
	return true
}
