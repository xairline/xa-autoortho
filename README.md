# xa-autoortho


>NOTE: WIP MAJOR REFACTORING IN PROGRESS

>NOTE: ONLY WORKS ON MAC SILICON

>NOTE: DO NOT USE UNLESS YOU HAVE TALKED TO ME

## Features

- [x] remove Python phase 1 (pack all dependencies in a binary)

- [x] auto mount with xplane loading

- [x] allow install new scenery

- [ ] add seasonal adjustment for ortho images

- [ ] new ui in xplane

- [ ] complete remove Python

- [ ] support new xplane format

## HOW TO INSTALL

1. Download [latest version](https://github.com/xairline/xa-autoortho/releases/latest) from github
2. Make sure [macfuse](https://osxfuse.github.io/) is properly installed
3. The zip file has 5 files, remove Mac quarantine flags on all of them. ``perm.sh`` can be used to do this.
4. Copy the unzipped folder into xplane plugins folder
5. (For now) use the autoortho (without icon) to open autoortho UI so you can install regions/change config. 

    > NOTE: UI IS ONLY FOR INSTALLING REGIONS OR CHANGING CONFIG
6. once 5 is done, launch xplane

## How to get support
Submit the following information:

1. XPlane Logs
2. AutoOrtho Logs
```shell
~/.autoortho-data/logs/autoortho.log
```
3. If you have UI issue, submit the console log of UI, aka output in the terminal when you run the UI