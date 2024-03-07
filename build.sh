#!/usr/bin/sh 
# purpose is to build the go program, setup systemctl unit and start running
echo "Now building and installing aquaponic control program.. "

sudo ln -sf /home/niranjan/source/github.com/eemsymachines-in/patio/aquapone.config.json /etc/aquapone.config.json
go build -o /usr/bin/eensymacaqupone .  && chmod 774 /usr/bin/eensymacaqupone
echo "done! built aquapone control \n run from /usr/bin/eensymacaqupone"

sudo systemctl enable ./aquapone.service
sudo systemctl daemon-reload
