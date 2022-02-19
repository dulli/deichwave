# Configuration Files

The `/config` folder holds all runtime configuration files that can be used by the compiled executables. It normally has to be shipped along side the executables.

## Main Configuration: `default.toml`

Application settings are configured using either `*.toml` files or environment variables (which have precedence over file-based configuration). They are loaded using [`cleanenv`](https://github.com/ilyakaznacheev/cleanenv) and follow the structure defined in the `common` package's `config.go` file. `default.toml` is both an example configuration file where everything that was commented out is an optional setting with the respective default values. LED group configuration has to always be supplied in the configuration file as it has no sensible default value and is not really fit to map to environment variables.

## TODOs

- [ ] Add a `systemd` service unit file
