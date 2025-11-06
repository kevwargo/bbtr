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
	$(INSTALL) 10_linux /etc/grub.d

.PHONY: update-initramfs
update-initramfs: install
	update-initramfs -u -k all

.PHONY: update-grub
update-grub: install
	update-grub
