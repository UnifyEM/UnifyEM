#!/bin/sh
cd ~/source/UnifyEM/agent
go build -o /tmp/uem-agent
sudo codesign -s - -f --deep /tmp/uem-agent
sudo /tmp/uem-agent upgrade

