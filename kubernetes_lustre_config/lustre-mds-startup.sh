#!/bin/sh

#kubectl create secret generic vmi-lustre-mds-secret --from-file=userdata=lustre-mds-startup.sh
#export NEW_USER="foo"
#export SSH_PUB_KEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQDmkaDRsJMcUQGV2euLicgMxVhSDGN+sae/QoLkDVyZLAV6d/8tST3ZOevVoWtg4N10R1x3WlzbjeR4G59rKS2iHyz76GmlHBYFrwD7SXI23IIxlDgwnndyn33Qobw/RxHcfqc8YlDR/xGexqGvKqDg37RMyooirXj0WkfzT6PKOnSHGIKWYzoT6Qipq4yvJVxTVF6uSo2jt1H7PrGamLNO1Dk0F6EPF0T88z4kYcBVNL+/P4YkvFJ4uZRTg92B9ROfHa6X+Wy8fmoZG2ieJwmySl5yNbUWqNegNlpQqU4n/Pp/FJcZWAPPBqH8XPTLhrYXYeKrxJDR1WsRJgH9nW8/ moc-lustre"

export SSH_PUB_KEY="ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQDakNvL1Umeyu7Uqfx4J5z5I9B48hpcDc7T9EAX/WS3A3qOnRBwA1OAq1e5yoLG/9rCTb1xje6YPTJQxPOSamFHWjNzru3ugDvJSe1Pfwbf1ETzSmW8VwU7LYNyb+ydJN+1H9xeTyr/sevIqYHxQraF4fab9no76PtRE9ERikVaO9cHpBL0iXF1PBT0/Y8SUQQ2dJhZaZJOMsa9kUmNrqoAcpulVsRuotUomBXOmHI1mhFU9j34K2Sgfpiav2BKyL7AyB3BM0kzF7s94DCsl8+ikRGjrWXfzzkXlc/2POoJsgI5JUQAjZhuIp8cyRYLZKI7/23ufRVO5vR6V7VyMC+bDEDRjuzJYR6cE2q5TljqNBuGFccObEzVzPZzgymYDdbvAL8WO1WinwL7pYwC1ksPKsubtcrEgPk9+C31AKZ+p2fCYv31sA9XEQLihHpI7qkguamamlZQf/DSPcZ4vPrGtzV+5IkXtRRIpd6PwHxSKRwKn8CoWnOkgB2t3Q9g4wk= master@lustreproj"

#sudo adduser -U -m $NEW_USER
export DEFAULT_USER="centos"
echo "$DEFAULT_USER:centos" | chpasswd
sudo mkdir /home/$DEFAULT_USER/.ssh
sudo echo "$SSH_PUB_KEY" > /home/$DEFAULT_USER/.ssh/authorized_keys
sudo chown -R ${DEFAULT_USER}: /home/$DEFAULT_USER/.ssh

# enable lnet
if [ ! -c /dev/lnet ] ; then
   sudo exec /sbin/modprobe -v lnet >/dev/null 2>&1
fi

# enable lustre
/sbin/lsmod | /bin/grep lustre 1>/dev/null 2>&1
if [ $? -ne 0 ] ; then
   sudo /sbin/modprobe -v lustre >/dev/null 2>&1
fi

# enable ZFS
/sbin/lsmod | /bin/grep zfs 1>/dev/null 2>&1
if [ $? -ne 0 ] ; then
   sudo /sbin/modprobe -v zfs >/dev/null 2>&1
fi

# create and mount MGT
sudo /usr/sbin/mkfs.lustre --fsname=lustrefs --mgsnode=lustre-mgs.default-lustre@tcp0 --mdt --index=0 /dev/vdb > /dev/null 2>&1
sudo /usr/bin/mkdir /mdt
sudo /usr/sbin/mount.lustre /dev/vdb /mdt
