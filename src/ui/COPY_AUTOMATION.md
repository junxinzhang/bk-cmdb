# Build Output Copy Automation

This directory contains automation scripts to copy the frontend build output from the enterprise build location to the local development web directory.

## Purpose

After running `yarn build` in the `/src/ui` directory, the build output is placed in `src/bin/enterprise/cmdb/web`. For local development convenience, these files need to be copied to `src/web`.

## Available Scripts

### 1. Build and Copy (Recommended)
```bash
yarn build:copy
```
This command will:
1. Run the full build process (`yarn build`)
2. Automatically copy the output to `src/web`

### 2. Copy Only
```bash
yarn copy-to-web
```
This command only copies existing build output from `src/bin/enterprise/cmdb/web` to `src/web`. Use this if you've already built the project and just want to update the local web directory.

### 3. Manual Script Execution
```bash
# Using Node.js script (cross-platform)
node copy-to-web.js

# Using bash script (Unix/Linux/macOS only)
./copy-to-web.sh
```

## How It Works

1. **Source**: `src/bin/enterprise/cmdb/web` (build output location)
2. **Destination**: `src/web` (local development web directory)
3. **Operation**: Complete recursive copy, replacing existing files

## File Structure

```
src/ui/
├── copy-to-web.js          # Cross-platform Node.js copy script
├── copy-to-web.sh          # Bash copy script (Unix systems)
├── package.json            # Updated with new npm scripts
└── COPY_AUTOMATION.md      # This documentation
```

## Prerequisites

- The build output must exist in `src/bin/enterprise/cmdb/web`
- Run `yarn build` first if the build output doesn't exist

## Troubleshooting

### Error: "Source directory does not exist"
This means you need to run the build process first:
```bash
yarn build
```

### Permission Issues (Unix systems)
If you get permission errors, make sure the scripts are executable:
```bash
chmod +x copy-to-web.sh
chmod +x copy-to-web.js
```

## Implementation Details

- The Node.js script (`copy-to-web.js`) is the recommended approach as it works on all platforms
- The bash script (`copy-to-web.sh`) is provided as an alternative for Unix systems
- Both scripts completely replace the destination directory contents
- The scripts include error checking and informative output