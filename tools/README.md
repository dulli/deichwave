# Tools

Supporting tools and scripts used for the project. These are not part of the actual codebase but rather standalone commands.

## Generate HTTPS Certificate: `generate_certificate.sh`

Run from project root to (re-)generate the self-signed certificate used for the HTTPS server (which will be embedded into the binary at compile time and can be installed on a client by navigation to the `/certificate` endpoint on the web interface).

```shell
./tools/generate_certificate.sh
```

All certificate related files are located in the `web/tls` directory.

## Prepare Sound Effects: `process_sounds.py`

This will pre-process all sound files in a given directory and copy them into the sound effect folder defined in your configuration file (if one exists, or the default folder otherwise). The pre-processing steps include de-noising, de-rumbling, dynamic range compression, loudness normalization and vocal isolation where requested. Call with a target directory that holds your original sound files as the first argument, `--target` is the targeted loudness level:

```shell
poetry run python ./tools/process_sounds.py "./data/sounds/original/" --target=-15
```

If a sound files path includes the string `-random`, a dotfile that indicates to `Deichwave` that all sounds in that sound's folder should be randomized will be created. If the path includes the magic string `-vocals`, a vocal isolation using high- and lowpass filtering will be attempted.
