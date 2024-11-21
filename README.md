# xa-autoortho

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

## How to debug/run manually

use `autoortho` and click `run`

## How to fix broken AO tiles

### Symptom

You are successfully running AutoOrtho (AO) when you suddenly start flying over water that shouldn’t be there,
OR…. you start X-Plane (XP) at an airport and your aircraft appears to crash into water.

### Diagnosis

You have flown over, or started and an airport that AO has loaded a broken scenery tile.

### Solution

AO Patch (starting in version .5.1

### Implementation

See the Youtube video https://www.youtube.com/watch?v=f3gU3oHv4BA

Do not go further unless you know that your AO installation is working!

#### Highlights (X-Airline version of AO)

- Works by mounting an additional FUSE volume where the patched tiles are located. When XP and AO start, the patch
  volume also mounts. The patched volume is referenced in the Scenery_Packs.ini file.

- Get the patch from https://github.com/xairline/xa-autoortho/releases/tag/v0.5.1  (y_ao_patches)

- Follow the directions for install. The below may be helpful with the details of installation.

- Once you have placed the patch files in the AO installation directory/folder, start XP so the AO Patch gets invoked.
  Shutdown XP (dont fly). Start XP again so that both FUSE AO and AO Patch are running.  (You should see one more FUSE
  volumes in your XP Custom Scenery folder.

- Modify the Scenery_Packs.ini file. The AO Patch (y_ao_patches) will be at the top of the Scenery_Packs.ini file as
  “SCENERY_PACK Custom Scenery/y_ao_patches/“. Move this line to just above the SCENERY_PACK Custom Scenery/z_ao_xx/
  line in the Scenery_Packs.ini file. it should look something like this (using North America in this example).

```
 SCENERY_PACK Custom Scenery/simHeaven_X-World_America-7-forests/
 SCENERY_PACK Custom Scenery/simHeaven_X-World_America-8-network/
 SCENERY_PACK Custom Scenery/------------------Orthos----------------/
 SCENERY_PACK Custom Scenery/yAutoOrtho_Overlays/
 SCENERY_PACK Custom Scenery/y_ao_patches/
 SCENERY_PACK Custom Scenery/z_ao_na/
 ...
```

- Save the Scenery_Packs.ini.
- Restart XP

> Note
>
> Just like the the AO volume, the AO patch volume will not be visible from Finder when FUSE volumes are mounted .
> You may can use Terminal to see it (use “ls” command.)


