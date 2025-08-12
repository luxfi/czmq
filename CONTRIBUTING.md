# Contributing to LuxFi CZMQ

The contributors are listed in AUTHORS (add yourself). This project uses the MPL v2 license, see LICENSE.

# Our Process
Before you send a pull request, please familiarize yourself with the [C4.1 Collective Code Construction Contract](http://rfc.zeromq.org/spec:22) process. A quick summary (but please, do read the process document):
* A Pull Request should be described in the form of a problem statement.
* The code included with the pull request should be a proposed solution to that problem.
* The submitted code should adhere to our style guidelines (described below).
* The submitted code should include tests.
* The submitted code should not break any existing tests.

"A Problem" should be one single clear problem. Large complex problems should be broken down into a series of smaller problems when ever possible.

**Please be aware** that LuxFi CZMQ follows semantic versioning and is currently at **v4.2.0**. We merge to master. We deploy from master. Master is expected to be working, at all times. We strive to do our very best to never break public API in this library. Changes can be additive, but they can not break the existing API. This fork maintains the `/v4` module path for Go module compatibility.

# Style Guide
* Your code must be formatted with [Gofmt](https://blog.golang.org/go-fmt-your-code)
* Your code should pass [golint](https://github.com/golang/lint). If for some reason it cannot, please provide an explanation.
* Your code should pass [go vet](https://golang.org/cmd/vet/)


