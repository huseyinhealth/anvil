# Anvil
A terminal-based instance and mod manager for Minecraft. (Only supports Fabric for now)

## Installation
```sh
curl -fsSL https://raw.githubusercontent.com/huseyinhealth/anvil/main/install.sh | bash # install to ~/.local/bin
```
or

```sh
curl -fsSL https://raw.githubusercontent.com/huseyinhealth/anvil/main/install.sh | bash -s -- --system # install to /usr/local/bin
```

## Usage

### Help

```sh
anvil help
```

### Account
```sh
anvil login           # Sign in with Microsoft
anvil logout          # Sign out
anvil status          # Show current account and selected instance
```

### Instance Management
```sh
anvil new <name> <version>        # Create a new Fabric instance
anvil select <name>               # Select an instance
anvil run                         # Launch the selected instance
anvil list                        # List all instances
anvil destroy <name>              # Delete an instance
```

Aliases: `new` → `create`, `select` → `switch`

### Mod Management
```sh
anvil install <slug> [slug...]    # Install one or more mods
anvil uninstall <slug> [slug...]  # Remove one or more mods
anvil modlist                     # List installed mods
anvil search <query>              # Search mods on Modrinth
```

Aliases: `install` → `add`, `uninstall` → `remove`

## Configuration
Anvil stores all data in `~/.anvil/`:
```
~/.anvil/
├── instances/          # Game instances
│   └── <name>/
│       ├── mods/
│       ├── saves/
│       ├── config/
│       ├── assets/
│       ├── libraries/
│       ├── versions/
│       └── anvil.json  # Instance metadata
├── jre/                # Java runtimes (managed by Anvil)
├── profile.json        # Microsoft account info
├── .selected           # Currently selected instance
└── filecache.json      # Download cache
```


## Roadmap
- [ ] `anvil update` — update installed mods
- [ ] `anvil autoremove` — remove unused dependencies
- [ ] Forge support
- [ ] CurseForge support

## Contributing
Contributions are welcome! Feel free to open a pull request for anything you'd like to add or fix. Just make sure your PR has a clear description of what it does and why.

## Acknowledgements
- [Modrinth](https://modrinth.com) — mod search and downloads
- [FabricMC](https://fabricmc.net) — mod loader and meta API
- [Adoptium](https://adoptium.net) — Java runtime downloads
- [Mojang](https://minecraft.net) — Minecraft and launcher meta API

## License
GPL v3
