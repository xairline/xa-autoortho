mkfile_path := $(abspath $(lastword $(MAKEFILE_LIST)))
current_dir := $(notdir $(patsubst %/,%,$(dir $(mkfile_path))))

dev:
	cargo build
	mv target/debug/libxa_autoortho.dylib build/XA-autoortho/mac.xpl

mac:
	cargo build --release --target aarch64-apple-darwin
	mv target/aarch64-apple-darwin/release/libxa_autoortho.dylib build/XA-autoortho/mac_arm.xpl
	cargo build --release --target x86_64-apple-darwin
	mv target/x86_64-apple-darwin/release/libxa_autoortho.dylib build/XA-autoortho/mac_amd.xpl
	lipo build/XA-autoortho/mac_arm.xpl build/XA-autoortho/mac_amd.xpl -create -output build/XA-autoortho/mac.xpl