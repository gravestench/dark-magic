package models

// SkillData represents the data fields in the skills.txt file.
type SkillData struct {
	ID                string `csv:"Id"`                // Unique identifier for the skill.
	CharClass         string `csv:"charclass"`         // Character class associated with the skill.
	SkillDesc         string `csv:"skilldesc"`         // Description of the skill.
	SrvStFunc         string `csv:"srvstfunc"`         // Server Start function used by the skill.
	SrvDoFunc         string `csv:"srvdofunc"`         // Server Do function used by the skill.
	PrgStack          string `csv:"prgstack"`          // Progressive stack.
	SrvPrgFunc1       string `csv:"srvprgfunc1"`       // Server Progressive function 1.
	SrvPrgFunc2       string `csv:"srvprgfunc2"`       // Server Progressive function 2.
	SrvPrgFunc3       string `csv:"srvprgfunc3"`       // Server Progressive function 3.
	PrgCalc1          string `csv:"prgcalc1"`          // Progressive calculation 1.
	PrgCalc2          string `csv:"prgcalc2"`          // Progressive calculation 2.
	PrgCalc3          string `csv:"prgcalc3"`          // Progressive calculation 3.
	PrgDam            string `csv:"prgdam"`            // Progressive damage.
	SrvMissile        string `csv:"srvmissile"`        // Server missile.
	DecQuant          string `csv:"decquant"`          // Decrement quantity.
	Lob               string `csv:"lob"`               // Lob skill.
	SrvMissileA       string `csv:"srvmissilea"`       // Server missile A.
	SrvMissileB       string `csv:"srvmissileb"`       // Server missile B.
	SrvMissileC       string `csv:"srvmissilec"`       // Server missile C.
	SrvOverlay        string `csv:"srvoverlay"`        // Server overlay.
	AuraFilter        string `csv:"aurafilter"`        // Aura filter.
	AuraState         string `csv:"aurastate"`         // Aura state.
	AuraTargetState   string `csv:"auratargetstate"`   // Aura target state.
	AuraLenCalc       string `csv:"auralencalc"`       // Aura length calculation.
	AuraRangeCalc     string `csv:"aurarangecalc"`     // Aura range calculation.
	AuraStat1         string `csv:"aurastat1"`         // Aura stat 1.
	AuraStatCalc1     string `csv:"aurastatcalc1"`     // Aura stat calculation 1.
	AuraStat2         string `csv:"aurastat2"`         // Aura stat 2.
	AuraStatCalc2     string `csv:"aurastatcalc2"`     // Aura stat calculation 2.
	AuraStat3         string `csv:"aurastat3"`         // Aura stat 3.
	AuraStatCalc3     string `csv:"aurastatcalc3"`     // Aura stat calculation 3.
	AuraStat4         string `csv:"aurastat4"`         // Aura stat 4.
	AuraStatCalc4     string `csv:"aurastatcalc4"`     // Aura stat calculation 4.
	AuraStat5         string `csv:"aurastat5"`         // Aura stat 5.
	AuraStatCalc5     string `csv:"aurastatcalc5"`     // Aura stat calculation 5.
	AuraStat6         string `csv:"aurastat6"`         // Aura stat 6.
	AuraStatCalc6     string `csv:"aurastatcalc6"`     // Aura stat calculation 6.
	AuraEvent1        string `csv:"auraevent1"`        // Aura event 1.
	AuraEventFunc1    string `csv:"auraeventfunc1"`    // Aura event function 1.
	AuraEvent2        string `csv:"auraevent2"`        // Aura event 2.
	AuraEventFunc2    string `csv:"auraeventfunc2"`    // Aura event function 2.
	AuraEvent3        string `csv:"auraevent3"`        // Aura event 3.
	AuraEventFunc3    string `csv:"auraeventfunc3"`    // Aura event function 3.
	AuraTgtEvent      string `csv:"auratgtevent"`      // Aura target event.
	AuraTgtEventFunc  string `csv:"auratgteventfunc"`  // Aura target event function.
	PassiveState      string `csv:"passivestate"`      // Passive state.
	PassiveIType      string `csv:"passiveitype"`      // Passive item type.
	PassiveStat1      string `csv:"passivestat1"`      // Passive stat 1.
	PassiveCalc1      string `csv:"passivecalc1"`      // Passive calculation 1.
	PassiveStat2      string `csv:"passivestat2"`      // Passive stat 2.
	PassiveCalc2      string `csv:"passivecalc2"`      // Passive calculation 2.
	PassiveStat3      string `csv:"passivestat3"`      // Passive stat 3.
	PassiveCalc3      string `csv:"passivecalc3"`      // Passive calculation 3.
	PassiveStat4      string `csv:"passivestat4"`      // Passive stat 4.
	PassiveCalc4      string `csv:"passivecalc4"`      // Passive calculation 4.
	PassiveStat5      string `csv:"passivestat5"`      // Passive stat 5.
	PassiveCalc5      string `csv:"passivecalc5"`      // Passive calculation 5.
	PassiveEvent      string `csv:"passiveevent"`      // Passive event.
	PassiveEventFunc  string `csv:"passiveeventfunc"`  // Passive event function.
	Summon            string `csv:"summon"`            // Summon.
	PetType           string `csv:"pettype"`           // Pet type.
	PetMax            string `csv:"petmax"`            // Maximum number of pets.
	SumMode           string `csv:"summode"`           // Summoning mode.
	SumSkill1         string `csv:"sumskill1"`         // Summon skill 1.
	SumSk1Calc        string `csv:"sumsk1calc"`        // Summon skill 1 calculation.
	SumSkill2         string `csv:"sumskill2"`         // Summon skill 2.
	SumSk2Calc        string `csv:"sumsk2calc"`        // Summon skill 2 calculation.
	SumSkill3         string `csv:"sumskill3"`         // Summon skill 3.
	SumSk3Calc        string `csv:"sumsk3calc"`        // Summon skill 3 calculation.
	SumSkill4         string `csv:"sumskill4"`         // Summon skill 4.
	SumSk4Calc        string `csv:"sumsk4calc"`        // Summon skill 4 calculation.
	SumSkill5         string `csv:"sumskill5"`         // Summon skill 5.
	SumSk5Calc        string `csv:"sumsk5calc"`        // Summon skill 5 calculation.
	SumUMod           string `csv:"sumumod"`           // Summon unique modifier.
	SumOverlay        string `csv:"sumoverlay"`        // Summon overlay.
	StSuccessOnly     string `csv:"stsuccessonly"`     // Single target success only.
	StSound           string `csv:"stsound"`           // Single target sound.
	StSoundClass      string `csv:"stsoundclass"`      // Single target sound class.
	StSoundDelay      string `csv:"stsounddelay"`      // Single target sound delay.
	WeaponSnd         string `csv:"weaponsnd"`         // Weapon sound.
	DoSound           string `csv:"dosound"`           // Do sound.
	DoSoundA          string `csv:"dosound a"`         // Do sound A.
	DoSoundB          string `csv:"dosound b"`         // Do sound B.
	TgtOverlay        string `csv:"tgtoverlay"`        // Target overlay.
	TgtSound          string `csv:"tgtsound"`          // Target sound.
	PrgOverlay        string `csv:"prgoverlay"`        // Progressive overlay.
	PrgSound          string `csv:"prgsound"`          // Progressive sound.
	CastOverlay       string `csv:"castoverlay"`       // Cast overlay.
	CltOverlayA       string `csv:"cltoverlaya"`       // Client overlay A.
	CltOverlayB       string `csv:"cltoverlayb"`       // Client overlay B.
	CltStFunc         string `csv:"cltstfunc"`         // Client Start function.
	CltDoFunc         string `csv:"cltdofunc"`         // Client Do function.
	CltPrgFunc1       string `csv:"cltprgfunc1"`       // Client Progressive function 1.
	CltPrgFunc2       string `csv:"cltprgfunc2"`       // Client Progressive function 2.
	CltPrgFunc3       string `csv:"cltprgfunc3"`       // Client Progressive function 3.
	CltMissile        string `csv:"cltmissile"`        // Client missile.
	CltMissileA       string `csv:"cltmissilea"`       // Client missile A.
	CltMissileB       string `csv:"cltmissileb"`       // Client missile B.
	CltMissileC       string `csv:"cltmissilec"`       // Client missile C.
	CltMissileD       string `csv:"cltmissiled"`       // Client missile D.
	CltCalc1          string `csv:"cltcalc1"`          // Client calculation 1.
	CltCalc2          string `csv:"cltcalc2"`          // Client calculation 2.
	CltCalc3          string `csv:"cltcalc3"`          // Client calculation 3.
	Warp              string `csv:"warp"`              // Warp skill.
	Immediate         string `csv:"immediate"`         // Immediate skill.
	Enhanceable       string `csv:"enhanceable"`       // Enhanceable skill.
	AttackRank        string `csv:"attackrank"`        // Attack rank.
	NoAmmo            string `csv:"noammo"`            // No ammo skill.
	Range             string `csv:"range"`             // Skill range.
	WeapSel           string `csv:"weapsel"`           // Weapon selection.
	ITypeA1           string `csv:"itypea1"`           // Item type A1.
	ITypeA2           string `csv:"itypea2"`           // Item type A2.
	ITypeA3           string `csv:"itypea3"`           // Item type A3.
	ETypeA1           string `csv:"etypea1"`           // Elemental type A1.
	ETypeA2           string `csv:"etypea2"`           // Elemental type A2.
	ITypeB1           string `csv:"itypeb1"`           // Item type B1.
	ITypeB2           string `csv:"itypeb2"`           // Item type B2.
	ITypeB3           string `csv:"itypeb3"`           // Item type B3.
	ETypeB1           string `csv:"etypeb1"`           // Elemental type B1.
	ETypeB2           string `csv:"etypeb2"`           // Elemental type B2.
	Anim              string `csv:"anim"`              // Animation.
	SeqTrans          string `csv:"seqtrans"`          // Sequence transition.
	MonAnim           string `csv:"monanim"`           // Monster animation.
	SeqNum            string `csv:"seqnum"`            // Sequence number.
	SeqInput          string `csv:"seqinput"`          // Sequence input.
	Durability        string `csv:"durability"`        // Durability.
	UseAttackRate     string `csv:"UseAttackRate"`     // Use attack rate.
	LineOfSight       string `csv:"LineOfSight"`       // Line of sight.
	TargetableOnly    string `csv:"TargetableOnly"`    // Targetable only.
	SearchEnemyXY     string `csv:"SearchEnemyXY"`     // Search enemy XY.
	SearchEnemyNear   string `csv:"SearchEnemyNear"`   // Search enemy near.
	SearchOpenXY      string `csv:"SearchOpenXY"`      // Search open XY.
	SelectProc        string `csv:"SelectProc"`        // Select proc.
	TargetCorpse      string `csv:"TargetCorpse"`      // Target corpse.
	TargetPet         string `csv:"TargetPet"`         // Target pet.
	TargetAlly        string `csv:"TargetAlly"`        // Target ally.
	TargetItem        string `csv:"TargetItem"`        // Target item.
	AttackNoMana      string `csv:"AttackNoMana"`      // Attack no mana.
	TgtPlaceCheck     string `csv:"TgtPlaceCheck"`     // Target place check.
	ItemEffect        string `csv:"ItemEffect"`        // Item effect.
	ItemCltEffect     string `csv:"ItemCltEffect"`     // Item client effect.
	ItemTgtDo         string `csv:"ItemTgtDo"`         // Item target do.
	ItemTarget        string `csv:"ItemTarget"`        // Item target.
	ItemCheckStart    string `csv:"ItemCheckStart"`    // Item check start.
	ItemCltCheckStart string `csv:"ItemCltCheckStart"` // Item client check start.
	ItemCastSound     string `csv:"ItemCastSound"`     // Item cast sound.
	ItemCastOverlay   string `csv:"ItemCastOverlay"`   // Item cast overlay.
	SkPoints          string `csv:"skpoints"`          // Skill points.
	ReqLevel          string `csv:"reqlevel"`          // Required level.
	MaxLevel          string `csv:"maxlvl"`            // Maximum level.
	ReqStr            string `csv:"reqstr"`            // Required strength.
	ReqDex            string `csv:"reqdex"`            // Required dexterity.
	ReqInt            string `csv:"reqint"`            // Required intelligence.
	ReqVit            string `csv:"reqvit"`            // Required vitality.
	ReqSkill1         string `csv:"reqskill1"`         // Required skill 1.
	ReqSkill2         string `csv:"reqskill2"`         // Required skill 2.
	ReqSkill3         string `csv:"reqskill3"`         // Required skill 3.
	Restrict          string `csv:"restrict"`          // Skill restrictions.
	State1            string `csv:"State1"`            // Skill state 1.
	State2            string `csv:"State2"`            // Skill state 2.
	State3            string `csv:"State3"`            // Skill state 3.
	Delay             string `csv:"delay"`             // Skill delay.
	LeftSkill         string `csv:"leftskill"`         // Left skill.
	Repeat            string `csv:"repeat"`            // Repeat skill.
	CheckFunc         string `csv:"checkfunc"`         // Check function.
	NoCostInState     string `csv:"nocostinstate"`     // No cost in state.
	UseManaOnDo       string `csv:"usemanaondo"`       // Use mana on do.
	StartMana         string `csv:"startmana"`         // Starting mana.
	MinMana           string `csv:"minmana"`           // Minimum mana.
	ManaShift         string `csv:"manashift"`         // Mana shift.
	Mana              string `csv:"mana"`              // Base mana.
	Interrupt         string `csv:"interrupt"`         // Interruptible skill.
	InTown            string `csv:"InTown"`            // Usable in town.
	Aura              string `csv:"aura"`              // Aura skill.
	Periodic          string `csv:"periodic"`          // Periodic skill.
	PerDelay          string `csv:"perdelay"`          // Periodic delay.
	Finishing         string `csv:"finishing"`         // Finishing skill.
	Passive           string `csv:"passive"`           // Passive skill.
	Progressive       string `csv:"progressive"`       // Progressive skill.
	General           string `csv:"general"`           // General skill.
	Scroll            string `csv:"scroll"`            // Scroll skill.
	Calc1             string `csv:"calc1"`             // Calculation 1.
	Calc2             string `csv:"calc2"`             // Calculation 2.
	Calc3             string `csv:"calc3"`             // Calculation 3.
	Calc4             string `csv:"calc4"`             // Calculation 4.
	Param1            string `csv:"Param1"`            // Parameter 1.
	Param2            string `csv:"Param2"`            // Parameter 2.
	Param3            string `csv:"Param3"`            // Parameter 3.
	Param4            string `csv:"Param4"`            // Parameter 4.
	Param5            string `csv:"Param5"`            // Parameter 5.
	Param6            string `csv:"Param6"`            // Parameter 6.
	Param7            string `csv:"Param7"`            // Parameter 7.
	Param8            string `csv:"Param8"`            // Parameter 8.
	InGame            string `csv:"InGame"`            // Enabled in-game.
	ToHit             string `csv:"ToHit"`             // Base bonus Attack Rating at level 1.
	LevToHit          string `csv:"LevToHit"`          // Additional bonus Attack Rating per level.
	ToHitCalc         string `csv:"ToHitCalc"`         // Attack Rating calculation.
	ResultFlags       string `csv:"ResultFlags"`       // Controls different flags for the target's reaction after being hit.
	HitFlags          string `csv:"HitFlags"`          // Controls different flags for the damage dealt when the target is hit.
	HitClass          string `csv:"HitClass"`          // Defines the skill's damage routines when hitting.
	Kick              string `csv:"Kick"`              // Separate function for calculating physical damage when kicking.
	HitShift          string `csv:"HitShift"`          // Percentage modifier for the skill's damage.
	SrcDam            string `csv:"SrcDam"`            // Percentage modifier for weapon damage transferred to the skill's damage.
	MinDam            string `csv:"MinDam"`            // Minimum baseline physical damage dealt by the skill.
	MinLevDam1        string `csv:"MinLevDam1"`        // Additional minimum physical damage dealt at level thresholds.
	MinLevDam2        string `csv:"MinLevDam2"`        // Additional minimum physical damage dealt at level thresholds.
	MinLevDam3        string `csv:"MinLevDam3"`        // Additional minimum physical damage dealt at level thresholds.
	MinLevDam4        string `csv:"MinLevDam4"`        // Additional minimum physical damage dealt at level thresholds.
	MinLevDam5        string `csv:"MinLevDam5"`        // Additional minimum physical damage dealt at level thresholds.
	MaxDam            string `csv:"MaxDam"`            // Maximum baseline physical damage dealt by the skill.
	MaxLevDam1        string `csv:"MaxLevDam1"`        // Additional maximum physical damage dealt at level thresholds.
	MaxLevDam2        string `csv:"MaxLevDam2"`        // Additional maximum physical damage dealt at level thresholds.
	MaxLevDam3        string `csv:"MaxLevDam3"`        // Additional maximum physical damage dealt at level thresholds.
	MaxLevDam4        string `csv:"MaxLevDam4"`        // Additional maximum physical damage dealt at level thresholds.
	MaxLevDam5        string `csv:"MaxLevDam5"`        // Additional maximum physical damage dealt at level thresholds.
	DmgSymPerCalc     string `csv:"DmgSymPerCalc"`     // Damage symmetrical percentage calculation.
	EType             string `csv:"EType"`             // Elemental type.
	EMin              string `csv:"EMin"`              // Minimum baseline elemental damage dealt by the skill.
	EMinLev1          string `csv:"EMinLev1"`          // Additional minimum elemental damage dealt at level thresholds.
	EMinLev2          string `csv:"EMinLev2"`          // Additional minimum elemental damage dealt at level thresholds.
	EMinLev3          string `csv:"EMinLev3"`          // Additional minimum elemental damage dealt at level thresholds.
	EMinLev4          string `csv:"EMinLev4"`          // Additional minimum elemental damage dealt at level thresholds.
	EMinLev5          string `csv:"EMinLev5"`          // Additional minimum elemental damage dealt at level thresholds.
	EMax              string `csv:"EMax"`              // Maximum baseline elemental damage dealt by the skill.
	EMaxLev1          string `csv:"EMaxLev1"`          // Additional maximum elemental damage dealt at level thresholds.
	EMaxLev2          string `csv:"EMaxLev2"`          // Additional maximum elemental damage dealt at level thresholds.
	EMaxLev3          string `csv:"EMaxLev3"`          // Additional maximum elemental damage dealt at level thresholds.
	EMaxLev4          string `csv:"EMaxLev4"`          // Additional maximum elemental damage dealt at level thresholds.
	EMaxLev5          string `csv:"EMaxLev5"`          // Additional maximum elemental damage dealt at level thresholds.
	EDmgSymPerCalc    string `csv:"EDmgSymPerCalc"`    // Elemental damage symmetrical percentage calculation.
	ELen              string `csv:"ELen"`              // Elemental length.
	ELevLen1          string `csv:"ELevLen1"`          // Additional elemental length at level thresholds.
	ELevLen2          string `csv:"ELevLen2"`          // Additional elemental length at level thresholds.
	ELevLen3          string `csv:"ELevLen3"`          // Additional elemental length at level thresholds.
	ELenSymPerCalc    string `csv:"ELenSymPerCalc"`    // Elemental length symmetrical percentage calculation.
	AIType            string `csv:"aitype"`            // AI type.
	AIBonus           string `csv:"aibonus"`           // AI bonus.
	CostMult          string `csv:"cost mult"`         // Cost multiplier.
	CostAdd           string `csv:"cost add"`          // Cost addition.
}

