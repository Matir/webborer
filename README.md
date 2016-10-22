## GoBuster ##

GoBuster is a directory-enumeration tool written in Go and targeting CLI usage.

### Features ###

* Highly portable -- requires no runtime once compiled.
* No GUI required.
* Supports Socks 4, 4a, and 5 proxies.
* Supports excluding entire subpaths.
* Capable of parsing returned HTML for additional directories to parse.
* Highly scalable -- Go's parallel model allows for many workers at once.

### Contributing ###

Please see the CONTRIBUTING file in this directory.

[![Build Status](https://travis-ci.org/Matir/gobuster.svg?branch=master)](https://travis-ci.org/Matir/gobuster)
[![codecov](https://codecov.io/gh/Matir/gobuster/branch/master/graph/badge.svg)](https://codecov.io/gh/Matir/gobuster)

### Copyright ###
Copyright 2015-2016 Google Inc.

GoBuster is not an official Google product (experimental or otherwise), it is
just code that happens to be owned by Google.

### Contact ###
For questions about GoBuster, contact David Tomaschik
<<davidtomaschik@google.com>>
