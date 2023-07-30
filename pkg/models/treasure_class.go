package models

// TreasureClass represents the Treasure Class from the original game.
type TreasureClass struct {
	TreasureClass string  `csv:"Treasure Class"` // Unique Treasure Class ID.
	Picks         int     `csv:"Picks"`          // Number of item drop chances rolled or total guaranteed quantity of item drops.
	Unique        float64 `csv:"Unique"`         // Modifies the item ratio drop for a Unique Quality item.
	Set           float64 `csv:"Set"`            // Modifies the item ratio drop for a Set Quality item.
	Rare          float64 `csv:"Rare"`           // Modifies the item ratio drop for a Rare Quality item.
	Magic         float64 `csv:"Magic"`          // Modifies the item ratio drop for a Magic Quality item.
	NoDrop        float64 `csv:"NoDrop"`         // Probability of no item dropping by the Treasure Class.
	Item1         string  `csv:"Item1"`          // Potential Item Type for drop.
	Item2         string  `csv:"Item2"`          // Additional potential Item Type for drop.
	Item3         string  `csv:"Item3"`          // Additional potential Item Type for drop.
	Item4         string  `csv:"Item4"`          // Additional potential Item Type for drop.
	Item5         string  `csv:"Item5"`          // Additional potential Item Type for drop.
	Item6         string  `csv:"Item6"`          // Additional potential Item Type for drop.
	Item7         string  `csv:"Item7"`          // Additional potential Item Type for drop.
	Item8         string  `csv:"Item8"`          // Additional potential Item Type for drop.
	Item9         string  `csv:"Item9"`          // Additional potential Item Type for drop.
	Item10        string  `csv:"Item10"`         // Additional potential Item Type for drop.
	Prob1         float64 `csv:"Prob1"`          // Probability for the drop of Item1.
	Prob2         float64 `csv:"Prob2"`          // Probability for the drop of Item2.
	Prob3         float64 `csv:"Prob3"`          // Probability for the drop of Item3.
	Prob4         float64 `csv:"Prob4"`          // Probability for the drop of Item4.
	Prob5         float64 `csv:"Prob5"`          // Probability for the drop of Item5.
	Prob6         float64 `csv:"Prob6"`          // Probability for the drop of Item6.
	Prob7         float64 `csv:"Prob7"`          // Probability for the drop of Item7.
	Prob8         float64 `csv:"Prob8"`          // Probability for the drop of Item8.
	Prob9         float64 `csv:"Prob9"`          // Probability for the drop of Item9.
	Prob10        float64 `csv:"Prob10"`         // Probability for the drop of Item10.
	Term          int     `csv:"Term"`           // Indicates the end of the Treasure Class data.
}

// TreasureClassEx represents the Treasure Class linked to a monster drop.
type TreasureClassEx struct {
	TreasureClass     string  `csv:"Treasure Class"`    // Unique Treasure Class ID.
	Group             int     `csv:"group"`             // Treasure Class group ID for item drop calculation.
	Level             int     `csv:"level"`             // Level of the Treasure Class.
	Picks             int     `csv:"Picks"`             // Number of item drop chances rolled or total guaranteed quantity of item drops.
	Unique            float64 `csv:"Unique"`            // Modifies the item ratio drop for a Unique Quality item.
	Set               float64 `csv:"Set"`               // Modifies the item ratio drop for a Set Quality item.
	Rare              float64 `csv:"Rare"`              // Modifies the item ratio drop for a Rare Quality item.
	Magic             float64 `csv:"Magic"`             // Modifies the item ratio drop for a Magic Quality item.
	NoDrop            float64 `csv:"NoDrop"`            // Probability of no item dropping by the Treasure Class.
	Item1             string  `csv:"Item1"`             // Potential Item Type or other Treasure Class that can drop.
	Item2             string  `csv:"Item2"`             // Additional potential Item Type or Treasure Class.
	Item3             string  `csv:"Item3"`             // Additional potential Item Type or Treasure Class.
	Item4             string  `csv:"Item4"`             // Additional potential Item Type or Treasure Class.
	Item5             string  `csv:"Item5"`             // Additional potential Item Type or Treasure Class.
	Item6             string  `csv:"Item6"`             // Additional potential Item Type or Treasure Class.
	Item7             string  `csv:"Item7"`             // Additional potential Item Type or Treasure Class.
	Item8             string  `csv:"Item8"`             // Additional potential Item Type or Treasure Class.
	Item9             string  `csv:"Item9"`             // Additional potential Item Type or Treasure Class.
	Item10            string  `csv:"Item10"`            // Additional potential Item Type or Treasure Class.
	Prob1             float64 `csv:"Prob1"`             // Probability for the drop of Item1.
	Prob2             float64 `csv:"Prob2"`             // Probability for the drop of Item2.
	Prob3             float64 `csv:"Prob3"`             // Probability for the drop of Item3.
	Prob4             float64 `csv:"Prob4"`             // Probability for the drop of Item4.
	Prob5             float64 `csv:"Prob5"`             // Probability for the drop of Item5.
	Prob6             float64 `csv:"Prob6"`             // Probability for the drop of Item6.
	Prob7             float64 `csv:"Prob7"`             // Probability for the drop of Item7.
	Prob8             float64 `csv:"Prob8"`             // Probability for the drop of Item8.
	Prob9             float64 `csv:"Prob9"`             // Probability for the drop of Item9.
	Prob10            float64 `csv:"Prob10"`            // Probability for the drop of Item10.
	FirstLadderSeason int     `csv:"firstLadderSeason"` // Starting ladder season for rolling in ladder games (inclusive).
	LastLadderSeason  int     `csv:"lastLadderSeason"`  // Last ladder season for which the treasure class is ladder-only (inclusive).
	NoAlwaysDrop      bool    `csv:"noAlwaysDrop"`      // If true, this treasure class will roll normally when being forced to always drop.
}
