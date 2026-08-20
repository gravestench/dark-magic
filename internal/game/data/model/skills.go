package models

// SkillData represents the data fields in the skills.txt file.
type SkillData struct {
	ID                string `csv:"Id"`                // Unique identifier for the skill.
	CharClass         string `csv:"charclass"`         // Character class associated with the skill.
	SkillName         string `csv:"skill"`             // Description of the skill.
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
	// Controls different flags for the target's reaction after being hit.
	ResultFlags string `csv:"ResultFlags"`
	// Controls different flags for the damage dealt when the target is hit.
	HitFlags string `csv:"HitFlags"`
	HitClass string `csv:"HitClass"` // Defines the skill's damage routines when hitting.
	Kick     string `csv:"Kick"`     // Separate function for calculating physical damage when kicking.
	HitShift string `csv:"HitShift"` // Percentage modifier for the skill's damage.
	// Percentage modifier for weapon damage transferred to the skill's damage.
	SrcDam         string `csv:"SrcDam"`
	MinDam         string `csv:"MinDam"`         // Minimum baseline physical damage dealt by the skill.
	MinLevDam1     string `csv:"MinLevDam1"`     // Additional minimum physical damage dealt at level thresholds.
	MinLevDam2     string `csv:"MinLevDam2"`     // Additional minimum physical damage dealt at level thresholds.
	MinLevDam3     string `csv:"MinLevDam3"`     // Additional minimum physical damage dealt at level thresholds.
	MinLevDam4     string `csv:"MinLevDam4"`     // Additional minimum physical damage dealt at level thresholds.
	MinLevDam5     string `csv:"MinLevDam5"`     // Additional minimum physical damage dealt at level thresholds.
	MaxDam         string `csv:"MaxDam"`         // Maximum baseline physical damage dealt by the skill.
	MaxLevDam1     string `csv:"MaxLevDam1"`     // Additional maximum physical damage dealt at level thresholds.
	MaxLevDam2     string `csv:"MaxLevDam2"`     // Additional maximum physical damage dealt at level thresholds.
	MaxLevDam3     string `csv:"MaxLevDam3"`     // Additional maximum physical damage dealt at level thresholds.
	MaxLevDam4     string `csv:"MaxLevDam4"`     // Additional maximum physical damage dealt at level thresholds.
	MaxLevDam5     string `csv:"MaxLevDam5"`     // Additional maximum physical damage dealt at level thresholds.
	DmgSymPerCalc  string `csv:"DmgSymPerCalc"`  // Damage symmetrical percentage calculation.
	EType          string `csv:"EType"`          // Elemental type.
	EMin           string `csv:"EMin"`           // Minimum baseline elemental damage dealt by the skill.
	EMinLev1       string `csv:"EMinLev1"`       // Additional minimum elemental damage dealt at level thresholds.
	EMinLev2       string `csv:"EMinLev2"`       // Additional minimum elemental damage dealt at level thresholds.
	EMinLev3       string `csv:"EMinLev3"`       // Additional minimum elemental damage dealt at level thresholds.
	EMinLev4       string `csv:"EMinLev4"`       // Additional minimum elemental damage dealt at level thresholds.
	EMinLev5       string `csv:"EMinLev5"`       // Additional minimum elemental damage dealt at level thresholds.
	EMax           string `csv:"EMax"`           // Maximum baseline elemental damage dealt by the skill.
	EMaxLev1       string `csv:"EMaxLev1"`       // Additional maximum elemental damage dealt at level thresholds.
	EMaxLev2       string `csv:"EMaxLev2"`       // Additional maximum elemental damage dealt at level thresholds.
	EMaxLev3       string `csv:"EMaxLev3"`       // Additional maximum elemental damage dealt at level thresholds.
	EMaxLev4       string `csv:"EMaxLev4"`       // Additional maximum elemental damage dealt at level thresholds.
	EMaxLev5       string `csv:"EMaxLev5"`       // Additional maximum elemental damage dealt at level thresholds.
	EDmgSymPerCalc string `csv:"EDmgSymPerCalc"` // Elemental damage symmetrical percentage calculation.
	ELen           string `csv:"ELen"`           // Elemental length.
	ELevLen1       string `csv:"ELevLen1"`       // Additional elemental length at level thresholds.
	ELevLen2       string `csv:"ELevLen2"`       // Additional elemental length at level thresholds.
	ELevLen3       string `csv:"ELevLen3"`       // Additional elemental length at level thresholds.
	ELenSymPerCalc string `csv:"ELenSymPerCalc"` // Elemental length symmetrical percentage calculation.
	AIType         string `csv:"aitype"`         // AI type.
	AIBonus        string `csv:"aibonus"`        // AI bonus.
	CostMult       string `csv:"cost mult"`      // Cost multiplier.
	CostAdd        string `csv:"cost add"`       // Cost addition.
}
