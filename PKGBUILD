

pkgname=snapd
pkgdesc="Service and tools for management of snap packages."
depends=('squashfs-tools' 'libseccomp' 'libsystemd' 'apparmor')
optdepends=('bash-completion: bash completion support'
            'xdg-desktop-portal: desktop integration')
pkgver=2.59.1
pkgrel=1
arch=('x86_64' 'i686' 'armv7h' 'aarch64')
url="https://github.com/snapcore/snapd"
license=('GPL3')
makedepends=('git' 'go' 'go-tools' 'libseccomp' 'libcap' 'systemd' 'xfsprogs' 'python-docutils' 'apparmor')
conflicts=('snap-confine')
options=('!strip' 'emptydirs' '!lto')
install=snapd.install
source=(
    "$pkgname-$pkgver.tar.xz::https://github.com/snapcore/${pkgname}/releases/download/${pkgver}/${pkgname}_${pkgver}.vendor.tar.xz"
)
sha256sums=('67d24dd43e963007443d287ee8517cc8c1ab6c8d8a2431564e81e79a5f454ae0')


_gourl=github.com/snapcore/snapd

prepare() {
  cd "$pkgname-$pkgver"

  export GOPATH="$srcdir/go"
  mkdir -p "$GOPATH"

  # Have snapd checkout appear in a place suitable for subsequent GOPATH. This
  # way we don't have to go get it again and it is exactly what the tag/hash
  # above describes.
  mkdir -p "$(dirname "$GOPATH/src/${_gourl}")"
  ln --no-target-directory -fs "$srcdir/$pkgname-$pkgver" "$GOPATH/src/${_gourl}"

  for name in "${source[@]}"; do
      if [[ "${name%.patch}" == "$name" ]]; then
          # not a patch
          continue
      fi
      msg2 "applying $name"
      patch -p1 -i "$srcdir/$name"
  done
}

build() {
  cd "$pkgname-$pkgver"
  export GOPATH="$srcdir/go"

  # GOFLAGS may be modified by CI tools
  # GOFLAGS are the go build flags for all binaries, GOFLAGS_SNAP are for snap
  # build only.
  GOFLAGS=""
  GOFLAGS_SNAP="-tags nomanagers"

  export CGO_ENABLED="1"
  export CGO_CFLAGS="${CFLAGS}"
  export CGO_CPPFLAGS="${CPPFLAGS}"
  export CGO_CXXFLAGS="${CXXFLAGS}"
  export CGO_LDFLAGS="${LDFLAGS}"

  ./mkversion.sh $pkgver-$pkgrel

  # because argument expansion with quoting in bash is hard, and -ldflags=-extldflags='-foo'
  # is not exactly the same as -ldflags "-extldflags '-foo'" use the array trick
  # to pass exactly what we want
  flags=(-buildmode=pie -ldflags "-w -s -linkmode external -extldflags '$LDFLAGS'" -trimpath -mod=vendor)
  staticflags=(-buildmode=pie -ldflags "-w -s -linkmode external -extldflags '$LDFLAGS -static'" -trimpath -mod=vendor)
  # Build/install snap and snapd
  go build "${flags[@]}" -o "$srcdir/go/bin/snap" $GOFLAGS_SNAP "${_gourl}/cmd/snap"
  go build "${flags[@]}" -o "$srcdir/go/bin/snapd" $GOFLAGS "${_gourl}/cmd/snapd"
  go build "${flags[@]}" -o "$srcdir/go/bin/snap-seccomp" $GOFLAGS "${_gourl}/cmd/snap-seccomp"
  go build "${flags[@]}" -o "$srcdir/go/bin/snap-failure" $GOFLAGS "${_gourl}/cmd/snap-failure"
  go build "${flags[@]}" -o "$srcdir/go/bin/snapd-apparmor" $GOFLAGS "${_gourl}/cmd/snapd-apparmor"
  # build snap-exec and snap-update-ns completely static for base snaps
  go build "${staticflags[@]}" -o "$srcdir/go/bin/snap-update-ns" $GOFLAGS "${_gourl}/cmd/snap-update-ns"
  go build "${staticflags[@]}" -o "$srcdir/go/bin/snap-exec" $GOFLAGS "${_gourl}/cmd/snap-exec"
  go build "${staticflags[@]}" -o "$srcdir/go/bin/snapctl" $GOFLAGS "${_gourl}/cmd/snapctl"

  # Generate data files such as real systemd units, dbus service, environment
  # setup helpers out of the available templates
  make -C data \
       BINDIR=/bin \
       LIBEXECDIR=/usr/lib \
       SYSTEMDSYSTEMUNITDIR=/usr/lib/systemd/system \
       SNAP_MOUNT_DIR=/var/lib/snapd/snap \
       SNAPD_ENVIRONMENT_FILE=/etc/default/snapd

  cd cmd
  autoreconf -i -f
  ./configure \
    --prefix=/usr \
    --libexecdir=/usr/lib/snapd \
    --with-snap-mount-dir=/var/lib/snapd/snap \
    --enable-apparmor \
    --enable-nvidia-biarch \
    --enable-merged-usr
  make $MAKEFLAGS
}

check() {
    export GOPATH="$srcdir/go"
    cd "$srcdir/go/src/${_gourl}"

    # make sure the binaries that need to be built statically really are
    for binary in snap-exec snap-update-ns snapctl; do
        if ! LC_ALL=C ldd "$srcdir/go/bin/$binary" 2>&1 | grep -q 'not a dynamic executable'; then
            echo "$binary is not a static binary"
            exit 1
        fi
    done
}

