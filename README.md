## WebBorer ##

WebBorer is a directory-enumeration tool written in Go and targeting CLI usage.

This project was formerly named 'GoBuster', but that had a namespace collision
with OJ Reeves' excellent tool (which was released about the same time as I was
preparing this for release).

### Features ###

* Highly portable -- requires no runtime once compiled.
* No GUI required.
* Supports Socks 4, 4a, and 5 proxies.
* Supports excluding entire subpaths.
* Capable of parsing returned HTML for additional directories to parse.
* Highly scalable -- Go's parallel model allows for many workers at once.

### Contributing ###

Please see the CONTRIBUTING file in this directory.

[![Build Status](https://travis-ci.org/Matir/webborer.svg?branch=master)](https://travis-ci.org/Matir/webborer)
[![codecov](https://codecov.io/gh/Matir/webborer/branch/master/graph/badge.svg)](https://codecov.io/gh/Matir/webborer)

### Copyright ###
Copyright 2015-2017 Google Inc.

WebBorer is not an official Google product (experimental or otherwise), it is
just code that happens to be owned by Google.

### Contact ###
For questions about WebBorer, contact David Tomaschik
<<davidtomaschik@google.com>>
