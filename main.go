package main

import (
	"anvil/cmd"
	"anvil/internal"
	"fmt"
	"os"
	"slices"
)

type Command = internal.Command

var commands []Command

func registerCommand(name string, alias []string, minArgs int, f func(...string)) {
	commands = append(commands, internal.Command{
        Name:  name,
        Alias: alias,
        MinArgs:  minArgs,
        F:     f,
    })
}

func checkArgLen(c *Command) bool {
	return len(os.Args) >= c.MinArgs + 2
}

func runCommand(command *Command) {
	if checkArgLen(command) {
		command.F(os.Args[2:]...)
		return
	}

	fmt.Fprintf(os.Stderr, "Expected minimum of %d arguments. (got %d)\n", command.MinArgs, len(os.Args[2:]))
	os.Exit(1)
}

func parseArgs() {
	commandName := os.Args[1]
	for i, j := range commands {
		if j.Name == commandName || slices.Contains(j.Alias, commandName){
			runCommand(&commands[i])
			return
		}
	}

	fmt.Fprintln(os.Stderr, "Type \"anvil help\" for help.")
	os.Exit(1)
}

func main() {
	internal.Init()
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "Type \"anvil help\" for help.")
		os.Exit(1)
	}

	registerCommand("destroy",   []string{},           1, cmd.Destroy)
	registerCommand("help",      []string{},           0, cmd.Help)
	registerCommand("install",   []string{"add"},      1, cmd.Install)
	registerCommand("list",      []string{},           0, cmd.List)
	registerCommand("login",     []string{"signin"},   0, cmd.Login)
	registerCommand("logout",    []string{"signout"},  0, cmd.Logout)
	registerCommand("modlist",   []string{},           0, cmd.ModList)
	registerCommand("new",       []string{"create"},   2, cmd.New)
	registerCommand("run",       []string{},           0, cmd.Run)
	registerCommand("search",    []string{"find"},     1, cmd.Search)
	registerCommand("select",    []string{"switch"},   1, cmd.Select)
	registerCommand("status",    []string{},           0, cmd.Status)
	registerCommand("uninstall", []string{"remove"},   1, cmd.Remove)

	parseArgs()
}
