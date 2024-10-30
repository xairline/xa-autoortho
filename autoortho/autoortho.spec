# -*- mode: python ; coding: utf-8 -*-


a = Analysis(
    ['autoortho/autoortho.py'],
    pathex=[],
    binaries=[
    ('autoortho/imgs/splash.png','imgs'),
    ('autoortho/imgs/banner1.png','imgs')],
    datas=[],
    hiddenimports=['FreeSimpleGUI'],
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
app = BUNDLE(
    exe,
    name='autoortho.app',
    icon='autoortho/imgs/ao-icon.ico',
    bundle_identifier=None,
)
