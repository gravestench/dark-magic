package models

// ObjectGroup represents a group of possible Objects to spawn in a part of an area level.
type ObjectGroup struct {
	// GroupName is a reference field to define the Object Group name.
	GroupName string `csv:"GroupName"`

	// ObjectIDs contains the IDs of Objects assigned to this Object Group.
	ObjectID0 int `csv:"ID0"`
	ObjectID1 int `csv:"ID1"`
	ObjectID2 int `csv:"ID2"`
	ObjectID3 int `csv:"ID3"`
	ObjectID4 int `csv:"ID4"`
	ObjectID5 int `csv:"ID5"`
	ObjectID6 int `csv:"ID6"`
	ObjectID7 int `csv:"ID7"`

	// ObjectDensities controls the number of Objects to spawn in the area level.
	// This is also affected by the Object's populate function defined by the "PopulateFn" field from the objects.txt file.
	// The maximum value allowed is 128.
	ObjectDesnity0 int `csv:"DENSITY0"`
	ObjectDesnity1 int `csv:"DENSITY1"`
	ObjectDesnity2 int `csv:"DENSITY2"`
	ObjectDesnity3 int `csv:"DENSITY3"`
	ObjectDesnity4 int `csv:"DENSITY4"`
	ObjectDesnity5 int `csv:"DENSITY5"`
	ObjectDesnity6 int `csv:"DENSITY6"`
	ObjectDesnity7 int `csv:"DENSITY7"`

	// ObjectProbabilities control the probability that the Object will spawn in the area level.
	// This is calculated in order, so the first probability that is successful will be chosen.
	// These field values should add up to exactly 100 in total to guarantee that one of the objects spawns.
	ObjectProbability0 int `csv:"PROB0"`
	ObjectProbability1 int `csv:"PROB1"`
	ObjectProbability2 int `csv:"PROB2"`
	ObjectProbability3 int `csv:"PROB3"`
	ObjectProbability4 int `csv:"PROB4"`
	ObjectProbability5 int `csv:"PROB5"`
	ObjectProbability6 int `csv:"PROB6"`
	ObjectProbability7 int `csv:"PROB7"`
}
