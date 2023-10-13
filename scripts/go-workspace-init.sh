#!/bin/bash

test ! -f go.work && go work init
go work use -r .
