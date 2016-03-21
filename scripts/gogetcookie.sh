#!/bin/bash

# https://github.com/golang/go/issues/12933

touch ~/.gitcookies
chmod 0600 ~/.gitcookies

git config --global http.cookiefile ~/.gitcookies

tr , \\t <<\__END__ >>~/.gitcookies
go.googlesource.com,FALSE,/,TRUE,2147483647,o,git-celrenheit.gmail.com=1/ayC7N4Y01-O8fhm2H1BJdpXZwi1tKlIOutPnw3kw2R4
go-review.googlesource.com,FALSE,/,TRUE,2147483647,o,git-celrenheit.gmail.com=1/ayC7N4Y01-O8fhm2H1BJdpXZwi1tKlIOutPnw3kw2R4
__END__
