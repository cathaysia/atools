#!/bin/python3
import os
import logging
import argparse
import subprocess
from typing import List


def setLogger(level: str):
    bad = False
    if level is None:
        level = "Warning"
    level_int = 0
    if level == 'Info':
        level_int = logging.INFO
    elif level == 'Debug':
        level_int = logging.DEBUG
    elif level == 'Warning':
        level_int = logging.WARNING
    elif level == 'Error':
        level_int = logging.ERROR
    elif level == 'Critical':
        level_int = logging.CRITICAL
    else:
        bad = True

    logging.basicConfig(
        level=level_int,
        format="%(levelname)s: %(message)s"
    )
    if bad:
        logging.error("log level error: %s", level)
        exit(1)


def parse_cmdline() -> argparse.Namespace:
    parser = argparse.ArgumentParser()

    parser.add_argument('-B', '--build-type', required=False,
                        help='set build type: Debug|Release')
    parser.add_argument('apps', nargs='*', help='apps need be build')

    parser.add_argument('--log', required=False, help="Set log level")

    arg_res = parser.parse_args()

    return arg_res


def get_build_flags(build_type: str) -> str:
    if build_type == None:
        build_type = 'Release'
    flags = ''
    if build_type == 'Release':
        flags = '-gcflags "-N -l"'
    elif build_type == 'Debug':
        flags = '-ldflags "-s -w"'
    else:
        logging.critical("build type error: %s", build_type)
        exit(1)
    return flags


# 检查需要构建的 app
def get_build_apps(apps: List[str]) -> List[str]:
    build_apps = list()
    for app in apps:
        app_path = os.path.join(os.getcwd(), 'cmd/'+app)
        if not os.path.exists(app_path):
            logging.error("app don't exists: %s", app)
            exit(1)
        build_apps.append(app)
    return build_apps


def build_app(app: str, flag: str) -> None:
    cmd = 'go build {} ./cmd/{}'.format(flag, app)
    logging.debug('run command: "{}"'.format(cmd))
    subprocess.call(cmd, shell=True)


if __name__ == "__main__":
    args = parse_cmdline()
    setLogger(args.log)

    apps = get_build_apps(args.apps)
    flag = get_build_flags(args.build_type)

    for app in apps:
        build_app(app, flag)
