package rules

import "fmt"

type DecisionTable struct {
	Inputs  []string
	Outputs []string
	Rules   []DecisionRow
}

type DecisionRow struct {
	Conditions []bool
	Actions    []string
}

func EvaluateTable(table DecisionTable, facts map[string]interface{}) (map[string]string, error) {
	for _, row := range table.Rules {
		match := true
		for i, cond := range table.Inputs {
			passes, err := evaluateCondition(cond, facts)
			if err != nil {
				return nil, err
			}
			if passes != row.Conditions[i] {
				match = false
				break
			}
		}
		if match {
			result := make(map[string]string)
			for j, output := range table.Outputs {
				result[output] = row.Actions[j]
			}
			return result, nil
		}
	}
	return nil, nil
}

func evaluateCondition(expr string, facts map[string]interface{}) (bool, error) {
	// Parse simple expressions like "age >= 65"
	var field string
	var op string
	var value float64

	_, err := fmt.Sscanf(expr, "%s %s %f", &field, &op, &value)
	if err != nil {
		return false, err
	}

	actual, ok := facts[field]
	if !ok {
		return false, fmt.Errorf("missing fact: %s", field)
	}

	num, ok := actual.(float64)
	if !ok {
		return false, fmt.Errorf("fact %s is not numeric", field)
	}

	switch op {
	case ">=":
		return num >= value, nil
	case ">":
		return num > value, nil
	case "<":
		return num < value, nil
	case "<=":
		return num <= value, nil
	case "==":
		return num == value, nil
	case "!=":
		return num != value, nil
	default:
		return false, fmt.Errorf("unsupported operator: %s", op)
	}
}
