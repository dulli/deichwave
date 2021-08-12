#!/usr/bin/env python3
import os
import subprocess
import logging
import time
import datetime
from threading import Timer

import pigpio

BOOT_WAIT     = 1 # [s]
RESTART_LIMIT = 5
MUTE_TTS      = False
ENABLE_USB    = True
KEY_GPIO      = 4
KEY_DEGLITCH  = 25000
KEY_GRACETIME = 30 # [s]

sdcard_sounds = '/home/dietpi/bbycrgo/sounds/'
sdcard_music  = '/home/dietpi/bbycrgo/music/'
usb_music     = '/media/usb/bbycr/'

_process_list = []
_pi           = pigpio.pi()
_killtimer    = None
_warntimer    = None
_started      = False
_stopped      = False
_waiting      = True

def _logid():
    return datetime.datetime.now().strftime('%Y-%m-%d-%H-%M')

def _run(cmd):
    logging.debug('RUN "%s"', cmd)
    subprocess.run('{} >> /home/dietpi/bbycrgo/log/run.log'.format(cmd), check=True, shell=True)

def _tts(text):
    if MUTE_TTS:
        return

    logging.debug('TTS "%s"', text)
    _run('pico2wave -w tts.wav "{}" | paplay --volume=32500'.format(text))

def _sound(file, blocking=False):
    logging.debug('SND "%s"', file)
    cmd = 'paplay --volume=40000 {}'.format(file)
    if blocking:
        subprocess.run(cmd, shell=True)
    else:
        subprocess.Popen(cmd, shell=True)


def _link(source, target):
    _run('ln -s "{}" "{}"'.format(source, target))
    logging.info('Linked "%s" to "%s"', source, target)

def _process(name, cmd):
    global _process_list
    process = subprocess.Popen('exec "{}" > "/home/dietpi/bbycrgo/log/{}-{}.log" 2>&1'.format(cmd, name, _logid()), shell=True)
    _process_list.append((process, cmd, name, 0))
    logging.info('Started process "%s" (%s)', name, cmd)

def _watchdog(interval=5):
    global _process_list
    time.sleep(interval)
    while True:
        for i in range(len(_process_list)):
            (process, cmd, name, restarts) = _process_list[i]
            if process.poll() != None:
                restarts = restarts + 1
                if restarts < RESTART_LIMIT:
                    _tts('Restarting {}'.format(name))
                    process = subprocess.Popen('exec "{}" > "/home/dietpi/bbycrgo/log/{}-{}.log" 2>&1'.format(cmd, name, _logid()), shell=True)
                    _process_list[i] = (process, cmd, name, restarts)
                    logging.error('Restarted process "%s" (%s)', name, cmd)
                else:
                    logging.critical('Restarted process "%s" too many times in a row, exiting!', name)
                    _killall()
                    return
            elif restarts > 0:
                restarts = restarts - 1

        time.sleep(interval)

def _killall():
    global _process_list
    for i in range(len(_process_list)):
        (process, cmd, name, restarts) = _process_list[i]
        process.kill()
    _process_list = []
    time.sleep(BOOT_WAIT)

def start():
    global _started
    global _stopped
    _started = True

    # Perform general setup steps and restart PulseAudio
    _killall()
    _run('chmod +x /home/dietpi/bbycrgo/bbycr-*')
    _process('pulse', 'pulseaudio')
    time.sleep(BOOT_WAIT)

    # Initialize MPD and run boot loop sound
    _run('mpc clear')
    _run('mpc add "file:///home/dietpi/bbycrgo/sounds/kitt.mp3"')
    _run('mpc play')

    # Prepare the working directory
    os.chdir('/home/dietpi/bbycrgo/env')
    _run('touch start')
    _run('rm *')
    _run('ln -s /dev/stdout tts.wav')

    # Determine if a USB drive with music is connected
    usb_connected = os.path.isdir(usb_music)
    usb_valid     = os.path.isdir(os.path.join(usb_music, 'playlists')) and len([f for f in os.scandir(os.path.join(usb_music, 'playlists')) if f.is_dir()]) == 2
    use_usb       = ENABLE_USB and usb_valid

    # Output the USB status
    if usb_connected:
        _tts('USB found')
        logging.info('USB drive is connected')
    if usb_connected and not usb_valid:
        _tts('No Playlists')
        logging.warn('However, the USB drive does not contain the correct folder structure')

    # Link the proper sound and music files into the current working directory
    _link(sdcard_sounds, 'sounds')
    _link(usb_music if use_usb else sdcard_music, 'music')

    # Allow some time for an interactive user to read the outputs
    _sound('/home/dietpi/bbycrgo/sounds/logon.wav')    
    logging.info('Waiting %d seconds before starting components...', BOOT_WAIT)
    time.sleep(BOOT_WAIT)

    # Launch the BBYCRgo Components and keep track of them
    _process('lights', '/home/dietpi/bbycrgo/bbycr-lights')
    _process('engine', '/home/dietpi/bbycrgo/bbycr-engine')
    _process('inputs', '/home/dietpi/bbycrgo/bbycr-inputs.py')

    # Keep track of the components and restart them if needed
    _watchdog()
    logging.warn('Stopping watchdog loop!')
    _stopped = True

def stop():
    _killall()
    _run('/sbin/poweroff')

def warn():
    global _warntimer
    _warntimer = None
    _sound('/home/dietpi/bbycrgo/sounds/shutdown.wav')

def _callback(gpio, level, tick):
    global _killtimer
    global _warntimer
    global _waiting

    logging.debug('KEY State for pin %d changed to "%d".', gpio, level)

    if not level:
        if _killtimer is not None:
            logging.warn('Key was turned back on, stopping shutdown!')
            _killtimer.cancel()
            _killtimer = None

        if _warntimer is not None:   
            logging.warn('Key was turned back on, stopping warning!')         
            _warntimer.cancel()
            _warntimer = None

        if not _started:
            logging.info('Key was turned to on, starting!')
            _waiting = False

    elif level:
        _sound('/home/dietpi/bbycrgo/sounds/logoff.wav')
        logging.warn('Key was turned to off, stopping in %d s!', KEY_GRACETIME)        
        _killtimer = Timer(KEY_GRACETIME, stop)
        _warntimer = Timer(KEY_GRACETIME/2, warn)
        _killtimer.start()

def _wait_for_key(interval=0.5):
    global _pi
    global _started
    global _stopped
    global _waiting

    logging.info('Waiting for key to turn on (%d)', KEY_GPIO)
    _pi.set_pull_up_down(KEY_GPIO, pigpio.PUD_UP)
    _pi.set_mode(KEY_GPIO, pigpio.INPUT)
    _pi.set_glitch_filter(KEY_GPIO, KEY_DEGLITCH)
    _pi.callback(KEY_GPIO, pigpio.EITHER_EDGE, _callback)

    while _waiting:
        time.sleep(interval)
    start()

if __name__ == "__main__":
    try:
        if os.geteuid() != 0:
            exit('Root required, run again with "sudo"')
        logging.basicConfig(level=logging.INFO)
        #logging.basicConfig(level=logging.DEBUG)

        _process('pulse', 'pulseaudio')
        time.sleep(BOOT_WAIT)
        _sound('/home/dietpi/bbycrgo/sounds/boot.wav')
        _wait_for_key()
    except:
        _killall()
        logging.exception('BBYCR stopped due to an unhandled exception')
