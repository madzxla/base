package evaluator

import (
	"base/object"
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	serverMuxes      = map[int64]*http.ServeMux{}
	serverStarted    = map[int64]bool{}
	serverMu         sync.Mutex
	serverMiddleware []*object.Function
	serverRouteGroup string
	serverStateMu    sync.RWMutex
)

func getOrCreateMux(port int64) *http.ServeMux {
	serverMu.Lock()
	defer serverMu.Unlock()
	if mux, ok := serverMuxes[port]; ok {
		return mux
	}
	mux := http.NewServeMux()
	serverMuxes[port] = mux
	return mux
}

func startServerOnce(port int64, mux *http.ServeMux) {
	serverMu.Lock()
	if serverStarted[port] {
		serverMu.Unlock()
		return
	}
	serverStarted[port] = true
	serverMu.Unlock()

	go func() {
		KeepAlive = true
		fmt.Printf("B.A.S.E. Server listening on :%d\n", port)
		server := &http.Server{
			Addr:              fmt.Sprintf(":%d", port),
			Handler:           mux,
			ReadHeaderTimeout: 20 * time.Second,
			ReadTimeout:       1 * time.Minute,
			WriteTimeout:      2 * time.Minute,
			IdleTimeout:       120 * time.Second,
		}
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("Server error: %s\n", err.Error())
		}
	}()
}

func getStatusCode(arg object.Object) (int, bool) {
	switch v := arg.(type) {
	case *object.Integer:
		return int(v.Value), true
	case *object.Float:
		return int(v.Value), true
	}
	return 0, false
}

