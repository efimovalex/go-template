#!/bin/bash

# skip /vendor/, /docs/swagger, /cmd/ and /internal/testhelpers
echo "Get list of project pkgs, skipping vendor, docs, cmd"
# NOTE: patterns must be sepparated with |
PKG_LIST=$(go list ./... | grep -Ev "vendor|docs")
COVERAGE_DIR=${COVERAGE_DIR:-.coverage}

echo "Run go mod tidy ..."
go mod tidy -v

echo "Run formating ..."
go fmt ./...

echo "Running tests and code coverage ..."
# Create a coverage file for each package
# test minim coverage
MINCOVERAGE=70

if [ -d $COVERAGE_DIR ]; then rm -rf $COVERAGE_DIR/*; else mkdir $COVERAGE_DIR; fi;

for package in $PKG_LIST; do 
    pkgcov=$(go test -covermode=atomic -race -coverprofile="$COVERAGE_DIR/$(basename $package).cov" "$package"); 
    retVal=$?;
    if [ $retVal -ne 0 ]; then 
        echo "🚨 TEST FAIL";
        echo "$pkgcov";
        echo;
        exit $retVal;
    fi;
    pcoverage=$(echo $pkgcov| grep "coverage" | sed -E "s/.*coverage: ([0-9]*\.[0-9]+)\% of statements/\1/g");
    if [ ! -z "$pcoverage" ]; then 
        if [ $(echo ${pcoverage%%.*}) -lt ${MINCOVERAGE} ]; then 
            echo "🚨 COVERAGE FAIL";
            echo "🚨 Test coverage of $package is $pcoverage%";
            echo;
            exit 1;
        else 
            echo "✅ Test coverage of $package is $pcoverage%";
        fi 
    else 
        echo "➖ No tests for $package";
    fi 
done
echo 'mode: atomic' > "$COVERAGE_DIR"/coverage.cov;
for fcov in "$COVERAGE_DIR"/*.cov; do 
    if [ $fcov != "$COVERAGE_DIR/coverage.cov" ]; then 
        tail -q -n +2 $fcov >> $COVERAGE_DIR/coverage.cov;
    fi 
done
pcoverage=$(go tool cover -func=$COVERAGE_DIR/coverage.cov | grep 'total' | awk '{print substr($3, 1, length($3)-1)}');
echo ;
if [ $(echo ${pcoverage%%.*}) -lt $MINCOVERAGE ]; then 
    echo "🚨 Test coverage of project is $pcoverage%";
    echo "FAIL";
    exit 1;
else 
    echo ">> ✅ Test coverage of project is $pcoverage%";
fi