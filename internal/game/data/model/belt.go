package models

// BeltData holds the statistics for belts and their various item slots
type BeltData struct {
	// Reference field to define the belt type
	Name     string `csv:"name" lua:"name"`
	NumBoxes int    `csv:"numboxes" lua:"numboxes"` // Number of item slots in the belt

	// Belt slot 1 left side coordinates (for Server verification)
	Box1Left  int `csv:"box1left" lua:"box1left"`
	Box2Left  int `csv:"box2left" lua:"box2left"`
	Box3Left  int `csv:"box3left" lua:"box3left"`
	Box4Left  int `csv:"box4left" lua:"box4left"`
	Box5Left  int `csv:"box5left" lua:"box5left"`
	Box6Left  int `csv:"box6left" lua:"box6left"`
	Box7Left  int `csv:"box7left" lua:"box7left"`
	Box8Left  int `csv:"box8left" lua:"box8left"`
	Box9Left  int `csv:"box9left" lua:"box9left"`
	Box10Left int `csv:"box10left" lua:"box10left"`
	Box11Left int `csv:"box11left" lua:"box11left"`
	Box12Left int `csv:"box12left" lua:"box12left"`
	Box13Left int `csv:"box13left" lua:"box13left"`
	Box14Left int `csv:"box14left" lua:"box14left"`
	Box15Left int `csv:"box15left" lua:"box15left"`
	Box16Left int `csv:"box16left" lua:"box16left"`

	// Belt slot 1 right side coordinates (for Server verification)
	Box1Right  int `csv:"box1right" lua:"box1right"`
	Box2Right  int `csv:"box2right" lua:"box2right"`
	Box3Right  int `csv:"box3right" lua:"box3right"`
	Box4Right  int `csv:"box4right" lua:"box4right"`
	Box5Right  int `csv:"box5right" lua:"box5right"`
	Box6Right  int `csv:"box6right" lua:"box6right"`
	Box7Right  int `csv:"box7right" lua:"box7right"`
	Box8Right  int `csv:"box8right" lua:"box8right"`
	Box9Right  int `csv:"box9right" lua:"box9right"`
	Box10Right int `csv:"box10right" lua:"box10right"`
	Box11Right int `csv:"box11right" lua:"box11right"`
	Box12Right int `csv:"box12right" lua:"box12right"`
	Box13Right int `csv:"box13right" lua:"box13right"`
	Box14Right int `csv:"box14right" lua:"box14right"`
	Box15Right int `csv:"box15right" lua:"box15right"`
	Box16Right int `csv:"box16right" lua:"box16right"`

	// Belt slot 1 top coordinates (for Server verification)
	Box1Top  int `csv:"box1top" lua:"box1top"`
	Box2Top  int `csv:"box2top" lua:"box2top"`
	Box3Top  int `csv:"box3top" lua:"box3top"`
	Box4Top  int `csv:"box4top" lua:"box4top"`
	Box5Top  int `csv:"box5top" lua:"box5top"`
	Box6Top  int `csv:"box6top" lua:"box6top"`
	Box7Top  int `csv:"box7top" lua:"box7top"`
	Box8Top  int `csv:"box8top" lua:"box8top"`
	Box9Top  int `csv:"box9top" lua:"box9top"`
	Box10Top int `csv:"box10top" lua:"box10top"`
	Box11Top int `csv:"box11top" lua:"box11top"`
	Box12Top int `csv:"box12top" lua:"box12top"`
	Box13Top int `csv:"box13top" lua:"box13top"`
	Box14Top int `csv:"box14top" lua:"box14top"`
	Box15Top int `csv:"box15top" lua:"box15top"`
	Box16Top int `csv:"box16top" lua:"box16top"`

	// Belt slot 1 bottom coordinates (for Server verification)
	Box1Bottom  int `csv:"box1bottom" lua:"box1bottom"`
	Box2Bottom  int `csv:"box2bottom" lua:"box2bottom"`
	Box3Bottom  int `csv:"box3bottom" lua:"box3bottom"`
	Box4Bottom  int `csv:"box4bottom" lua:"box4bottom"`
	Box5Bottom  int `csv:"box5bottom" lua:"box5bottom"`
	Box6Bottom  int `csv:"box6bottom" lua:"box6bottom"`
	Box7Bottom  int `csv:"box7bottom" lua:"box7bottom"`
	Box8Bottom  int `csv:"box8bottom" lua:"box8bottom"`
	Box9Bottom  int `csv:"box9bottom" lua:"box9bottom"`
	Box10Bottom int `csv:"box10bottom" lua:"box10bottom"`
	Box11Bottom int `csv:"box11bottom" lua:"box11bottom"`
	Box12Bottom int `csv:"box12bottom" lua:"box12bottom"`
	Box13Bottom int `csv:"box13bottom" lua:"box13bottom"`
	Box14Bottom int `csv:"box14bottom" lua:"box14bottom"`
	Box15Bottom int `csv:"box15bottom" lua:"box15bottom"`
	Box16Bottom int `csv:"box16bottom" lua:"box16bottom"`

	// Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemTypeCol1 string `csv:"defaultItemTypeCol1" lua:"defaultItemTypeCol1"`
	DefaultItemTypeCol2 string `csv:"defaultItemTypeCol2" lua:"defaultItemTypeCol2"`
	DefaultItemTypeCol3 string `csv:"defaultItemTypeCol3" lua:"defaultItemTypeCol3"`
	DefaultItemTypeCol4 string `csv:"defaultItemTypeCol4" lua:"defaultItemTypeCol4"`
	DefaultItemCodeCol1 string `csv:"defaultItemCodeCol1" lua:"defaultItemCodeCol1"`
	DefaultItemCodeCol2 string `csv:"defaultItemCodeCol2" lua:"defaultItemCodeCol2"`
	DefaultItemCodeCol3 string `csv:"defaultItemCodeCol3" lua:"defaultItemCodeCol3"`
	DefaultItemCodeCol4 string `csv:"defaultItemCodeCol4" lua:"defaultItemCodeCol4"`
}
