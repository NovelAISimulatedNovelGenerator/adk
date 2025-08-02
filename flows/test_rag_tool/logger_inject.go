package main

import "go.uber.org/zap"

var plgLog *zap.Logger = zap.NewNop()

func SetLogger(l *zap.Logger) { plgLog = l }

func L() *zap.Logger { return plgLog }
