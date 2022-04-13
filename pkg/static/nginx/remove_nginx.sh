#!/bin/bash

PRINT="echo -e"
RED="\033[31m"
GREEN="\033[32m"
CNone="\033[0m"

$PRINT "checking usable package manager"

if command -v yum >/dev/null; then
  PKGM="yum"
elif command -v apt >/dev/null; then
  PKGM="apt"
elif command -v apt-get >/dev/null; then
  PKGM="apt-get"
else
  echo "No support package manager was found"
  exit 1
fi
$PRINT "${GREEN}package manager found: ${PKGM}${CNone}"

$PRINT "Removing nginx by${PKGM}..."
$PKGM remove -y nginx
ret=$?
if [ $ret -ne 0 ]; then
  $PRINT "${RED}Fail to remove nginx${CNone}"
else
  $PRINT "${GREEN}Successfully remove nginx${CNone}"
fi
