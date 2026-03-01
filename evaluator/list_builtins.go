package evaluator

import (
	"base/object"
	"sort"
	"strings"
)

func RegisterListBuiltins() {
	builtins["list.length"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `list.length` must be ARRAY, got %s", args[0].Type())
			}
			return &object.Integer{Value: int64(len(arr.Elements))}
		},
	}

	builtins["list.map"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok1 := args[0].(*object.Array)
			fn, ok2 := args[1].(*object.Function)
			if !ok1 || !ok2 {
				return newError("arguments to `list.map` must be (ARRAY, FUNCTION)")
			}

			newElements := make([]object.Object, len(arr.Elements))
			for i, el := range arr.Elements {
				newElements[i] = applyFunction(env, fn, []object.Object{el})
			}
			return &object.Array{Elements: newElements}
		},
	}

	builtins["list.filter"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok1 := args[0].(*object.Array)
			fn, ok2 := args[1].(*object.Function)
			if !ok1 || !ok2 {
				return newError("arguments to `list.filter` must be (ARRAY, FUNCTION)")
			}

			newElements := []object.Object{}
			for _, el := range arr.Elements {
				res := applyFunction(env, fn, []object.Object{el})
				if res == TRUE {
					newElements = append(newElements, el)
				}
			}
			return &object.Array{Elements: newElements}
		},
	}

	builtins["list.contains"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `list.contains` must be ARRAY")
			}
			target := args[1]
			for _, el := range arr.Elements {
				if evalEquals(el, target) {
					return TRUE
				}
			}
			return FALSE
		},
	}

	builtins["list.sort"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `list.sort` must be ARRAY")
			}
			
			newElements := make([]object.Object, len(arr.Elements))
			copy(newElements, arr.Elements)
			sort.Slice(newElements, func(i, j int) bool {
				return newElements[i].Inspect() < newElements[j].Inspect()
			})
			return &object.Array{Elements: newElements}
		},
	}

	builtins["list.reverse"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `list.reverse` must be ARRAY")
			}
			newElements := make([]object.Object, len(arr.Elements))
			for i, el := range arr.Elements {
				newElements[len(arr.Elements)-1-i] = el
			}
			return &object.Array{Elements: newElements}
		},
	}

	builtins["list.reduce"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			arr, ok1 := args[0].(*object.Array)
			fn, ok2 := args[1].(*object.Function)
			if !ok1 || !ok2 {
				return newError("arguments to `list.reduce` must be (ARRAY, FUNCTION, initial)")
			}
			acc := args[2]
			for _, el := range arr.Elements {
				acc = applyFunction(env, fn, []object.Object{acc, el})
				if isError(acc) {
					return acc
				}
			}
			return acc
		},
	}

	builtins["list.join"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok1 := args[0].(*object.Array)
			delim, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `list.join` must be (ARRAY, STRING)")
			}
			parts := make([]string, len(arr.Elements))
			for i, el := range arr.Elements {
				parts[i] = el.Inspect()
			}
			return &object.String{Value: strings.Join(parts, delim.Value)}
		},
	}

	builtins["list.slice"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) < 2 {
				return newError("wrong number of arguments. got=%d, want=2 or 3", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `list.slice` must be ARRAY")
			}
			start, ok := args[1].(*object.Integer)
			if !ok {
				return newError("second argument to `list.slice` must be INTEGER")
			}
			startVal := int(start.Value)
			endVal := len(arr.Elements)
			if len(args) == 3 {
				if end, ok := args[2].(*object.Integer); ok {
					endVal = int(end.Value)
				}
			}
			if startVal < 0 {
				startVal = 0
			}
			if endVal > len(arr.Elements) {
				endVal = len(arr.Elements)
			}
			if startVal > endVal {
				return &object.Array{Elements: []object.Object{}}
			}
			newElements := make([]object.Object, endVal-startVal)
			copy(newElements, arr.Elements[startVal:endVal])
			return &object.Array{Elements: newElements}
		},
	}

	builtins["list.flat"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("argument to `list.flat` must be ARRAY")
			}
			var result []object.Object
			for _, el := range arr.Elements {
				if inner, ok := el.(*object.Array); ok {
					result = append(result, inner.Elements...)
				} else {
					result = append(result, el)
				}
			}
			return &object.Array{Elements: result}
		},
	}

	builtins["list.push"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `list.push` must be ARRAY")
			}
			newElements := make([]object.Object, len(arr.Elements)+1)
			copy(newElements, arr.Elements)
			newElements[len(arr.Elements)] = args[1]
			return &object.Array{Elements: newElements}
		},
	}

	builtins["list.index_of"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			arr, ok := args[0].(*object.Array)
			if !ok {
				return newError("first argument to `list.index_of` must be ARRAY")
			}
			target := args[1]
			for i, el := range arr.Elements {
				if evalEquals(el, target) {
					return &object.Integer{Value: int64(i)}
				}
			}
			return &object.Integer{Value: -1}
		},
	}
}
