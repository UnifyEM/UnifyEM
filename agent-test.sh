#!/bin/sh
cd ~/source/UnifyEM/agent
rm -f /tmp/uem-agent
go build -o /tmp/uem-agent
#sudo codesign -s - -f --deep /tmp/uem-agent
codesign -s "Developer ID Application: Tenebris Technologies Inc. (76F27732FD)" -f --timestamp -o runtime -i "com.tenebris.uem-agent" /tmp/uem-agent
codesign --verify --deep --strict --verbose=2 /tmp/uem-agent
sudo /tmp/uem-agent upgrade

