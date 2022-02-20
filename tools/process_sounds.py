import logging
import os
import subprocess
from pathlib import Path
from typing import Optional

import lib.cleanenv as cleanenv
import typer
from rich.logging import RichHandler
from rich.progress import track

DEFAULT_DRY_RUN = False
app = typer.Typer()


def configure(debug):
    cfg = {
        "config": "config/default.toml",
        "audio": {
            "rate": 48000,
        },
        "sounds": {
            "path": "data/sounds/effects",
            "ext": ".ogg",
            "randomizer": ".random",
        },
        "_prefix": "SPP_",
    }
    cfg = cleanenv.configure(cfg)
    cfg["debug"] = debug

    fmt = "%(message)s"
    if debug:
        logging.basicConfig(level=logging.DEBUG, format=fmt, handlers=[RichHandler(markup=True)])
    else:
        logging.basicConfig(level=logging.INFO, format=fmt, handlers=[RichHandler(markup=True)])
    return cfg


filters = {
    "isolate_vocals": "lowpass=f=4000,highpass=f=250",
    "compression": "dynaudnorm=p=0.5:s=5",
    "denoise": "anlmdn=s=0.0001:p=0.01:m=15",
    "derumble": "highpass=f=100",
    "pad": "adelay=10000|10000",
    "trim": "silenceremove=1:0:-50dB",
}

codecs = {
    ".ogg": "libvorbis",
}


def normalize(cfg, fin, fout, dry_run):
    t = ["-t", str(cfg["target"])]
    ar = ["-ar", str(cfg["audio"]["rate"])]
    ca = ["-c:a", codecs[os.path.splitext(fout)[1]]]
    ba = ["-b:a", "96k"]
    nt = ["-nt", "rms"]
    f = ["-f"] if cfg["overwrite"] else []
    v = ["-v"] if cfg["debug"] else []
    e = ["-e=-ac 2"]
    prf_keys = ["denoise", "derumble", "compression"]
    pof_keys = []
    if "-vocals" in fin:
        pof_keys.append("isolate_vocals")
    prf = ["-prf", ",".join(filters[key] for key in prf_keys)] if prf_keys else []
    pof = ["-pof", ",".join(filters[key] for key in pof_keys)] if pof_keys else []
    cmd = ["ffmpeg-normalize", fin, *t, *ar, *prf, *pof, *ca, *ba, *nt, *f, *v, *e, "-o", fout]
    if not dry_run:
        result = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE)
        if result.stdout:
            logging.info("ffmpeg-normalize (%s): %s", fout, result.stdout.decode())
        if result.stderr:
            logging.warning("ffmpeg-normalize (%s): %s", fout, result.stderr.decode())


@app.command()
def start(
    folder: Optional[str] = typer.Argument("data/sounds/original"),
    target: int = -10,
    overwrite: bool = False,
    debug: bool = False,
    dry_run: bool = DEFAULT_DRY_RUN,
):
    cfg = configure(debug)
    audio_files = list(Path(folder).rglob("*.[wW][aA][vV]"))

    cfg["target"] = target
    cfg["overwrite"] = overwrite
    if cfg["overwrite"]:
        logging.warning("Overwriting files if they already exist!")

    logging.info("Pre-processing %d audio files in %s", len(audio_files), folder)
    for fin in track(audio_files):
        fout = fin = str(fin)
        fout = fout.replace(os.path.normpath(folder), "")
        fout = fout.replace("-vocals", "")
        fout = fout.replace("-random", "")

        fname = os.path.splitext(os.path.basename(fout))[0].strip()
        fdir = os.path.join(
            os.path.normpath(cfg["sounds"]["path"]),
            os.path.dirname(fout)[1:].strip(),
        )
        fout = os.path.join(
            fdir,
            fname + cfg["sounds"]["ext"],
        )
        logging.debug("%s -> %s", fin, fout)

        if not os.path.exists(fdir):
            if not dry_run:
                os.makedirs(fdir)
            logging.debug("Creating sound folder %s", str(fdir))

            frnd = os.path.join(fdir, cfg["sounds"]["randomizer"])
            if "-random" in str(fin):
                if not os.path.exists(frnd):
                    if not dry_run:
                        Path(frnd).touch()
                    logging.debug("Creating randomizer %s", str(frnd))
            else:
                if os.path.exists(frnd):
                    if not dry_run:
                        os.remove(frnd)
                    logging.warning("Removing randomizer %s", str(frnd))

        if not os.path.exists(fout) or cfg["overwrite"]:
            normalize(cfg, fin, fout, dry_run)


def main():
    try:
        app()
    except SystemExit as e:
        logging.info("Script was terminated")


if __name__ == "__main__":
    main()
