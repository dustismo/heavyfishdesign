#!/bin/sh
# This simple script for doing a production release 
# Assumes the docs site is in a sibling directory called heavyfishdesign-docs

# echo Running tests
go test ./...

read -p "Continue? " -n 1 -r
echo    # (optional) move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

# echo Running diff_Test

go run main.go diff_test

read -p "Continue? " -n 1 -r
echo    # (optional) move to a new line
if [[ ! $REPLY =~ ^[Yy]$ ]]
then
    exit 1
fi

# echo Updating designs

go run main.go designs_updated

# echo Rendering Documentation

cd docs && make html && cd ..

# now create bundle
mkdir build_tmp
mkdir build_tmp/heavyfishdesign

echo Building windows binaries
GOOS=windows GOARCH=amd64 go build main.go
mv main.exe build_tmp/heavyfishdesign/hfd-windows.exe

echo Building mac binaries
GOOS=darwin GOARCH=amd64 go build main.go
mv main build_tmp/heavyfishdesign/hfd-mac

cp -R designs build_tmp/heavyfishdesign/designs

cd build_tmp
zip -r heavyfishdesign.zip heavyfishdesign
mv heavyfishdesign.zip ../../heavyfishdesign-docs/html/_static
cd ..
rm -R build_tmp

echo All Done!