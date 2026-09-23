# commander-comedor

CLI to check the UGR (Universidad de Granada) comedores universitarios menu.

## Install

```
go install github.com/oppermax/commander-comedor@latest
```

## Usage

```
commander-comedor              # today's menu (default)
commander-comedor tomorrow     # tomorrow's menu
commander-comedor week         # full week

commander-comedor --comedor pts   # only a specific comedor
commander-comedor --veggy         # only the vegetarian/vegan option (Menú 2)
```

Comedores with identical (or near-identical, typos included) menus are merged
in the output.
