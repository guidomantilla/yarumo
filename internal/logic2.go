package main

import (
	"errors"
	"fmt"

	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/engine"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/parser"
	"github.com/guidomantilla/yarumo/pkg/common/maths/logic2/props"
)

var (
	User          = UserType{Name: "Ana", Age: 22, Active: true, Email: "", Country: "CO", IsAdmin: true, Has2FA: false, TermsAccepted: false}
	UserInferable = UserType{Name: "Ana", Age: 17, Active: true, Email: "hey", Country: "CO", IsAdmin: true, Has2FA: true, TermsAccepted: true}
	Rules         = []engine.Rule{
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

	for _, rule := range rules {
		_, why := eng.Query(parser.MustParse(rule.String()))
		fmt.Println(engine.PrettyExplain(why))
	}

	return nil
}
