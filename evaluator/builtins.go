package evaluator

import (
	"base/object"
	"fmt"
	"strings"
)

var builtins = map[string]*object.Builtin{
	"print": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			out := make([]string, len(args))
			for i, arg := range args {
				out[i] = arg.Inspect()
			}
			fmt.Println(strings.Join(out, " "))
			return NULL
		},
	},
	"len": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}

			switch arg := args[0].(type) {
			case *object.String:
				return &object.Integer{Value: int64(len(arg.Value))}
			case *object.Array:
				return &object.Integer{Value: int64(len(arg.Elements))}
			case *object.Hash:
				return &object.Integer{Value: int64(len(arg.Pairs))}
			default:
				return newError("argument to `len` not supported, got %s", args[0].Type())
			}
		},
	},
	"type": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			return &object.String{Value: string(args[0].Type())}
		},
	},
	"range": {
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 || len(args) > 3 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}
			start, ok1 := args[0].(*object.Integer)
			end, ok2 := args[1].(*object.Integer)
			if !ok1 || !ok2 {
				return newError("arguments to `range` must be INTEGER")
			}
			step := int64(1)
			if len(args) == 3 {
				s, ok := args[2].(*object.Integer)
				if !ok {
					return newError("step argument to `range` must be INTEGER")
				}
				step = s.Value
			}
			if step == 0 {
				return newError("step cannot be zero")
			}

			// OOM defense
			if (end.Value-start.Value)/step > 100000 || (end.Value-start.Value)/step < -100000 {
				return newError("range results in too many elements (max 100,000)")
			}

			var elements []object.Object
			if step > 0 {
				for i := start.Value; i < end.Value; i += step {
					elements = append(elements, &object.Integer{Value: i})
				}
			} else {
				for i := start.Value; i > end.Value; i += step {
					elements = append(elements, &object.Integer{Value: i})
				}
			}
			return &object.Array{Elements: elements}
		},
	},
}
