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
	server string
}

func NewCLI(args []string) *CLI {
	return &CLI{
		args:   args,
		stdout: os.Stdout,
		stderr: os.Stderr,
		getenv: os.Getenv,
	}
}

// initServer инициализирует server из флага --server или переменной окружения
func (c *CLI) initServer() {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // не выводим ошибки, если флаг не найден

	serverFlag := fs.String("server", c.getenv("ANTIBRUTEFORCE_SERVER"), "")

	// Парсим аргументы, игнорируя ошибки (флаг может отсутствовать)
	_ = fs.Parse(c.args[1:])

	if *serverFlag == "" {
		c.server = "http://localhost:8080"
	} else {
		c.server = *serverFlag
	}
}

func (c *CLI) Run() int {
	// Инициализируем server один раз
	c.initServer()

	if len(c.args) < 2 {
		c.printUsage()
		return 1
	}

	/*
		// Удаляем флаг --server из аргументов для дальнейшего парсинга
		args := c.removeServerFlag()

		// Передаем очищенные аргументы в команды
		switch c.args[1] {
		case "check":
			return c.runCheck(args)
		case "reset":
			return c.runReset(args)
		case "whitelist":
			return c.runWhitelist(args)
		case "blacklist":
			return c.runBlacklist(args)
		default:
			c.printUsage()
			return 1
		}
	*/

	c.printUsage()
	return 1

}

/*
// removeServerFlag удаляет --server и его значение из аргументов
func (c *CLI) removeServerFlag() []string {
	var result []string
	skip := false
	for i := 1; i < len(c.args); i++ {
		if skip {
			skip = false
			continue
		}
		if c.args[i] == "--server" && i+1 < len(c.args) {
			skip = true
			continue
		}
		result = append(result, c.args[i])
	}
	return result
}

func (c *CLI) runCheck(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	login := fs.String("login", "", "")
	password := fs.String("password", "", "")
	ip := fs.String("ip", "", "")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintln(c.stderr, err)
		return 1
	}

	if *login == "" || *password == "" || *ip == "" {
		fmt.Fprintln(c.stderr, "Error: --login, --password, --ip are required")
		return 1
	}

	reqBody := map[string]string{
		"login":    *login,
		"password": *password,
		"ip":       *ip,
	}

	resp, err := c.postJSON(c.server+"/check", reqBody)
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

func (c *CLI) runReset(args []string) int {
	fs := flag.NewFlagSet("reset", flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	login := fs.String("login", "", "")
	ip := fs.String("ip", "", "")

	if err := fs.Parse(args[1:]); err != nil {
		fmt.Fprintln(c.stderr, err)
		return 1
	}

	if *login == "" && *ip == "" {
		fmt.Fprintln(c.stderr, "Error: need --login or --ip")
		return 1
	}

	reqBody := map[string]string{}
	if *login != "" {
		reqBody["login"] = *login
	}
	if *ip != "" {
		reqBody["ip"] = *ip
	}

	resp, err := c.postJSON(c.server+"/reset", reqBody)
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

func (c *CLI) runWhitelist(args []string) int {
	if len(args) < 2 {
		fmt.Fprintln(c.stderr, "Usage: cli whitelist <add|remove|list> [subnet]")
		return 1
	}

	subcmd := args[1]
	switch subcmd {
	case "add":
		return c.whitelistAdd()
	case "remove":
		return c.whitelistRemove(args[2:])
	case "list":
		return c.whitelistList(args[2:])
	default:
		fmt.Fprintf(c.stderr, "Unknown whitelist command: %s\n", subcmd)
		return 1
	}
}

*/

func (c *CLI) parseSubnetCommand(name string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	if err := fs.Parse(c.args); err != nil {
		fmt.Fprintln(c.stderr, err)
		return nil
	}

	// Позиционные аргументы после флагов
	if len(fs.Args()) < 1 {
		fmt.Fprintf(c.stderr, "Usage: cli %s <subnet>\n", name)
		return nil
	}

	return fs
}

