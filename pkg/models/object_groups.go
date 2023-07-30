package models

// ObjectGroup represents a group of possible Objects to spawn in a part of an area level.
type ObjectGroup struct {
	// GroupName is a reference field to define the Object Group name.
	GroupName string `csv:"GroupName"`

	// ObjectIDs contains the IDs of Objects assigned to this Object Group.
	ObjectIDs [8]int `csv:"ID0,ID1,ID2,ID3,ID4,ID5,ID6,ID7"`

	// ObjectDensities controls the number of Objects to spawn in the area level.
	// This is also affected by the Object's populate function defined by the "PopulateFn" field from the objects.txt file.
	// The maximum value allowed is 128.
	ObjectDensities [8]int `csv:"DENSITY0,DENSITY1,DENSITY2,DENSITY3,DENSITY4,DENSITY5,DENSITY6,DENSITY7"`

	// ObjectProbabilities control the probability that the Object will spawn in the area level.
	// This is calculated in order, so the first probability that is successful will be chosen.
	// These field values should add up to exactly 100 in total to guarantee that one of the objects spawns.
	ObjectProbabilities [8]int `csv:"PROB0,PROB1,PROB2,PROB3,PROB4,PROB5,PROB6,PROB7"`
}
