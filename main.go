package main

import (
	"base/evaluator"
	"base/lexer"
	"base/object"
	"base/parser"
	"base/repl"
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"os/signal"
	"os/user"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Purple = "\033[35m"
	Cyan   = "\033[36m"
	Gray   = "\033[37m"
)

func main() {
	if len(os.Args) < 2 {
		startREPL()
		return
	}

	arg := os.Args[1]

	switch arg {
	case "-v", "--version":
		fmt.Printf("B.A.S.E. version %s\n", object.VERSION)
		checkVersion(false)
	case "update", "--update":
		force := len(os.Args) > 2 && os.Args[2] == "--force"
		updateBase(force)
	case "-e":
		if len(os.Args) < 3 {
			fmt.Println("Usage: base -e \"code\"")
			os.Exit(1)
		}
		evalString(os.Args[2])
	case "help":
		printHelp()
	case "check":
		if len(os.Args) < 3 {
			fmt.Println("Usage: base check <file.base>")
			os.Exit(1)
		}
		checkFile(os.Args[2])
	case "uninstall":
		uninstallBase()
	case "new":
		if len(os.Args) < 3 {
			fmt.Println("Usage: base new <project-name>")
			os.Exit(1)
		}
		scaffoldProject(os.Args[2])
	case "run":
		runFromConfig()
	default:
		if strings.HasSuffix(arg, ".base") {
			runFile(arg)
		} else {
			fmt.Printf("Unknown command: %s\n", arg)
			fmt.Println("Run 'base help' for usage information.")
			os.Exit(1)
		}
	}
}

func startREPL() {
	user, err := user.Current()
	if err != nil {
		panic(err)
	}
	fmt.Printf("%sHello %s!%s This is the %sB.A.S.E.%s programming language (%sv%s%s)!\n", Cyan, user.Username, Reset, Purple, Reset, Gray, object.VERSION, Reset)
	fmt.Printf("Type %sexit%s to quit. Use %sbase help%s for commands.\n", Red, Reset, Yellow, Reset)
	checkVersion(true)
	registerAllBuiltins()
	registerImportHandler()
	repl.Start(os.Stdin, os.Stdout)
}

func registerAllBuiltins() {
	evaluator.RegisterBackendBuiltins()
	evaluator.RegisterJSONBuiltins()
	evaluator.RegisterStdBuiltins()
	evaluator.RegisterListBuiltins()
	evaluator.RegisterDBBuiltins()
	evaluator.RegisterCryptoBuiltins()
	evaluator.RegisterDataBuiltins()
	evaluator.RegisterSSHBuiltins()
	evaluator.RegisterServerBuiltins()
	evaluator.RegisterSystemBuiltins()
	evaluator.RegisterNotifyBuiltins()
	evaluator.RegisterChannelBuiltins()
	evaluator.RegisterWSBuiltins()
	evaluator.RegisterRegexBuiltins()
}

