#!/bin/bash

wget https://www.foundationdb.org/downloads/5.1.7/rhel7/installers/foundationdb-clients-5.1.7-1.el7.x86_64.rpm
wget https://www.foundationdb.org/downloads/5.1.7/rhel7/installers/foundationdb-server-5.1.7-1.el7.x86_64.rpm
rpm -Uvh foundationdb-clients-5.1.7-1.el7.x86_64.rpm foundationdb-server-5.1.7-1.el7.x86_64.rpm

wget https://dl.google.com/go/go1.10.2.linux-amd64.tar.gz
tar -C /usr/local -xzf go1.10.2.linux-amd64.tar.gz
echo 'PATH=$PATH:/usr/local/go/bin' >> ~/.bash_profile
echo 'PATH=$PATH:$HOME/go/bin' >> ~/.bash_profile
echo "export PATH" >> ~/.bash_profile
source ~/.bash_profile

yum install -y fuse fuse-devel git gcc
mkdir -p $HOME/go/src/github.com/advanderveer
ssh-keygen -f ~/.ssh/id_rsa -t rsa -N ''
cat ~/.ssh/id_rsa.pub
git clone git@github.com:advanderveer/dfs.git $HOME/go/src/github.com/advanderveer/dfs

go get github.com/Masterminds/glide
cd $HOME/go/src/github.com/advanderveer/dfs
glide install

cat >/etc/systemd/system/ffs.service <<EOL
[Unit]
Description=ffs Service
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/go/bin/go run /root/go/src/github.com/advanderveer/dfs/ffsvr/main.go /tmp/ffsdata4 0.0.0.0:10105
Restart=on-abort


[Install]
WantedBy=multi-user.target
EOL

systemctl daemon-reload
systemctl enable ffs
systemctl restart ffs
