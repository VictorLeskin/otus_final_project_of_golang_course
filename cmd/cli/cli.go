package main

import (
    "bytes"
    "encoding/json"
    "flag"
    "fmt"
    "io"
    "net/http"
    "os"
)

type CLI struct {
    args   []string
    stdout io.Writer
    stderr io.Writer
    getenv func(string) string
}

func NewCLI(args []string) *CLI {
    return &CLI{
        args:   args,
        stdout: os.Stdout,
        stderr: os.Stderr,
        getenv: os.Getenv,
    }
}

// Run запускает CLI
func (c *CLI) Run() int {
    if len(c.args) < 2 {
        c.printUsage()
        return 1
    }

    switch c.args[1] {
    case "check":
        return c.runCheck()
    case "reset":
        return c.runReset()
    case "whitelist":
        return c.runWhitelist()
    case "blacklist":
        return c.runBlacklist()
    default:
        c.printUsage()
        return 1
    }
}

func (c *CLI) runCheck() int {
    fs := flag.NewFlagSet("check", flag.ContinueOnError)
    fs.SetOutput(c.stderr)
    
    login := fs.String("login", "", "")
    password := fs.String("password", "", "")
    ip := fs.String("ip", "", "")
    server := fs.String("server", c.getenv("ANTIBRUTEFORCE_SERVER"), "")
    
    if err := fs.Parse(c.args[2:]); err != nil {
        fmt.Fprintln(c.stderr, err)
        return 1
    }
    
    if *login == "" || *password == "" || *ip == "" {
        fmt.Fprintln(c.stderr, "Error: --login, --password, --ip are required")
        return 1
    }
    
    if *server == "" {
        *server = "http://localhost:8080"
    }
    
    reqBody := map[string]string{
        "login":    *login,
        "password": *password,
        "ip":       *ip,
    }
    
    resp, err := c.postJSON(*server+"/check", reqBody)
    if err != nil {
        fmt.Fprintf(c.stderr, "Error: %v\n", err)
        return 1
    }
    defer resp.Body.Close()
    
    if resp.StatusCode != http.StatusOK {
        fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
        return 1
    }
    
    var result map[string]bool
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        fmt.Fprintf(c.stderr, "Error parsing response: %v\n", err)
        return 1
    }
    
    if result["ok"] {
        fmt.Fprintln(c.stdout, "OK: allowed")
        return 0
    } else {
        fmt.Fprintln(c.stdout, "DENIED: brute-force detected")
        return 1
    }
}

