package query

import (
	"fmt"
	"strconv"
	"strings"
)

func stringHumanPattern(nodes []Node) string {
	var result []string
	for _, node := range nodes {
		switch n := node.(type) {
		case Pattern:
			v := n.Value
			if n.Annotation.Labels.isSet(Quoted) {
				v = strconv.Quote(v)
			}
			if n.Negated {
				v = fmt.Sprintf("(not %s)", v)
			}
			result = append(result, v)
		case Operator:
			var nested []string
			for _, operand := range n.Operands {
				nested = append(nested, stringHumanPattern([]Node{operand}))
			}
			var separator string
			switch n.Kind {
			case Or:
				separator = " or "
			case And:
				separator = " and "
			}
			result = append(result, "("+strings.Join(nested, separator)+")")
		}
	}
	return strings.Join(result, "")
}

func stringHumanParameters(nodes []Parameter) string {
	var result []string
	for _, n := range nodes {
		v := n.Value
		if n.Annotation.Labels.isSet(Quoted) {
			v = strconv.Quote(v)
		}
		if n.Negated {
			return fmt.Sprintf("-%s:%s", n.Field, v)
		}
		result = append(result, fmt.Sprintf("%s:%s", n.Field, v))
	}
	return strings.Join(result, " ")
}

// StringHuman creates a valid query string from a parsed query. It is used in
// contexts like query suggestions where we take the original query string of a
// user, parse it to a tree, modify the tree, and return a valid string
// representation. To faithfully preserve the meaning of the original tree,
// we need to consider whether to add operators like "and" contextually and must
// process the tree as a whole:
//
// repo:foo file:bar a and b -> preserve 'and', but do not insert 'and' between 'repo:foo file:bar'.
// repo:foo file:bar a b     -> do not insert any 'and', especially not between 'a b'.
//
// It strives to be syntax preserving, but may in some cases affect whitespace,
// operator capitalization, or parenthesized groupings. In very complex queries,
// additional 'and' operators may be inserted to segment parameters
// from patterns to preserve the original meaning.
func StringHuman(nodes []Node) string {
	basic, err := PartitionSearchPattern(nodes)
	if err != nil {
		// We couldn't partition at this level in the tree, so recurse on operators until we can.
		var v []string
		for _, node := range nodes {
			if term, ok := node.(Operator); ok {
				var s []string
				for _, operand := range term.Operands {
					s = append(s, StringHuman([]Node{operand}))
				}
				if term.Kind == Or {
					v = append(v, "("+strings.Join(s, " or ")+")")
				} else if term.Kind == And {
					v = append(v, "("+strings.Join(s, " and ")+")")
				}
			}
		}
		return strings.Join(v, "")
	}
	if basic.Pattern == nil {
		return stringHumanParameters(basic.Parameters)
	}
	if len(basic.Parameters) == 0 {
		return stringHumanPattern([]Node{basic.Pattern})
	}
	return stringHumanParameters(basic.Parameters) + " " + stringHumanPattern([]Node{basic.Pattern})
}

// toString returns a string representation of a query's structure.
func toString(nodes []Node) string {
	var result []string
	for _, node := range nodes {
		result = append(result, node.String())
	}
	return strings.Join(result, " ")
}

// toString returns a string representation of a Basic query's structure.
func toStringBasic(basic *Basic) string {
	var result []string
	for _, node := range basic.Parameters {
		result = append(result, node.String())
	}
	if basic.Pattern != nil {
		result = append(result, basic.Pattern.String())
	}
	return strings.Join(result, " ")
}
