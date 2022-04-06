#!/bin/python3
import os
import sys
import logging
import argparse
import subprocess
from typing import List

ALLOWED_APPS = [
    'aghc',
    'anslookup',
    'aping',
    'aproxy',
]


def setLogger(level: str):
    log_level = {
        'Debug': logging.DEBUG,
        'Info': logging.INFO,
        'Warning': logging.WARNING,
        'Error': logging.ERROR,
        'Critical': logging.CRITICAL
    }

    logging.basicConfig(
        level=log_level[level],
        format="%(levelname)s: %(message)s"
    )


def parse_cmdline() -> argparse.Namespace:
    parser = argparse.ArgumentParser()
    apps = ALLOWED_APPS[:]
    apps.append('all')

    parser.add_argument('-B', '--build-type', type=str, choices=['Debug', 'Release'], default='Debug', required=False,
                        help='set build type')
    parser.add_argument('-i', '--install', action='store_true',
                        required=False, help='Install apps')
    parser.add_argument('apps', nargs='*', choices=apps,
                        help='apps need be build')

    parser.add_argument('--log', required=False, choices=[
                        'Debug', 'Info', 'Warning', 'Error', 'Critical'], default='Warning', help="Set log level")

    if len(sys.argv) < 2:
        print(parser.format_help())
        exit(0)
    arg_res = parser.parse_args()

    return arg_res


def get_build_flags(build_type: str) -> str:
    flags = ''
    if build_type == 'Debug':
        flags = '-gcflags "-N -l"'
    elif build_type == 'Release':
        flags = '-ldflags "-s -w"'
    return flags


# 检查需要构建的 app
def get_build_apps(apps: List[str]) -> List[str]:
    build_apps = list()
    if 'all' in apps:
        return ALLOWED_APPS
    for app in apps:
        app_path = os.path.join(os.getcwd(), 'cmd/'+app)
        if not os.path.exists(app_path):
            logging.error("app don't exists: %s", app)
            exit(1)
        build_apps.append(app)
    return build_apps


def tidy_project():
    cmd = 'go mod tidy'
    res = subprocess.call(cmd, shell=True)
    if res != 0:
        exit(res)


def build_app(app: str, build_type: str) -> None:
    flag = get_build_flags(build_type)
    cmd = 'go build {} ./cmd/{}'.format(flag, app)
    logging.debug('run command: %s', cmd)
    res = subprocess.call(cmd, shell=True)
    if res != 0:
        logging.error("Fail to build %s", app)
        exit(1)
    if not build_type == "Release":
        return
    app_path = os.path.join(os.getcwd(), app)
    strip_app(app_path)


def move_app(app: str):
    file_path = os.path.join(os.getcwd(), app)
    if not os.path.exists(file_path):
        return
    bin_path = os.path.join(os.getcwd(), 'bin')
    if not os.path.exists(bin_path):
        os.mkdir(bin_path)

    res = subprocess.call('mv {} {}'.format(file_path, bin_path), shell=True)
    if res != 0:
        logging.error("移动 app 失败： %s", app)


def install_app(app: str):
    des_path = os.getenv('GOPATH')
    if des_path is None:
        logging.error("请设置 GOPATH 环境变量")
        return
    des_path = os.path.join(des_path, 'bin')
    if not os.path.exists(des_path):
        os.mkdir(des_path)
    file_path = os.path.join(os.getcwd(), 'bin/'+app)
    if not os.path.exists(file_path):
        logging.error("app 安装失败: %s", app)
        return

    res = subprocess.call('mv {} {}'.format(file_path, des_path), shell=True)
    if res != 0:
        logging.error("安装 app 失败： %s", app)


def strip_app(file_path: str):
    strip_path = '/usr/bin/strip'
    if not os.path.exists(strip_path):
        logging.error("strip 不存在： %s", strip_path)
        return
    res = subprocess.call('{} "{}"'.format(strip_path, file_path), shell=True)
    if res != 0:
        logging.error("strip 失败")
        return


if __name__ == "__main__":
    args = parse_cmdline()
    setLogger(args.log)
    tidy_project()

    for app in get_build_apps(args.apps):
        build_app(app, args.build_type)
        move_app(app)
        if args.install is True:
            logging.debug("安装 app: %s", app)
            install_app(app)
