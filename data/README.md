# Data

This folder holds all user data and is normally deployed alongside the executable commands.

## Sound Effects: `/sounds/processed`

Sound effects need to be pre-process (see `/tools`) and are then put here, individual files directly in `/sounds/processed` are treated as single sound files that always play in the same way. If multiple files are bundled into subdirectories, each subdirectory is a single sound with multiple variations, that play either randomly or sequentially.

## Playlists: `/music/playlists`

Each folder in `/music/playlists` is treated as an individual playlist of music files. As resampling is not implemented for music yet, they all need have the same format and sampling rate.

## Light Effects: `/lights/effects`

A light effect is a `*.tengo` script[^0] that exports a function to render the next effect frame, using the following signature:

```golang
export {
    info: {
        maxtick: 256,
        frametime: 0.05 // [s]
    },
    frame: func(leds, tick){
        for group in leds {
            for idx:=0; idx<group.count; idx++ {
                group.brightness[idx] = ...
                group.color[idx] = ...
            }
        }
        return leds
    }
}
```

Most included light effects are the usual suspects, like Larson scanners, rainbows and color fades[^1].

### Config Variables

- `maxtick`: Maximum number to which the animation frame index will be counted before resetting to `0`, if `0`, it won't ever be increased (e.g. for constant effects like solid colors)
- `frametime`: Seconds for which each frame will be visible, if `0` it is shown until the next frame is requested manually (e.g. for constant effects like solid colors)

### Input Variables

- `tick`: Index number of the current animation frame, if this reaches `maxtick`, it resets to `0`
- `leds`: An object defining the LED setup on which the effect will be ultimately displayed, they are split into named groups with a fixed LED count to each of which a brightness (where 0.0 is black, 2.0 is white and 1.0 is the nominal color) and a color index (that maps to an internal rainbow color palette with 256 colors) is attached

```golang
leds = {
    "front-left": {
        count: 5,
        brightness: [1.0, 1.0, 1.0, 1.0, 1.0],
        color: [0, 0, 0, 0, 0]
    }
}
```

## References

[^0]: [The Tengo Language](https://github.com/d5/tengo)
[^1]: [Tweaking4All - LEDStrip effects for NeoPixel and FastLED](https://www.tweaking4all.com/hardware/arduino/adruino-led-strip-effects/)
