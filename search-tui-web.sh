#!/bin/bash
go build -o mark-dev --tags "fts5" . && ttyd --writable -t disableLeaveAlert=true -t disableResizeOverlay=true ./mark-dev search
