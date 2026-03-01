package evaluator

import (
	"base/object"
	"regexp"
)

func RegisterRegexBuiltins() {
	builtins["regex.match"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			pattern, ok1 := args[0].(*object.String)
			input, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `regex.match` must be STRING")
			}
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex: %s", err.Error())
			}
			if re.MatchString(input.Value) {
				return TRUE
			}
			return FALSE
		},
	}

	builtins["regex.find"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			pattern, ok1 := args[0].(*object.String)
			input, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `regex.find` must be STRING")
			}
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex: %s", err.Error())
			}
			match := re.FindString(input.Value)
			if match == "" {
				return NULL
			}
			return &object.String{Value: match}
		},
	}

	builtins["regex.find_all"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			pattern, ok1 := args[0].(*object.String)
			input, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `regex.find_all` must be STRING")
			}
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex: %s", err.Error())
			}
			matches := re.FindAllString(input.Value, 10000)
			elements := make([]object.Object, len(matches))
			for i, m := range matches {
				elements[i] = &object.String{Value: m}
			}
			if len(matches) == 10000 && len(re.FindAllString(input.Value, 10001)) > 10000 {
				return newError("regex.find_all matched over 10,000 times (OOM protected)")
			}
			return &object.Array{Elements: elements}
		},
	}

	builtins["regex.replace"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			pattern, ok1 := args[0].(*object.String)
			input, ok2 := args[1].(*object.String)
			replacement, ok3 := args[2].(*object.String)
			if !ok1 || !ok2 || !ok3 {
				return newError("arguments to `regex.replace` must be STRING")
			}
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex: %s", err.Error())
			}
			return &object.String{Value: re.ReplaceAllString(input.Value, replacement.Value)}
		},
	}

	builtins["regex.split"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			pattern, ok1 := args[0].(*object.String)
			input, ok2 := args[1].(*object.String)
			if !ok1 || !ok2 {
				return newError("arguments to `regex.split` must be STRING")
			}
			re, err := regexp.Compile(pattern.Value)
			if err != nil {
				return newError("invalid regex: %s", err.Error())
			}
			parts := re.Split(input.Value, 10001)
			if len(parts) > 10000 {
				return newError("regex.split generated over 10,000 fragments (OOM protected)")
			}
			elements := make([]object.Object, len(parts))
			for i, p := range parts {
				elements[i] = &object.String{Value: p}
			}
			return &object.Array{Elements: elements}
		},
	}
}
