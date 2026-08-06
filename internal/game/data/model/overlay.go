package models

// Overlay represents overlay graphics related to states, auras, cast animations, curses, and buffs.
type Overlay struct {
	Overlay       string `csv:"overlay"`       // Defines the name of the overlay, used in other data files.
	Filename      string `csv:"Filename"`      // Defines which DCC file to use for the Overlay.
	Version       int    `csv:"version"`       // Defines which game version to use this Overlay (0 = Classic mode | 100 = Expansion mode).
	Character     string `csv:"Character"`     // Used for name categorizing Overlays for unit translation mapping.
	PreDraw       bool   `csv:"PreDraw"`       // Boolean field. If equals 1, then display the Overlay in front of sprites. If equals 0, then display the Overlay behind sprites.
	OneOfN        int    `csv:"1ofN"`          // Controls how to randomly display Overlays. This value will randomly add to the current index of the Overlay to possibly use another Overlay that is indexed after this current Overlay.
	XOffset       int    `csv:"Xoffset"`       // Sets the horizontal offset of the overlay on the unit. Positive values move it toward the left and negative values move it towards the right.
	YOffset       int    `csv:"Yoffset"`       // Sets the vertical offset of the overlay on the unit. Positive values move it down and negative values move it up.
	Height1       int    `csv:"Height1"`       // Additional value added to "Yoffset" depending on "OverlayHeight" field value from monstats2.txt.
	Height2       int    `csv:"Height2"`       // Additional value added to "Yoffset" for player unit types.
	AnimRate      int    `csv:"AnimRate"`      // Controls the animation frame rate of the Overlay. The value is the number of frames that will update per second.
	LoopWaitTime  int    `csv:"LoopWaitTime"`  // Controls the number of periodic frames to wait until redrawing the Overlay. This only works with Overlays that are a loop type.
	Trans         int    `csv:"Trans"`         // Controls the alpha mode for how the Overlay is displayed, which can affect transparency and blending.
	InitRadius    int    `csv:"InitRadius"`    // Controls the starting Light Radius value for the Overlay (Max = 18).
	Radius        int    `csv:"Radius"`        // Controls the maximum Light Radius value for the Overlay. This can only be greater than or equal to "InitRadius".
	Red           int    `csv:"Red"`           // Controls the Red color gradient of the Light Radius.
	Green         int    `csv:"Green"`         // Controls the Green color gradient of the Light Radius.
	Blue          int    `csv:"Blue"`          // Controls the Blue color gradient of the Light Radius.
	NumDirections int    `csv:"NumDirections"` // The number of directions in the cell file.
	LocalBlood    int    `csv:"LocalBlood"`    // Controls how to display green blood or VFX on a unit.
}
