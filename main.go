package main

import (
	"anvil/cmd"
	"anvil/internal"
	"fmt"
	"os"
	"slices"
)

type Command = internal.Command

func registerCommand(name string, description string, alias []string, minArgs int, f func(...string)) {
	internal.Commands = append(internal.Commands, internal.Command{
        Name: name,
		Description: description,
        Alias: alias,
        MinArgs: minArgs,
        F: f,
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
	for i, j := range internal.Commands {
		if j.Name == commandName || slices.Contains(j.Alias, commandName){
			runCommand(&internal.Commands[i])
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

	registerCommand("destroy", 		"Destroy (delete) instance.",			[]string{},           		1, cmd.Destroy)
	registerCommand("help", 		"Show help.",      						[]string{"-h", "--help"},  	0, cmd.Help)
	registerCommand("install", 		"Install a mod by slug.",   			[]string{"add", "i"}, 		1, cmd.Install)
	registerCommand("list", 		"List instances.",     					[]string{},           		0, cmd.List)
	registerCommand("login", 		"Login with a Microsoft account.", 		[]string{"signin"},   		0, cmd.Login)
	registerCommand("logout", 		"Log out from Microsoft account.",		[]string{"signout"},  		0, cmd.Logout)
	registerCommand("modlist", 		"List installed mods.",   				[]string{},           		0, cmd.ModList)
	registerCommand("new", 			"Create a new instance.",       		[]string{"create", "n"},   	2, cmd.New)
	registerCommand("run", 			"Run selected instance.",       		[]string{},           		0, cmd.Run)
	registerCommand("search", 		"Search mods on Modrinth.",    			[]string{"find", "s"},     	1, cmd.Search)
	registerCommand("select", 		"Select instance.",    					[]string{"switch"},   		1, cmd.Select)
	registerCommand("status", 		"Show account and selected instance.",  []string{},           		0, cmd.Status)
	registerCommand("uninstall", 	"Uninstall a mod.", 					[]string{"remove", "r"},   	1, cmd.Remove)

	parseArgs()
}
