package evaluator

import (
	"base/object"
	"bytes"
	"encoding/csv"
	"os"

	"gopkg.in/yaml.v3"
)

func RegisterDataBuiltins() {
	builtins["csv.read"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `csv.read` must be STRING")
			}
			content, err := os.ReadFile(path.Value)
			if err != nil {
				return newError("could not read file: %s", err.Error())
			}
			r := csv.NewReader(bytes.NewReader(content))
			records, err := r.ReadAll()
			if err != nil {
				return newError("csv parse error: %s", err.Error())
			}
			rows := make([]object.Object, len(records))
			for i, record := range records {
				cols := make([]object.Object, len(record))
				for j, col := range record {
					cols[j] = &object.String{Value: col}
				}
				rows[i] = &object.Array{Elements: cols}
			}
			return &object.Array{Elements: rows}
		},
	}

	builtins["csv.write"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			path, ok1 := args[0].(*object.String)
			data, ok2 := args[1].(*object.Array)
			if !ok1 || !ok2 {
				return newError("arguments to `csv.write` must be (STRING, ARRAY)")
			}
			var buf bytes.Buffer
			w := csv.NewWriter(&buf)
			for _, rowObj := range data.Elements {
				row, ok := rowObj.(*object.Array)
				if !ok {
					return newError("each row in csv.write must be an ARRAY")
				}
				record := make([]string, len(row.Elements))
				for j, col := range row.Elements {
					record[j] = col.Inspect()
				}
				w.Write(record)
			}
			w.Flush()
			if err := os.WriteFile(path.Value, buf.Bytes(), 0644); err != nil {
				return newError("could not write file: %s", err.Error())
			}
			return TRUE
		},
	}

	builtins["yaml.read"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			path, ok := args[0].(*object.String)
			if !ok {
				return newError("argument to `yaml.read` must be STRING")
			}
			content, err := os.ReadFile(path.Value)
			if err != nil {
				return newError("could not read file: %s", err.Error())
			}
			var result interface{}
			if err := yaml.Unmarshal(content, &result); err != nil {
				return newError("yaml parse error: %s", err.Error())
			}
			return goTypeToBaseObject(result)
		},
	}

	builtins["yaml.write"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			path, ok1 := args[0].(*object.String)
			data := args[1]
			if !ok1 {
				return newError("first argument to `yaml.write` must be STRING")
			}
			goData := baseObjectToGoType(data)
			yamlBytes, err := yaml.Marshal(goData)
			if err != nil {
				return newError("yaml marshal error: %s", err.Error())
			}
			err = os.WriteFile(path.Value, yamlBytes, 0644)
			if err != nil {
				return newError("could not write file: %s", err.Error())
			}
			return TRUE
		},
	}
}
