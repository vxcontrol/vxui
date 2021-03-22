[ "$PACKAGE_VER" ] || PACKAGE_VER=$(git describe --tags `git rev-list --tags --max-count=1`)
[ "$PACKAGE_REV" ] || PACKAGE_REV=$(git rev-parse --short HEAD)

# for debugging -gcflags="all=-N -l" 
go build -ldflags "-X main.PackageVer=$PACKAGE_VER -X main.PackageRev=$PACKAGE_REV" -o app main.go
