# Stampdalf

![stampdalf](./icon/stampdalf.png)

A filesystem timestamp preservation tool that executes commands while keeping
original file access and modification times intact.

🧙 it's a wizard of time and stamps 🧙

Pronounced: stamp-dalf (stamp like time-stamp, dalf like Gandalf the wizard)

## What it does

`stampdalf` allows you to run any command that modifies files in a directory
tree, then automatically resets all timestamps back to their original values. 
Any new files created during command execution are set to Unix epoch 
(1970-01-01) or a custom timestamp via `SOURCE_DATE_EPOCH`.

## Why?

- **Reproducible builds**: Ensure consistent timestamps across build artifacts
- **Build system optimization**: Prevent unnecessary rebuilds triggered by timestamp changes

## Installation

```bash
go install github.com/yourusername/stampdalf@latest
```

Or build from source:

```bash
go build main.go
```

## Usage

```bash
stampdalf <directory> [command...]
```

### Examples

```bash
# Format all Go files without changing timestamps
stampdalf --cd ./src gofmt -w .

# Build a project with reproducible timestamps
stampdalf --cd ./project make build

# Process images without affecting modification times
stampdalf ./images mogrify -resize 50% images/*.jpg

# Set new files to specific timestamp (for reproducible builds)
SOURCE_DATE_EPOCH=1609459200 stampdalf --cd ./dist npm run build
```

## How it works

1. **Scan**: Records all file timestamps (atime/mtime) in the target directory tree with nanosecond precision
2. **Execute**: Runs the specified command with the directory as working directory
3. **Reset**: Restores original timestamps for all pre-existing files; sets new files to Unix epoch or `SOURCE_DATE_EPOCH`
