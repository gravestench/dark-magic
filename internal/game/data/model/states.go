package models

// State represents different passive behaviors applied to units that can apply various effects in the game.
type State struct {
	State           string `csv:"state"`             // Unique name ID for the state
	Group           string `csv:"group"`             // Group ID value to limit one active state with the same group on a unit
	RemHit          int    `csv:"remhit"`            // Boolean field, 1: State removed when the unit is hit, 0: Ignore
	NoSend          int    `csv:"nosend"`            // Boolean field, 1: State change not sent to the client, 0: Ignore
	Transform       int    `csv:"transform"`         // Boolean field, 1: Changes unit's appearance and resets animations, 0: Ignore
	Aura            int    `csv:"aura"`              // Boolean field, 1: Treated as an aura, 0: Ignore
	Curable         int    `csv:"curable"`           // Boolean field, 1: State can be cured, 0: Ignore
	Curse           int    `csv:"curse"`             // Boolean field, 1: State flagged as a curse, 0: Ignore
	Active          int    `csv:"active"`            // Boolean field, 1: State classified as an active state, 0: Ignore
	Restrict        int    `csv:"restrict"`          // Boolean field, 1: State restricts usage of certain skills, 0: Ignore
	Disguise        int    `csv:"disguise"`          // Boolean field, 1: State flagged as a disguise, 0: Ignore
	AttBlue         int    `csv:"attblue"`           // Boolean field, 1: Affects Attack Rating value in blue color, 0: Ignore
	DamBlue         int    `csv:"damblue"`           // Boolean field, 1: Affects Damage value in blue color, 0: Ignore
	ArmBlue         int    `csv:"armblue"`           // Boolean field, 1: Affects Defense value (Armor) in blue color, 0: Ignore
	RFBlue          int    `csv:"rfblue"`            // Boolean field, 1: Affects Fire Resistance value in blue color, 0: Ignore
	RLBlue          int    `csv:"rlblue"`            // Boolean field, 1: Affects Lightning Resistance value in blue color, 0: Ignore
	RCBlue          int    `csv:"rcblue"`            // Boolean field, 1: Affects Cold Resistance value in blue color, 0: Ignore
	StamBarBlue     int    `csv:"stambarblue"`       // Boolean field, 1: Affects Stamina Bar UI in blue color, 0: Ignore
	RPBlue          int    `csv:"rpblue"`            // Boolean field, 1: Affects Poison Resistance value in blue color, 0: Ignore
	AttRed          int    `csv:"attred"`            // Boolean field, 1: Affects Attack Rating value in red color, 0: Ignore
	DamRed          int    `csv:"damred"`            // Boolean field, 1: Affects Damage value in red color, 0: Ignore
	ArmRed          int    `csv:"armred"`            // Boolean field, 1: Affects Defense value (Armor) in red color, 0: Ignore
	RFRed           int    `csv:"rfred"`             // Boolean field, 1: Affects Fire Resistance value in red color, 0: Ignore
	RLRed           int    `csv:"rlred"`             // Boolean field, 1: Affects Lightning Resistance value in red color, 0: Ignore
	RCRed           int    `csv:"rcred"`             // Boolean field, 1: Affects Cold Resistance value in red color, 0: Ignore
	RPRed           int    `csv:"rpred"`             // Boolean field, 1: Affects Poison Resistance value in red color, 0: Ignore
	Exp             int    `csv:"exp"`               // Boolean field, 1: Unit with this state gives/gains exp, 0: Ignore
	PlrStayDeath    int    `csv:"plrstaydeath"`      // Boolean field, 1: State persists on player after death, 0: Ignore
	MonStayDeath    int    `csv:"monstaydeath"`      // Boolean field, 1: State persists on monster (non-boss) after death, 0: Ignore
	BossStayDeath   int    `csv:"bossstaydeath"`     // Boolean field, 1: State persists on boss after death, 0: Ignore
	Hide            int    `csv:"hide"`              // Boolean field, 1: Hides the unit when dead (corpse and death animations not drawn), 0: Ignore
	HideDead        int    `csv:"hidedead"`          // Boolean field, 1: Used to destroy units with invisible corpses, 0: Ignore
	Shatter         int    `csv:"shatter"`           // Boolean field, 1: Causes ice shatter missiles when the unit dies, 0: Ignore
	UDead           int    `csv:"udead"`             // Boolean field, 1: Flags the unit as a used dead corpse, 0: Ignore
	Life            int    `csv:"life"`              // Boolean field, 1: Cancels monster's normal life regeneration, 0: Ignore
	Green           int    `csv:"green"`             // Boolean field, 1: Overrides unit's color changes and makes it green, 0: Ignore
	PGSV            int    `csv:"pgsv"`              // Boolean field, 1: Part of a progressive skill (charge-up skill), 0: Ignore
	NoOverlays      int    `csv:"nooverlays"`        // Boolean field, 1: Disables standard way of adding overlays, 0: Ignore
	NoClear         int    `csv:"noclear"`           // Boolean field, 1: Applied state does not clear stats from the state's previous application, 0: Ignore
	BossInv         int    `csv:"bossinv"`           // Boolean field, 1: Uses the boss's inventory for the unit's equipped item graphics, 0: Ignore
	MeleeOnly       int    `csv:"meleeonly"`         // Boolean field, 1: Makes all unit's attacks melee attacks, 0: Ignore
	NotOnDead       int    `csv:"notondead"`         // Boolean field, 1: Does not play On function if unit is dead, 0: Ignore
	Overlay1        string `csv:"overlay1"`          // Controls which overlay to use for displaying the state (Front Start)
	Overlay2        string `csv:"overlay2"`          // Controls which overlay to use for displaying the state (Back Start)
	Overlay3        string `csv:"overlay3"`          // Controls which overlay to use for displaying the state (Front End)
	Overlay4        string `csv:"overlay4"`          // Controls which overlay to use for displaying the state (Back End)
	PGSVOverlay     string `csv:"pgsvoverlay"`       // Controls which overlay to use for progressive charges (charge-up skills)
	CastOverlay     string `csv:"castoverlay"`       // Controls which overlay to use when the state is initially applied
	RemOverlay      string `csv:"removerlay"`        // Controls which overlay to use when the state is removed
	Stat            string `csv:"stat"`              // Controls the associated stat for the state
	SetFunc         int    `csv:"setfunc"`           // Controls client-side set functions when the state is applied
	RemFunc         int    `csv:"remfunc"`           // Controls client-side remove functions when the state is removed
	Missile         string `csv:"missile"`           // Possible parameter for "setfunc" field, specifies missile from Missiles.txt
	Skill           string `csv:"skill"`             // Possible parameter for "setfunc" field, specifies skill from skills.txt
	ItemType        string `csv:"itemtype"`          // Potential Item Type affected by the state's color change (see ItemTypes.txt)
	ItemTrans       string `csv:"itemtrans"`         // Controls color change of the item when the unit has this state (uses color codes)
	ColorPri        int    `csv:"colorpri"`          // Priority of state's color change compared to other states on the unit
	ColorShift      int    `csv:"colorshift"`        // Controls index of the color shift palette to use
	LightR          int    `csv:"light-r"`           // State's change of the red color value of Light radius (0-255)
	LightG          int    `csv:"light-g"`           // State's change of the green color value of Light radius (0-255)
	LightB          int    `csv:"light-b"`           // State's change of the blue color value of Light radius (0-255)
	OnSound         string `csv:"onsound"`           // Sound played when the state is initially applied
	OffSound        string `csv:"offsound"`          // Sound played when the state is removed
	GfxType         int    `csv:"gfxtype"`           // Handles unit graphics transformation based on unit type
	GfxClass        int    `csv:"gfxclass"`          // Controls unit class used for handling graphics transformation
	CltEvent        string `csv:"cltevent"`          // Controls event to check on the client side for "clteventfunc" field
	CltEventFunc    int    `csv:"clteventfunc"`      // Client Unit event function called when event is determined in "cltevent" field
	CltActiveFunc   int    `csv:"cltactivefunc"`     // Client Do function called every frame while the state is active
	SrvActiveFunc   int    `csv:"srvactivefunc"`     // Server Do function called every frame while the state is active
	CanStack        int    `csv:"canstack"`          // Boolean field, 1: State can stack with duplicates (only for "poison" state), 0: Ignore
	SunderFull      int    `csv:"sunderfull"`        // Boolean field, 1: Reapply negative resistance stats at full potential when calculating pierce immunity if broken, 0: Ignore
	SunderResReduce int    `csv:"sunder-res-reduce"` // Boolean field, 1: Apply pierce resistance at reduced effectiveness when calculating pierce resistance if immunity was broken, 0: Ignore
}
