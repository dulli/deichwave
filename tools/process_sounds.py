import os
import subprocess
from pathlib import Path

FFMPEG_PATH = '"ffmpeg"'
SOUNDS_PATH = os.path.join(Path().resolve(), "data", "sounds", "original")

print("This script will look for WAV-files in {0}...".format(SOUNDS_PATH))
print("It will first normalize them and then remove any silence from the beginning...")
print('If a path contains "-vocals", a bandpass filter will be used to isolate vocals.')
print("")
print("Looking for ffmpeg at path: {0}".format(FFMPEG_PATH))
print("")
input("Press Enter to continue...")

os.listdir(SOUNDS_PATH)

for wav in Path(SOUNDS_PATH).rglob("*.[wW][aA][vV]"):
    new_wav = str(wav).replace("original", "processed")
    new_wav = new_wav.replace(" -vocals", "")
    new_wav = new_wav.replace("-vocals", "")
    new_wav = new_wav.replace(" -random", "")
    new_wav = new_wav.replace("-random", "")

    pad_wav = new_wav.replace(".wav", ".pad.wav")
    norm_wav = new_wav.replace(".wav", ".norm.wav")
    trim_wav = new_wav.replace(".wav", ".trim.wav")
    rate_wav = new_wav.replace(".wav", ".rate.wav")
    new_out = new_wav.replace(".wav", ".ogg")

    new_dir = os.path.dirname(new_wav)
    if not os.path.exists(new_dir):
        os.makedirs(new_dir)

    new_randomizer = os.path.join(new_dir, ".random")
    if "-random" in str(wav):
        if not os.path.exists(new_randomizer):
            Path(new_randomizer).touch()
    else:
        if os.path.exists(new_randomizer):
            os.remove(new_randomizer)

    subprocess.run(
        'ffmpeg -i "{0}" -af "adelay=10000|10000" -y "{1}"'.format(wav, pad_wav), shell=True
    )

    subprocess.run('ffmpeg-normalize "{0}" -t -15 -o "{1}"'.format(pad_wav, norm_wav), shell=True)
    os.remove(pad_wav)

    subprocess.run(
        'ffmpeg -i "{0}" -af silenceremove=1:0:-50dB -y "{1}"'.format(norm_wav, trim_wav),
        shell=True,
    )
    os.remove(norm_wav)

    subprocess.run('ffmpeg -i "{0}" -ar 48000 -y "{1}"'.format(trim_wav, rate_wav), shell=True)
    os.remove(trim_wav)

    if "-vocal" in str(wav):
        subprocess.run(
            'ffmpeg -i "{0}" -af lowpass=4000,highpass=250 -y "{1}"'.format(rate_wav, new_wav),
            shell=True,
        )
        os.remove(rate_wav)

    else:
        if os.path.exists(new_wav):
            os.remove(new_wav)
        os.rename(rate_wav, new_wav)

    subprocess.run(
        'ffmpeg -i "{0}" -c:a libvorbis -b:a 96k -ar 48000 -ac 2 "{1}"'.format(new_wav, new_out),
        shell=True,
    )
    os.remove(new_wav)
    print(new_out)

input("Press Enter to close...")
