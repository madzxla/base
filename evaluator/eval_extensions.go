package evaluator

import (
	"base/ast"
	"base/object"
	"strings"
)

func evalTemplateLiteral(tl *ast.TemplateLiteral, env *object.Environment) object.Object {
	var out strings.Builder
	for _, part := range tl.Parts {
		val := Eval(part, env)
		if isError(val) {
			return val
		}
		out.WriteString(val.Inspect())
	}
	return &object.String{Value: out.String()}
}

func evalArrowFunction(af *ast.ArrowFunctionLiteral, env *object.Environment) object.Object {
	if af.Body != nil {
		return &object.Function{
			Parameters: af.Parameters,
			Body:       af.Body,
			Env:        env,
		}
	}
	// Single expression: wrap in a return statement block
	body := &ast.BlockStatement{
		Statements: []ast.Statement{
			&ast.ReturnStatement{
				ReturnValue: af.Expression,
			},
		},
	}
	return &object.Function{
		Parameters: af.Parameters,
		Body:       body,
		Env:        env,
	}
}

func evalMatchExpression(me *ast.MatchExpression, env *object.Environment) object.Object {
	subject := Eval(me.Subject, env)
	if isError(subject) {
		return subject
	}

	for _, mc := range me.Cases {
		for _, val := range mc.Values {
			caseVal := Eval(val, env)
			if isError(caseVal) {
				return caseVal
			}
			if evalEquals(subject, caseVal) {
				return Eval(mc.Body, env)
			}
		}
	}

	if me.Default != nil {
		return Eval(me.Default, env)
	}

	return NULL
}

func evalDestructureLetStatement(ds *ast.DestructureLetStatement, env *object.Environment) object.Object {
	val := Eval(ds.Value, env)
	if isError(val) {
		return val
	}

	if ds.IsHash {
		hash, ok := val.(*object.Hash)
		if !ok {
			return newError("destructure: expected HASH, got %s", val.Type())
		}
		for _, name := range ds.Names {
			if v, exists := hash.Pairs[name]; exists {
				env.Set(name, v)
			} else {
				env.Set(name, NULL)
			}
		}
	} else {
		arr, ok := val.(*object.Array)
		if !ok {
			return newError("destructure: expected ARRAY, got %s", val.Type())
		}
		for i, name := range ds.Names {
			if i < len(arr.Elements) {
				env.Set(name, arr.Elements[i])
			} else {
				env.Set(name, NULL)
			}
		}
	}

	return NULL
}

func evalEnumStatement(es *ast.EnumStatement, env *object.Environment) object.Object {
	enumHash := &object.Hash{Pairs: map[string]object.Object{}}
	for i, member := range es.Members {
		enumHash.Pairs[member] = &object.Integer{Value: int64(i)}
	}
	env.Set(es.Name, enumHash)
	return NULL
}
