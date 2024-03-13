#!/usr/bin/sh 
# purpose is to build the go program, setup systemctl unit and start running
echo "Now building and installing aquaponic control program.. "

sudo ln -sf /home/niranjan/source/github.com/eensymachines-in/aquapone/aquapone.config.json /etc/aquapone.config.json
sudo go build -o /usr/bin/eensymacaqupone .  && sudo chmod 774 /usr/bin/eensymacaqupone

echo 'building systemctl unit..'
sudo systemctl enable $(pwd)/$NAME_SYSCTLSERVICE
sudo systemctl daemon-reload
sudo systemctl start $NAME_SYSCTLSERVICE
echo 'done..\nrun from /usr/bin/eensymacaqupone'
