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
  docker build -t avanderveer/ffs -f dep.Dockerfile .
}

case $1 in
	"build-svr") run_build-svr ;;
	*) print_help ;;
esac
