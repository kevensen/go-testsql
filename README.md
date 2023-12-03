# go-testsql

## Background

The ideal unit test is hermetic. The code is self contained and isolated.  A 
number of libraries exist that provide for the mocking of complicated 
dependencies, like those with a SQL server (see 
https://github.com/DATA-DOG/go-sqlmock).  These are appropriate to use in many
cases.

However, there are some cases where the interaction between the code under test
and the server must be tested.  For example, if you have a custom SQL query
statement in your code under test, it cannot quite so easily mocked.

## Purpose

This library is meant to address those corner cases, providing a mechanism
to start a database container and expose the connection.  The container
port is not bound to the host port and should not create a conflict
with any existing containers.

## Requirements

* Docker