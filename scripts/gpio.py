#!/usr/bin/env python3
import copy
import logging
import time
from urllib import parse, request

import pigpio

# TODO remove this and replace it with a Go implementation to actually pack everything into one executable

# def configure(debug):
#     import lib.cleanenv as cleanenv

#     cfg = {
#         "config": "config/default.toml",
#         "audio": {
#             "rate": 48000,
#         },
#         "sounds": {
#             "path": "data/sounds/effects",
#             "ext": ".ogg",
#             "randomizer": ".random",
#         },
#         "_prefix": "GPIO_",
#     }
#     cfg = cleanenv.configure(cfg)
#     cfg["debug"] = debug

#     fmt = "%(message)s"
#     if debug:
#         logging.basicConfig(level=logging.DEBUG, format=fmt)
#     else:
#         logging.basicConfig(level=logging.INFO, format=fmt)
#     return cfg


inputs = {
    "buttons": {
        "left": [
            ((1, 1), "i2c"),
            ((1, 2), "i2c"),
            ((1, 3), "i2c"),
            ((1, 4), "i2c"),
            ((1, 5), "i2c"),
            ((1, 6), "i2c"),
        ],
        "right": [
            ((2, 1), "i2c"),
            ((2, 2), "i2c"),
            ((2, 3), "i2c"),
            ((2, 4), "i2c"),
            ((2, 5), "i2c"),
            ((2, 6), "i2c"),
        ],
    },
    "switches": {
        "left": [((1, 7), "i2c"), ((1, 8), "i2c")],
        "right": [((2, 7), "i2c"), ((2, 8), "i2c")],
    },
    "rotaries": {"left": [(7, 25, 8, "gpio")], "right": [(9, 11, 10, "gpio")]},
}

actions = copy.deepcopy(inputs)

actions["buttons"]["left"][0] = ("sounds/Schnapps/play", "")
actions["buttons"]["left"][1] = ("sounds/Halligalli/play", "")
actions["buttons"]["left"][2] = ("sounds/Airhorn/play", "")
actions["buttons"]["left"][3] = ("sounds/Assi/play", "")
actions["buttons"]["left"][4] = ("sounds/Assi Toni/play", "")
actions["buttons"]["left"][5] = ("lights/Strobe/start", "lights/Strobe/stop")

actions["switches"]["left"][0] = ("sounds/Polizei/loop", "sounds/Polizei/unloop")
actions["switches"]["left"][1] = ("lights/Police/start", "lights/Police/stop")

actions["rotaries"]["left"][0] = (
    "system/volume/-5",
    "system/volume/5",
    "music/next",
)

actions["buttons"]["right"][0] = ("sounds/Wolfgang pass auf!/play", "")
actions["buttons"]["right"][1] = ("sounds/Halligalli/play", "")
actions["buttons"]["right"][2] = ("sounds/Meme/play", "")
actions["buttons"]["right"][3] = ("sounds/Hey, geh weg!/play", "")
actions["buttons"]["right"][4] = ("lights/Strobe/start", "lights/Strobe/stop")
actions["buttons"]["right"][5] = ("sounds/Abfahrt/play", "")

actions["switches"]["right"][0] = ("sounds/Pokemon Battle/loop", "sounds/Pokemon Battle/unloop")
actions["switches"]["right"][1] = ("lights/Battle/start", "lights/Battle/stop")

actions["rotaries"]["right"][0] = (
    "system/intensity/-5",
    "system/intensity/5",
    "sounds/Abfahrt/play",
)

# DO NOT EDIT AFTER THIS LINE
DEGLITCH_BUTTON = 10000
DEGLITCH_SWITCH = 10000
DEGLITCH_ROTARY = 2500
DEGLITCH_EXPAND = 10000

PORT_EXPANDERS = [(0x21, 17), (0x20, 27)]

DEICHWAVE_API = "http://127.0.0.1:3000/api/v0/"

_pi = pigpio.pi()
_actions = {}
_callbacks = []
_rotflag = 0
_activated = 0


def _send(cmds):
    if not cmds:
        return

    if not isinstance(cmds, list):
        cmds = [cmds]

    for cmd in cmds:
        url = f"{DEICHWAVE_API}{parse.quote(cmd)}"
        logging.debug(url)
        try:
            req = request.Request(url, method="POST")
            request.urlopen(req)

        except:
            logging.exception("No connection to the Deichwave server")


def _rotary_high(gpio, level, tick):
    global _activated
    global _actions
    global _rotflag

    if not _activated:
        return

    if not gpio in _actions:
        return

    event_left = _actions[gpio][0]
    event_right = _actions[gpio][1]
    if not level:
        _rotflag = 0
        return
    if _rotflag:
        logging.debug("Rotary: Turned left.")
        if event_left:
            _send(event_left)
        return
    if not _rotflag:
        logging.debug("Rotary: Turned right.")
        if event_right:
            _send(event_right)
        return


def _rotary_low(gpio, level, tick):
    global _activated
    global _actions
    global _rotflag

    if not _activated:
        return

    if level:
        _rotflag = 1
    else:
        _rotflag = 0


