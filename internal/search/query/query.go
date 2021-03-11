package query

// pipline should take a parser (literal or regex), which is a function, and produce
// a parsetree.

// then, i also want it to have validation functions, that just carry forward the parse tree,
// but throw an error if they fail (for settings, etc).

// Pipeline parses the query string and performs query expansion to generate
// possibly multiple disjoint queries to execute (results should be unioned).
func Pipeline(in string, options ParserOptions) (Plan, error) {
	parseTree, err := Parse(in, options)
	if err != nil {
		return nil, err
	}

	var plan Plan
	for _, disjunct := range Dnf(parseTree) {
		err = validate(disjunct)
		if err != nil {
			return nil, err
		}
		plan = append(plan, Q(disjunct))
	}
	return plan, nil
}
