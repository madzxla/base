<p align="center">
  <img src="logo.png" alt="B.A.S.E. Logo" width="300">
</p>

# B.A.S.E.

B.A.S.E. (Backend Automation & Scripting Environment) is a simple, standalone programming language for writing backend logic and server scripts.

It's built in Go and compiles to a single 22MB binary. There are no dependencies to install, no `npm`, and no `pip`. Just download the binary and you can run HTTP servers, query databases, send Discord webhooks, and automate your machine.

## Installation

Run this to download the binary and install it to `/usr/local/bin`:

```bash
curl -fsSL https://raw.githubusercontent.com/igorkalen/base/main/install.sh | bash
```

Check if it works:
```bash
base --version
```

## How to use it

Start the interactive REPL to test things out:
```bash
base
```

Run a file:
```bash
base my_script.base
```

Create a new web server project (creates a folder with a `server.base` file):
```bash
base new my-api
cd my-api
base run
```

Run a quick one-liner:
```bash
base -e 'print("Hello world")'
```

---

## What can it do?

B.A.S.E. comes with a big standard library. Everything below works out of the box.

### Web Server (Express-style)
Build APIs with an Express-like routing system. Define routes, then start your server.

```base
// Express-style routing — define routes, then start
server.route("/", function(req, res) {
    res.html(200, "<h1>Welcome to B.A.S.E.</h1>")
})

server.route("/api/users", function(req, res) {
    res.cors()  // enable CORS
    res.json(200, {"users": ["Igor", "Alice"]})
})

server.route("/health", function(req, res) {
    res.json(200, {"status": "ok"})
})

server.start(3000)
```

Method-specific routes and static file serving:

```base
// Serve a folder of HTML/CSS/JS files
server.static(8080, "./public")

// Method-specific routes (auto-start)
server.get(3000, "/api/items", function(req, res) {
    res.json(200, [{"id": 1, "name": "Item 1"}])
})

server.post(3000, "/api/items", function(req, res) {
    let body = req.body
    res.json(201, {"created": body})
})
```

Response helpers available in every handler:

| Helper | Description |
|---|---|
| `res.send(status, body)` | JSON response |
| `res.json(status, data)` | Pretty JSON response |
| `res.html(status, html)` | HTML response |
| `res.text(status, text)` | Plain text response |
| `res.file(status, path)` | File with auto MIME type |
| `res.redirect(status, url)` | HTTP redirect |
| `res.header(key, value)` | Set custom header |
| `res.cors(origin?)` | Enable CORS (`*` by default) |

Request object available in every handler:

| Property | Description |
|---|---|
| `req.method` | HTTP method (GET, POST, etc.) |
| `req.path` | Request URL path |
| `req.query` | Query parameters as hash |
| `req.headers` | Request headers as hash |
| `req.body` | Parsed JSON body |
| `req.body_raw` | Raw body as string |

### HTTP Client
Send requests to other APIs. It supports custom headers, timeouts, and retries.

```base
let data = http.get("https://api.example.com/data", {
    "headers": {"Authorization": "Bearer token123"},
    "timeout": 5,
    "retries": 3
})
```

### JSON
Parse and stringify JSON data.

```base
let obj = json.parse('{"name": "Igor", "age": 25}')
print(obj.name)

let str = json.stringify({"key": "value"})
print(str)
```

### Databases
Connect to SQLite, MySQL, PostgreSQL, or MongoDB natively.

```base
db.connect("main", "sqlite", "./data.db")
db.exec("main", "CREATE TABLE users (name TEXT)")
db.insert("main", "users", {"name": "Igor"})
db.insert_many("main", "users", [{"name": "Alice"}, {"name": "Bob"}])

let users = db.query("main", "SELECT * FROM users")
```

### Notifications (Discord & Email)
Send messages directly to a Discord channel or via an SMTP email server.

```base
// Send a message to a Discord webhook
notify.discord("https://discord.com/api/webhooks/...", "Server deployment finished!")

// Send an email (host, port, username, password, to_email, subject, body)
notify.email("smtp.gmail.com", "587", "you@gmail.com", "password", "target@email.com", "Alert", "Server is down!")
```

### Background Tasks (Concurrency)
Spawn tasks to run in the background. `chan()` creates a thread-safe channel to collect the results.

```base
let results = chan()

spawn function() {
    let ping = http.ping("https://google.com")
    results.send(ping)
}()

wait_all()
print(results.read_all())
```

> **Note on Background Tasks:** B.A.S.E. has an **Auto Keep-Alive** system. If you start an HTTP server or schedule a cron job, the engine automatically detects it and keeps your script running forever. You never have to write messy `while true` or `wait()` loops to prevent your app from exiting!

### Files & Encryption
Read/write files, generate UUIDs, and encrypt files using AES-256-GCM.

```base
let id = crypto.uuid()

// Encrypt a file
crypto.encrypt_file("aes-256-gcm", "secrets.json", "my-password")

// Read a file
let content = file.read("config.json")
```

### Cron Jobs
Run functions automatically on a schedule.

