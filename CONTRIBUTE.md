# Contributing

The `go-python` project (and `gopy`) eagerly accepts contributions from the community.

## Introduction


The `go-python` project provides libraries and tools in Go for the Go community to better integrate with Python projects and libraries, and we would like you to join us in improving `go-python`'s quality and scope.
This document is for contributors or those interested in contributing.
Questions about `go-python` and the use of its libraries can be directed to the [go-python](mailto:go-python@googlegroups.com) mailing list.

## Contributing

### Working Together

When contributing or otherwise participating, please:

- Be friendly and welcoming
- Be patient
- Be thoughtful
- Be respectful
- Be charitable
- Avoid destructive behavior

Excerpted from the [Go conduct document](https://golang.org/conduct).

### Reporting Bugs

When you encounter a bug, please open an issue on the corresponding repository.
Start the issue title with the repository/sub-repository name, like `bind: issue name`.
Be specific about the environment you encountered the bug in (_e.g.:_ operating system, Go compiler version, ...).
If you are able to write a reproducer for the bug, please include it in the issue.
As a rule, we keep all tests OK and try to increase code coverage.

### Suggesting Enhancements

If the scope of the enhancement is small, open an issue.
If it is large, such as suggesting a new repository, sub-repository, or interface refactoring, then please start a discussion on [the go-python list](https://groups.google.com/forum/#!forum/go-python).

### Your First Code Contribution

If you are a new contributor, *thank you!*
Before your first merge, you will need to be added to the [CONTRIBUTORS](https://github.com/go-python/license/blob/master/CONTRIBUTORS) and [AUTHORS](https://github.com/go-python/license/blob/master/AUTHORS) files.
Open a pull request adding yourself to these files.
All `go-python` code follows the BSD license in the [license document](https://github.com/go-python/license/blob/master/LICENSE).
We prefer that code contributions do not come with additional licensing.
For exceptions, added code must also follow a BSD license.

### Code Contribution

If it is possible to split a large pull request into two or more smaller pull requests, please try to do so.
Pull requests should include tests for any new code before merging.
It is ok to start a pull request on partially implemented code to get feedback, and see if your approach to a problem is sound.
You don't need to have tests, or even have code that compiles to open a pull request, although both will be needed before merge.
When tests use magic numbers, please include a comment explaining the source of the number.
Benchmarks are optional for new features, but if you are submitting a pull request justified by performance improvement, you will need benchmarks to measure the impact of your change, and the pull request should include a report from [benchcmp](https://godoc.org/golang.org/x/tools/cmd/benchcmp) or, preferably, [benchstat](https://godoc.org/golang.org/x/perf/cmd/benchstat).

Commit messages also follow some rules.
They are best explained at the official [Go](https://golang.org) "Contributing guidelines" document:

[golang.org/doc/contribute.html](https://golang.org/doc/contribute.html#commit_changes)

For example:

```
bind: add support for cffi

This CL adds support for the cffi python backend.
The existing implementation, cpython, only generates code for the
CPython-2 C API.
Now, with cffi, we support generation of python modules for Python-2,
Python-3 and PyPy VMs.

Fixes go-python/gopy#42.
```

If the `CL` modifies multiple packages at the same time, include them in the commit message:

```
py,bind: implement wrapping of Go interfaces

bla-bla

Fixes go-python/gopy#40.
```

Please always format your code with [goimports](https://godoc.org/golang.org/x/tools/cmd/goimports).
Best is to have it invoked as a hook when you save your `.go` files.

Files in the `go-python` repository don't list author names, both to avoid clutter and to avoid having to keep the lists up to date.
Instead, your name will appear in the change log and in the [CONTRIBUTORS](https://github.com/go-python/license/blob/master/CONTRIBUTORS) and [AUTHORS](https://github.com/go-python/license/blob/master/AUTHORS) files.

New files that you contribute should use the standard copyright header:

```
// Copyright 20xx The go-python Authors.  All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
```

Files in the repository are copyright the year they are added.
Do not update the copyright year on files that you change.

### Code Review

If you are a contributor, please be welcoming to new contributors.
[Here](http://sarah.thesharps.us/2014/09/01/the-gentle-art-of-patch-review/) is a good guide.

There are several terms code reviewers may use that you should become familiar with.

  * ` LGTM ` — looks good to me
  * ` SGTM ` — sounds good to me
  * ` PTAL ` — please take another look
  * ` CL ` — change list; a single commit in the repository
  * ` s/foo/bar/ ` — please replace ` foo ` with ` bar `; this is [sed syntax](http://en.wikipedia.org/wiki/Sed#Usage)
  * ` s/foo/bar/g ` — please replace ` foo ` with ` bar ` throughout your entire change

We follow the convention of requiring at least 1 reviewer to say LGTM before a merge.
When code is tricky or controversial, submitters and reviewers can request additional review from others and more LGTMs before merge.
You can ask for more review by saying PTAL in a comment in a pull request.
You can follow a PTAL with one or more @someone to get the attention of particular people.
If you don't know who to ask, and aren't getting enough review after saying PTAL, then PTAL @go-python/developers will get more attention.
Also note that you do not have to be the pull request submitter to request additional review.

### What Can I Do to Help?

If you are looking for some way to help the `go-python` project, there are good places to start, depending on what you are comfortable with.

- You can [search](https://github.com/issues?utf8=%E2%9C%93&q=is%3Aopen+is%3Aissue+user%3Ago-python) for open issues in need of resolution.
- You can improve documentation, or improve examples.
- You can add and improve tests.
- You can improve performance, either by improving accuracy, speed, or both.
- You can suggest and implement new features that you think belong in `go-python`.

### Style

We use [Go style](https://github.com/golang/go/wiki/CodeReviewComments).

---

This _"Contributing"_ guide has been extracted from the [Gonum](https://gonum.org) project.
Its guide is [here](https://github.com/gonum/license/blob/master/CONTRIBUTING.md).
