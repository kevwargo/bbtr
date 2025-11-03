#!/bin/sh

PREREQ="btrfs"

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

/sbin/bootsnap
