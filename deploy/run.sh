#!/bin/sh

wget --show-progress -qO agent https://github.com/last9/k8stream/releases/${VERSION:-latest}/download/k8stream
chmod a+x agent
