<!-- markdownlint-configure-file {
  "MD013": false,
  "MD033": false,
  "MD041": false
} -->

<div align="center">
<img alt="Deichwave Project Logo" src="web/public/static/logo.svg">

# Deichwave

[![GitHub license](https://badgen.net/github/license/dulli/deichwave)](https://github.com/dulli/deichwave/blob/master/LICENSE)
[![GitHub release](https://img.shields.io/github/release/dulli/deichwave.svg)](https://github.com/dulli/deichwave/releases/)
[![GoReportCard](https://goreportcard.com/badge/github.com/dulli/deichwave)](https://goreportcard.com/report/github.com/dulli/deichwave)

Deichwave is an **automated**, but interactive, **media player** powering your parties with music, light and sound-effects.

Even on-the-go.

[Getting started](#getting-started) •
[Background](#background) •
[Structure](#structure) •
[GUI](#web-interface) •
[BBYCR](#the-original-vehicle)

</div>

## Getting Started

### Setup

Copy the compiled executable as to your target computer and run it from the directory that contains your `data/sounds` and `data/music` folders (see [below](#runtime-environment)), e.g. using the command line:

```shell
./deichwave
```

Then monitor the log output to see if everything is up and running and for instructions to access [the web interface](#web-interface).

<!-- ## Configuration -->

## Background

The **BBYCR** is a [purpose-built vehicle](#the-original-vehicle) used for outdoor parties on the move (e.g. for [Kohltouren](https://en.wikipedia.org/wiki/Gr%C3%BCnkohlessen) or the traditional hiking tour on [German Father's Days](https://en.m.wikipedia.org/wiki/Father%27s_Day#Germany)). It is equipped with speakers to play music (continously, without requiring interaction while also being able to easily adjust what playlists are used) and sound effects (on demand), as well as being lit up with addressable LED strips (in addition to providing a way to bring along sufficient amounts of beer). It uses a `Raspberry Pi 3` as its on-board computer.

To power all this on a software level, `Deichwave` was born after multiple previous attempts that simple bolted together a bunch of Python scripts with existing software like `MPD`, `pigpio` and others. As this proved to be too difficult to setup (and use for non-technical friends that borrowed the vehicle) and also not robust enough to survive the, sometimes day-long, deployments, it was first turned into multiple `Go` microservices that were again bolted onto the existing scripts and external software before being rewritten into its current form from scratch after two years of disappointing mid-party music failures.

This time it is supposed to be a portable monolith that can compile into a single executable, so that it is as easy as possible to deploy and can even be used directly from a thumb drive that contains the music and sound files that should be played. It also cross-compiles on (and for) Linux and Windows running on `ARM64` and `x64` platforms. So aside from the targeted `Raspberry Pi 3` on the original **BBYCR** you can even use it for your home-party on your regular laptop or wherever else you like.

<!-- ## Development

### Compiling

... [^1] -->

## Structure

This project is structured into multiple folders with different purposes, most with individual additional readmes:

### Runtime Environment

- [`/config`](config/README.md): Application configuration files
- [`/data`](data/README.md): Static assets used by the application

(All of these have to be made available, according to your configuration, to the application when you want to start it)

### Development Environment

- [`/cmd`](cmd/README.md): Source code for the executable commands
- [`/pkg`](pkg/README.md): Library code that houses the actual functionality
- [`/tools`](tools/README.md): Additional tools and scripts
- `/web`: Web interface used to control all important features, will get bundled into the executable during compilation

(These are only required if you want to contribute to the development of `Deichwave`)

### Miscellaneous

- [`/api`](api/README.md): API specifications
- `/bin`: Target folder for the compiled binaries, if the supplied `VS Code` build tasks are used, the resulting binaries will be ordered into subfolders of the format `<os>/<architecture>`
- `/docs`: Assets that contain or support the project's documentation
- `.github`, `.gocc`, `.vscode`: Contain configurations for the development infrastructure

## Web Interface

By default, `Deichwave` can be remote controlled using a web interface hosted at port `:3000`:

<div align="center">
<img alt="Web Interface Screenshot" src="docs/photos/web-screenshot.png">
<small>
Screenshot of the web interface[^0]
</small>
</div>

## The Original Vehicle

Outfitted with main speakers between its wheel base, additional tweeters pointed at the front, a subwoofer at the back, underbody lights, an integrated foldable beer pong table and built-in bar counters, you almost forget that the **BBYCR**'s main purpose is to make sure that you don't run out of drinks on your tour. All of this is controlled by a `Rasbperry Pi` in one of it's two trunks, which is also connected to two side-panels that houses multiple buttons, rotary encoders and switches allowing the participants of the tour to access the entertainment functions at any time.

<div align="center">
<img alt="BBYCR Photo at Night" src="docs/photos/bbycr-at-night.png">
</div>

## TODOs

- [x] Add a way to run the main executable automatically at boot
- [x] Create pre-compiled releases
- [ ] Add signed OTA updates/patches
- [ ] Create a Flutter frontend
- [ ] Add tests

## References

[^0]: [Screenshot created with `Screely`](https://www.screely.com/)

<!-- [^1]: [Embedding resources in Windows executables](https://github.com/tc-hib/go-winres) -->
