# Protobufs

This is the public protocol buffers API for [slothd](https://github.com/gjermundgaraba/slothchain).

## Download

The `buf` CLI comes with an export command. Use `buf export -h` for details

#### Examples:

Download cosmwasm protos for a commit:
```bash
buf export buf.build/cosmwasm/slothd:${commit} --output ./tmp
```

Download all project protos:
```bash
buf export . --output ./tmp
```