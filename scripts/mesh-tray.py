#!/usr/bin/env python3
"""Mesh system tray icon using AppIndicator3 (works on both X11 and Wayland)."""

import os
import subprocess
import sys

import gi
gi.require_version("Gtk", "3.0")
gi.require_version("AppIndicator3", "0.1")
from gi.repository import AppIndicator3, Gtk


def run(cmd):
    subprocess.Popen(cmd, shell=True, stdout=subprocess.DEVNULL, stderr=subprocess.DEVNULL)


def on_open(_):
    run("xdg-open http://localhost:8080")


def on_start(_):
    run("systemctl --user start mesh")


def on_stop(_):
    run("systemctl --user stop mesh")


def on_restart(_):
    run("systemctl --user restart mesh")
    run('notify-send -i "{}" "Mesh" "Services restarted"'.format(icon_path))


def on_quit(_):
    Gtk.main_quit()


def build_menu():
    menu = Gtk.Menu()

    for label, handler in [
        ("Open Dashboard", on_open),
        ("Start Services", on_start),
        ("Stop Services", on_stop),
        ("Restart Services", on_restart),
        (None, None),
        ("Quit", on_quit),
    ]:
        if label is None:
            menu.append(Gtk.SeparatorMenuItem())
        else:
            item = Gtk.MenuItem(label=label)
            item.connect("activate", handler)
            menu.append(item)

    menu.show_all()
    return menu


if __name__ == "__main__":
    icon_path = sys.argv[1] if len(sys.argv) > 1 else "dialog-information"

    indicator = AppIndicator3.Indicator.new(
        "mesh-tray",
        icon_path,
        AppIndicator3.IndicatorCategory.APPLICATION_STATUS,
    )
    indicator.set_status(AppIndicator3.IndicatorStatus.ACTIVE)
    indicator.set_title("Mesh Knowledge Engine")
    indicator.set_menu(build_menu())

    Gtk.main()
