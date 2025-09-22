package main

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/engine"
	parser2 "github.com/guidomantilla/yarumo/pkg/common/maths/logic2/parser"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

var (
	Adult2         = props.Var("Adult")
	Active2        = props.Var("Active")
	Colombian2     = props.Var("Colombian")
	EmailValid2    = props.Var("EmailValid")
	Has2FA2        = props.Var("Has2FA")
	TermsAccepted2 = props.Var("TermsAccepted")
	Admin2         = props.Var("Admin")
	CanLogin2      = props.Var("CanLogin")

	Rules = []engine.Rule{
		engine.BuildRule("r1", "Colombian & Adult", "Active"),
		engine.BuildRule("r2", "Active & Admin", "TermsAccepted"),
		engine.BuildRule("r3", "Active & Admin => Has2FA", "Has2FA"),
		engine.BuildRule("r3", "Active & EmailValid => CanLogin", "CanLogin"),
	}
)

func process(rules []engine.Rule, assertions ...props.Var) error {

	eng := engine.Engine{Facts: engine.FactBase{}, Rules: rules}

	for _, a := range assertions {
		eng.Assert(a)
	}

	fired := eng.RunToFixpoint(3)
	if len(fired) == 0 {
		return errors.New("no rules fired")
	}

	_, why := eng.Query(parser2.MustParse(Rules[0].String()))
	fmt.Print(engine.PrettyExplain(why))
	return nil
}
