package models

// Object represents the functionalities of all objects found in area levels.
type Object struct {
	// Defines the unique type class of the object which is used to reference this object. These are also defined in the objpreset.txt file.
	Class int `csv:"Class"`

	// String key. Used as the display name of the object when being highlighted by the player.
	Name string `csv:"Name"`

	// Determines what files to use to display the graphics of the object. These are defined by the ObjType.txt file.
	Token string `csv:"Token"`

	// Boolean Field. If equals 1, then the object can be selected by the player and highlighted when hovered on by the mouse cursor.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	Selectable0 bool `csv:"Selectable0"`
	Selectable1 bool `csv:"Selectable1"`
	Selectable2 bool `csv:"Selectable2"`
	Selectable3 bool `csv:"Selectable3"`
	Selectable4 bool `csv:"Selectable4"`
	Selectable5 bool `csv:"Selectable5"`
	Selectable6 bool `csv:"Selectable6"`
	Selectable7 bool `csv:"Selectable7"`

	// Controls the amount of sub tiles that the object occupies using X and Y coordinates. This is generally used for measuring the object's size when trying to spawn objects in rooms and controlling their distances apart.
	SizeX int `csv:"SizeX"`
	SizeY int `csv:"SizeY"`

	// Controls the frame length of the object's mode. If this equals 0, then that mode will be skipped.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	FrameCnt0 int `csv:"FrameCnt0"`
	FrameCnt1 int `csv:"FrameCnt1"`
	FrameCnt2 int `csv:"FrameCnt2"`
	FrameCnt3 int `csv:"FrameCnt3"`
	FrameCnt4 int `csv:"FrameCnt4"`
	FrameCnt5 int `csv:"FrameCnt5"`
	FrameCnt6 int `csv:"FrameCnt6"`
	FrameCnt7 int `csv:"FrameCnt7"`

	// Controls the animation frame rate of how many frames to update per delta (Measured in 256ths).
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	FrameDelta0 int `csv:"FrameDelta0"`
	FrameDelta1 int `csv:"FrameDelta1"`
	FrameDelta2 int `csv:"FrameDelta2"`
	FrameDelta3 int `csv:"FrameDelta3"`
	FrameDelta4 int `csv:"FrameDelta4"`
	FrameDelta5 int `csv:"FrameDelta5"`
	FrameDelta6 int `csv:"FrameDelta6"`
	FrameDelta7 int `csv:"FrameDelta7"`

	// Boolean Field. If equals 1, then the object's current animation will loop back to play again when it finishes.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	CycleAnim0 bool `csv:"CycleAnim0"`
	CycleAnim1 bool `csv:"CycleAnim1"`
	CycleAnim2 bool `csv:"CycleAnim2"`
	CycleAnim3 bool `csv:"CycleAnim3"`
	CycleAnim4 bool `csv:"CycleAnim4"`
	CycleAnim5 bool `csv:"CycleAnim5"`
	CycleAnim6 bool `csv:"CycleAnim6"`
	CycleAnim7 bool `csv:"CycleAnim7"`

	// Controls the Light Radius distance value for the object. If this value equals 0, then the object will not emit a Light Radius.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	Lit0 int `csv:"Lit0"`
	Lit1 int `csv:"Lit1"`
	Lit2 int `csv:"Lit2"`
	Lit3 int `csv:"Lit3"`
	Lit4 int `csv:"Lit4"`
	Lit5 int `csv:"Lit5"`
	Lit6 int `csv:"Lit6"`
	Lit7 int `csv:"Lit7"`

	// Boolean Field. If equals 1, then the object will draw a shadow. If equals 0, then the object will not draw a shadow.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	BlocksLight0 int `csv:"BlocksLight0"`
	BlocksLight1 int `csv:"BlocksLight1"`
	BlocksLight2 int `csv:"BlocksLight2"`
	BlocksLight3 int `csv:"BlocksLight3"`
	BlocksLight4 int `csv:"BlocksLight4"`
	BlocksLight5 int `csv:"BlocksLight5"`
	BlocksLight6 int `csv:"BlocksLight6"`
	BlocksLight7 int `csv:"BlocksLight7"`

	// Boolean Field. If equals 1, then the object will have collision. If equals 0, then the object will not have collision, and units can walk through it.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	HasCollision0 bool `csv:"HasCollision0"`
	HasCollision1 bool `csv:"HasCollision1"`
	HasCollision2 bool `csv:"HasCollision2"`
	HasCollision3 bool `csv:"HasCollision3"`
	HasCollision4 bool `csv:"HasCollision4"`
	HasCollision5 bool `csv:"HasCollision5"`
	HasCollision6 bool `csv:"HasCollision6"`
	HasCollision7 bool `csv:"HasCollision7"`

	// Boolean Field. If equals 1, then the player can target this object to be attacked, and the player will use the Kick skill when operating the object.
	// If the object has the Class equal to "CompellingOrb" or "SoulStoneForge", then instead of using the Kick skill, players will use the Attack skill when operating the object.
	// If equals 0, then ignore this, and the player will not use a skill or animation when operating the object.
	IsAttackable0 bool `csv:"IsAttackable0"`

	// Controls the frame for where the object will start playing the next animation.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	Start0 int `csv:"Start0"`
	Start1 int `csv:"Start1"`
	Start2 int `csv:"Start2"`
	Start3 int `csv:"Start3"`
	Start4 int `csv:"Start4"`
	Start5 int `csv:"Start5"`
	Start6 int `csv:"Start6"`
	Start7 int `csv:"Start7"`

	// Boolean Field. If equals 1, then enable the object to update its mode based on the game's time of day.
	// If equals 0, then the object will not update its mode based on the time of day.
	EnvEffect bool `csv:"EnvEffect"`

	// Boolean Field. If equals 1, then the object will be treated as a door when the game handles its collision, animation properties, tooltips, and commands.
	// If equals 0, then ignore this.
	IsDoor bool `csv:"IsDoor"`

	// Boolean Field. If equals 1, then the object will block the player's line of sight to see anything beyond the object.
	// If equals 0, then ignore this. This field relies on the "IsDoor" field being enabled.
	BlocksVis bool `csv:"BlocksVis"`

	// Determines the object's orientation type, which can affect mouse selection priority of the object when a unit is being rendered in front of or behind the object (such as a door object covering a unit and how the mouse selection should handle that).
	// This also affects the randomization of the coordinates when spawning the object near the edge of a room.
	Orientation int `csv:"Orientation"`

	// Controls how the object's sprite is drawn, which can affect how it is displayed in Perspective game camera mode.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	OrderFlag0 int `csv:"OrderFlag0"`
	OrderFlag1 int `csv:"OrderFlag1"`
	OrderFlag2 int `csv:"OrderFlag2"`
	OrderFlag3 int `csv:"OrderFlag3"`
	OrderFlag4 int `csv:"OrderFlag4"`
	OrderFlag5 int `csv:"OrderFlag5"`
	OrderFlag6 int `csv:"OrderFlag6"`
	OrderFlag7 int `csv:"OrderFlag7"`

	// Boolean Field. If equals 1, then enable a random chance that the object will spawn in already in Opened mode.
	// The game will choose a 1/14 chance that this can happen when the object is spawned. If equals 0, then ignore this.
	PreOperate bool `csv:"PreOperate"`

	// Boolean Field. If equals 1, then confirm that this object has the correlating mode.
	// If equals 0, then this object will not have the correlating mode. This flag can affect how the object functions work.
	// Each field is numbered, correlating to 1 of 8 Object Modes that the object uses (See Overview section, or ObjMode.txt).
	Mode0 int `csv:"Mode0"`
	Mode1 int `csv:"Mode1"`
	Mode2 int `csv:"Mode2"`
	Mode3 int `csv:"Mode3"`
	Mode4 int `csv:"Mode4"`
	Mode5 int `csv:"Mode5"`
	Mode6 int `csv:"Mode6"`
	Mode7 int `csv:"Mode7"`

	// Controls the offset values in the X and Y directions for the object's visual graphics. This is measured in game pixels.
	Xoffset int `csv:"Xoffset"`
	Yoffset int `csv:"Yoffset"`

	// Boolean Field. If equals 1, then draw the object's shadows. If equals 0, then do not draw the object's shadows.
	Draw bool `csv:"Draw"`

	// Controls the Red color gradient of the object's Light Radius. This field depends on the "Lit#" field having a value greater than 0.
	Red int `csv:"Red"`

	// Controls the Green color gradient of the object's Light Radius. This field depends on the "Lit#" field having a value greater than 0.
	Green int `csv:"Green"`

	// Controls the Blue color gradient of the object's Light Radius. This field depends on the "Lit#" field having a value greater than 0.
	Blue int `csv:"Blue"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Head composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	HD bool `csv:"HD"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Torso composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	TR bool `csv:"TR"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Legs composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	LG bool `csv:"LG"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Right Arm composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	RA bool `csv:"RA"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Left Arm composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	LA bool `csv:"LA"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Right Hand composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	RH bool `csv:"RH"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Left Hand composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	LH bool `csv:"LH"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Shield composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	SH bool `csv:"SH"`

	// Boolean Field. If equals 1, then the object will be flagged to have a Special # composite piece, and the game will use the component system to handle the object's mouse selection collision box.
	// If equals 0, then ignore this.
	S1 bool `csv:"S1"`
	S2 bool `csv:"S2"`
	S3 bool `csv:"S3"`
	S4 bool `csv:"S4"`
	S5 bool `csv:"S5"`
	S6 bool `csv:"S6"`
	S7 bool `csv:"S7"`
	S8 bool `csv:"S8"`

	// Defines the total amount of composite pieces. If this value is greater than 1, then the game will treat the object with the multiple composite piece system, and the player can hover the mouse over and select the object's different components.
	TotalPieces int `csv:"TotalPieces"`

	// Determines the object's class type by declaring a specific value. This is used by the various functions ("InitFn", "OperateFn", "PopulateFn") for knowing how to handle specific types of objects.
	SubClass int `csv:"SubClass"`

	// Controls the X and Y distance delta values between adjacent objects when they are being populated together.
	// This field is only used by the Populate Function ("PopulateFn") values 3 and 4, for the Add Barrels and Add Crates functions.
	Xspace int `csv:"Xspace"`
	Yspace int `csv:"Yspace"`

	// Controls the vertical offset of the name tooltip's position above the object when the object is being selected. This is measured in pixels.
	NameOffset int `csv:"NameOffset"`

	// Boolean Field. If equals 1, then if a monster operates the object, then the object will run its operate function.
	// If equals 0, then if a monster operates the object, then the object will not run its operate function.
	MonsterOK bool `csv:"MonsterOK"`

	// Controls what shrine function to use (See "Code" field in shrines.txt) when the object is told to do its Skill command.
	ShrineFunction int `csv:"ShrineFunction"`

	// Boolean Field. If equals 1, the game will restore the object in an inactive state when the area level repopulates after a player loads back into it.
	// If equals 0, then the game will not restore the object.
	Restore bool `csv:"Restore"`

	// Used as possible parameters for various functions for the object.
	Parm0 int `csv:"Parm0"`
	Parm1 int `csv:"Parm1"`
	Parm2 int `csv:"Parm2"`
	Parm3 int `csv:"Parm3"`
	Parm4 int `csv:"Parm4"`

	// Boolean Field. If equals 1, then the object will have a random chance to spawn with the locked attribute and have a display tooltip name with the "lockedchest" string key.
	// This only works when the object has the Init Function ("InitFn") value equal to 3. If equals 0, then ignore this.
	Lockable bool `csv:"Lockable"`

	// Controls if an object should call its Populate function ("PopulateFn") when it is chosen as an object that can spawn in a room.
	// Objects with a gore value greater than 2 will not be populated in rooms.
	Gore int `csv:"Gore"`

	// Boolean Field. If equals 1, then the object's animation rate will always match the "FrameDelta#" field (depending on the object's mode) which means the client and server will have synced animations.
	// If equals 0, then the animation rate will have random visual variation.
	Sync bool `csv:"Sync"`

	// Controls the amount of damage dealt by the object when it performs an Operate Function ("OperateFn") that deals damage such as triggering a pulse trap or an explosion.
	Damage int `csv:"Damage"`

	// Boolean Field. If equals 1, then add and remove an overlay on the object based on its current mode. If equals 0, then ignore this.
	// This field will only work with specific object Classes and will use specific Overlays for those objects.
	Overlay bool `csv:"Overlay"`

	// Boolean Field. If equals 1, then the game will handle the bounding box around the object for mouse selection.
	// The game will use the object's pixel size and "Left", "Top", "Width", "Height" field values to determine the collision size.
	// If equals 0, then ignore this.
	CollisionSubst bool `csv:"CollisionSubst"`

	// Controls the starting X position offset value for drawing the bounding collision box around the object for mouse selection.
	// This field depends on the "CollisionSubst" field being enabled.
	Left int `csv:"Left"`

	// Controls the starting Y position offset value for drawing the bounding collision box around the object for mouse selection.
	// This field depends on the "CollisionSubst" field being enabled.
	Top int `csv:"Top"`

	// Controls the ending X position offset value for drawing the bounding collision box around the object for mouse selection.
	// This field depends on the "CollisionSubst" field being enabled.
	Width int `csv:"Width"`

	// Controls the ending Y position offset value for drawing the bounding collision box around the object for mouse selection.
	// This field depends on the "CollisionSubst" field being enabled.
	Height int `csv:"Height"`

	// Defines a function that the game will use when the player clicks on the object.
	OperateFn ObjectOperateFunction `csv:"OperateFn"`

	// Defines a function that the game will use to spawn this object.
	PopulateFn ObjectPopulateFunction `csv:"PopulateFn"`

	// Defines a function to control how the object works while active and when initially activated by a player.
	InitFn ObjectInitFunction `csv:"InitFn"`

	// Defines a function that runs on the object from the game's client side.
	ClientFn ObjectClientFunction `csv:"ClientFn"`

	// Boolean Field. If equals 1, then when the object has been used, the game will not restore the object in an inactive state when the area level repopulates after a player loads back into it.
	// If equals 0, then ignore this.
	RestoreVirgins bool `csv:"RestoreVirgins"`

	// Boolean Field. If equals 1, then missiles can collide with this object. If equals 0, then missiles will ignore and fly through this object.
	BlockMissile bool `csv:"BlockMissile"`

	// Controls the targeting priority of the object.
	// Code	Description
	// 0	The object will not change its targeting priority.
	// 1	The object's target priority will equal a corpse only when the object is opened.
	// 2	The object's target priority always equals a corpse.
	ObjectTargetPriority `csv:"DrawUnder"`

	// Boolean Field. If equals 1, then this object will be classified as an
	// object that can be opened to warp to another area, and the UI will be
	// notified to display a tooltip for opening or entering, based on the
	// object’s mode. If equals 0, then ignore this.
	OpenWarp bool `csv:"OpenWarp"`

	// Used to display a tile in the Automap to represent the object. Defines
	// which cell number to use in the tile list for the Automap. If this value
	// equals 0, then this object will not display on the Automap.
	// (See Automap.txt)
	AutoMap string `csv:"AutoMap"`
}
