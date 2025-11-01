#!/bin/sh

PREREQ=""

prereqs()
{
	echo "$PREREQ"
}

case "$1" in
    prereqs)
	    prereqs
	    exit 0
	    ;;
esac

. /usr/share/initramfs-tools/hook-functions

if [ -x /sbin/bootsnap ]; then
    copy_exec /sbin/bootsnap
fi
