import os

import toml
from dotenv import load_dotenv


def visit_nested(d, p=[]):
    for k, v in d.items():
        if not isinstance(v, dict):
            yield p + [k], v
        else:
            yield from visit_nested(v, p + [k])


def set_nested(d, k, v):
    for key in k[:-1]:
        d = d.setdefault(key, {})
    d[k[-1]] = v


def configure(cfg):
    # Prepare environment variables
    env_prefix = cfg["_prefix"].upper()
    del cfg["_prefix"]
    load_dotenv(".env")

    for cfg_path, _ in visit_nested(cfg):
        env_path = (env_prefix + "_".join(cfg_path)).upper()
        if env_path in os.environ:
            set_nested(cfg, cfg_path, os.environ[env_path])

    # Prepare config file variables
    cfg_file = toml.load(cfg["config"])
    del cfg["config"]

    for cfg_path, value in visit_nested(cfg_file):
        set_nested(cfg, cfg_path, value)

    return cfg
