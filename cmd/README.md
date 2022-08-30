# Commands

All executables (i.e. `main` packages, also those used for testing purposes) are located in this folder. The `.go`-files themselves as well as their parent folders are named after the desired executable names to simplify the `VS Code` build tasks used to compile them.

## Start Deichwave with a REST Server enabled: `deichwave`

Initializes all modules and necessary hardware before starting a `REST` API server that can be used to control the modules.

## Test and preview the light effects in a GUI: `deichwave-lighttest`

Initializes the `lights` module and creates a GUI to preview a light effect (which has to be supplied as the first argument).
