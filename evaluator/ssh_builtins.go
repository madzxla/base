package evaluator

import (
	"base/object"
	"net"
	"os"

	"golang.org/x/crypto/ssh"
)

func RegisterSSHBuiltins() {
	builtins["ssh.exec"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 4 {
				return newError("wrong number of arguments. got=%d, want=4", len(args))
			}
			hostObj, ok1 := args[0].(*object.String)
			userObj, ok2 := args[1].(*object.String)
			keyPathObj, ok3 := args[2].(*object.String)
			commandObj, ok4 := args[3].(*object.String)
			if !ok1 || !ok2 || !ok3 || !ok4 {
				return newError("all arguments to `ssh.exec` must be STRING")
			}
			host := hostObj.Value
			user := userObj.Value
			keyPath := keyPathObj.Value
			command := commandObj.Value

			key, err := os.ReadFile(keyPath)
			if err != nil {
				return newError("unable to read private key: %v", err)
			}

			signer, err := ssh.ParsePrivateKey(key)
			if err != nil {
				return newError("unable to parse private key: %v", err)
			}

			config := &ssh.ClientConfig{
				User: user,
				Auth: []ssh.AuthMethod{
					ssh.PublicKeys(signer),
				},
				HostKeyCallback: ssh.InsecureIgnoreHostKey(),
			}

			client, err := ssh.Dial("tcp", net.JoinHostPort(host, "22"), config)
			if err != nil {
				return newError("unable to connect: %v", err)
			}
			defer client.Close()

			session, err := client.NewSession()
			if err != nil {
				return newError("unable to create session: %v", err)
			}
			defer session.Close()

			output, err := session.CombinedOutput(command)
			if err != nil {
				return &object.Hash{
					Pairs: map[string]object.Object{
						"exit_code": &object.Integer{Value: 1},
						"output":    &object.String{Value: string(output)},
						"error":     &object.String{Value: err.Error()},
					},
				}
			}

			return &object.Hash{
				Pairs: map[string]object.Object{
					"exit_code": &object.Integer{Value: 0},
					"output":    &object.String{Value: string(output)},
				},
			}
		},
	}
}