package() {
  cd "$pkgname-$pkgver"
  export GOPATH="$srcdir/go"

  # Install bash completion
  install -Dm644 data/completion/bash/snap \
    "$pkgdir/usr/share/bash-completion/completions/snap"
  install -Dm644 data/completion/bash/complete.sh \
    "$pkgdir/usr/lib/snapd/complete.sh"
  install -Dm644 data/completion/bash/etelpmoc.sh \
    "$pkgdir/usr/lib/snapd/etelpmoc.sh"
  # Install zsh completion
  install -Dm644 data/completion/zsh/_snap \
    "$pkgdir/usr/share/zsh/site-functions/_snap"

  # Install systemd units, dbus services and a script for environment variables
  make -C data/ install \
     DBUSSERVICESDIR=/usr/share/dbus-1/services \
     BINDIR=/usr/bin \
     SYSTEMDSYSTEMUNITDIR=/usr/lib/systemd/system \
     SNAP_MOUNT_DIR=/var/lib/snapd/snap \
     DESTDIR="$pkgdir"

  # Install polkit policy
  install -Dm644 data/polkit/io.snapcraft.snapd.policy \
    "$pkgdir/usr/share/polkit-1/actions/io.snapcraft.snapd.policy"

  # Install executables
  install -Dm755 "$srcdir/go/bin/snap" "$pkgdir/usr/bin/snap"
  install -Dm755 "$srcdir/go/bin/snapctl" "$pkgdir/usr/lib/snapd/snapctl"
  install -Dm755 "$srcdir/go/bin/snapd" "$pkgdir/usr/lib/snapd/snapd"
  install -Dm755 "$srcdir/go/bin/snap-seccomp" "$pkgdir/usr/lib/snapd/snap-seccomp"
  install -Dm755 "$srcdir/go/bin/snap-failure" "$pkgdir/usr/lib/snapd/snap-failure"
  install -Dm755 "$srcdir/go/bin/snapd-apparmor" "$pkgdir/usr/lib/snapd/snapd-apparmor"
  install -Dm755 "$srcdir/go/bin/snap-update-ns" "$pkgdir/usr/lib/snapd/snap-update-ns"
  install -Dm755 "$srcdir/go/bin/snap-exec" "$pkgdir/usr/lib/snapd/snap-exec"
  # Ensure /usr/bin/snapctl is a symlink to /usr/libexec/snapd/snapctl
  ln -s /usr/lib/snapd/snapctl "$pkgdir/usr/bin/snapctl"

  # pre-create directories
  install -dm755 "$pkgdir/var/lib/snapd/snap"
  install -dm755 "$pkgdir/var/cache/snapd"
  install -dm755 "$pkgdir/var/lib/snapd/apparmor"
  install -dm755 "$pkgdir/var/lib/snapd/assertions"
  install -dm755 "$pkgdir/var/lib/snapd/dbus-1/services"
  install -dm755 "$pkgdir/var/lib/snapd/dbus-1/system-services"
  install -dm755 "$pkgdir/var/lib/snapd/desktop/applications"
  install -dm755 "$pkgdir/var/lib/snapd/device"
  install -dm755 "$pkgdir/var/lib/snapd/hostfs"
  install -dm755 "$pkgdir/var/lib/snapd/mount"
  install -dm755 "$pkgdir/var/lib/snapd/seccomp/bpf"
  install -dm755 "$pkgdir/var/lib/snapd/snap/bin"
  install -dm755 "$pkgdir/var/lib/snapd/snaps"
  install -dm755 "$pkgdir/var/lib/snapd/inhibit"
  install -dm755 "$pkgdir/var/lib/snapd/lib/gl"
  install -dm755 "$pkgdir/var/lib/snapd/lib/gl32"
  install -dm755 "$pkgdir/var/lib/snapd/lib/vulkan"
  install -dm755 "$pkgdir/var/lib/snapd/lib/glvnd"
  # these dirs have special permissions
  install -dm111 "$pkgdir/var/lib/snapd/void"
  install -dm700 "$pkgdir/var/lib/snapd/cookie"
  install -dm700 "$pkgdir/var/lib/snapd/cache"

  make -C cmd install DESTDIR="$pkgdir/"

  # Install man file
  mkdir -p "$pkgdir/usr/share/man/man8"
  "$srcdir/go/bin/snap" help --man > "$pkgdir/usr/share/man/man8/snap.8"

  # Install the "info" data file with snapd version
  install -m 644 -D "$srcdir/go/src/${_gourl}/data/info" \
          "$pkgdir/usr/lib/snapd/info"

  # Remove snappy core specific units
  rm -fv "$pkgdir/usr/lib/systemd/system/snapd.system-shutdown.service"
  rm -fv "$pkgdir/usr/lib/systemd/system/snapd.autoimport.service"
  rm -fv "$pkgdir/usr/lib/systemd/system/snapd.recovery-chooser-trigger.service"
  rm -fv "$pkgdir"/usr/lib/systemd/system/snapd.snap-repair.*
  rm -fv "$pkgdir"/usr/lib/systemd/system/snapd.core-fixup.*
  # and scripts
  rm -fv "$pkgdir/usr/lib/snapd/snapd.core-fixup.sh"
  rm -fv "$pkgdir/usr/bin/ubuntu-core-launcher"
  rm -fv "$pkgdir/usr/lib/snapd/system-shutdown"
}
