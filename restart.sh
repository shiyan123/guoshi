#!/bin/bash
cd /data/go/src/gm-background
go install
sudo supervisorctl restart gm-background
tail -f /data/go/logs/gm-background.log