def _callback(gpio, level, tick):
    global _activated
    global _actions

    if not _activated:
        return

    logging.debug('Button: State for pin %d changed to "%d".', gpio, level)
    if not gpio in _actions:
        return

    if len(_actions[gpio]) == 2:
        event_on = _actions[gpio][1]
        event_off = _actions[gpio][0]
    else:
        event_on = _actions[gpio][2]
        event_off = ""
    if level and event_on:
        _send(event_on)
    if not level and event_off:
        _send(event_off)


def _callback_i2c(gpio, level, tick):
    global _exp
    global _pi

    if not _activated:
        return

    logging.debug("I2C: Received interrupt (%d)", gpio)
    if gpio not in _exp:
        logging.error("I2C: Interrupt (%d) is not linked to an expander", gpio)
        return

    (idx, handle, prev_state) = _exp[gpio]
    new_state = _convert_i2c(_pi.i2c_read_byte(handle))
    for n in range(8):
        if new_state[n] != prev_state[n]:
            _callback(_spoof_pin_i2c(idx, n + 1), int(new_state[n]), tick)
    _exp[gpio] = (idx, handle, new_state)


def _spoof_pin_i2c(idx, pin):
    return idx * 100 + pin


def _convert_i2c(data):
    return [bool(data & (1 << n)) for n in range(8)]


_exp = {}


def _setup():
    global _activated
    global _actions
    global _rotflag
    global _pi
    global _callbacks
    global _exp

    # logging.basicConfig(level=logging.INFO)
    logging.basicConfig(level=logging.DEBUG)

    # Initialize the I2C handlers
    for idx, expander in enumerate(PORT_EXPANDERS):
        try:
            (addr, interrupt) = expander
            handle = _pi.i2c_open(1, addr)
            logging.debug("Registered I2C %d", addr)
            _pi.i2c_write_byte(handle, 0xFF)
            _exp[interrupt] = (idx + 1, handle, _convert_i2c(0xFF))
            _pi.set_pull_up_down(interrupt, pigpio.PUD_UP)
            _pi.set_mode(interrupt, pigpio.INPUT)
            _pi.set_glitch_filter(interrupt, DEGLITCH_EXPAND)
            _pi.callback(interrupt, pigpio.EITHER_EDGE, _callback_i2c)
        except:
            logging.error("Could not register I2C %d, check connection and address", addr)

    # Initialize the buttons
    for side_name, side in inputs["buttons"].items():
        for idx, hw in enumerate(side):
            (hwid, hwtype) = hw
            if hwtype == "i2c":
                hwid = _spoof_pin_i2c(hwid[0], hwid[1])
                _actions[hwid] = actions["buttons"][side_name][idx]
            elif hwtype == "gpio":
                _actions[hwid] = actions["buttons"][side_name][idx]
                _pi.set_pull_up_down(hwid, pigpio.PUD_UP)
                _pi.set_mode(hwid, pigpio.INPUT)
                _pi.set_glitch_filter(hwid, DEGLITCH_BUTTON)
                _callbacks.append(_pi.callback(hwid, pigpio.EITHER_EDGE, _callback))
            else:
                logging.error('Unsupported input type for buttons: "%s"', hwtype)

    # Initialize the switches
    for side_name, side in inputs["switches"].items():
        for idx, hw in enumerate(side):
            (hwid, hwtype) = hw
            if hwtype == "i2c":
                hwid = _spoof_pin_i2c(hwid[0], hwid[1])
                _actions[hwid] = actions["switches"][side_name][idx]
            elif hwtype == "gpio":
                _actions[hwid] = actions["switches"][side_name][idx][::-1]
                _pi.set_pull_up_down(hwid, pigpio.PUD_UP)
                _pi.set_mode(hwid, pigpio.INPUT)
                _pi.set_glitch_filter(hwid, DEGLITCH_SWITCH)
                _callbacks.append(_pi.callback(hwid, pigpio.EITHER_EDGE, _callback))
            else:
                logging.error('Unsupported input type for switches: "%s"', hwtype)

    # Initialize the rotary encoders
    cb = [_rotary_low, _rotary_high, _callback]
    for side_name, side in inputs["rotaries"].items():
        for idx, hw in enumerate(side):
            (hwleft, hwright, hwpress, hwtype) = hw
            if hwtype == "i2c":
                hwid = _spoof_pin_i2c(hwid[0], hwid[1])
                _actions[hwid] = actions["rotaries"][side_name][idx]
            elif hwtype == "gpio":
                pins = [hwleft, hwright, hwpress]
                for cbidx, hwid in enumerate(pins):
                    _actions[hwid] = actions["rotaries"][side_name][idx]
                    _pi.set_pull_up_down(hwid, pigpio.PUD_UP)
                    _pi.set_mode(hwid, pigpio.INPUT)
                    _pi.set_glitch_filter(hwid, DEGLITCH_ROTARY)
                    _callbacks.append(_pi.callback(hwid, pigpio.EITHER_EDGE, cb[cbidx]))
            else:
                logging.error('Unsupported input type for rotary encoders: "%s"', hwtype)

    time.sleep(1)
    _activated = 1
    logging.info("Waiting for GPIO input")
    _loop()


def _loop(interval=10):
    while True:
        time.sleep(interval)


if __name__ == "__main__":
    try:
        _setup()
    except:
        logging.exception("GPIO stopped due to an unhandled exception")
