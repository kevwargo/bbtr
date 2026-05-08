package stream

import (
	"fmt"
	"strings"
)

const (
	C_UNSPEC        = 0
	C_SUBVOL        = 1
	C_SNAPSHOT      = 2
	C_MKFILE        = 3
	C_MKDIR         = 4
	C_MKNOD         = 5
	C_MKFIFO        = 6
	C_MKSOCK        = 7
	C_SYMLINK       = 8
	C_RENAME        = 9
	C_LINK          = 10
	C_UNLINK        = 11
	C_RMDIR         = 12
	C_SET_XATTR     = 13
	C_REMOVE_XATTR  = 14
	C_WRITE         = 15
	C_CLONE         = 16
	C_TRUNCATE      = 17
	C_CHMOD         = 18
	C_CHOWN         = 19
	C_UTIMES        = 20
	C_END           = 21
	C_UPDATE_EXTENT = 22
	C_FALLOCATE     = 23
	C_FILEATTR      = 24
	C_ENCODED_WRITE = 25
)

var cmdTypes = map[int]string{
	C_UNSPEC:        "UNSPEC",
	C_SUBVOL:        "SUBVOL",
	C_SNAPSHOT:      "SNAPSHOT",
	C_MKFILE:        "MKFILE",
	C_MKDIR:         "MKDIR",
	C_MKNOD:         "MKNOD",
	C_MKFIFO:        "MKFIFO",
	C_MKSOCK:        "MKSOCK",
	C_SYMLINK:       "SYMLINK",
	C_RENAME:        "RENAME",
	C_LINK:          "LINK",
	C_UNLINK:        "UNLINK",
	C_RMDIR:         "RMDIR",
	C_SET_XATTR:     "SET_XATTR",
	C_REMOVE_XATTR:  "REMOVE_XATTR",
	C_WRITE:         "WRITE",
	C_CLONE:         "CLONE",
	C_TRUNCATE:      "TRUNCATE",
	C_CHMOD:         "CHMOD",
	C_CHOWN:         "CHOWN",
	C_UTIMES:        "UTIMES",
	C_END:           "END",
	C_UPDATE_EXTENT: "UPDATE_EXTENT",
	C_FALLOCATE:     "FALLOCATE",
	C_FILEATTR:      "FILEATTR",
	C_ENCODED_WRITE: "ENCODED_WRITE",
}

const (
	A_UNSPEC             = 0
	A_UUID               = 1
	A_CTRANSID           = 2
	A_INO                = 3
	A_SIZE               = 4
	A_MODE               = 5
	A_UID                = 6
	A_GID                = 7
	A_RDEV               = 8
	A_CTIME              = 9
	A_MTIME              = 10
	A_ATIME              = 11
	A_OTIME              = 12
	A_XATTR_NAME         = 13
	A_XATTR_DATA         = 14
	A_PATH               = 15
	A_PATH_TO            = 16
	A_PATH_LINK          = 17
	A_FILE_OFFSET        = 18
	A_DATA               = 19
	A_CLONE_UUID         = 20
	A_CLONE_CTRANSID     = 21
	A_CLONE_PATH         = 22
	A_CLONE_OFFSET       = 23
	A_CLONE_LEN          = 24
	A_FALLOCATE_MODE     = 25
	A_FILEATTR           = 26
	A_UNENCODED_FILE_LEN = 27
	A_UNENCODED_LEN      = 28
	A_UNENCODED_OFFSET   = 29
	A_COMPRESSION        = 30
	A_ENCRYPTION         = 31
)

var attrTypes = map[int]string{
	A_UNSPEC:             "unspec",
	A_UUID:               "uuid",
	A_CTRANSID:           "ctransid",
	A_INO:                "ino",
	A_SIZE:               "size",
	A_MODE:               "mode",
	A_UID:                "uid",
	A_GID:                "gid",
	A_RDEV:               "rdev",
	A_CTIME:              "ctime",
	A_MTIME:              "mtime",
	A_ATIME:              "atime",
	A_OTIME:              "otime",
	A_XATTR_NAME:         "xattr_name",
	A_XATTR_DATA:         "xattr_data",
	A_PATH:               "path",
	A_PATH_TO:            "path_to",
	A_PATH_LINK:          "path_link",
	A_FILE_OFFSET:        "file_offset",
	A_DATA:               "data",
	A_CLONE_UUID:         "clone_uuid",
	A_CLONE_CTRANSID:     "clone_ctransid",
	A_CLONE_PATH:         "clone_path",
	A_CLONE_OFFSET:       "clone_offset",
	A_CLONE_LEN:          "clone_len",
	A_FALLOCATE_MODE:     "fallocate_mode",
	A_FILEATTR:           "fileattr",
	A_UNENCODED_FILE_LEN: "unencoded_file_len",
	A_UNENCODED_LEN:      "unencoded_len",
	A_UNENCODED_OFFSET:   "unencoded_offset",
	A_COMPRESSION:        "compression",
	A_ENCRYPTION:         "encryption",
}

func (a Attribute) String() string {
	return fmt.Sprintf("%s:%d[%d]", attrTypes[int(a.Type)], a.Type, a.Length)
}

func (c Command) String() string {
	var attrs []string
	for _, a := range c.Attrs {
		attrs = append(attrs, a.String())
	}

	return fmt.Sprintf("%s:%d[%d]: %s", cmdTypes[int(c.Type)], c.Type, c.Length, strings.Join(attrs, "; "))
}
