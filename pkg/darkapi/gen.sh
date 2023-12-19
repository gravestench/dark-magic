#!/bin/bash

go run github.com/traefik/yaegi/internal/cmd/extract github.com/gravestench/dark-magic/pkg/models

# NOTE: work-around for bug in yaegi/internal/cmd/extract; replace "pkg/models/models" with "pkg/models".
sed -i 's/\/models\/models/\/models/' 'go1_21_github.com_gravestench_dark-magic_pkg_models.go'
