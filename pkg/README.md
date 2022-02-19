# Packages

Library code broken up into modular components that supply the different functions.

## Common Modules: `/common`

Modules that are shared by multiple other packages are put here, e.g. basic audio functions that both the music and sound-effect subsystem rely on or a really simple event queue that provides a way to dispatch and listen to events across packages.

## Light Effects: `/lights`

Provides everything required to render light effects, i.e. output one-dimensional arrays of color based on some kind of ruleset defining what they look like. Actually displaying these effects is then up to the specific implementation using this package (see e.g. the `bbycr-lighttest` command). For flexibility, the light effects are defined as `Tengo` scripts[^0] that calculate each frame at runtime, so they can be changed, added and deleted without recompilation. See the `/data/lights/effects` subdirectory for examples.

## Sound Board: `/sounds`

Sound effects can be played once, or looped and un-looped. If multiple sound files make up one sound effect, they are either played sequentially or at random. They are taken from `/data/sounds/processed` and buffered into memory for minimum latency. See `/tools/process_sounds.py` for the pre-processing that is applied.

## Music Player: `/music`

The music player is meant to play random songs from multiple playlists, choosing the next playlist at pre-defined, adjustable probabilities. It provides basic functionalities like forwarding, pausing and skipping songs for each playlist. By default, every playlist is a separate folder full with audio files (normally `*.mp3`) in the `/data/music/playlists` subdirectory.

## REST API Server: `/rest`

The `REST` API server is based on an `OpenAPI` specification (see `/api`) and exposes all necessary functions of the other packages over a network/for HTTP requests. This powers the web interface located in `/web`, which is basically a static `HTML` site that performs all actions by client-side calls to the API.

## Hardware drivers: `/hardware`

Specific hardware implementations, e.g. to address LEDs and make them output the rendered light effects, are platform specific and stored in this package. If a driver is not implemented for a platform, it is replaced by a stub that returns the appropriate errors when attempting the hardware setup.

## TODOs

- [ ] Move to `/internal`?

## References

[^0]: [The Tengo Language](https://github.com/d5/tengo)
