#!/bin/bash
cd ~/task
grep -Eo "[a-zA-Z0-9+-.]*[a-zA-Z][a-zA-Z0-9+-.]*://[a-zA-Z0-9./?=_%:-]+" in.txt>output.txt