func printHelp() {
	fmt.Printf("\n%sB.A.S.E.%s - %sBackend Automation & Scripting Environment%s (%sv%s%s)\n\n", Purple, Reset, Gray, Reset, Cyan, object.VERSION, Reset)

	fmt.Printf("%sUSAGE:%s\n", Yellow, Reset)
	fmt.Printf("  base                          Start interactive REPL\n")
	fmt.Printf("  base <script.base>            Run a script file\n")
	fmt.Printf("  base -e \"code\"                Evaluate a one-liner\n")
	fmt.Printf("  base -v, --version            Print version\n")
	fmt.Printf("  base update                   Update B.A.S.E. to latest version\n")
	fmt.Printf("  base update --force           Force reinstall current version\n")
	fmt.Printf("  base help                     Show this help menu\n")
	fmt.Printf("  base check <file.base>        Check syntax without executing\n")
	fmt.Printf("  base new <name>               Scaffold a new project\n")
	fmt.Printf("  base run                      Run project from base.json\n")
	fmt.Printf("  base uninstall                Remove base from system\n\n")

	fmt.Printf("%sLANGUAGE FEATURES:%s\n", Yellow, Reset)
	fmt.Printf("  %snull%s             null keyword, null coalescing (??)\n", Cyan, Reset)
	fmt.Printf("  %stemplate strings%s `Hello ${name}` with expression interpolation\n", Cyan, Reset)
	fmt.Printf("  %sarrow functions%s  (x) => x * 2, (a, b) => { return a + b }\n", Cyan, Reset)
	fmt.Printf("  %sdefault params%s   function(x, msg = \"Hello\") { ... }\n", Cyan, Reset)
	fmt.Printf("  %sdestructuring%s    let {a, b} = hash, let [x, y] = array\n", Cyan, Reset)
	fmt.Printf("  %sspread operator%s  [...list1, ...list2]\n", Cyan, Reset)
	fmt.Printf("  %smatch/switch%s     match value { case 1: { ... } default: { ... } }\n", Cyan, Reset)
	fmt.Printf("  %senums%s            enum Color { RED, GREEN, BLUE }\n", Cyan, Reset)
	fmt.Printf("  %srange%s            range(0, 10), range(0, 10, 2)\n\n", Cyan, Reset)

	fmt.Printf("%sCORE MODULES:%s\n", Yellow, Reset)
	fmt.Printf("  %shttp%s      get, post, put, patch, delete, ping\n", Cyan, Reset)
	fmt.Printf("  %sdb%s        connect, query, insert, insert_many, update, delete, exec, aggregate\n", Cyan, Reset)
	fmt.Printf("  %sserver%s    listen, static, route, start, get, post, put, patch, delete, middleware, group\n", Cyan, Reset)
	fmt.Printf("  %sfile%s      read, write, append, exists, delete, list, mkdir, replace, json_update\n", Cyan, Reset)
	fmt.Printf("  %scrypto%s    uuid, hash, encrypt_file, decrypt_file\n", Cyan, Reset)
	fmt.Printf("  %ssys%s       exec, timestamp, version\n", Cyan, Reset)
	fmt.Printf("  %smath%s      abs, sqrt, pow, round, sin, cos, log, min, max, floor, ceil, random\n", Cyan, Reset)
	fmt.Printf("  %sstring%s    upper, lower, replace, slice, pad_left, split, trim, contains, starts_with, ends_with, index_of\n", Cyan, Reset)
	fmt.Printf("  %slist%s      map, filter, sort, contains, length, reverse, reduce, join, slice, flat, push, index_of\n", Cyan, Reset)
	fmt.Printf("  %sjson%s      parse, stringify\n", Cyan, Reset)
	fmt.Printf("  %sencode%s    base64\n", Cyan, Reset)
	fmt.Printf("  %sdecode%s    base64\n", Cyan, Reset)
	fmt.Printf("  %scsv%s       read, write\n", Cyan, Reset)
	fmt.Printf("  %syaml%s      read, write\n", Cyan, Reset)
	fmt.Printf("  %sregex%s     match, find, find_all, replace, split\n", Cyan, Reset)
	fmt.Printf("  %sssh%s       exec remote commands\n", Cyan, Reset)
	fmt.Printf("  %sws%s        connect (WebSocket client)\n", Cyan, Reset)
	fmt.Printf("  %snotify%s    discord, email\n", Cyan, Reset)
	fmt.Printf("  %sschedule%s  recurring jobs\n", Cyan, Reset)
	fmt.Printf("  %schan%s      thread-safe channels\n", Cyan, Reset)
	fmt.Printf("  %sarchive%s   zip\n\n", Cyan, Reset)

	fmt.Printf("%sSERVER (Express-style):%s\n", Yellow, Reset)
	fmt.Printf("  server.route(path, handler)        Register a route\n")
	fmt.Printf("  server.start(port)                 Start server with all routes\n")
	fmt.Printf("  server.get(port, path, handler)    GET-only route (auto-starts)\n")
	fmt.Printf("  server.post(port, path, handler)   POST-only route (auto-starts)\n")
	fmt.Printf("  server.put(port, path, handler)    PUT-only route (auto-starts)\n")
	fmt.Printf("  server.patch(port, path, handler)  PATCH-only route (auto-starts)\n")
	fmt.Printf("  server.delete(port, path, handler) DELETE-only route (auto-starts)\n")
	fmt.Printf("  server.static(port, dir)           Serve static files\n")
	fmt.Printf("  server.middleware(handler)          Add global middleware\n")
	fmt.Printf("  server.group(prefix, fn)           Group routes under a prefix\n")
	fmt.Printf("  server.listen(port, path, handler) Route + auto-start (legacy)\n\n")

	fmt.Printf("%sRESPONSE HELPERS (in handler):%s\n", Yellow, Reset)
	fmt.Printf("  res.send(status, body)     JSON response\n")
	fmt.Printf("  res.json(status, data)     Pretty JSON response\n")
	fmt.Printf("  res.html(status, content)  HTML response\n")
	fmt.Printf("  res.text(status, content)  Plain text response\n")
	fmt.Printf("  res.file(status, path)     File response (auto MIME)\n")
	fmt.Printf("  res.redirect(status, url)  Redirect response\n")
	fmt.Printf("  res.header(key, value)     Set response header\n")
	fmt.Printf("  res.cors(origin?)          Enable CORS headers\n\n")

	fmt.Printf("%sUTILITIES:%s\n", Yellow, Reset)
	fmt.Printf("  log(msg, lvl?)   wait(sec)      type(v)      range(start, end, step?)\n")
	fmt.Printf("  print(args..)    wait_all()     env.get(n)   len(str)\n\n")

	fmt.Printf("%sEXAMPLES:%s\n", Yellow, Reset)
	fmt.Printf("  base script.base              Run a script\n")
	fmt.Printf("  base -e \"print(1 + 2)\"        Quick math\n\n")

	fmt.Printf("For full documentation visit: %shttps://github.com/igorkalen/base%s\n\n", Blue, Reset)
}

func checkFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", filename, err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Printf("Found %d syntax error(s) in %s:\n", len(p.Errors()), filename)
		for _, msg := range p.Errors() {
			fmt.Printf("  ✗ %s\n", msg)
		}
		os.Exit(1)
	}

	fmt.Printf("✓ %s — no syntax errors found\n", filename)
}

func uninstallBase() {
	paths := []string{"/usr/local/bin/base"}
	fmt.Println("Uninstalling B.A.S.E....")
	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			err := os.Remove(p)
			if err != nil {
				fmt.Printf("Error removing %s: %s\n", p, err)
				fmt.Println("Try running with sudo: sudo base uninstall")
				os.Exit(1)
			}
			fmt.Printf("Removed %s\n", p)
		}
	}
	fmt.Println("B.A.S.E. has been uninstalled.")
}

func scaffoldProject(name string) {
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		fmt.Printf("Error: project name must not contain path separators or '..'\n")
		os.Exit(1)
	}
	dir := name
	os.MkdirAll(dir, 0755)

	config := map[string]string{
		"name":  name,
		"entry": "main.base",
	}
	configBytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Printf("Error: failed to generate project config: %s\n", err.Error())
		os.Exit(1)
	}
	os.WriteFile(filepath.Join(dir, "base.json"), configBytes, 0644)

	mainContent := `log("Hello from " + "` + name + `!");
`
	os.WriteFile(filepath.Join(dir, "main.base"), []byte(mainContent), 0644)

	publicDir := filepath.Join(dir, "public")
	os.MkdirAll(publicDir, 0755)
	htmlContent := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>` + name + `</title>
</head>
<body>
    <h1>` + name + `</h1>
    <p>Powered by B.A.S.E.</p>
</body>
</html>`
	os.WriteFile(filepath.Join(publicDir, "index.html"), []byte(htmlContent), 0644)

	serverContent := `server.static(3000, "./public");
