package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type CLI struct {
	args   []string
	stdout io.Writer
	stderr io.Writer
	server string

	httpClient *http.Client
	ctx        context.Context
	cancel     context.CancelFunc
}

type CheckResult struct {
	Result bool   `json:"result"`
	Error  string `json:"error,omitempty"`
}

func NewCLI(args []string) *CLI {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	return &CLI{
		args:       args,
		stdout:     os.Stdout,
		stderr:     os.Stderr,
		httpClient: &http.Client{},
		ctx:        ctx,
		cancel:     cancel,
	}
}

// initServer парсит флаг --server из аргументов командной строки.
// server является обязательным параметром.
func (c *CLI) initServer() int {
	fs := flag.NewFlagSet("", flag.ContinueOnError)
	fs.SetOutput(io.Discard) // подавляем вывод ошибок.

	serverFlag := fs.String("server", "", "")

	// Парсим аргументы, ищем --server.
	if err := fs.Parse(c.args); err != nil {
		fmt.Fprintln(c.stderr, "failed to parse --server flag")
		return 1
	}

	if *serverFlag == "" {
		fmt.Fprintln(c.stderr, "--server is required")
		return 1
	}

	c.server = *serverFlag
	return 0
}

func (c *CLI) Run() int {
	c.args = c.args[1:] // remove programm name.

	// Инициализируем server один раз.
	if c.initServer() == 1 {
		return 1
	}

	_ = c.removeServerFlag() // --server URL presented. No need to process return code.

	return c.runServerCommand()
}

// removeServerFlag удаляет --server и его значение из аргументов.
// возвращает 0 если нет ошибок, 1 в противном.
func (c *CLI) removeServerFlag() int {
	for i := 0; i < len(c.args); i++ {
		if c.args[i] == "--server" {
			if i+1 < len(c.args) {
				c.args = append(c.args[:i], c.args[i+2:]...)
				return 0
			}
		}
	}
	return 1
}

func (c *CLI) runServerCommand() int {
	if len(c.args) < 1 {
		fmt.Fprintln(c.stderr, "Wrong command line parameters")
		return 1
	}

	// Передаем очищенные аргументы в команды.
	switch c.args[0] {
	case "check":
		c.args = c.args[1:]
		return c.runCheck()
	case "reset":
		c.args = c.args[1:]
		return c.runReset()
	case "whitelist":
		c.args = c.args[1:]
		return c.runWhitelist()
	case "blacklist":
		c.args = c.args[1:]
		return c.runBlacklist()
	default:
		c.printUsage()
		return 1
	}
}

func (c *CLI) runReset() int {
	if len(c.args) < 1 {
		fmt.Fprintln(c.stderr, "Wrong 'cli reset' command line parameters")
		return 1
	}
	subcmd := c.args[0]
	switch subcmd {
	case "login":
		c.args = c.args[1:]
		return c.resetLogin()
	case "ip":
		c.args = c.args[1:]
		return c.resetIP()
	default:
		fmt.Fprintf(c.stderr, "Unknown reset command: %s\n", subcmd)
		return 1
	}
}

func (c *CLI) runWhitelist() int {
	if len(c.args) < 1 {
		fmt.Fprintln(c.stderr, "Wrong 'cli whitelist' command line parameters")
		return 1
	}
	subcmd := c.args[0]
	switch subcmd {
	case "add":
		c.args = c.args[1:]
		return c.whitelistAdd()
	case "remove":
		c.args = c.args[1:]
		return c.whitelistRemove()
	case "list":
		c.args = c.args[1:]
		return c.whitelistList()
	default:
		fmt.Fprintf(c.stderr, "Unknown whitelist command: %s\n", subcmd)
		return 1
	}
}

func (c *CLI) runBlacklist() int {
	if len(c.args) < 1 {
		fmt.Fprintln(c.stderr, "Wrong 'cli blacklist' command line parameters")
		return 1
	}
	subcmd := c.args[0]
	switch subcmd {
	case "add":
		c.args = c.args[1:]
		return c.blacklistAdd()
	case "remove":
		c.args = c.args[1:]
		return c.blacklistRemove()
	case "list":
		c.args = c.args[1:]
		return c.blacklistList()
	default:
		fmt.Fprintf(c.stderr, "Unknown blacklist command: %s\n", subcmd)
		return 1
	}
}

// parseSubnetCommand парсит команды add/remove/reset для whitelist/blacklist/login/ip.
// name - имя команды для справки (например "whitelist add").
// datatype - тип данных: "subnet" или "ip".
// Возвращает FlagSet с распарсенными флагами или nil при ошибке.
func (c *CLI) parseSubnetCommand(name string, parameterType string) *flag.FlagSet {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	if err := fs.Parse(c.args); err != nil {
		fmt.Fprintln(c.stderr, err)
		return nil
	}

	// Позиционные аргументы после флагов.
	if len(fs.Args()) < 1 {
		fmt.Fprintf(c.stderr, "Usage: cli %s <%s>\n", name, parameterType)
		return nil
	}

	return fs
}

func (c *CLI) processServerStatus(resp *http.Response, reportSuccess func()) int {
	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
		return 1
	}

	reportSuccess()
	return 0
}

func (c *CLI) resetLogin() int {
	fs := c.parseSubnetCommand("reset login", "login")
	if fs == nil {
		return 1
	}
	login := fs.Args()[0]

	reqBody := map[string]string{"login": login}
	resp, err := c.postJSON(c.server+"/reset", reqBody)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	return c.processServerStatus(resp, func() {
		fmt.Fprintf(c.stdout, "Reset successful for login: %s\n", login)
	})
}