func (c *CLI) whitelistAdd() int {
	fs := c.parseSubnetCommand("whitelist add")
	if fs == nil {
		return 1
	}
	subnet := fs.Args()[0]

	reqBody := map[string]string{"subnet": subnet}
	resp, err := c.postJSON(c.server+"/whitelist/add", reqBody)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
		return 1
	}

	fmt.Fprintf(c.stdout, "Added %s to whitelist\n", subnet)
	return 0
}

/*
func (c *CLI) whitelistRemove(args []string) int {
	fs := flag.NewFlagSet("whitelist remove", flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(c.stderr, err)
		return 1
	}

	positional := fs.Args()
	if len(positional) < 1 {
		fmt.Fprintln(c.stderr, "Usage: cli whitelist remove <subnet>")
		return 1
	}

	subnet := positional[0]

	reqBody := map[string]string{"subnet": subnet}
	resp, err := c.postJSON(c.server+"/whitelist/remove", reqBody)
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

func (c *CLI) whitelistList(args []string) int {
	fs := flag.NewFlagSet("whitelist list", flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(c.stderr, err)
		return 1
	}

	resp, err := http.Get(c.server + "/whitelist")
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

func (c *CLI) runBlacklist(args []string) int {
	// аналогично whitelist
	if len(args) < 2 {
		fmt.Fprintln(c.stderr, "Usage: cli blacklist <add|remove|list> [subnet]")
		return 1
	}

	subcmd := args[1]
	switch subcmd {
	case "add":
		return c.blacklistAdd(args[2:])
	case "remove":
		return c.blacklistRemove(args[2:])
	case "list":
		return c.blacklistList(args[2:])
	default:
		fmt.Fprintf(c.stderr, "Unknown blacklist command: %s\n", subcmd)
		return 1
	}
}

func (c *CLI) blacklistAdd(args []string) int {
	fs := flag.NewFlagSet("blacklist add", flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(c.stderr, err)
		return 1
	}

	positional := fs.Args()
	if len(positional) < 1 {
		fmt.Fprintln(c.stderr, "Usage: cli blacklist add <subnet>")
		return 1
	}

	subnet := positional[0]

	reqBody := map[string]string{"subnet": subnet}
	resp, err := c.postJSON(c.server+"/blacklist/add", reqBody)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Fprintf(c.stdout, "Added %s to blacklist\n", subnet)
		return 0
	}

	fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
	return 1
}

func (c *CLI) blacklistRemove(args []string) int {
	fs := flag.NewFlagSet("blacklist remove", flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(c.stderr, err)
		return 1
	}

	positional := fs.Args()
	if len(positional) < 1 {
		fmt.Fprintln(c.stderr, "Usage: cli blacklist remove <subnet>")
		return 1
	}

	subnet := positional[0]

	reqBody := map[string]string{"subnet": subnet}
	resp, err := c.postJSON(c.server+"/blacklist/remove", reqBody)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		fmt.Fprintf(c.stdout, "Removed %s from blacklist\n", subnet)
		return 0
	}

	fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
	return 1
}

func (c *CLI) blacklistList(args []string) int {
	fs := flag.NewFlagSet("blacklist list", flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	if err := fs.Parse(args); err != nil {
		fmt.Fprintln(c.stderr, err)
		return 1
	}

	resp, err := http.Get(c.server + "/blacklist")
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
		fmt.Fprintln(c.stdout, "Blacklist is empty")
	} else {
		fmt.Fprintln(c.stdout, "Blacklist:")
		for _, subnet := range result["subnets"] {
			fmt.Fprintf(c.stdout, "  %s\n", subnet)
		}
	}
	return 0
}

*/

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
    ANTIBRUTEFORCE_SERVER  API server address (default: http://localhost:8080)`)
}
