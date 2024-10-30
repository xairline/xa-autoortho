# -*- mode: python ; coding: utf-8 -*-


a1 = Analysis(
    ['autoortho/autoortho.py'],
    pathex=[],
    binaries=[
    ('autoortho/imgs/splash.png','imgs'),
    ('autoortho/imgs/banner1.png','imgs')],
    datas=[
         (certifi.where(), '.'),
    ],
    hiddenimports=['FreeSimpleGUI'],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    noarchive=False,
    optimize=0,
)
pyz1 = PYZ(a1.pure)

exe1 = EXE(
    pyz1,
    a1.scripts,
    a1.binaries,
    a1.datas,
    [],
    name='autoortho',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=False,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
    icon=['autoortho/imgs/ao-icon.ico'],
)
app1 = BUNDLE(
    exe1,
    name='autoortho.app',
    icon='autoortho/imgs/ao-icon.ico',
    bundle_identifier=None,
)


a = Analysis(
    ['autoortho/autoortho_fuse.py'],
    pathex=[],
    binaries=[('autoortho/lib/darwin_arm/libispc_texcomp.dylib', '.'),  # Adjust destination path if needed
                      ('autoortho/aoimage/aoimage.dylib', '.')],
    datas=[
         (certifi.where(), '.'),
    ],
    hiddenimports=[],
    hookspath=[],
    hooksconfig={},
    runtime_hooks=[],
    excludes=[],
    noarchive=False,
    optimize=0,
)
pyz = PYZ(a.pure)

exe = EXE(
    pyz,
    a.scripts,
    a.binaries,
    a.datas,
    [],
    name='autoortho_fuse',
    debug=False,
    bootloader_ignore_signals=False,
    strip=False,
    upx=True,
    upx_exclude=[],
    runtime_tmpdir=None,
    console=True,
    disable_windowed_traceback=False,
    argv_emulation=False,
    target_arch=None,
    codesign_identity=None,
    entitlements_file=None,
)