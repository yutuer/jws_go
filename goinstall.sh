#!/bin/bash
#will generate pkg
go install $(go list ./... | grep -vf buildignore)
