markdown

# Anvil
A terminal-based instance and mod manager for FabricMC.

## Installation
```sh
curl -fsSL https://raw.githubusercontent.com/huseyinhealth/anvil/main/install.sh -o /tmp/anvil-install.sh
bash /tmp/anvil-install.sh
```

For system-wide installation:
```sh
bash /tmp/anvil-install.sh --system
```

## Usage
```sh
anvil help
```

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
├── cache/              # Download cache
├── profile.json        # Microsoft account info
└── .selected           # Currently selected instance
```

## Roadmap
- [ ] `anvil update` — update installed mods
- [ ] `anvil autoremove` — remove unused dependencies
- [ ] Forge support
- [ ] CurseForge support
- [ ] Server support

## Contributing
Contributions are welcome! Feel free to open a pull request for anything you'd like to add or fix. Just make sure your PR has a clear description of what it does and why.

## Acknowledgements
- [Modrinth](https://modrinth.com) — mod search and downloads
- [FabricMC](https://fabricmc.net) — mod loader and meta API
- [Adoptium](https://adoptium.net) — Java runtime downloads
- [Mojang](https://minecraft.net) — Minecraft and launcher meta API

## License
GPL-3.0