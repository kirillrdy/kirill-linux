set -xe

NEWROOT=/home/kirillvr/newroot2

rm -rfv glibc-2.30
curl -O http://ftp.gnu.org/gnu/glibc/glibc-2.30.tar.xz
tar xf glibc-2.30.tar.xz
cd glibc-2.30
mkdir build
cd build
../configure --prefix=/usr                          \
             --disable-werror                       \
             --enable-kernel=3.2                    \
             --enable-stack-protector=strong        \
             --with-headers=/usr/include            \
             libc_cv_slibdir=/lib
make -j10

make install_root=${NEWROOT} install

curl -O http://ftp.gnu.org/gnu/binutils/binutils-2.32.tar.xz
rm -rfv binutils-2.32
tar xf binutils-2.32.tar.xz
cd binutils-2.32
mkdir build
cd build
../configure --prefix=/usr       \
             --enable-gold       \
             --enable-ld=default \
             --enable-plugins    \
             --enable-shared     \
             --disable-werror    \
             --enable-64-bit-bfd \
             --with-system-zlib
make -j10
make install DESTDIR=$NEWROOT


curl -O http://ftp.gnu.org/gnu/ncurses/ncurses-6.1.tar.gz
rm -rfv ncurses-6.1
tar xf ncurses-6.1.tar.gz
cd ncurses-6.1
./configure --prefix=/usr           \
            --mandir=/usr/share/man \
            --with-shared           \
            --without-debug         \
            --without-normal        \
            --enable-pc-files       \
            --enable-widec
make -j10
make install DESTDIR=$NEWROOT

curl -O http://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz
rm -rfv bash-5.0
tar xf bash-5.0.tar.gz
cd bash-5.0
./configure --prefix=/usr                    \
            --docdir=/usr/share/doc/bash-5.0 \
            --without-bash-malloc            \
            --with-installed-readline
make -j10
make install DESTDIR=$NEWROOT
