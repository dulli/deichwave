# Configuration Files

The `/config` folder holds all runtime configuration files that can be used by the compiled executables. It normally has to be shipped along side the executables.

## Main Configuration: `default.toml`

Application settings are configured using either `*.toml` files or environment variables (which have precedence over file-based configuration). They are loaded using [`cleanenv`](https://github.com/ilyakaznacheev/cleanenv) and follow the structure defined in the `common` package's `config.go` file. `default.toml` is both an example configuration file where everything that was commented out is an optional setting with the respective default values. LED group configuration has to always be supplied in the configuration file as it has no sensible default value and is not really fit to map to environment variables.

## Linux Platform Config

### Device Tree

A device tree config to initialize the `PCF2574` I²C-GPIO-expanders, used to add additional buttons and other inputs, is included in this repository. It can be activated from the project root using the the `dtoverlay` tool, e.g. for two expanders with the I²C addresses `0x20` and `0x21` that have their interrupt lines connected to `GPIO17` and `GPIO27`:

```bash
sudo dtoverlay -d ./config/device-tree/ gpio-expander addr=0x20 irq=17
sudo dtoverlay -d ./config/device-tree/ gpio-expander addr=0x21 irq=27
```

For the interrupts to work correctly on an _Raspberry Pi_ you may have to activate the internal pull ups for these pins by adding the following line to your `/boot/config.txt`:

```bash
gpio=17,27=ip,pu
```

After changes to the `gpio-expander.dts` source, it has to be recompiled using `dtc` (install using `sudo apt install device-tree-compiler`):

```bash
dtc -O dtb -o config/device-tree/gpio-expander.dtbo config/device-tree/gpio-expander.dts
```

To permanently apply the device tree overlays for usage, copy the compiled `gpio-expander.dtbo` file to `/boot/overlays/` and add the following lines to the `/boot/config.txt` as well:

```bash
dtoverlay=gpio-expander,addr=0x20,irq=17
dtoverlay=gpio-expander,addr=0x21,irq=27
```
