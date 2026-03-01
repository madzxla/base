package evaluator

import (
	"base/object"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/smtp"
)

func RegisterNotifyBuiltins() {
	builtins["notify.discord"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			webhookURL, ok1 := args[0].(*object.String)
			message, ok2 := args[1].(*object.String)

			if !ok1 || !ok2 {
				return newError("arguments to `notify.discord` must be (STRING, STRING)")
			}

			payload := map[string]string{"content": message.Value}
			jsonPayload, err := json.Marshal(payload)
			if err != nil {
				return newError("discord: failed to marshal payload: %s", err.Error())
			}

			resp, err := http.Post(webhookURL.Value, "application/json", bytes.NewBuffer(jsonPayload))
			if err != nil {
				return newError("discord post error: %s", err.Error())
			}
			defer resp.Body.Close()

			if resp.StatusCode >= 400 {
				return newError("discord returned status: %d", resp.StatusCode)
			}

			return TRUE
		},
	}

	builtins["notify.email"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 7 {
				return newError("wrong number of arguments. got=%d, want=7", len(args))
			}
			hostObj, ok1 := args[0].(*object.String)
			portObj, ok2 := args[1].(*object.String)
			userObj, ok3 := args[2].(*object.String)
			passObj, ok4 := args[3].(*object.String)
			toObj, ok5 := args[4].(*object.String)
			subjectObj, ok6 := args[5].(*object.String)
			bodyObj, ok7 := args[6].(*object.String)
			if !ok1 || !ok2 || !ok3 || !ok4 || !ok5 || !ok6 || !ok7 {
				return newError("all arguments to `notify.email` must be STRING")
			}
			host := hostObj.Value
			port := portObj.Value
			user := userObj.Value
			pass := passObj.Value
			to := toObj.Value
			subject := subjectObj.Value
			body := bodyObj.Value

			auth := smtp.PlainAuth("", user, pass, host)
			msg := []byte(fmt.Sprintf("To: %s\r\nSubject: %s\r\n\r\n%s\r\n", to, subject, body))

			err := smtp.SendMail(host+":"+port, auth, user, []string{to}, msg)
			if err != nil {
				return newError("email send error: %s", err.Error())
			}

			return TRUE
		},
	}
}