func (c *CLI) runReset() int {
    fs := flag.NewFlagSet("reset", flag.ContinueOnError)
    fs.SetOutput(c.stderr)
    
    login := fs.String("login", "", "")
    ip := fs.String("ip", "", "")
    server := fs.String("server", c.getenv("ANTIBRUTEFORCE_SERVER"), "")
    
    if err := fs.Parse(c.args[2:]); err != nil {
        fmt.Fprintln(c.stderr, err)
        return 1
    }
    
    if *login == "" && *ip == "" {
        fmt.Fprintln(c.stderr, "Error: need --login or --ip")
        return 1
    }
    
    if *server == "" {
        *server = "http://localhost:8080"
    }
    
    reqBody := map[string]string{}
    if *login != "" {
        reqBody["login"] = *login
    }
    if *ip != "" {
        reqBody["ip"] = *ip
    }
    
    resp, err := c.postJSON(*server+"/reset", reqBody)
    if err != nil {
        fmt.Fprintf(c.stderr, "Error: %v\n", err)
        return 1
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusOK {
        fmt.Fprintln(c.stdout, "Reset successful")
        return 0
    }
    
    fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
    return 1
}

func (c *CLI) runWhitelist() int {
    if len(c.args) < 3 {
        fmt.Fprintln(c.stderr, "Usage: cli whitelist <add|remove|list> [subnet]")
        return 1
    }
    
    subcmd := c.args[2]
    switch subcmd {
    case "add":
        return c.whitelistAdd()
    case "remove":
        return c.whitelistRemove()
    case "list":
        return c.whitelistList()
    default:
        fmt.Fprintf(c.stderr, "Unknown whitelist command: %s\n", subcmd)
        return 1
    }
}

func (c *CLI) whitelistAdd() int {
    fs := flag.NewFlagSet("whitelist add", flag.ContinueOnError)
    fs.SetOutput(c.stderr)
    server := fs.String("server", c.getenv("ANTIBRUTEFORCE_SERVER"), "")
    
    if err := fs.Parse(c.args[3:]); err != nil {
        fmt.Fprintln(c.stderr, err)
        return 1
    }
    
    if fs.NArg() < 1 {
        fmt.Fprintln(c.stderr, "Usage: cli whitelist add <subnet>")
        return 1
    }
    
    subnet := fs.Arg(0)
    if *server == "" {
        *server = "http://localhost:8080"
    }
    
    reqBody := map[string]string{"subnet": subnet}
    resp, err := c.postJSON(*server+"/whitelist/add", reqBody)
    if err != nil {
        fmt.Fprintf(c.stderr, "Error: %v\n", err)
        return 1
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusOK {
        fmt.Fprintf(c.stdout, "Added %s to whitelist\n", subnet)
        return 0
    }
    
    fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
    return 1
}

func (c *CLI) whitelistRemove() int {
    fs := flag.NewFlagSet("whitelist remove", flag.ContinueOnError)
    fs.SetOutput(c.stderr)
    server := fs.String("server", c.getenv("ANTIBRUTEFORCE_SERVER"), "")
    
    if err := fs.Parse(c.args[3:]); err != nil {
        fmt.Fprintln(c.stderr, err)
        return 1
    }
    
    if fs.NArg() < 1 {
        fmt.Fprintln(c.stderr, "Usage: cli whitelist remove <subnet>")
        return 1
    }
    
    subnet := fs.Arg(0)
    if *server == "" {
        *server = "http://localhost:8080"
    }
    
    reqBody := map[string]string{"subnet": subnet}
    resp, err := c.postJSON(*server+"/whitelist/remove", reqBody)
    if err != nil {
        fmt.Fprintf(c.stderr, "Error: %v\n", err)
        return 1
    }
    defer resp.Body.Close()
    
    if resp.StatusCode == http.StatusOK {
        fmt.Fprintf(c.stdout, "Removed %s from whitelist\n", subnet)
        return 0
    }
    
    fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
    return 1
}

func (c *CLI) whitelistList() int {
    fs := flag.NewFlagSet("whitelist list", flag.ContinueOnError)
    fs.SetOutput(c.stderr)
    server := fs.String("server", c.getenv("ANTIBRUTEFORCE_SERVER"), "")
    
    if err := fs.Parse(c.args[3:]); err != nil {
        fmt.Fprintln(c.stderr, err)
        return 1
    }
    
    if *server == "" {
        *server = "http://localhost:8080"
    }
    
    resp, err := http.Get(*server + "/whitelist")
    if err != nil {
        fmt.Fprintf(c.stderr, "Error: %v\n", err)
        return 1
    }
    defer resp.Body.Close()
    
    var result map[string][]string
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        fmt.Fprintf(c.stderr, "Error parsing response: %v\n", err)
        return 1
    }
    
    if len(result["subnets"]) == 0 {
        fmt.Fprintln(c.stdout, "Whitelist is empty")
    } else {
        fmt.Fprintln(c.stdout, "Whitelist:")
        for _, subnet := range result["subnets"] {
            fmt.Fprintf(c.stdout, "  %s\n", subnet)
        }
    }
    return 0
}

func (c *CLI) runBlacklist() int {
    // аналогично whitelist, только пути /blacklist/*
    // ... (можно реализовать позже)
    return 0
}

func (c *CLI) postJSON(url string, body interface{}) (*http.Response, error) {
    jsonData, err := json.Marshal(body)
    if err != nil {
        return nil, err
    }
    return http.Post(url, "application/json", bytes.NewBuffer(jsonData))
}

func (c *CLI) printUsage() {
    fmt.Fprintln(c.stdout, `Anti-BruteForce CLI

Usage:
    cli check --login <login> --password <password> --ip <ip>
    cli reset [--login <login>] [--ip <ip>]
    cli whitelist add <subnet>
    cli whitelist remove <subnet>
    cli whitelist list
    cli blacklist add <subnet>
    cli blacklist remove <subnet>
    cli blacklist list

Environment:
    ANTIBRUTEFORCE_SERVER  API server address (default: http://localhost:8080)
`)
}