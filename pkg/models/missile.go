package models

// Missile represents a projectile used throughout the game for attacks, skills, and special effects.
type Missile struct {
	Missile           string `csv:"Missile"`       // Defines the unique name ID for the missile.
	PCltDoFunc        string `csv:"pCltDoFunc"`    // ID value to select a function for the missile's behavior on the client side while active every frame.
	PCltHitFunc       string `csv:"pCltHitFunc"`   // ID value to select a specialized function for the missile's behavior when hitting something on the client side.
	PSrvDoFunc        string `csv:"pSrvDoFunc"`    // ID value to select a specialized function for the missile's behavior on the server side every frame.
	PSrvHitFunc       string `csv:"pSrvHitFunc"`   // ID value to select a specialized function for the missile's behavior when hitting something on the server side.
	PSrvDmgFunc       string `csv:"pSrvDmgFunc"`   // ID value to select a specialized function called before damaging a unit on the server side.
	SrvCalc1          string `csv:"SrvCalc1"`      // Numeric calculation field used as a parameter for "pSrvDoFunc".
	Param1            int    `csv:"Param1"`        // Integer field used as a parameter for "pSrvDoFunc".
	Param2            int    `csv:"Param2"`        // Integer field used as a parameter for "pSrvDoFunc".
	Param3            int    `csv:"Param3"`        // Integer field used as a parameter for "pSrvDoFunc".
	Param4            int    `csv:"Param4"`        // Integer field used as a parameter for "pSrvDoFunc".
	Param5            int    `csv:"Param5"`        // Integer field used as a parameter for "pSrvDoFunc".
	CltCalc1          string `csv:"CltCalc1"`      // Numeric calculation field used as a parameter for "pCltDoFunc".
	CltParam1         int    `csv:"CltParam1"`     // Integer field used as a parameter for "pCltDoFunc".
	CltParam2         int    `csv:"CltParam2"`     // Integer field used as a parameter for "pCltDoFunc".
	CltParam3         int    `csv:"CltParam3"`     // Integer field used as a parameter for "pCltDoFunc".
	CltParam4         int    `csv:"CltParam4"`     // Integer field used as a parameter for "pCltDoFunc".
	CltParam5         int    `csv:"CltParam5"`     // Integer field used as a parameter for "pCltDoFunc".
	SHitCalc1         string `csv:"SHitCalc1"`     // Numeric calculation field used as a parameter for "pSrvHitFunc".
	SHitPar1          int    `csv:"sHitPar1"`      // Integer field used as a parameter for "pSrvHitFunc".
	SHitPar2          int    `csv:"sHitPar2"`      // Integer field used as a parameter for "pSrvHitFunc".
	SHitPar3          int    `csv:"sHitPar3"`      // Integer field used as a parameter for "pSrvHitFunc".
	CHitCalc1         string `csv:"CHitCalc1"`     // Numeric calculation field used as a parameter for "pCltHitFunc".
	CHitPar1          int    `csv:"cHitPar1"`      // Integer field used as a parameter for "pCltHitFunc".
	CHitPar2          int    `csv:"cHitPar2"`      // Integer field used as a parameter for "pCltHitFunc".
	CHitPar3          int    `csv:"cHitPar3"`      // Integer field used as a parameter for "pCltHitFunc".
	DmgCalc1          string `csv:"DmgCalc1"`      // Numeric calculation field used as a parameter for "pSrvDmgFunc".
	DParam1           int    `csv:"dParam1"`       // Integer field used as a parameter for "pSrvDmgFunc".
	DParam2           int    `csv:"dParam2"`       // Integer field used as a parameter for "pSrvDmgFunc".
	Vel               int    `csv:"Vel"`           // Baseline velocity of the missile in pixels traveled per frame.
	MaxVel            int    `csv:"MaxVel"`        // Maximum velocity of the missile.
	VelLev            int    `csv:"VelLev"`        // Extra velocity based on the caster unit's level.
	Accel             int    `csv:"Accel"`         // Acceleration of the missile's movement.
	Range             int    `csv:"Range"`         // Baseline duration that the missile will exist for after it is created, measured in frames where 25 Frames = 1 second.
	LevRange          int    `csv:"LevRange"`      // Extra duration based on the caster unit's level.
	Light             int    `csv:"Light"`         // Missile's Light Radius size measured in grid sub-tiles.
	Flicker           int    `csv:"Flicker"`       // If greater than 0, the Light Radius randomly changes in size every 4th frame while the missile is active.
	Red               int    `csv:"Red"`           // Red color value of the missile's Light Radius.
	Green             int    `csv:"Green"`         // Green color value of the missile's Light Radius.
	Blue              int    `csv:"Blue"`          // Blue color value of the missile's Light Radius.
	InitSteps         int    `csv:"InitSteps"`     // Number of frames the missile needs to be alive until it becomes visible on the game client.
	Activate          int    `csv:"Activate"`      // Number of frames the missile needs to be alive until it becomes active.
	LoopAnim          int    `csv:"LoopAnim"`      // If true, the missile's animation will repeat once the previous animation finishes.
	CelFile           string `csv:"CelFile"`       // DCC missile file to use for the visual graphics of the missile.
	AnimRate          int    `csv:"animrate"`      // Visual speed of the missile's animation graphics.
	AnimLen           int    `csv:"AnimLen"`       // Length of the missile's animation in frames where 25 Frames = 1 second.
	AnimSpeed         int    `csv:"AnimSpeed"`     // Visual speed of the missile's animation graphics on the client side, measured in 16ths.
	RandStart         int    `csv:"RandStart"`     // If greater than 0, the missile will start at a random frame between 0 and this value when it begins its animation.
	SubLoop           bool   `csv:"SubLoop"`       // If true, the missile will use a specific sequence of its animation while it is alive.
	SubStart          int    `csv:"SubStart"`      // The starting frame of the sequence animation. Requires "SubLoop" field to be enabled.
	SubStop           int    `csv:"SubStop"`       // The ending frame of the sequence animation. Requires "SubLoop" field to be enabled.
	CollideType       int    `csv:"CollideType"`   // Missile's collision type, controlling what units or objects the missile can impact.
	CollideKill       bool   `csv:"CollideKill"`   // If true, the missile will be destroyed when it collides with something.
	CollideFriend     bool   `csv:"CollideFriend"` // If true, the missile can collide with friendly units, including the caster.
	LastCollide       bool   `csv:"LastCollide"`   // If true, the missile will track the last unit that it collided with.
	Collision         bool   `csv:"Collision"`     // If true, the missile will have a missile type path placement collision mask when it is initialized or moved.
	ClientCol         bool   `csv:"ClientCol"`     // If true, the missile will check collision on the client based on the "CollideType" field.
	ClientSend        bool   `csv:"ClientSend"`    // If true, the server will create the missile on the client.
	NextHit           bool   `csv:"NextHit"`       // If true, the missile will use the next delay.
	NextDelay         int    `csv:"NextDelay"`     // Delay in frame length until the missile is allowed to hit the same unit again. Requires "NextHit" field to be enabled.
	XOffset           int    `csv:"xoffset"`       // X location coordinate (measured in pixels) to offset the visual graphics of the missile.
	YOffset           int    `csv:"yoffset"`       // Y location coordinate (measured in pixels) to offset the visual graphics of the missile.
	ZOffset           int    `csv:"zoffset"`       // Z location coordinate (measured in pixels) to visually draw the missile based on its actual location.
	Size              int    `csv:"Size"`          // Diameter in sub-tiles (X and Y axis) that the missile will occupy.
	SrcTown           bool   `csv:"SrcTown"`       // If true, the missile will be destroyed if the caster unit is located in an act town.
	CltSrcTown        int    `csv:"CltSrcTown"`    // If greater than 0 and "LoopAnim" field is disabled, this controls which frame to set the missile's animation when the player is in town.
	CanDestroy        int    `csv:"CanDestroy"`    // If true, the missile can be attacked and destroyed.
	ToHit             int    `csv:"ToHit"`         // If true, this missile will use the caster's Attack Rating stat to determine if the missile should hit its target.
	AlwaysExplode     int    `csv:"AlwaysExplode"` // If true, the missile will always process an explosion when killed.
	Explosion         int    `csv:"Explosion"`     // If true, the missile will be classified as an explosion.
	Town              int    `csv:"Town"`          // If true, the missile is allowed to be alive when in a town area.
	NoUniqueMod       int    `csv:"NoUniqueMod"`   // If true, the missile will not receive bonuses from Unique monster modifiers.
	NoMultiShot       int    `csv:"NoMultiShot"`   // If true, the missile will not be affected by the Multi-Shot monster modifier.
	Holy              int    `csv:"Holy"`          // Bit field flag where each value is a code to allow the missile to damage a certain type of monster.
	CanSlow           int    `csv:"CanSlow"`       // If true, the missile can be affected by the "slowmissiles" state.
	ReturnFire        int    `csv:"ReturnFire"`    // If true, missile can trigger the Sorceress Chilling Armor event function.
	GetHit            int    `csv:"GetHit"`        // If true, the missile will cause the target unit to enter the Get Hit mode (GH), which acts as the hit recovery mode.
	SoftHit           int    `csv:"SoftHit"`       // If true, the missile will cause a soft hit on the unit.
	KnockBack         int    `csv:"KnockBack"`     // Percentage chance (out of 100) that the target unit will be knocked back when hit by the missile.
	Trans             int    `csv:"Trans"`         // Alpha mode for how the missile is displayed, affecting transparency and blending.
	Pierce            bool   `csv:"Pierce"`        // If true, allow the Pierce modifier function to work with this missile.
	MissileSkill      bool   `csv:"MissileSkill"`  // If true, the missile will use the skill's damage instead of its defined damage.
	Skill             string `csv:"Skill"`         // Links to the "skill" field from the skills.txt file.
	ResultFlags       string `csv:"ResultFlags"`   // Controls different flags affecting how the target reacts after being hit by the missile.
	HitFlags          string `csv:"HitFlags"`      // Controls different flags affecting the damage dealt when the target is hit by the missile.
	HitShift          int    `csv:"HitShift"`      // Percentage modifier for the missile's damage.
	ApplyMastery      bool   `csv:"ApplyMastery"`  // If true, apply the caster's elemental mastery bonus modifiers to the missile's elemental damage.
	SrcDamage         int    `csv:"SrcDamage"`     // Controls how much of the source unit's damage should be added to the missile's damage.
	Half2HSrc         bool   `csv:"Half2HSrc"`     // If true and the source unit is currently wielding a 2-Handed weapon, the source damage is reduced by 50%.
	SrcMissDmg        int    `csv:"SrcMissDmg"`    // If the missile was created by another missile, this controls how much of the source missile's damage should be added to this missile's damage.
	MinDamage         int    `csv:"MinDamage"`     // Minimum baseline physical damage dealt by the missile.
	MinLevDam1        string `csv:"MinLevDam1"`    // Additional minimum physical damage dealt by the missile, calculated using the leveling formula.
	MinLevDam2        string `csv:"MinLevDam2"`
	MinLevDam3        string `csv:"MinLevDam3"`
	MinLevDam4        string `csv:"MinLevDam4"`
	MinLevDam5        string `csv:"MinLevDam5"`
	MaxDamage         int    `csv:"MaxDamage"`  // Maximum baseline physical damage dealt by the missile.
	MaxLevDam1        string `csv:"MaxLevDam1"` // Additional maximum physical damage dealt by the missile, calculated using the leveling formula.
	MaxLevDam2        string `csv:"MaxLevDam2"`
	MaxLevDam3        string `csv:"MaxLevDam3"`
	MaxLevDam4        string `csv:"MaxLevDam4"`
	MaxLevDam5        string `csv:"MaxLevDam5"`
	DmgSymPerCalc     string `csv:"DmgSymPerCalc"` // Calculation field determining the percentage increase to the physical damage dealt by the missile.
	EType             string `csv:"EType"`         // Type of elemental damage dealt by the missile. Possible values: "None", "fire", "ltng", "mag", "cold", "pois", "life", "mana", "stam", "stun", "rand", "burn", "frze".
	EMin              int    `csv:"EMin"`          // Minimum baseline elemental damage dealt by the missile.
	MinELev1          int    `csv:"MinELev1"`      // Additional minimum elemental damage dealt by the missile, calculated using the leveling formula.
	MinELev2          int    `csv:"MinELev2"`
	MinELev3          int    `csv:"MinELev3"`
	EMax              int    `csv:"EMax"`     // Maximum baseline elemental damage dealt by the missile.
	MaxELev1          int    `csv:"MaxELev1"` // Additional maximum elemental damage dealt by the missile, calculated using the leveling formula.
	MaxELev2          int    `csv:"MaxELev2"`
	MaxELev3          int    `csv:"MaxELev3"`
	ELen              int    `csv:"ELen"`     // Baseline elemental duration dealt by the missile in frames. Only applicable to appropriate elemental types with duration.
	ELevLen1          int    `csv:"ELevLen1"` // Additional elemental duration added by the missile, calculated using the leveling formula.
	ELevLen2          int    `csv:"ELevLen2"`
	ELevLen3          int    `csv:"ELevLen3"`
	HitClass          int    `csv:"HitClass"`         // Missile's own hit class into the damage routines, used for determining hit sound effects and overlays.
	NumDirections     int    `csv:"NumDirections"`    // Number of directions allowed by the missile, based on the DCC file used.
	LocalBlood        bool   `csv:"LocalBlood"`       // If true, change the color of blood missiles to green.
	DamageRate        int    `csv:"DamageRate"`       // Controls the "damage_framerate" stat, acting as a percentage multiplier for damage reduction stat modifiers during damage resistance calculations.
	TravelSound       string `csv:"TravelSound"`      // Points to a "Sound" field defined in the sounds.txt file, used when the missile is created and while it is alive.
	HitSound          string `csv:"HitSound"`         // Points to a "Sound" field defined in the sounds.txt file, used when the missile collides with a target.
	ProgSound         string `csv:"ProgSound"`        // Points to a "Sound" field defined in the sounds.txt file, used for a programmed special event based on the client function.
	ProgOverlay       string `csv:"ProgOverlay"`      // Points to the "overlay" field defined in the Overlay.txt file, used for a programmed special event based on the server or client function.
	ExplosionMissile  string `csv:"ExplosionMissile"` // Points to the "Missile" field for another missile, used for the missile created on the client when this missile explodes.
	SubMissile1       string `csv:"SubMissile1"`      // Points to the "Missile" field for another missile, used for creating a new missile based on the server function used.
	SubMissile2       string `csv:"SubMissile2"`
	SubMissile3       string `csv:"SubMissile3"`
	HitSubMissile1    string `csv:"HitSubMissile1"` // Points to the "Missile" field for another missile, used for a new missile after a collision, based on the server function used.
	HitSubMissile2    string `csv:"HitSubMissile2"`
	HitSubMissile3    string `csv:"HitSubMissile3"`
	HitSubMissile4    string `csv:"HitSubMissile4"`
	CltSubMissile1    string `csv:"CltSubMissile1"` // Points to the "Missile" field for another missile, used for creating a new missile based on the client function used.
	CltSubMissile2    string `csv:"CltSubMissile2"`
	CltSubMissile3    string `csv:"CltSubMissile3"`
	CltHitSubMissile1 string `csv:"CltHitSubMissile1"` // Points to the "Missile" field for another missile, used for a new missile after a collision, based on the client function used.
	CltHitSubMissile2 string `csv:"CltHitSubMissile2"`
	CltHitSubMissile3 string `csv:"CltHitSubMissile3"`
}
