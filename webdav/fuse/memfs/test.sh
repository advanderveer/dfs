#!/bin/sh

rm -fr /tmp/t/secfs.test
git clone https://github.com/billziss-gh/secfs.test.git /tmp/t/secfs.test
git -C /tmp/t/secfs.test checkout -q e7edfb503c95998c1b0eb6b3f83ef203b1ed95dc
sed -e 's/^fs=.*$/fs="cgofuse"/' -i "" /tmp/t/secfs.test/fstest/fstest/tests/conf;
mv /tmp/t/secfs.test/fstest/fstest/tests/{,.}zzz_ResourceFork
make -C /tmp/t/secfs.test

#as seen in https://github.com/billziss-gh/cgofuse/blob/master/.travis.yml
go run memfs.go -o allow_other,default_permissions,use_ino,attr_timeout=0 /tmp/t/m &
# sleep 10
sleep 2
umount /tmp/t/m/
