echo "Setting version of package to: $1 ..."
git tag $1 && git push origin $1 && git ls-remote --tags origin | grep $1 || true