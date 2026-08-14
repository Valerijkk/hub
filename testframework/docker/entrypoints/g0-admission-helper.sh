#!/bin/sh
set -eu

if [ "$#" -ne 0 ]; then
    exit 64
fi

exec python3 /opt/hub/admission/admit_local_dependency_assets.py
