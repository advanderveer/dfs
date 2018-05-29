#!/bin/bash
set -e

dev_profile="nerd-cli-dev"

function print_help {
	printf "Available Commands:\n";
	awk -v sq="'" '/^function run_([a-zA-Z0-9-]*)\s*/ {print "-e " sq NR "p" sq " -e " sq NR-1 "p" sq }' make.sh \
		| while read line; do eval "sed -n $line make.sh"; done \
		| paste -d"|" - - \
		| sed -e 's/^/  /' -e 's/function run_//' -e 's/#//' -e 's/{/	/' \
		| awk -F '|' '{ print "  " $2 "\t" $1}' \
		| expand -t 30
}

function run_build-svr { #build the ffs server
  xgo --image=billziss/xgo-cgofuse --targets=darwin/amd64,windows/amd64 --dest bin .
}

function run_win-test { #test on windows
  winfsp-tests-x64.exe --external --resilient --share-prefix=\gomemfs\share -create_allocation_test -create_fileattr_test -getfileinfo_name_test -setfileinfo_test -delete_access_test -setsecurity_test -querydir_namelen_test -reparse* -stream*
}

case $1 in
	"build-svr") run_build-svr ;;
	"win-test") run_win-test ;;
	*) print_help ;;
esac