func buildResObject(env *object.Environment, w http.ResponseWriter) *object.Hash {
	customHeaders := map[string]string{}

	resSend := &object.Builtin{
		Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
			if len(innerArgs) < 2 {
				return newError("res.send needs (status, body)")
			}
			status, ok := getStatusCode(innerArgs[0])
			if !ok {
				return newError("res.send: status must be INTEGER")
			}
			for k, v := range customHeaders {
				w.Header().Set(k, v)
			}
			if w.Header().Get("Content-Type") == "" {
				w.Header().Set("Content-Type", "application/json")
			}
			w.WriteHeader(status)
			jsonBytes, err := json.Marshal(baseObjectToGoType(innerArgs[1]))
			if err != nil {
				jsonBytes = []byte("{}")
			}
			w.Write(jsonBytes)
			return NULL
		},
	}

	resJson := &object.Builtin{
		Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
			if len(innerArgs) < 2 {
				return newError("res.json needs (status, data)")
			}
			status, ok := getStatusCode(innerArgs[0])
			if !ok {
				return newError("res.json: status must be INTEGER")
			}
			for k, v := range customHeaders {
				w.Header().Set(k, v)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(status)
			jsonBytes, err := json.MarshalIndent(baseObjectToGoType(innerArgs[1]), "", "  ")
			if err != nil {
				jsonBytes = []byte("{}")
			}
			w.Write(jsonBytes)
			return NULL
		},
	}

	resHtml := &object.Builtin{
		Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
			if len(innerArgs) < 2 {
				return newError("res.html needs (status, content)")
			}
			status, ok := getStatusCode(innerArgs[0])
			if !ok {
				return newError("res.html: status must be INTEGER")
			}
			for k, v := range customHeaders {
				w.Header().Set(k, v)
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(status)
			w.Write([]byte(innerArgs[1].Inspect()))
			return NULL
		},
	}

	resText := &object.Builtin{
		Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
			if len(innerArgs) < 2 {
				return newError("res.text needs (status, content)")
			}
			status, ok := getStatusCode(innerArgs[0])
			if !ok {
				return newError("res.text: status must be INTEGER")
			}
			for k, v := range customHeaders {
				w.Header().Set(k, v)
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(status)
			w.Write([]byte(innerArgs[1].Inspect()))
			return NULL
		},
	}

	resHeader := &object.Builtin{
		Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
			if len(innerArgs) != 2 {
				return newError("res.header needs (key, value)")
			}
			customHeaders[innerArgs[0].Inspect()] = innerArgs[1].Inspect()
			return NULL
		},
	}

	resFile := &object.Builtin{
		Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
			if len(innerArgs) < 2 {
				return newError("res.file needs (status, filepath)")
			}
			status, ok := getStatusCode(innerArgs[0])
			if !ok {
				return newError("res.file: status must be INTEGER")
			}
			filePath, ok := innerArgs[1].(*object.String)
			if !ok {
				return newError("res.file filepath must be STRING")
			}

			absPath, err := filepath.Abs(filePath.Value)
			if err != nil {
				return newError("invalid path: %s", err.Error())
			}
			cwd, err := os.Getwd()
			if err != nil {
				return newError("could not determine working directory: %s", err.Error())
			}
			if !strings.HasPrefix(absPath, cwd) {
				return newError("security: access denied to file outside project root: %s", filePath.Value)
			}

			info, err := os.Stat(absPath)
			if err != nil {
				w.WriteHeader(404)
				w.Write([]byte("File not found"))
				return NULL
			}
			if info.Size() > 100*1024*1024 {
				return newError("res.file: file too large (max 100MB)")
			}

			content, err := os.ReadFile(absPath)
			if err != nil {
				w.WriteHeader(500)
				w.Write([]byte("Error reading file"))
				return NULL
			}
			ext := filepath.Ext(absPath)
			mimeType := mime.TypeByExtension(ext)
			if mimeType == "" {
				mimeType = "application/octet-stream"
			}
			for k, v := range customHeaders {
				w.Header().Set(k, v)
			}
			w.Header().Set("Content-Type", mimeType)
			w.WriteHeader(status)
			w.Write(content)
			return NULL
		},
	}

	resRedirect := &object.Builtin{
		Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
			if len(innerArgs) < 2 {
				return newError("res.redirect needs (status, url)")
			}
			status, ok := getStatusCode(innerArgs[0])
			if !ok {
				return newError("res.redirect: status must be INTEGER")
			}
			target, ok := innerArgs[1].(*object.String)
			if !ok {
				return newError("res.redirect: url must be STRING")
			}
			for k, v := range customHeaders {
				w.Header().Set(k, v)
			}
			w.Header().Set("Location", target.Value)
			w.WriteHeader(status)
			return NULL
		},
	}

	resCors := &object.Builtin{
		Fn: func(innerEnv *object.Environment, innerArgs ...object.Object) object.Object {
			origin := "*"
			if len(innerArgs) > 0 {
				if o, ok := innerArgs[0].(*object.String); ok {
					origin = o.Value
				}
			}
			customHeaders["Access-Control-Allow-Origin"] = origin
			customHeaders["Access-Control-Allow-Methods"] = "GET, POST, PUT, PATCH, DELETE, OPTIONS"
			customHeaders["Access-Control-Allow-Headers"] = "Content-Type, Authorization"
			return NULL
		},
	}

	return &object.Hash{
		Pairs: map[string]object.Object{
			"send":     resSend,
			"json":     resJson,
			"html":     resHtml,
			"text":     resText,
			"header":   resHeader,
			"file":     resFile,
			"redirect": resRedirect,
			"cors":     resCors,
		},
	}
}

