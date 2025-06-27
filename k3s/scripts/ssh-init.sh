#!/bin/bash

set -x

if [ ! -z "$SSH_KEY" ]; then
    mkdir -p /var/run/sshd
fi