```base
schedule("0 * * * *", function() {
    log("Running hourly task", "INFO")
})
```

---

## Syntax Basics

The syntax looks like JavaScript or C, but uses `let` for variables.

```base
// Variables
let username = "Igor"
let active = true
let config = {"debug": true, "retries": 3}
let tags = ["dev", "admin"]

// Null & null coalescing
let x = null
let val = x ?? "default"  // "default"

// String escape sequences: \n \t \r \\
let multiline = "line1\nline2\nline3"

// Template strings
let greeting = `Hello ${username}, you have ${len(tags)} tags!`

// Functions
let calculate = function(x, y) {
    return x + y
}

// Default parameters
let greet = function(who, msg = "Hello") {
    return msg + ", " + who + "!"
}

// Arrow functions
let double = (x) => x * 2
let add = (a, b) => a + b
let block = (x) => {
    let y = x * 10
    return y + 1
}

// If / else if / else
if active {
    print("Welcome " + username)
} else if username == "guest" {
    print("Hello guest")
} else {
    print("Unknown user")
}

// Ternary
let msg = active ? "online" : "offline"

// Match/switch expressions
let result = match msg {
    case "online": { "User is here" }
    case "offline", "away": { "User is gone" }
    default: { "Unknown" }
}

// Loops
foreach t in tags {
    print(t)
}

foreach i, t in tags {
    print(i + ": " + t)
}

for (let i = 0; i < 5; i = i + 1) {
    print(i)
}

foreach i in range(0, 10, 2) {
    print(i)  // 0, 2, 4, 6, 8
}

while active {
    print("running")
    active = false
}

// Destructuring
let {name, age} = {"name": "Igor", "age": 25}
let [first, second] = [10, 20]

// Spread operator
let merged = [...[1, 2], ...[3, 4]]  // [1, 2, 3, 4]

// Enums
enum Color { RED, GREEN, BLUE }
print(Color.RED)  // 0

// Try/catch + throw
try {
    throw "something went wrong"
} catch (err) {
    print(err.message)
}

// Imports
import "utils.base" as utils

// Concurrency
spawn someFunction()
wait_all()

// Globals (accessible across scopes)
global API_KEY = "abc123"
```

### Server Middleware & Route Groups

```base
// Global middleware — runs before every route handler
server.middleware(function(req, res) {
    res.header("X-Powered-By", "B.A.S.E.")
})

// Group routes under a prefix
server.group("/api/v1", function() {
    server.get(3000, "/users", function(req, res) {
        res.json(200, ["Igor", "Alice"])
    })
    server.get(3000, "/posts", function(req, res) {
        res.json(200, [])
    })
})

server.start(3000)
```

### Regex

```base
let matched = regex.match("[0-9]+", "abc123")  // true
let found = regex.find("[0-9]+", "abc123def")  // "123"
let all = regex.find_all("[0-9]+", "a1b2c3")   // ["1", "2", "3"]
let replaced = regex.replace("[0-9]", "abc123", "X")  // "abcXXX"
let parts = regex.split("[,;]", "a,b;c")       // ["a", "b", "c"]
```

## Full list of built-ins
To see everything the language can do, type `base help` in your terminal. It lists all modules for `http`, `db`, `file`, `math`, `list`, `crypto`, `string`, `csv`, `yaml`, `regex`, `ws`, `ssh`, and more.

### Standard Library Quick Reference

| Module | Functions |
|---|---|
| `http` | get, post, put, patch, delete, ping |
| `db` | connect, query, insert, insert_many, update, delete, exec, aggregate |
| `server` | route, start, get, post, put, patch, delete, static, listen, middleware, group |
| `file` | read, write, append, exists, delete, list, mkdir, replace, json_update |
| `crypto` | uuid, hash, encrypt_file, decrypt_file |
| `sys` | exec, timestamp, version |
| `math` | abs, sqrt, pow, round, sin, cos, log, min, max, floor, ceil, random |
| `string` | upper, lower, replace, slice, pad_left, split, trim, contains, starts_with, ends_with, index_of |
| `list` | map, filter, sort, contains, length, reverse, reduce, join, slice, flat, push, index_of |
| `json` | parse, stringify |
| `csv` | read, write |
| `yaml` | read, write |
| `regex` | match, find, find_all, replace, split |
| `encode` / `decode` | base64 |
| `ssh` | exec |
| `ws` | connect |
| `notify` | discord, email |
| `chan` | create thread-safe channels |
| `archive` | zip |
| `env` | get |
| Utilities | print, log, len, type, range, wait, wait_all, schedule |

## Contact Me

If you have any questions, feel free to reach out to me at [igor@igorkalen.dev](mailto:igor@igorkalen.dev).

## Contributing

1. Clone the repo: `git clone https://github.com/igorkalen/base.git`
2. Install dependencies: `go mod download`
3. Make changes (built-ins are in `/evaluator`, language rules are in `/parser` and `/lexer`)
4. Test it locally: `go run main.go tests/my_test.base`
5. Open a Pull Request!