func (c *CLI) resetIP() int {
	fs := c.parseSubnetCommand("reset ip", "ip")
	if fs == nil {
		return 1
	}
	ip := fs.Args()[0]

	reqBody := map[string]string{"ip": ip}
	resp, err := c.postJSON(c.server+"/reset", reqBody)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	return c.processServerStatus(resp, func() {
		fmt.Fprintf(c.stdout, "Reset successful for ip: %s\n", ip)
	})
}

func (c *CLI) whitelistAdd() int {
	fs := c.parseSubnetCommand("whitelist add", "subnet")
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

	return c.processServerStatus(resp, func() {
		fmt.Fprintf(c.stdout, "Added %s to whitelist\n", subnet)
	})
}

func (c *CLI) whitelistRemove() int {
	fs := c.parseSubnetCommand("whitelist remove", "subnet")
	if fs == nil {
		return 1
	}
	subnet := fs.Args()[0]

	reqBody := map[string]string{"subnet": subnet}
	resp, err := c.postJSON(c.server+"/whitelist/remove", reqBody)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	return c.processServerStatus(resp, func() {
		fmt.Fprintf(c.stdout, "Removed %s from whitelist\n", subnet)
	})
}

func (c *CLI) whitelistList() int {
	if len(c.args) != 0 {
		fmt.Fprintf(c.stderr, "Usage: cli whitelist list\n")
		return 1
	}

	resp, err := c.getJSON(c.server + "/whitelist")
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
		return 1
	}

	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(c.stderr, "Error parsing response: %v\n", err)
		return 1
	}

	fmt.Fprintf(c.stdout, "Whitelist: %v\n", result["whitelist"])
	return 0
}

func (c *CLI) blacklistAdd() int {
	fs := c.parseSubnetCommand("blacklist add", "subnet")
	if fs == nil {
		return 1
	}
	subnet := fs.Args()[0]

	reqBody := map[string]string{"subnet": subnet}
	resp, err := c.postJSON(c.server+"/blacklist/add", reqBody)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	return c.processServerStatus(resp, func() {
		fmt.Fprintf(c.stdout, "Added %s to blacklist\n", subnet)
	})
}

func (c *CLI) blacklistRemove() int {
	fs := c.parseSubnetCommand("blacklist remove", "subnet")
	if fs == nil {
		return 1
	}
	subnet := fs.Args()[0]

	reqBody := map[string]string{"subnet": subnet}
	resp, err := c.postJSON(c.server+"/blacklist/remove", reqBody)
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	return c.processServerStatus(resp, func() {
		fmt.Fprintf(c.stdout, "Removed %s from blacklist\n", subnet)
	})
}

func (c *CLI) blacklistList() int {
	if len(c.args) != 0 {
		fmt.Fprintf(c.stderr, "Usage: cli blacklist list\n")
		return 1
	}

	resp, err := c.getJSON(c.server + "/blacklist")
	if err != nil {
		fmt.Fprintf(c.stderr, "Error: %v\n", err)
		return 1
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(c.stderr, "Server error: %s\n", resp.Status)
		return 1
	}

	var result map[string][]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(c.stderr, "Error parsing response: %v\n", err)
		return 1
	}

	fmt.Fprintf(c.stdout, "Blacklist: %v\n", result["blacklist"])
	return 0
}

func (c *CLI) parseCheckCommand() (int, *string, *string, *string) {
	if len(c.args) != 6 {
		fmt.Fprintln(c.stderr, "Usage: cli check --login <login> --password <password> --ip <ip>")
		return 1, nil, nil, nil
	}

	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(c.stderr)

	login := fs.String("login", "", "")
	password := fs.String("password", "", "")
	ip := fs.String("ip", "", "")

	if err := fs.Parse(c.args); err != nil {
		fmt.Fprintln(c.stderr, err)
		return 1, nil, nil, nil
	}

	if *login == "" || *password == "" || *ip == "" {
		fmt.Fprintln(c.stderr, "Error: --login, --password, --ip are required")
		return 1, nil, nil, nil
	}

	return 0, login, password, ip
}

func (c *CLI) runCheck() int {
	ret, login, password, ip := c.parseCheckCommand()
	if ret == 1 {
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

	var result CheckResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		fmt.Fprintf(c.stderr, "Error parsing response: %v\n", err)
		return 1
	}

	if result.Error != "" {
		fmt.Fprintf(c.stderr, "%s\n", result.Error)
		return 1
	}

	if !result.Result {
		fmt.Fprintln(c.stdout, "DENIED: brute-force detected")
		return 1
	}

	fmt.Fprintln(c.stdout, "OK: allowed")
	return 0
}

func (c *CLI) postJSON(url string, body interface{}) (*http.Response, error) {
	jsonData, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(c.ctx, "POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	return c.httpClient.Do(req)
}

func (c *CLI) getJSON(url string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(c.ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}

	return c.httpClient.Do(req)
}

func (c *CLI) printUsage() {
	fmt.Fprintln(c.stdout, `Anti-BruteForce CLI

Usage:
    cli check --login <login> --password <password> --ip <ip>
    
    cli reset login <login>
    cli reset ip <ip>
    
    cli whitelist add <subnet>
    cli whitelist remove <subnet>
    cli whitelist list
    
    cli blacklist add <subnet>
    cli blacklist remove <subnet>
    cli blacklist list

Options:
    --server <url>  API server address (required)`)
}
