package models

// PetType represents the various statistics for each type of pet from all the classes summon Skills.
type PetType struct {
	PetType   string    `csv:"pet type"`                        // Defines the name of the pet type, used in the "pettype" column in skills.txt.
	Group     int       `csv:"group"`                           // Used as an ID field, where if pet types share the same group value, then only 1 pet of that group is allowed to be alive at any time. If equals 0 (or null), then ignore this.
	BaseMax   int       `csv:"basemax"`                         // This sets a baseline maximum number of pets allowed to be alive when skill levels are reset or changed.
	Warp      bool      `csv:"warp"`                            // Boolean field. If equals 1, then the Pet will teleport to the player when the player teleports or warps to another area. If equals 0, then the pet will die instead.
	Range     bool      `csv:"range"`                           // Boolean field. If equals 1, then the Pet will die if the player teleports or warps to another area and is located more than 40 grid tiles in distances from the Pet. If equals 0, then ignore this.
	PartySend bool      `csv:"partysend"`                       // Boolean field. If equals 1, then tell the Pet to do the Party Location Update command (find the location of its Player) when its health changes. If equals 0, then ignore this.
	Unsummon  bool      `csv:"unsummon"`                        // Boolean field. If equals 1, then the Pet can be unsummoned by the Unsummon skill function. If equals 0, then the Pet cannot be unsummoned.
	Automap   bool      `csv:"automap"`                         // Boolean field. If equals 1, then display the Pet on the Automap. If equals 0, then hide the pet on the Automap.
	Name      string    `csv:"name"`                            // String Key. Used to define the Pet's name on its party frame.
	DrawHP    bool      `csv:"drawhp"`                          // Boolean field. If equals 1, then display the Pet's Life bar under the party frame. If equals 0, then hide the Pet's Life bar under the party icon.
	IconType  int       `csv:"icontype"`                        // Controls the functionality for how to display the Pet Icon and the number of Pets counter. (0: Do not display, 1: Display icon, 2: Display icon and counter)
	BaseIcon  string    `csv:"baseicon"`                        // Define which DC6 file to use for the default Pet's icon in its party frame.
	MClass    [4]int    `csv:"mclass1,mclass2,mclass3,mclass4"` // Defines the alternative pets to use for the "pet type" by using their specific unit's "hcIdx" from Monstats.txt.
	MIcon     [4]string `csv:"micon1,micon2,micon3,micon4"`     // Defines which DC6 files to use for the related "mclass" Pets' icons in their party frame.
}
