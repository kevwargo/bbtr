NAME := bootsnap
BUILD_DIR := build
BIN := $(BUILD_DIR)/$(NAME)
INSTALL := install -o root -g root -m 0755

.PHONY: build
build:
	go build -o $(BIN) main.go

.PHONY: install
install:
	$(INSTALL) -s $(BIN) /sbin
	$(INSTALL) -T hook.sh /etc/initramfs-tools/hooks/$(NAME)
	$(INSTALL) -T script.sh /etc/initramfs-tools/scripts/local-premount/$(NAME)

.PHONY: initramfs
initramfs: install
	update-initramfs -k `uname -r` -u
