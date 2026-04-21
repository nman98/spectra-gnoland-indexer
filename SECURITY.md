# Security Policy

The Spectra Gnoland Indexer project is open source and all the code is available on GitHub. The main maintainer is Cogwheel Validator.
Any kind of security issues should be reported to the Cogwheel Validator at <info@cogwheel.zone>.

## Supported Versions

This project consists of 2 main components: the indexer and the API. The current project is considered not stable
but it will follow N-Only or Latest Point Release Policy. This means that the latest release will be supported with security updates. If you are using a version that is not the latest, you should upgrade to the latest version as soon as possible.

So when looking for a security issue or any other issue check the latest version of the project. This usually means latest release tag.

## What Is Considered a Security Vulnerability

This are some suggestions only but this is some general guidance on what is considered a security vulnerability.

On the API any sort of SQL injection, XSS, CSRF, etc. is considered a security vulnerability. This also includes any
dependency that is not up to date and has known vulnerabilities.

On the indexer any sort of code execution vulnerability, SQL injection, etc. is considered a security vulnerability.

Any sort of vulnerability that can be used to gain access to the database or the indexer is considered a security vulnerability.

## Reporting a Vulnerability

If you find a security vulnerability please report it to the Cogwheel Validator at <info@cogwheel.zone>.
Please provide a detailed description of the vulnerability and how to reproduce it.
If possible provide a proof of concept or a code snippet that demonstrates the vulnerability.

In the email include:

- A commit hash of the version you are using.
- A detailed description of the vulnerability.
- A proof of concept or a code snippet that demonstrates the vulnerability.
- The steps to reproduce the vulnerability.
- The expected behavior.
- The actual behavior.
- The affected components.
- The affected files.
- The affected lines of code.
- The affected functions.
