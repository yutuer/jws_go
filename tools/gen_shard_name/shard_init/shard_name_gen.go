package shard_init

import (
	"crypto/md5"
	"fmt"
	"io"
)

func MkShardName(id, gid int) string {
	h := md5.New()
	io.WriteString(h, "32fds-+a") //salt 这个可不要乱改哦
	io.WriteString(h, fmt.Sprintf("id:%d, gid:%d", id, gid))
	return fmt.Sprintf("%x", h.Sum(nil))
}
