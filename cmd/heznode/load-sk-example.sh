#!/bin/sh

# Non-Boot Coordinator
go run . importkey --mode coord --cfg cfg.buidler.toml --privatekey 0x30f5fddb34cd4166adb2c6003fa6b18f380fd2341376be42cf1c7937004ac7a3

# Boot Coordinator
go run . importkey --mode coord --cfg cfg.buidler.toml --privatekey 0xa8a54b2d8197bc0b19bb8a084031be71835580a01e70a45a13babd16c9bc1563

# FeeAccount
go run . importkey --mode coord --cfg cfg.buidler.toml --privatekey 0x3a9270c020e169097808da4b02e8d9100be0f8a38cfad3dcfc0b398076381fdd
