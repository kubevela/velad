#!/bin/bash

PRINT="echo -e"
RED="\033[31m"
GREEN="\033[32m"
CNone="\033[0m"

$PRINT "checking usable package manager..."

if command -v yum >/dev/null; then
  PKGM="yum"
elif command -v apt-get >/dev/null; then
  PKGM="apt-get"
  $PKGM update -y
else
  echo "No support package manager was found"
  exit 1
fi
$PRINT "${GREEN}package manager found: ${PKGM}${CNone}"

$PRINT "Installing nginx by${PKGM}..."
$PKGM install -y nginx
ret=$?
if [ $ret -ne 0 ]; then
  $PRINT "${RED}Fail to install nginx${CNone}"
else
  $PRINT "${GREEN}Successfully install nginx${CNone}"
fi

STEAM_MOD="nginx-mod-stream"
if [ $PKGM = "apt-get" ]; then
  STEAM_MOD="libnginx-mod-stream"
fi

$PRINT "Installing nginx stream modules by ${PKGM}..."
$PKGM install -y $STEAM_MOD
ret=$?
if [ $ret -ne 0 ]; then
  $PRINT "${RED}Fail to install nginx stream mod${CNone}"
else
  $PRINT "${GREEN}Successfully install nginx stream mod${CNone}"
fi

$PRINT "Configuring nginx user..."
if id "nginx" &>/dev/null; then
  echo 'user nginx found'
else
  echo 'user nginx not found, creating...'
  useradd nginx
fi