func buildReqObject(r *http.Request) *object.Hash {
	body, _ := io.ReadAll(r.Body)
	var bodyObj interface{}
	json.Unmarshal(body, &bodyObj)

	queryParams := &object.Hash{Pairs: map[string]object.Object{}}
	for k, v := range r.URL.Query() {
		if len(v) == 1 {
			queryParams.Pairs[k] = &object.String{Value: v[0]}
		} else {
			elements := make([]object.Object, len(v))
			for i, val := range v {
				elements[i] = &object.String{Value: val}
			}
			queryParams.Pairs[k] = &object.Array{Elements: elements}
		}
	}

	bodyVal := goTypeToBaseObject(bodyObj)
	if bodyVal == nil || bodyVal.Type() == object.ERROR_OBJ {
		bodyVal = &object.String{Value: string(body)}
	}

	return &object.Hash{
		Pairs: map[string]object.Object{
			"method":   &object.String{Value: r.Method},
			"path":     &object.String{Value: r.URL.Path},
			"query":    queryParams,
			"headers":  goTypeToBaseObject(parseHeaders(r.Header)),
			"body":     bodyVal,
			"body_raw": &object.String{Value: string(body)},
		},
	}
}

func RegisterServerBuiltins() {
	builtins["server.listen"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 3 {
				return newError("wrong number of arguments. got=%d, want=3", len(args))
			}
			port, ok1 := args[0].(*object.Integer)
			routePath, ok2 := args[1].(*object.String)
			fn, ok3 := args[2].(*object.Function)

			if !ok1 || !ok2 || !ok3 {
				return newError("arguments to `server.listen` must be (INTEGER, STRING, FUNCTION)")
			}

			mux := getOrCreateMux(port.Value)

			mux.HandleFunc(routePath.Value, func(w http.ResponseWriter, r *http.Request) {
				r.Body = http.MaxBytesReader(w, r.Body, 10<<20) // 10MB max
				reqObj := buildReqObject(r)
				resObj := buildResObject(env, w)
				applyFunction(env, fn, []object.Object{reqObj, resObj})
			})

			startServerOnce(port.Value, mux)
			return TRUE
		},
	}

	for _, method := range []string{"get", "post", "put", "patch", "delete"} {
		m := strings.ToUpper(method)
		builtins["server."+method] = &object.Builtin{
			Fn: func(env *object.Environment, args ...object.Object) object.Object {
				if len(args) != 3 {
					return newError("wrong number of arguments. got=%d, want=3", len(args))
				}
				port, ok1 := args[0].(*object.Integer)
				routePath, ok2 := args[1].(*object.String)
				fn, ok3 := args[2].(*object.Function)

				if !ok1 || !ok2 || !ok3 {
					return newError("arguments must be (INTEGER, STRING, FUNCTION)")
				}

				mux := getOrCreateMux(port.Value)

				// Use Go 1.22+ method-pattern syntax: "GET /path"
				pattern := m + " " + routePath.Value
				mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
					r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
					reqObj := buildReqObject(r)
					resObj := buildResObject(env, w)
					applyFunction(env, fn, []object.Object{reqObj, resObj})
				})

				startServerOnce(port.Value, mux)
				return TRUE
			},
		}
	}

	builtins["server.static"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			port, ok1 := args[0].(*object.Integer)
			dir, ok2 := args[1].(*object.String)

			if !ok1 || !ok2 {
				return newError("arguments to `server.static` must be (INTEGER, STRING)")
			}

			mux := getOrCreateMux(port.Value)
			fs := http.FileServer(http.Dir(dir.Value))

			mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
				fullPath := filepath.Join(dir.Value, filepath.Clean(r.URL.Path))

				info, err := os.Stat(fullPath)
				if err != nil {
					w.Header().Set("Content-Type", "text/html; charset=utf-8")
					w.WriteHeader(404)
					w.Write([]byte(default404Page()))
					return
				}

				if info.IsDir() {
					indexPath := filepath.Join(fullPath, "index.html")
					if _, err := os.Stat(indexPath); err == nil {
						r.URL.Path = path.Join(r.URL.Path, "index.html")
					}
				}

				fs.ServeHTTP(w, r)
			})

			startServerOnce(port.Value, mux)
			return TRUE
		},
	}

	builtins["server.route"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			routePath, ok1 := args[0].(*object.String)
			fn, ok2 := args[1].(*object.Function)

			if !ok1 || !ok2 {
				return newError("arguments to `server.route` must be (STRING, FUNCTION)")
			}

			// Apply route group prefix
			serverStateMu.RLock()
			fullPath := serverRouteGroup + routePath.Value
			serverStateMu.RUnlock()

			// Use port 0 as a staging area, will be moved on server.start
			mux := getOrCreateMux(0)

			serverStateMu.RLock()
			middlewareCopy := make([]*object.Function, len(serverMiddleware))
			copy(middlewareCopy, serverMiddleware)
			serverStateMu.RUnlock()

			mux.HandleFunc(fullPath, func(w http.ResponseWriter, r *http.Request) {
				r.Body = http.MaxBytesReader(w, r.Body, 10<<20)
				reqObj := buildReqObject(r)
				resObj := buildResObject(env, w)
				for _, mw := range middlewareCopy {
					result := applyFunction(env, mw, []object.Object{reqObj, resObj})
					if result != nil && result.Type() == object.ERROR_OBJ {
						return
					}
				}
				applyFunction(env, fn, []object.Object{reqObj, resObj})
			})

			return TRUE
		},
	}

	builtins["server.start"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			port, ok := args[0].(*object.Integer)
			if !ok {
				return newError("argument to `server.start` must be INTEGER")
			}

			// Move routes from staging mux (port 0) to the real port
			serverMu.Lock()
			if stagingMux, ok := serverMuxes[0]; ok {
				serverMuxes[port.Value] = stagingMux
				delete(serverMuxes, 0)
			}
			serverMu.Unlock()

			mux := getOrCreateMux(port.Value)
			startServerOnce(port.Value, mux)
			return TRUE
		},
	}

	builtins["server.middleware"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 1 {
				return newError("wrong number of arguments. got=%d, want=1", len(args))
			}
			fn, ok := args[0].(*object.Function)
			if !ok {
				return newError("argument to `server.middleware` must be FUNCTION")
			}
			serverStateMu.Lock()
			serverMiddleware = append(serverMiddleware, fn)
			serverStateMu.Unlock()
			return TRUE
		},
	}

	builtins["server.group"] = &object.Builtin{
		Fn: func(env *object.Environment, args ...object.Object) object.Object {
			if len(args) != 2 {
				return newError("wrong number of arguments. got=%d, want=2", len(args))
			}
			prefix, ok1 := args[0].(*object.String)
			fn, ok2 := args[1].(*object.Function)
			if !ok1 || !ok2 {
				return newError("arguments to `server.group` must be (STRING, FUNCTION)")
			}
			// Save and restore route group prefix
			serverStateMu.Lock()
			prevGroup := serverRouteGroup
			serverRouteGroup = prevGroup + prefix.Value
			serverStateMu.Unlock()

			applyFunction(env, fn, []object.Object{})

			serverStateMu.Lock()
			serverRouteGroup = prevGroup
			serverStateMu.Unlock()
			return TRUE
		},
	}
}

func default404Page() string {
	return `<!DOCTYPE html>
<html>
<head><title>404 - Not Found</title>
<style>
body{font-family:system-ui,sans-serif;display:flex;justify-content:center;align-items:center;height:100vh;margin:0;background:#0a0a0a;color:#fff}
.box{text-align:center}
h1{font-size:72px;margin:0;background:linear-gradient(135deg,#667eea,#764ba2);-webkit-background-clip:text;-webkit-text-fill-color:transparent}
p{color:#888;font-size:18px}
a{color:#667eea;text-decoration:none}
</style>
</head>
<body><div class="box"><h1>404</h1><p>This page doesn't exist.</p><p>Powered by <a href="#">B.A.S.E.</a></p></div></body>
</html>`
}

func parseQueryString(raw string) map[string]string {
	result := map[string]string{}
	parsed, _ := url.ParseQuery(raw)
	for k, v := range parsed {
		result[k] = strings.Join(v, ",")
	}
	return result
}
