package gameScene

import "github.com/gravestench/servicemesh"

import "github.com/gravestench/dark-magic/pkg/services/configManager"

var (
	_ servicemesh.Service            = &Service{}
	_ servicemesh.HasLogger          = &Service{}
	_ servicemesh.HasDependencies    = &Service{}
	_ configManager.HasConfiguration = &Service{}
)