// this might be the D2 resurrected format...
//// SkillsData represents data about skills.
//type SkillsData struct {
//	Skill                            string `csv:"skill"`                            // Defines the unique name ID for the skill, which is how other files can reference the skill.
//	CharClass                        string `csv:"charclass"`                        // Assigns the skill to a specific character class which affects how skill item modifiers work and what skills the class can learn.
//	SkillDesc                        string `csv:"skilldesc"`                        // Controls the skill’s tooltip and general UI display. Points to a “skilldesc” field in the skilldesc.txt file.
//	SrvStFunc                        string `csv:"srvstfunc"`                        // Server Start function. This controls how the skill works when it is starting to cast, on the server side. This uses a code value to call a function, affecting how certain fields are used. (e.g., BarStartDoubleSwing - Uses “calc2” to apply an attack rate bonus.)
//	SrvDoFunc                        string `csv:"srvdofunc"`                        // Server Do function. This controls how the skill works when it finishes being cast, on the server side. This uses a code value to call a function, affecting how certain fields are used. (e.g., MonDoDiabloLight - Shoot a missile from the caster to the target location, adhering to the caster’s inferno animation frames (See monstats2.txt). Use “calc1” to control the missile range, otherwise default to using the missile’s “Param2” value calculated with its current level. Use “calc3” to control the periodic frame count for how often to create the missile.)
//	SrvStopFunc                      string `csv:"srvstopfunc"`                      // Server Stop function. This controls how the skill cleans up after ending. This uses a code value to call a function, affecting how certain fields are used. (e.g., BarStopWhirlwind - Handles changing the collision, pathing, and aura state of the caster.)
//	PrgStack                         int    `csv:"prgstack"`                         // Boolean Field. If equals 1, then all “srvprgfunc#” functions will execute, based on the current number of progressive charges. If equals 0, then only the relative “srvprogfunc#” function will execute, based on the current number of progressive charges.
//	SrvPrgFunc1                      string `csv:"srvprgfunc1"`                      // Controls what Server Do function is used when executing the progressive skill with a charge number equal to 1. This field uses the same functions as the “srvdofunc” field.
//	SrvPrgFunc2                      string `csv:"srvprgfunc2"`                      // Controls what Server Do function is used when executing the progressive skill with a charge number equal to 2. This field uses the same functions as the “srvdofunc” field.
//	SrvPrgFunc3                      string `csv:"srvprgfunc3"`                      // Controls what Server Do function is used when executing the progressive skill with a charge number equal to 3. This field uses the same functions as the “srvdofunc” field.
//	PrgCalc1                         string `csv:"prgcalc1"`                         // Calculation Field. Used as a possible parameter for calculating values when executing the progressive skill with a charge number equal to 1.
//	PrgCalc2                         string `csv:"prgcalc2"`                         // Calculation Field. Used as a possible parameter for calculating values when executing the progressive skill with a charge number equal to 2.
//	PrgCalc3                         string `csv:"prgcalc3"`                         // Calculation Field. Used as a possible parameter for calculating values when executing the progressive skill with a charge number equal to 3.
//	PrgDam                           string `csv:"prgdam"`                           // Calls a function to modify the progressive damage dealt (e.g., ModifyProgressiveElementalConvert - Convert a percentage of the physical damage dealt to elemental damage, based on the “calc1” field. If the elemental type equals Cold, then at 3 charges, modify the Freeze Length based on the Cold Length and a divisor (using Param5)).
//	SrvMissile                       string `csv:"srvmissile"`                       // Used as a parameter for controlling what main missile is used for the server functions used (See “Missile” field in Missiles.txt).
//	DecQuant                         int    `csv:"decquant"`                         // Boolean Field. If equals 1, then the unit’s ammo quantity will be updated each time the skill’s Server Do function is called. If equals 0, then ignore this.
//	Lob                              int    `csv:"lob"`                              // Boolean Field. If equals 1, then missiles created by the skill’s functions will use the missile lobbing function, which will cause the missile to fly in an arc pattern. If equals 0, then missiles that are created will use the normal linear function.
//	SrvMissileA                      string `csv:"srvmissilea"`                      // Used as a parameter for controlling what secondary missile is used for the server functions used (See “Missile” field in Missiles.txt).
//	UseServerMissilesOnRemoteClients int    `csv:"useServerMissilesOnRemoteClients"` // Control new missile changes per player skill. Values of 1 will force enable it for that skill. Skills that have matching server/client missiles sets for the skill get auto-enabled. Setting this to a value greater than 1 will force it to skip this auto-enable logic. If equals 0, then ignore this. Note: this feature is disabled.
//	SrvOverlay                       string `csv:"srvoverlay"`                       // Creates an overlay on the target unit when the skill is used. This is a possible parameter used by various skill functions (See the “overlay” field in Overlay.txt).
//	AuraFilter                       int    `csv:"aurafilter"`                       // Controls different flags that affect how the skill’s aura will affect the different types of units. Uses an integer value to check against different bit fields. For example, if the value equals 4354 (binary = 1000100000010) then that returns true for the 4096 (binary = 1000000000000), 256 (binary = 0000100000000), and 2 (binary = 0000000000010) bit field values.
//	AuraState                        string `csv:"aurastate"`                        // Links to a state that can be applied to the caster unit when casting the skill, depending on the skill function used (See the “state” field in states.txt).
//	AuraTargetState                  string `csv:"auratargetstate"`                  // Links to a state that can be applied to the target unit when using the skill, depending on the skill function used (See the “state” field in states.txt).
//	AuraLenCalc                      string `csv:"auralencalc"`                      // Calculation Field. Controls the aura state duration on the unit (where 25 Frames = 1 second). If this value is empty, then the state duration will be controlled by other functions, or it will last forever. This can also be used as a parameter for certain skill functions.
//	AuraRangeCalc                    string `csv:"aurarangecalc"`                    // Calculation Field. Controls the aura state’s area radius size, measured in grid sub-tiles. This can also be used as a parameter for certain skill functions.
//	AuraStat1                        string `csv:"aurastat1"`                        // Controls which stat modifiers will be altered or added by the aura state (See the “Stat” field from ItemStatCost.txt).
//	AuraStatCalc1                    string `csv:"aurastatcalc1"`                    // Calculation Field. Controls the value for the relative “aurastat#” field.
//	AuraStat2                        string `csv:"aurastat2"`                        // Controls which stat modifiers will be altered or added by the aura state (See the “Stat” field from ItemStatCost.txt).
//	AuraStatCalc2                    string `csv:"aurastatcalc2"`                    // Calculation Field. Controls the value for the relative “aurastat#” field.
//	AuraStat3                        string `csv:"aurastat3"`                        // Controls which stat modifiers will be altered or added by the aura state (See the “Stat” field from ItemStatCost.txt).
//	AuraStatCalc3                    string `csv:"aurastatcalc3"`                    // Calculation Field. Controls the value for the relative “aurastat#” field.
//	AuraStat4                        string `csv:"aurastat4"`                        // Controls which stat modifiers will be altered or added by the aura state (See the “Stat” field from ItemStatCost.txt).
//	AuraStatCalc4                    string `csv:"aurastatcalc4"`                    // Calculation Field. Controls the value for the relative “aurastat#” field.
//	AuraStat5                        string `csv:"aurastat5"`                        // Controls which stat modifiers will be altered or added by the aura state (See the “Stat” field from ItemStatCost.txt).
//	AuraStatCalc5                    string `csv:"aurastatcalc5"`                    // Calculation Field. Controls the value for the relative “aurastat#” field.
//	AuraStat6                        string `csv:"aurastat6"`                        // Controls which stat modifiers will be altered or added by the aura state (See the “Stat” field from ItemStatCost.txt).
//	AuraStatCalc6                    string `csv:"aurastatcalc6"`                    // Calculation Field. Controls the value for the relative “aurastat#” field.
//	AuraEvent1                       string `csv:"auraevent1"`                       // Controls what event will trigger the relative “auraeventfunc#” field function. Links to an event listed in the events.txt file.
//	AuraEvent2                       string `csv:"auraevent2"`                       // Controls what event will trigger the relative “auraeventfunc#” field function. Links to an event listed in the events.txt file.
//	AuraEvent3                       string `csv:"auraevent3"`                       // Controls what event will trigger the relative “auraeventfunc#” field function. Links to an event listed in the events.txt file.
//	AuraEventFunc1                   string `csv:"auraeventfunc1"`                   // Controls the function used when the relative “auraevent#” event is triggered.
//	AuraEventFunc2                   string `csv:"auraeventfunc2"`                   // Controls the function used when the relative “auraevent#” event is triggered.
//	AuraEventFunc3                   string `csv:"auraeventfunc3"`                   // Controls the function used when the relative “auraevent#” event is triggered.
//	SkillActivateSubskill            string `csv:"SkillActivateSubskill"`            // Based on a random chance controlled by the “aurastatcalc1” value, cast a skill (determined by “sumskill1”) with a skill level controlled by “sumsk1calc”. If “aurastatcalc2” value is greater than 0, then factor the “passiveitype” and “passivereqweaponcount” fields for determining if the skill should be cast or not.
//	PassiveState                     string `csv:"passivestate"`                     // Links to a state that can be applied by the passive skill, depending on the skill function used (See the “state” field in states.txt).
//	PassiveIType                     string `csv:"passiveitype"`                     // Links to an Item Type to define what type of item needs to be equipped to enable the passive state (See the “ItemType” field in ItemTypes.txt).
//	PassiveReqWeaponCount            int    `csv:"passivereqweaponcount"`            // Controls how many equipped weapons are needed for this passive state to be enabled. If the value equals 1, then the player must have 1 weapon equipped for this passive state to be enabled. If the value equals 2, then the player must be dual-wielding weapons for this passive state to be enabled. If the value equals 0, then ignore this field.
//	PassiveStat1                     string `csv:"passivestat1"`                     // Assigns stat modifiers to the passive skill (See the “Stat” field from ItemStatCost.txt).
//	PassiveStatCalc1                 string `csv:"passivestatcalc1"`                 // Calculation Field. Controls the value for the relative “passivestat#” field.
//	PassiveStat2                     string `csv:"passivestat2"`                     // Assigns stat modifiers to the passive skill (See the “Stat” field from ItemStatCost.txt).
//	PassiveStatCalc2                 string `csv:"passivestatcalc2"`                 // Calculation Field. Controls the value for the relative “passivestat#” field.
//	PassiveStat3                     string `csv:"passivestat3"`                     // Assigns stat modifiers to the passive skill (See the “Stat” field from ItemStatCost.txt).
//	PassiveStatCalc3                 string `csv:"passivestatcalc3"`                 // Calculation Field. Controls the value for the relative “passivestat#” field.
//	PassiveStat4                     string `csv:"passivestat4"`                     // Assigns stat modifiers to the passive skill (See the “Stat” field from ItemStatCost.txt).
//	PassiveStatCalc4                 string `csv:"passivestatcalc4"`                 // Calculation Field. Controls the value for the relative “passivestat#” field.
//	PassiveStat5                     string `csv:"passivestat5"`                     // Assigns stat modifiers to the passive skill (See the “Stat” field from ItemStatCost.txt).
//	PassiveStatCalc5                 string `csv:"passivestatcalc5"`                 // Calculation Field. Controls the value for the relative “passivestat#” field.
//	PassiveStat6                     string `csv:"passivestat6"`                     // Assigns stat modifiers to the passive skill (See the “Stat” field from ItemStatCost.txt).
//	PassiveStatCalc6                 string `csv:"passivestatcalc6"`                 // Calculation Field. Controls the value for the relative “passivestat#” field.
//	Summon                           string `csv:"summon"`                           // Controls what monster is summoned by the skill (See the “Id” field in monstats.txt). This field’s usage will depend on the skill function used. This field can also be used as a reference for AI behaviors and for the skilldesc.txt file.
//	PetType                          string `csv:"pettype"`                          // Links to a pet type data to control how the summoned unit is displayed on the UI (See “pet type” field in pettype.txt).
//	PetMax                           string `csv:"petmax"`                           // Calculation Field. Used skill functions that summon pets to control how many summon units are allowed at a time.
//	SumMode                          string `csv:"summode"`                          // Defines the animation mode that the summoned monster will be initiated with.
//	SumSkill1                        string `csv:"sumskill1"`                        // Assigns a skill to the summoned monster. Points to another “skill” ID. This can be useful for adding a skill to a monster to transition its synergy bonuses.
//	SumSkill2                        string `csv:"sumskill2"`                        // Assigns a skill to the summoned monster. Points to another “skill” ID. This can be useful for adding a skill to a monster to transition its synergy bonuses.
//	SumSkill3                        string `csv:"sumskill3"`                        // Assigns a skill to the summoned monster. Points to another “skill” ID. This can be useful for adding a skill to a monster to transition its synergy bonuses.
//	SumSkill4                        string `csv:"sumskill4"`                        // Assigns a skill to the summoned monster. Points to another “skill” ID. This can be useful for adding a skill to a monster to transition its synergy bonuses.
//	SumSkill5                        string `csv:"sumskill5"`                        // Assigns a skill to the summoned monster. Points to another “skill” ID. This can be useful for adding a skill to a monster to transition its synergy bonuses.
//	SumSk1Calc                       string `csv:"sumsk1calc"`                       // Calculation Field. Controls the skill level for the designated “sumskill#” field when the skill is assigned to the monster.
//	SumSk2Calc                       string `csv:"sumsk2calc"`                       // Calculation Field. Controls the skill level for the designated “sumskill#” field when the skill is assigned to the monster.
//	SumSk3Calc                       string `csv:"sumsk3calc"`                       // Calculation Field. Controls the skill level for the designated “sumskill#” field when the skill is assigned to the monster.
//	SumSk4Calc                       string `csv:"sumsk4calc"`                       // Calculation Field. Controls the skill level for the designated “sumskill#” field when the skill is assigned to the monster.
//	SumSk5Calc                       string `csv:"sumsk5calc"`                       // Calculation Field. Controls the skill level for the designated “sumskill#” field when the skill is assigned to the monster.
//	SumUMod                          string `csv:"sumumod"`                          // Assigns a monster modifier to the summoned monster (See the “id” in monumod.txt).
//	SumOverlay                       string `csv:"sumoverlay"`                       // Creates an overlay on the summoned monster when it is first created (see the “overlay” field in Overlay.txt).
//	StSuccessOnly                    int    `csv:"stsuccessonly"`                    // Boolean Field. If equals 1, then the following sound and overlay fields will only play when the skill is successfully cast, instead of always being used even when the skill cast is interrupted. If equals 0, then the following sound and overlay fields will always be used when the skill is cast, regardless if the skill was interrupted or not.
//	StSound                          string `csv:"stsound"`                          // Controls what client sound is played when the skill is used, based on the client starting function (see the “Sound” field in sounds.txt).
//	StSoundClass                     string `csv:"stsoundclass"`                     // Controls what client sound is played when the skill is used by the skill’s assigned class (See “charclass” field), based on the client starting function (see the “Sound” field in sounds.txt). If the unit using the skill is not the same class as the “charclass” value for the skill, then this sound will not play.
//	StSoundDelay                     int    `csv:"stsounddelay"`                     // Boolean Field. If equals 1, then use the weapon’s hit class to determine the delay in frames (where 25 frames = 1 second) before playing the skill’s start sound. If equals 0, then the skill’s start sound will play immediately.
//	WeaponSnd                        int    `csv:"weaponsnd"`                        // Boolean Field. If equals 1, then play the weapon’s hit sound when hitting an enemy with this skill. The sound chosen is based on the weapon’s hit class. Also use the sound delay based on the weapon’s hit class to determine the delay in frames (where 25 frames = 1 second) before playing the weapon hit sound (See “stsounddelay” for the types of hit class sounds and delays used). If equals 0, then do not play the weapon hit sound when hitting an enemy with the skill attack.
//	DoSound                          string `csv:"dosound"`                          // Controls the sound for the skill each time the Client Do function is used (see the “Sound” field from sounds.txt).
//	DoSoundA                         string `csv:"dosound a"`                        // Used as a possible parameter for playing additional sounds based on the Client Do function used (see the “Sound” field in sounds.txt).
//	DoSoundB                         string `csv:"dosound b"`                        // Used as a possible parameter for playing additional sounds based on the Client Do function used (see the “Sound” field in sounds.txt).
//	TgtOverlay                       string `csv:"tgtoverlay"`                       // Used as a possible parameter for adding an Overlay on the target unit, based on the Client Do function used (see the “overlay” field in Overlay.txt).
//	TgtSound                         string `csv:"tgtsound"`                         // Used as a possible parameter for playing a sound located on the target unit, based on the Client Do function used (see the “Sound” field in sounds.txt).
//	PrgOverlay                       string `csv:"prgoverlay"`                       // Used as a possible parameter for adding an Overlay on the caster unit for progressive charge-up skill functions, based on the Client Do function used and how many progressive charges the caster unit has (see the “overlay” field in Overlay.txt).
//	PrgSound                         string `csv:"prgsound"`                         // Used as a possible parameter for playing a sound when using the skill for progressive charge-up skill functions, based on the Client Do function used and how many progressive charges the caster unit has (see the “Sound” field in sounds.txt).
//	CastOverlay                      string `csv:"castoverlay"`                      // Used as a possible parameter for adding an Overlay on the caster unit when using the skill, based on the Client Start/Do function used (see the “overlay” field in Overlay.txt).
//	CltOverlayA                      string `csv:"cltoverlaya"`                      // Used as a possible parameter for adding additional Overlays on the caster unit, based on the Client Start/Do function used (see the “overlay” field in Overlay.txt).
//	CltOverlayB                      string `csv:"cltoverlayb"`                      // Used as a possible parameter for adding additional Overlays on the caster unit, based on the Client Start/Do function used (see the “overlay” field in Overlay.txt).
//	CltStFunc                        string `csv:"cltstfunc"`                        // Client Start function. This controls how the skill works when it is starting to cast, on the client side. This uses a code value to call a function, affecting how certain fields are used.
//	CltStFuncDescription             string `csv:"cltstfuncdesc"`                    // Description of the Client Start function.
//	CltDoFunc                        string `csv:"cltdofunc"`                        // Client Do function. This controls how the skill works when it finishes being cast, on the client side. This uses a code value to call a function, affecting how certain fields are used.
//	CltDoFuncDescription             string `csv:"cltdofuncdesc"`                    // Description of the Client Do function.
//	CltStopFunc                      string `csv:"cltstopfunc"`                      // Client Stop function. This controls how the skill cleans up after ending. This uses a code value to call a function, affecting how certain fields are used.
//	CltStopFuncDescription           string `csv:"cltstopfuncdesc"`                  // Description of the Client Stop function.
//	CltPrgFunc1                      string `csv:"cltprgfunc1"`                      // Controls which Client Do function is used when the skill is executed while having a progressive charge number equal to 1 (uses “cltdofunc” values).
//	CltPrgFunc2                      string `csv:"cltprgfunc2"`                      // Controls which Client Do function is used when the skill is executed while having a progressive charge number equal to 2 (uses “cltdofunc” values).
//	CltPrgFunc3                      string `csv:"cltprgfunc3"`                      // Controls which Client Do function is used when the skill is executed while having a progressive charge number equal to 3 (uses “cltdofunc” values).
//	CltMissile                       string `csv:"cltmissile"`                       // Used as a parameter for controlling what main missile is used for the client functions used (See “Missile” field in Missiles.txt).
//	CltMissileA                      string `csv:"cltmissilea"`                      // Used as a parameter for controlling what secondary missile is used for the client functions used (See “Missile” field in Missiles.txt).
//	CltCalc1                         string `csv:"cltcalc1"`                         // Calculation Field. Use as a possible parameter for controlling values for the client functions.
//	Warp                             int    `csv:"warp"`                             // Boolean Field. If equals 1 and the skill uses a function that involves warping/teleporting, then check for a scene transition loading screen, if necessary. If equals 0, then ignore this.
//	Immediate                        int    `csv:"immediate"`                        // Boolean Field. If equals 1 and the skill has a periodic function, then immediately perform the skill’s function when the periodic skill is activated. If equals 0, then wait until the skill’s first periodic delay before performing the skill’s function.
//	Enhanceable                      int    `csv:"enhanceable"`                      // Boolean Field. If equals 1, then the skill will be included in the plus to all skills item modifier. If equals 0, then the skill will not be included in the plus to all skills item modifier.
//	AttackRank                       int    `csv:"attackrank"`                       // Controls the skill’s AI score value for determining what is the best target for the AI. The higher the value, then the more likely the AI will select a target with this skill equipped.
//	NoAmmo                           int    `csv:"noammo"`                           // Boolean Field. If equals 1, then the skill will not check that weapon’s ammo when performing an attack. This relies on the “Shoots” field from the ItemType.txt file. If equals 0, then the weapon will consume its ammo when being used by the skill.
//	Range                            string `csv:"range"`                            // Determines how the unit uses the skill, which affects the weapon requirements and the type of attack used.
//	WeaponSel                        string `csv:"weapsel"`                          // If the unit can dual wield weapons, then this field will control how the weapons are used for the skill.
//	ITypeA1                          string `csv:"itypea1"`                          // Controls what Item Types are included, or allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt).
//	ITypeA2                          string `csv:"itypea2"`                          // Controls what Item Types are included, or allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt).
//	ITypeA3                          string `csv:"itypea3"`                          // Controls what Item Types are included, or allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt).
//	ETypeA1                          string `csv:"etypea1"`                          // Controls what Item Types are excluded, or not allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt).
//	ETypeA2                          string `csv:"etypea2"`                          // Controls what Item Types are excluded, or not allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt).
//	ITypeB1                          string `csv:"itypeb1"`                          // Controls what alternate Item Types are included, or allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt). This acts as a second set of Item Types to check.
//	ITypeB2                          string `csv:"itypeb2"`                          // Controls what alternate Item Types are included, or allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt). This acts as a second set of Item Types to check.
//	ITypeB3                          string `csv:"itypeb3"`                          // Controls what alternate Item Types are included, or allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt). This acts as a second set of Item Types to check.
//	ETypeB1                          string `csv:"etypeb1"`                          // Controls what alternate Item Types are excluded, or not allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt). This acts as a second set of Item Types to check.
//	ETypeB2                          string `csv:"etypeb2"`                          // Controls what alternate Item Types are excluded, or not allowed, when determining if this skill can be used (See the “Code” field from ItemTypes.txt). This acts as a second set of Item Types to check.
//	Anim                             string `csv:"anim"`                             // Controls the animation mode that the player character will use when using this skill. Setting the mode to Sequence (SQ) will cause the player character to play a time controlled animation sequence, utilizing certain sequence fields.
//	SeqTrans                         string `csv:"seqtrans"`                         // Uses the same inputs as the “anim” field. If the “anim” field equals SQ (Sequence) and this field equals SC (Cast), then the sequence animation speed can be modified by the faster cast rate stat modifier.
//	MonAnim                          string `csv:"monanim"`                          // Controls the animation mode that the monster will use when using this skill. This is similar to the “anim” field except with monster units using the skill instead of player units.
//	SeqNum                           int    `csv:"seqnum"`                           // If the unit is a player and the “anim” used for the skill is a Sequence, then this field will control the index of which sequence animation to use. These sequences are specifically designed for certain skills, and each sequence has a set number of frames using certain animations.
//	SeqInput                         int    `csv:"seqinput"`                         // For skills that can repeat, this controls the number of frames to wait before the “Do” frame in the sequence. This acts as a delay in frames (25 Frames = 1 second) to wait within the sequence animation before it is allowed to be cast again or for looping back to the start of the sequence, such as for the Sorceress Inferno skill.
//	Durability                       int    `csv:"durability"`                       // Boolean Field. If equals 1 and when the player character ends an animation mode that was not Attack 1, Attack 2, or Throw, then check the quantity and durability of the player’s items to see if the valid weapons are out of ammo or are broken. If equals 0, then ignore this.
//	UseAttackRate                    int    `csv:"UseAttackRate"`                    // Boolean Field. If equals 1, then the current attack speed should be updated after using the skill, relative to the “attackrate” stat (See ItemStatCost.txt), and if the skill was an attacking skill. If equals 0, then the attack speed will not be updated after using the skill.
//	LineOfSight                      string `csv:"LineOfSight"`                      // Controls the skill’s collision filters for valid target locations when it is being cast.
//	TargetableOnly                   int    `csv:"TargetableOnly"`                   // Boolean Field. If equals 1, then the skill will require a target unit in order to be used. If equals 0, then ignore this.
//	SearchEnemyXY                    int    `csv:"SearchEnemyXY"`                    // Boolean Field. If equals 1, then when the skill is cast on a target location, it will automatically search in different directions in the target area to find the first available enemy target. If equals 0, then ignore this. This field can be overridden if the “SearchEnemyNear” field is enabled.
//	SearchEnemyNear                  int    `csv:"SearchEnemyNear"`                  // Boolean Field. If equals 1, then when the skill is cast on a target location, it will automatically find the nearest enemy target. If equals 0, then ignore this.
//	SearchOpenXY                     int    `csv:"SearchOpenXY"`                     // Boolean Field. If equals 1, then automatically search in a radius of size 7 in around the clicked target location to find an available unoccupied location to use the skill, testing for collision. If equals 0, then ignore this. This field can be overridden if the “SearchEnemyNear” or “SearchEnemyXY” field is enabled.
//	SelectProc                       string `csv:"SelectProc"`                       // Uses a function to check that the target is a corpse with various parameters before validating the usage of the skill.
//	TargetCorpse                     int    `csv:"TargetCorpse"`                     // Boolean Field. If equals 1, then the skill is allowed to target corpses. If equals 0, then skill cannot target corpses.
//	TargetPet                        int    `csv:"TargetPet"`                        // Boolean Field. If equals 1, then the skill is allowed to target pets (summons and mercenaries). If equals 0, then the skill cannot target pets.
//	TargetAlly                       int    `csv:"TargetAlly"`                       // Boolean Field. If equals 1, then the skill is allowed to target ally units. If equals 0, then the skill cannot target ally units.
//	TargetItem                       int    `csv:"TargetItem"`                       // Boolean Field. If equals 1, then the skill is allowed to target items on the ground. If equals 0, then the skill cannot target items.
//	AttackNoMana                     int    `csv:"AttackNoMana"`                     // Boolean Field. If equals 1, then then when the caster does not have enough mana to cast the skill and attempts to use the skill, the caster will default to using the Attack skill. If equals 0, then attempting to use the skill without enough mana will do nothing.
//	TgtPlaceCheck                    int    `csv:"TgtPlaceCheck"`                    // Boolean Field. If equals 1, then check that the target location is available for spawning a unit, testing collision. If equals 0, then ignore this. This is only used for skills that monsters use to spawn new units.
//	KeepCursorStateOnKill            int    `csv:"KeepCursorStateOnKill"`            // Boolean Field. If equals 1, then the mouse click hold state will continue to stay active after killing a selected target. If equals 0, then after killing the target, the mouse click hold state will reset.
//	ContinueCastUnselected           int    `csv:"ContinueCastUnselected"`           // Boolean Field. If equals 1, then while the mouse is held down and there is no more target selected, then the skill will continue being used at the mouse’s location. If equals 0, then while the mouse is held down and there is no more target selected, then the player character will cancel the skill and move to the mouse location.
//	ClearSelectedOnHold              int    `csv:"ClearSelectedOnHold"`              // Boolean Field. If equals 1, then when the mouse is held down, the target selection will be cleared. If equals 0, then ignore this.
//	ItemEffect                       string `csv:"ItemEffect"`                       // Uses a Server Do function (See “srvdofunc”) for handling how the skill is used when it is triggered by an item, on the server side.
//	ItemCltEffect                    string `csv:"ItemCltEffect"`                    // Uses a Client Do function (See “cltdofunc”) for handling how the skill is used when it is triggered by an item, on the client side.
//	ItemTgtDo                        int    `csv:"ItemTgtDo"`                        // Boolean Field. If equals 1, then use the skill from the enemy target instead of the caster. If equals 0, then ignore this.
//	ItemTarget                       string `csv:"ItemTarget"`                       // Controls the targeting behavior of the skill when it is triggered by an item.
//	ItemUseRestrict                  int    `csv:"ItemUseRetrict"`                   // Boolean Field. If equals 1, then use the state restriction defined in the “restrict” field when being triggered by an item.
//	ItemCheckStart                   int    `csv:"ItemCheckStart"`                   // Boolean Field. If equals 1, then use the skill’s Server Start function (See “srvstfunc”) when the skill is triggered by an item. If equals 0, then the skill’s Server Start function is ignored.
//	ItemCltCheckStart                int    `csv:"ItemCltCheckStart"`                // Boolean Field. If equals 1, then when the skill is triggered by an item, and if the target is dead and the skill has a Client Start function (See “cltstfunc”), then add the “corpse_noselect” to the target. If equals 0, then ignore this.
//	ItemCastSound                    string `csv:"ItemCastSound"`                    // Play a sound when the skill is used by an item event. Points to a “Sound” value in the sounds.txt file.
//	ItemCastOverlay                  string `csv:"ItemCastOverlay"`                  // Add a cast overlay when the skill is used by an item event. Points to an “overlay” value in the Overlay.txt file.
//	SkPoints                         int    `csv:"skpoints"`                         // Controls the number of Skill Points needed to level up the skill.
//	ReqLevel                         int    `csv:"reqlevel"`                         // Minimum character level required to be allowed to spend Skill Points on this skill.
//	MaxLevel                         int    `csv:"maxlvl"`                           // Maximum base level for the skill from spending Skill Points. Skill levels can be increased beyond this through item modifiers.
//	ReqStr                           int    `csv:"reqstr"`                           // Minimum Strength attribute required to be allowed to spend Skill Points on this skill.
//	ReqDex                           int    `csv:"reqdex"`                           // Minimum Dexterity attribute required to be allowed to spend Skill Points on this skill.
//	ReqInt                           int    `csv:"reqint"`                           // Minimum Intelligence attribute required to be allowed to spend Skill Points on this skill.
//	ReqVit                           int    `csv:"reqvit"`                           // Minimum Vitality attribute required to be allowed to spend Skill Points on this skill.
//	ReqSkill1                        string `csv:"reqskill1"`                        // Points to a “skill” field to act as a prerequisite skill. The prerequisite skill must be at least base level 1 before the player is allowed to spend Skill Points on this skill.
//	ReqSkill2                        string `csv:"reqskill2"`                        // Points to a “skill” field to act as a prerequisite skill. The prerequisite skill must be at least base level 1 before the player is allowed to spend Skill Points on this skill.
//	ReqSkill3                        string `csv:"reqskill3"`                        // Points to a “skill” field to act as a prerequisite skill. The prerequisite skill must be at least base level 1 before the player is allowed to spend Skill Points on this skill.
//	Restrict                         string `csv:"restrict"`                         // Controls how the skill is used when the unit has a restricted state (See the “restrict” field in states.txt).
//	PopState                         int    `csv:"PopState"`                         // The skill can be used at any time but when used, it will remove the unit’s restrict states.
//	State1                           string `csv:"State1"`                           // Points to a “state” field from the states.txt file. Used as parameters for the “restrict” field to control what specific states will restrict the usage of the skill.
//	State2                           string `csv:"State2"`                           // Points to a “state” field from the states.txt file. Used as parameters for the “restrict” field to control what specific states will restrict the usage of the skill.
//	State3                           string `csv:"State3"`                           // Points to a “state” field from the states.txt file. Used as parameters for the “restrict” field to control what specific states will restrict the usage of the skill.
//	LocalDelay                       int    `csv:"localdelay"`                       // Controls the Casting Delay duration for this skill after it is used (25 frames = 1 second).
//	GlobalDelay                      int    `csv:"globaldelay"`                      // Controls the Casting Delay duration for all other skills with Casting Delays after this skill is used (25 frames = 1 second).
//	LeftSkill                        int    `csv:"leftskill"`                        // Boolean Field. If equals 1, then the skill will appear on the list of skills to assign for the Left Mouse Button. If equals 0, then the skill will not appear on the Left Mouse Button skill list.
//	RightSkill                       int    `csv:"rightskill"`                       // Boolean Field. If equals 1, then the skill will appear on the list of skills to assign for the Right Mouse Button. If equals 0, then the skill will not appear on the Right Mouse Button skill list.
//	Repeat                           int    `csv:"repeat"`                           // Boolean Field. If equals 1 and the skill has no “anim” mode declared, then on the client side, the unit’s mode will repeat its current mode (this can also happen if the unit has the “inferno” state). If equals 0, then the unit will have its mode set to Neutral when starting to use the skill.
//	AlwaysHit                        int    `csv:"alwayshit"`                        // Boolean Field. If equals 1, then skills that rely on attack rating and defense will ignore those stats and will always hit enemies. If equals 0, then ignore this.
//	UseManaOnDo                      int    `csv:"usemanaondo"`                      // Boolean Field. If equals 1, then the skill will consume mana on its do function instead of its start function. If equals 0, then the skill will consume mana on its start function, which is the general case for skills.
//	StartMana                        int    `csv:"startmana"`                        // Controls the required amount of mana to start using the skill. This only works with certain “srvstfunc” functions such as SorStartInferno (Code = 11) or AssStartBladeFury (Code = 26).
//	MinMana                          int    `csv:"minmana"`                          // Controls the minimum amount of mana to use the skill. This can control skills that reduce in mana cost per level to not fall below this amount.
//	ManaShift                        string `csv:"manashift"`                        // This acts as a multiplicative modifier to control the precision of the mana cost after calculating the total mana cost with the “mana” and “lvlmana fields”. Mana is calculated in 256ths in code which equals 8 bits. This field acts as a bitwise shift value, which means it will modify the mana value by the power of 2. For example, if this value equals 5 then that means divide the mana value of 256ths by 2^5 = 32 (or multiply the mana by 32/256). A value equal to 8 means 256/256 which means that the mana of 256ths value is not shifted.
//	Mana                             int    `csv:"mana"`                             // Defines the base mana cost to use the skill at level 1.
//	LvlMana                          int    `csv:"lvlmana"`                          // Defines the change in mana cost per skill level gained.
//	Interrupt                        int    `csv:"interrupt"`                        // Boolean Field. If equals 1, then the casting the skill will be interruptible such as when the player is getting hit while casting a skill. If equals 0, then the skill should be uninterruptible.
//	InTown                           int    `csv:"InTown"`                           // Boolean Field. If equals 1, then the skill can be used while the unit is in town. If equals 0, then the skill cannot be used in town.
//	Aura                             int    `csv:"aura"`                             // Boolean Field. If equals 1, then the skill will be classified as an aura, which will make the skill execute its function periodically (using the “perdelay” field), based on if the “aurastate” state is active. Aura skills will also override a preexisting state if that state matches the skill’s “aurastate” state. If equals 0, then ignore this.
//	Periodic                         int    `csv:"periodic"`                         // Boolean Field. If equals 1, then the skill will execute its function periodically (using the “perdelay” field), based on if the “aurastate” state is active. If equals 0, then ignore this.
//	PerDelay                         string `csv:"perdelay"`                         // Calculation Field. Controls the periodic rate that the skill continuously executes its function. Minimum value equals 5. This field requires the “periodic” or “aura” field to be enabled.
//	Finishing                        int    `csv:"finishing"`                        // Boolean Field. If equals 1, then the skill will be classified as a finishing move, which can affect how progressive charges are consumed when using the skill and how the skill’s description tooltip is displayed. If equals 0, then ignore this.
//	PrgChargesToCast                 int    `csv:"prgchargestocast"`                 // Controls how many progressive charges are required to cast the skill.
//	PrgChargesConsumed               int    `csv:"prgchargesconsumed"`               // Controls how many progressive charges are consumed when the skill attack hits an enemy.
//	Passive                          int    `csv:"passive"`                          // Boolean Field. If equals 1, then the skill will be treated as a passive, which means that the skill will not show up in the skill selection menu and will not run a server do function. If equals 0, then the skill is an active skill that can be used.
//	Progressive                      string `csv:"progressive"`                      // Boolean Field. If equals 1, then the skill will use the progressive calculations to act as a charge-up skill that will add charges. This is only used for certain skill functions and will generally require the usage of the “progcalc#” fields and the “aurastat#” fields. If equals 0, then ignore this.
//	Scroll                           int    `csv:"scroll"`                           // Boolean Field. If equals 1, then the skill can appear as a scroll version in the skill selection UI if the skill allows for the scroll mechanics and if the player has the skill’s scroll item in the inventory. If equals 0, then ignore this.
//	Calc1                            string `csv:"calc1"`                            // Calculation Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Calc2                            string `csv:"calc2"`                            // Calculation Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Calc3                            string `csv:"calc3"`                            // Calculation Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Calc4                            string `csv:"calc4"`                            // Calculation Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Calc5                            string `csv:"calc5"`                            // Calculation Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Calc6                            string `csv:"calc6"`                            // Calculation Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Param1                           string `csv:"Param1"`                           // Integer Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Param2                           string `csv:"Param2"`                           // Integer Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Param3                           string `csv:"Param3"`                           // Integer Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Param4                           string `csv:"Param4"`                           // Integer Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Param5                           string `csv:"Param5"`                           // Integer Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	Param6                           string `csv:"Param6"`                           // Integer Field. It is used as a possible parameter for skill functions or as a numeric input for other calculation fields.
//	InGame                           int    `csv:"InGame"`                           // Boolean Field. If equals 1, then the skill is enabled to be used in-game. If equals 0, then the skill will be disabled in-game.
//	ToHit                            int    `csv:"ToHit"`                            // Baseline bonus Attack Rating that is added to the attack when using this skill at level 1.
//	LevToHit                         string `csv:"LevToHit"`                         // Additional bonus Attack Rating when using this skill, calculated per skill level.
//	ToHitCalc                        string `csv:"ToHitCalc"`                        // Calculation Field. Calculates the bonus Attack Rating when using the skill. This will override the “ToHit” and “LevToHit” fields if it is not blank.
//	ResultFlags                      int    `csv:"ResultFlags"`                      // Controls different flags that can affect how the target reacts after being hit by the skill attack. Uses an integer value to check against different bit fields by using the “&” operator.
//	HitFlags                         int    `csv:"HitFlags"`                         // Controls different flags that can affect the damage dealt when the target is hit by the skill. Uses an integer value to check against different bit fields by using the “&” operator.
//	HitClass                         int    `csv:"HitClass"`                         // Defines the skill’s damage routines when hitting, mainly used for determining hit sound effects and overlays. Uses an integer value to check against different bit fields by using the “&” operator.
//	Kick                             string `csv:"Kick"`                             // Boolean Field. If equals 1, then a separate function is used to calculate the physical damage dealt by the skill, ignoring the following damage fields.
//	HitShift                         string `csv:"HitShift"`                         // Controls the percentage modifier for the skill’s damage. This value cannot be less than 0 or greater than 8. This is calculated in 256ths.
//	SrcDam                           int    `csv:"SrcDam"`                           // Controls the percentage modifier for how much weapon damage is transferred to the skill’s damage (Out of 128).
//	MinDam                           int    `csv:"MinDam"`                           // Minimum baseline physical damage dealt by the skill.
//	MinLevDam1                       string `csv:"MinLevDam1"`                       // Controls the additional minimum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MinLevDam2                       string `csv:"MinLevDam2"`                       // Controls the additional minimum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MinLevDam3                       string `csv:"MinLevDam3"`                       // Controls the additional minimum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MinLevDam4                       string `csv:"MinLevDam4"`                       // Controls the additional minimum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MinLevDam5                       string `csv:"MinLevDam5"`                       // Controls the additional minimum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MaxDam                           int    `csv:"MaxDam"`                           // Maximum baseline physical damage dealt by the skill.
//	MaxLevDam1                       string `csv:"MaxLevDam1"`                       // Controls the additional maximum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MaxLevDam2                       string `csv:"MaxLevDam2"`                       // Controls the additional maximum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MaxLevDam3                       string `csv:"MaxLevDam3"`                       // Controls the additional maximum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MaxLevDam4                       string `csv:"MaxLevDam4"`                       // Controls the additional maximum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	MaxLevDam5                       string `csv:"MaxLevDam5"`                       // Controls the additional maximum physical damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	DmgSymPerCalc                    string `csv:"DmgSymPerCalc"`                    // Calculation Field. Determines the percentage increase to the physical damage dealt by the skill.
//	EType                            string `csv:"EType"`                            // Defines the type of elemental damage dealt by the skill. If this field is empty, then the related elemental fields below will not be used.
//	EMin                             int    `csv:"EMin"`                             // Minimum baseline elemental damage dealt by the skill.
//	EMinLev1                         string `csv:"EMinLev1"`                         // Controls the additional minimum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the skill’s current level.
//	EMinLev2                         string `csv:"EMinLev2"`                         // Controls the additional minimum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the skill’s current level.
//	EMinLev3                         string `csv:"EMinLev3"`                         // Controls the additional minimum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the skill’s current level.
//	EMinLev4                         string `csv:"EMinLev4"`                         // Controls the additional minimum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the skill’s current level.
//	EMinLev5                         string `csv:"EMinLev5"`                         // Controls the additional minimum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the skill’s current level.
//	EMax                             int    `csv:"EMax"`                             // Maximum baseline elemental damage dealt by the skill.
//	EMaxLev1                         string `csv:"EMaxLev1"`                         // Controls the additional maximum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	EMaxLev2                         string `csv:"EMaxLev2"`                         // Controls the additional maximum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	EMaxLev3                         string `csv:"EMaxLev3"`                         // Controls the additional maximum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	EMaxLev4                         string `csv:"EMaxLev4"`                         // Controls the additional maximum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	EMaxLev5                         string `csv:"EMaxLev5"`                         // Controls the additional maximum elemental damage dealt by the skill, calculated using the leveling formula between 5 level thresholds of the missile’s current level.
//	EDmgSymPerCalc                   string `csv:"EDmgSymPerCalc"`                   // Calculation Field. Determines the percentage increase to the total elemental damage dealt by the skill.
//	ELen                             string `csv:"ELen"`                             // The baseline elemental duration dealt by the skill. This is calculated in frame lengths where 25 Frames = 1 second. These fields only apply to appropriate elemental types with a duration.
//	ELevLen1                         string `csv:"ELevLen1"`                         // Controls the additional elemental duration added by the skill, calculated using the leveling formula between 3 level thresholds of the missile’s current level.
//	ELevLen2                         string `csv:"ELevLen2"`                         // Controls the additional elemental duration added by the skill, calculated using the leveling formula between 3 level thresholds of the missile’s current level.
//	ELevLen3                         string `csv:"ELevLen3"`                         // Controls the additional elemental duration added by the skill, calculated using the leveling formula between 3 level thresholds of the missile’s current level.
//	ELenSymPerCalc                   string `csv:"ELenSymPerCalc"`                   // Calculation Field. Determines the percentage increase to the total elemental duration dealt by the skill.
//	AIType                           string `csv:"aitype"`                           // Controls what the skill’s archetype for how the AI will handle using this skill. This mostly affects the mercenary AI and Shadow Warrior AI, but some types are used for general AI.
//	AIBonus                          int    `csv:"aibonus"`                          // This is only used with the Shadow Master AI. This value is a flat integer rating that gets added to the AI’s parameters when deciding which skill is most likely to be used next. The higher this value, then the more likely this skill will be used by the AI.
//	CostMult                         string `csv:"costmult"`                         // Multiplicative modifier of an item’s gold cost, only when the item has a stat modifier with this skill. This will affect the item’s buy, sell, and repair costs (Calculated in 1024ths).
//	CostAdd                          int    `csv:"costadd"`                          // Flat integer modification of an item’s gold cost, only when the item has a stat modifier with this skill. This will affect the item’s buy, sell, and repair costs. This is added after the “costmult” has modified the costs.
//}
