package models

type MiscItem struct {
	Name           string `csv:"name"`
	Version        int    `csv:"version"`
	CompactSave    int    `csv:"compactsave"`
	Rarity         int    `csv:"rarity"`
	Spawnable      int    `csv:"spawnable"`
	Speed          int    `csv:"speed"`
	ReqStr         int    `csv:"reqstr"`
	ReqDex         int    `csv:"reqdex"`
	Durability     int    `csv:"durability"`
	NoDurability   int    `csv:"nodurability"`
	Level          int    `csv:"level"`
	ShowLevel      int    `csv:"ShowLevel"`
	LevelReq       int    `csv:"levelreq"`
	Cost           int    `csv:"cost"`
	GambleCost     int    `csv:"gamble cost"`
	Code           string `csv:"code"`
	NameStr        string `csv:"namestr"`
	MagicLvl       int    `csv:"magic lvl"`
	AutoPrefix     int    `csv:"auto prefix"`
	AlternateGfx   string `csv:"alternategfx"`
	NormCode       string `csv:"normcode"`
	UberCode       string `csv:"ubercode"`
	UltraCode      string `csv:"ultracode"`
	Component      string `csv:"component"`
	InvWidth       int    `csv:"invwidth"`
	InvHeight      int    `csv:"invheight"`
	HasInv         int    `csv:"hasinv"`
	GemSockets     int    `csv:"gemsockets"`
	GemApplyType   string `csv:"gemapplytype"`
	FlippyFile     string `csv:"flippyfile"`
	InvFile        string `csv:"invfile"`
	UniqueInvFile  string `csv:"uniqueinvfile"`
	SetInvFile     string `csv:"setinvfile"`
	Useable        int    `csv:"useable"`
	Stackable      int    `csv:"stackable"`
	MinStack       int    `csv:"minstack"`
	MaxStack       int    `csv:"maxstack"`
	SpawnStack     int    `csv:"spawnstack"`
	Transmogrify   int    `csv:"Transmogrify"`
	TMogType       string `csv:"TMogType"`
	TMogMin        int    `csv:"TMogMin"`
	TMogMax        int    `csv:"TMogMax"`
	Type           string `csv:"type"`
	Type2          string `csv:"type2"`
	DropSound      string `csv:"dropsound"`
	DropSfxFrame   int    `csv:"dropsfxframe"`
	UseSound       string `csv:"usesound"`
	Unique         int    `csv:"unique"`
	Transparent    int    `csv:"transparent"`
	TransTbl       string `csv:"transtbl"`
	LightRadius    int    `csv:"lightradius"`
	Belt           int    `csv:"belt"`
	Quest          int    `csv:"quest"`
	QuestDiffCheck int    `csv:"questdiffcheck"`
	MissileType    string `csv:"missiletype"`
	DurWarning     int    `csv:"durwarning"`
	QntWarning     int    `csv:"qntwarning"`
	MinDam         int    `csv:"mindam"`
	MaxDam         int    `csv:"maxdam"`
	StrBonus       int    `csv:"StrBonus"`
	DexBonus       int    `csv:"DexBonus"`
	GemOffset      int    `csv:"gemoffset"`
	BitField1      int    `csv:"bitfield1"`
	Autobelt       int    `csv:"autobelt"`
	BetterGem      string `csv:"bettergem"`
	Multibuy       int    `csv:"multibuy"`
	SpellIcon      int    `csv:"spellicon"`
	PSpell         int    `csv:"pspell"`
	State          string `csv:"state"`
	CState1        string `csv:"cstate1"`
	CState2        string `csv:"cstate2"`
	Len            int    `csv:"len"`
	Stat1          string `csv:"stat1"`
	Stat2          string `csv:"stat2"`
	Stat3          string `csv:"stat3"`
	Calc1          int    `csv:"calc1"`
	Calc2          int    `csv:"calc2"`
	Calc3          int    `csv:"calc3"`
	SpellDesc      int    `csv:"spelldesc"`
	SpellDescStr   string `csv:"spelldescstr"`
	SpellDescStr2  string `csv:"spelldescstr2"`
	SpellDescCalc  int    `csv:"spelldesccalc"`
	SpellDescColor int    `csv:"spelldesccolor"`

	CharsiMin      int `csv:"CharsiMin"`
	CharsiMax      int `csv:"CharsiMax"`
	CharsiMagicMin int `csv:"CharsiMagicMin"`
	CharsiMagicMax int `csv:"CharsiMagicMax"`
	CharsiMagicLvl int `csv:"CharsiMagicLvl"`

	GheedMin      int `csv:"GheedMin"`
	GheedMax      int `csv:"GheedMax"`
	GheedMagicMin int `csv:"GheedMagicMin"`
	GheedMagicMax int `csv:"GheedMagicMax"`
	GheedMagicLvl int `csv:"GheedMagicLvl"`

	AkaraMin      int `csv:"AkaraMin"`
	AkaraMax      int `csv:"AkaraMax"`
	AkaraMagicMin int `csv:"AkaraMagicMin"`
	AkaraMagicMax int `csv:"AkaraMagicMax"`
	AkaraMagicLvl int `csv:"AkaraMagicLvl"`

	FaraMin      int `csv:"FaraMin"`
	FaraMax      int `csv:"FaraMax"`
	FaraMagicMin int `csv:"FaraMagicMin"`
	FaraMagicMax int `csv:"FaraMagicMax"`
	FaraMagicLvl int `csv:"FaraMagicLvl"`

	LysanderMin      int `csv:"LysanderMin"`
	LysanderMax      int `csv:"LysanderMax"`
	LysanderMagicMin int `csv:"LysanderMagicMin"`
	LysanderMagicMax int `csv:"LysanderMagicMax"`
	LysanderMagicLvl int `csv:"LysanderMagicLvl"`

	DrognanMin      int `csv:"DrognanMin"`
	DrognanMax      int `csv:"DrognanMax"`
	DrognanMagicMin int `csv:"DrognanMagicMin"`
	DrognanMagicMax int `csv:"DrognanMagicMax"`
	DrognanMagicLvl int `csv:"DrognanMagicLvl"`

	HratliMin      int `csv:"HratliMin"`
	HratliMax      int `csv:"HratliMax"`
	HratliMagicMin int `csv:"HratliMagicMin"`
	HratliMagicMax int `csv:"HratliMagicMax"`
	HratliMagicLvl int `csv:"HratliMagicLvl"`

	AlkorMin      int `csv:"AlkorMin"`
	AlkorMax      int `csv:"AlkorMax"`
	AlkorMagicMin int `csv:"AlkorMagicMin"`
	AlkorMagicMax int `csv:"AlkorMagicMax"`
	AlkorMagicLvl int `csv:"AlkorMagicLvl"`

	OrmusMin      int `csv:"OrmusMin"`
	OrmusMax      int `csv:"OrmusMax"`
	OrmusMagicMin int `csv:"OrmusMagicMin"`
	OrmusMagicMax int `csv:"OrmusMagicMax"`
	OrmusMagicLvl int `csv:"OrmusMagicLvl"`

	ElzixMin      int `csv:"ElzixMin"`
	ElzixMax      int `csv:"ElzixMax"`
	ElzixMagicMin int `csv:"ElzixMagicMin"`
	ElzixMagicMax int `csv:"ElzixMagicMax"`
	ElzixMagicLvl int `csv:"ElzixMagicLvl"`

	AshearaMin      int `csv:"AshearaMin"`
	AshearaMax      int `csv:"AshearaMax"`
	AshearaMagicMin int `csv:"AshearaMagicMin"`
	AshearaMagicMax int `csv:"AshearaMagicMax"`
	AshearaMagicLvl int `csv:"AshearaMagicLvl"`

	CainMin      int `csv:"CainMin"`
	CainMax      int `csv:"CainMax"`
	CainMagicMin int `csv:"CainMagicMin"`
	CainMagicMax int `csv:"CainMagicMax"`
	CainMagicLvl int `csv:"CainMagicLvl"`

	HalbuMin      int `csv:"HalbuMin"`
	HalbuMax      int `csv:"HalbuMax"`
	HalbuMagicMin int `csv:"HalbuMagicMin"`
	HalbuMagicMax int `csv:"HalbuMagicMax"`
	HalbuMagicLvl int `csv:"HalbuMagicLvl"`

	JamellaMin      int `csv:"JamellaMin"`
	JamellaMax      int `csv:"JamellaMax"`
	JamellaMagicMin int `csv:"JamellaMagicMin"`
	JamellaMagicMax int `csv:"JamellaMagicMax"`
	JamellaMagicLvl int `csv:"JamellaMagicLvl"`

	LarzukMin      int `csv:"LarzukMin"`
	LarzukMax      int `csv:"LarzukMax"`
	LarzukMagicMin int `csv:"LarzukMagicMin"`
	LarzukMagicMax int `csv:"LarzukMagicMax"`
	LarzukMagicLvl int `csv:"LarzukMagicLvl"`

	MalahMin      int `csv:"MalahMin"`
	MalahMax      int `csv:"MalahMax"`
	MalahMagicMin int `csv:"MalahMagicMin"`
	MalahMagicMax int `csv:"MalahMagicMax"`
	MalahMagicLvl int `csv:"MalahMagicLvl"`

	AnyaMin      int `csv:"AnyaMin"`
	AnyaMax      int `csv:"AnyaMax"`
	AnyaMagicMin int `csv:"AnyaMagicMin"`
	AnyaMagicMax int `csv:"AnyaMagicMax"`
	AnyaMagicLvl int `csv:"AnyaMagicLvl"`
}