log("Server running at http://localhost:3000");
`
	os.WriteFile(filepath.Join(dir, "server.base"), []byte(serverContent), 0644)

	fmt.Printf("\n%s✓%s Created project '%s%s%s'\n", Green, Reset, Cyan, name, Reset)
	fmt.Printf("  %sFiles:%s\n", Yellow, Reset)
	fmt.Printf("    %sbase.json%s       — project config\n", Green, Reset)
	fmt.Printf("    %smain.base%s       — entry point\n", Green, Reset)
	fmt.Printf("    %sserver.base%s     — web server\n", Green, Reset)
	fmt.Printf("    %spublic/index.html%s — simple landing page\n", Green, Reset)
	fmt.Printf("\n  %sGet started:%s\n", Yellow, Reset)
	fmt.Printf("    cd %s && %sbase main.base%s\n", name, Cyan, Reset)
	fmt.Printf("    cd %s && %sbase server.base%s\n\n", name, Cyan, Reset)
}

func runFromConfig() {
	content, err := os.ReadFile("base.json")
	if err != nil {
		fmt.Println("No base.json found in current directory.")
		fmt.Println("Run 'base new <name>' to create a project, or create base.json manually.")
		os.Exit(1)
	}

	var config map[string]string
	if err := json.Unmarshal(content, &config); err != nil {
		fmt.Printf("Error parsing base.json: %s\n", err)
		os.Exit(1)
	}

	entry, ok := config["entry"]
	if !ok {
		fmt.Println("base.json missing 'entry' field.")
		os.Exit(1)
	}

	runFile(entry)
}

func evalString(input string) {
	l := lexer.New(input)
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Woops! We ran into some B.A.S.E. parse errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	registerAllBuiltins()
	registerImportHandler()
	env := object.NewEnvironment()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("B.A.S.E Engine Panic: %v\n", r)
				os.Exit(1)
			}
		}()
		evaluated := evaluator.Eval(program, env)
		if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
			fmt.Println(evaluated.Inspect())
			os.Exit(1)
		}
	}()

	if evaluator.KeepAlive {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
	}
}

func runFile(filename string) {
	content, err := os.ReadFile(filename)
	if err != nil {
		fmt.Printf("Error reading file %s: %s\n", filename, err)
		os.Exit(1)
	}

	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()

	if len(p.Errors()) != 0 {
		fmt.Println("Woops! We ran into some B.A.S.E. parse errors:")
		for _, msg := range p.Errors() {
			fmt.Printf("\t%s\n", msg)
		}
		os.Exit(1)
	}

	registerAllBuiltins()
	registerImportHandler()
	env := object.NewEnvironment()

	func() {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("B.A.S.E Engine Panic: %v\n", r)
				os.Exit(1)
			}
		}()
		evaluated := evaluator.Eval(program, env)
		if evaluated != nil && evaluated.Type() == object.ERROR_OBJ {
			fmt.Println(evaluated.Inspect())
			os.Exit(1)
		}
	}()

	if evaluator.KeepAlive {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
	}
}

func registerImportHandler() {
	evaluator.ImportHandler = func(path string) (object.Object, error) {
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		l := lexer.New(string(content))
		p := parser.New(l)
		program := p.ParseProgram()

		if len(p.Errors()) != 0 {
			return nil, fmt.Errorf("parse errors in %s", path)
		}

		env := object.NewEnvironment()

		err = func() (err error) {
			defer func() {
				if r := recover(); r != nil {
					err = fmt.Errorf("B.A.S.E Engine Panic: %v", r)
				}
			}()
			evaluator.Eval(program, env)
			return nil
		}()

		if err != nil {
			return nil, err
		}

		return env.Export(), nil
	}
}

func checkVersion(quiet bool) {
	client := &http.Client{
		Timeout: 3 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/repos/igorkalen/base/releases/latest", nil)
	req.Header.Set("User-Agent", "B.A.S.E.-CLI")

	resp, err := client.Do(req)
	if err != nil {
		if !quiet {
			fmt.Printf("%sError checking for updates: %s%s\n", Red, err, Reset)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(object.VERSION, "v")

	if latest != current {
		fmt.Printf("\n%s🚀 Update available: %s%s -> %s%s%s\n", Yellow, Reset, current, Green, latest, Reset)
		fmt.Printf("Run %sbase update%s to upgrade to the latest version.\n\n", Cyan, Reset)
	} else if !quiet {
		fmt.Printf("%sB.A.S.E. is already up to date (v%s).%s\n", Green, object.VERSION, Reset)
	}
}

func updateBase(force bool) {
	fmt.Printf("%sChecking for updates...%s\n", Blue, Reset)

	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://api.github.com/repos/igorkalen/base/releases/latest", nil)
	req.Header.Set("User-Agent", "B.A.S.E.-CLI")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("%sError checking for updates: %s%s\n", Red, err, Reset)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Printf("%sError: GitHub API returned status %d%s\n", Red, resp.StatusCode, Reset)
		return
	}

	var release struct {
		TagName string `json:"tag_name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		fmt.Printf("%sError parsing update info.%s\n", Red, Reset)
		return
	}

	latest := strings.TrimPrefix(release.TagName, "v")
	current := strings.TrimPrefix(object.VERSION, "v")

	if latest == current && !force {
		fmt.Printf("%sB.A.S.E. is already at the latest version (v%s).%s\n", Green, object.VERSION, Reset)
		fmt.Printf("To reinstall anyway, run: %sbase update --force%s\n", Cyan, Reset)
		return
	}

	if force {
		fmt.Printf("%sForce reinstalling v%s...%s\n", Blue, current, Reset)
	} else {
		fmt.Printf("A new version is available: %s%s%s\n", Green, latest, Reset)
		fmt.Printf("Do you want to update? (Y/n): ")
		reader := bufio.NewReader(os.Stdin)
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(strings.ToLower(input))
		if input != "" && input != "y" && input != "yes" {
			fmt.Println("Update cancelled.")
			return
		}
	}

	fmt.Printf("%sUpdating B.A.S.E....%s\n", Blue, Reset)
	cmd := exec.Command("bash", "-c", "curl -fsSL https://raw.githubusercontent.com/igorkalen/base/main/install.sh | bash")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		fmt.Printf("%sUpdate failed: %s%s\n", Red, err, Reset)
		return
	}
}
