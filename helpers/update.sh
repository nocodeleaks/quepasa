#!/bin/bash

su - quepasa
git pull
systemctl
exit
systemctl daemon-reload && systemctl restart quepasa

exit 0