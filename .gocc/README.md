# Cross-Compilation

The `/.gocc` folder holds configuration files used to set up cross-compilation environments. These are invoked e.g. by the build tasks defined in `/.vscode/tasks.json`.

## `linux/amd64` to `linux/arm64`

To cross-compile on Linux for `ARM64` devices (i.e. a `Raspberry Pi`, which is the targeted platform of this project), the `gcc-aarch64-linux-gnu` toolchain is used, as defined in `/linux/arm64.env`. To install it on a Debian based system, use:

```shell
apt install gcc-aarch64-linux-gnu
```

### Linux Dependencies

On Linux, the CGO modules used for audio playback[^0] and user interfaces[^1] have same dependencies that need to be installed. This is done for both, the host architecture and the cross-compilation target in a similar way, if `Multiarch` support has been enabled[^2]:

```shell
# Host
apt install libasound2-dev libgl1-mesa-dev xorg-dev

# Multiarch
dpkg --add-architecture arm64
apt update
apt install libasound2-dev:arm64 libxxf86vm-dev:arm64 libxinerama-dev:arm64 libxi-dev:arm64 libxcursor-dev:arm64 libxrandr-dev:arm64
```

Additionally, on `Raspberry Pi` devices, outputting the light effects on `ws281x` LEDs requires the `go-rpi-ws281x` libraries to be installed, this is done manually using `scons`[^3]:

```shell
apt install scons
git clone https://github.com/jgarff/rpi_ws281x
cd rpi_ws281x

scons V=yes TOOLCHAIN=aarch64-linux-gnu

cp ws2811.h aarch64-linux-gnu/include/ws2811.h
cp rpihw.h aarch64-linux-gnu/include/rpihw.h
cp pwm.h aarch64-linux-gnu/include/pwm.h
cp libws2811.a aarch64-linux-gnu/lib/libws2811.a
```

## `linux/amd64` to `windows/amd64`

To cross-compile on Linux for `windows` platforms, the compilation environment in `/windows/amd64.env` is adapted to use the `mingw-w64` toolchain[^4]. To install it on a Debian based system, use:

```shell
apt install gcc-aarch64-linux-gnu
```

### Windows Dependencies

As the `Linux to Windows` cross-compilation is only used for testing purposes, dependencies may be missing for some commands at the moment.

## TODOs

- [ ] Find and fix missing windows dependencies

## References

[^0]: [Oto](https://github.com/hajimehoshi/oto)
[^1]: [Fyne](https://github.com/fyne-io/fyne)
[^2]: [Cross-Compiling CGO Projects](https://dh1tw.de/2019/12/cross-compiling-golang-cgo-projects/)
[^3]: [Forked version of go-rpi-ws281x that supports recent `rpi_ws281x` releases](https://github.com/dulli/go-rpi-ws281x)
[^4]: [Compiling for windows on Linux](https://stackoverflow.com/a/47061145)
