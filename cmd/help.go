package cmd

import "fmt"

func Help(...string) {
    fmt.Println(`
 █████╗ ███╗   ██╗██╗   ██╗██╗██╗     
██╔══██╗████╗  ██║██║   ██║██║██║     
███████║██╔██╗ ██║██║   ██║██║██║     
██╔══██║██║╚██╗██║╚██╗ ██╔╝██║██║     
██║  ██║██║ ╚████║ ╚████╔╝ ██║███████╗
╚═╝  ╚═╝╚═╝  ╚═══╝  ╚═══╝  ╚═╝╚══════╝

Instance + Mod Manager for FabricMC

USAGE:
  anvil <command> [arguments]

ACCOUNT:
  login                     Sign in with Microsoft
  logout                    Sign out
  status                    Show account and selected instance

INSTANCE MANAGEMENT:
  new, create <name> <ver>  Create a new Fabric instance
  select, switch <name>     Select an instance
  run                       Launch the selected instance
  list                      List all instances
  destroy <name>            Delete an instance

MOD MANAGEMENT:
  install, add <slug...>    Install one or more mods
  uninstall, remove <slug>  Remove a mod
  modlist                   List installed mods
  search, find <query>      Search mods on Modrinth

HELP:
  anvil help                Show this message`)
}
