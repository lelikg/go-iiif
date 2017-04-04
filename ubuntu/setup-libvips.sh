#!/bin/sh

# http://www.vips.ecs.soton.ac.uk/index.php?title=Development

WHOAMI=`python -c 'import os, sys; print os.path.realpath(sys.argv[1])' $0`
ROOT=`dirname $WHOAMI`

apt-get install -y build-essential pkg-config glib2.0-dev libxml2-dev libjpeg-dev libpng-dev libgif-dev libwebp-dev libtiff-dev libmagick-dev librsvg2-dev

git clone git://github.com/jcupitt/libvips.git

cd ${ROOT}

if [ -d ${ROOT}/libvips ]
then
    rm -rf ${ROOT}/libvips
fi

cd ${ROOT}/libvips

./bootstrap.sh
./configure 
make
make install
