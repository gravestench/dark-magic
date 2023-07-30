package models

// BeltData holds the statistics for belts and their various item slots
type BeltData struct {
	Name                string `csv:"name"`                // Reference field to define the belt type
	NumBoxes            int    `csv:"numboxes"`            // Number of item slots in the belt
	Box1Left            int    `csv:"box1left"`            // Belt slot 1 left side coordinate (for Server verification)
	Box2Left            int    `csv:"box2left"`            // Belt slot 2 left side coordinate (for Server verification)
	Box3Left            int    `csv:"box3left"`            // Belt slot 3 left side coordinate (for Server verification)
	Box4Left            int    `csv:"box4left"`            // Belt slot 4 left side coordinate (for Server verification)
	Box5Left            int    `csv:"box5left"`            // Belt slot 5 left side coordinate (for Server verification)
	Box6Left            int    `csv:"box6left"`            // Belt slot 6 left side coordinate (for Server verification)
	Box7Left            int    `csv:"box7left"`            // Belt slot 7 left side coordinate (for Server verification)
	Box8Left            int    `csv:"box8left"`            // Belt slot 8 left side coordinate (for Server verification)
	Box9Left            int    `csv:"box9left"`            // Belt slot 9 left side coordinate (for Server verification)
	Box10Left           int    `csv:"box10left"`           // Belt slot 10 left side coordinate (for Server verification)
	Box11Left           int    `csv:"box11left"`           // Belt slot 11 left side coordinate (for Server verification)
	Box12Left           int    `csv:"box12left"`           // Belt slot 12 left side coordinate (for Server verification)
	Box13Left           int    `csv:"box13left"`           // Belt slot 13 left side coordinate (for Server verification)
	Box14Left           int    `csv:"box14left"`           // Belt slot 14 left side coordinate (for Server verification)
	Box15Left           int    `csv:"box15left"`           // Belt slot 15 left side coordinate (for Server verification)
	Box16Left           int    `csv:"box16left"`           // Belt slot 16 left side coordinate (for Server verification)
	Box1Right           int    `csv:"box1right"`           // Belt slot 1 right side coordinate (for Server verification)
	Box2Right           int    `csv:"box2right"`           // Belt slot 2 right side coordinate (for Server verification)
	Box3Right           int    `csv:"box3right"`           // Belt slot 3 right side coordinate (for Server verification)
	Box4Right           int    `csv:"box4right"`           // Belt slot 4 right side coordinate (for Server verification)
	Box5Right           int    `csv:"box5right"`           // Belt slot 5 right side coordinate (for Server verification)
	Box6Right           int    `csv:"box6right"`           // Belt slot 6 right side coordinate (for Server verification)
	Box7Right           int    `csv:"box7right"`           // Belt slot 7 right side coordinate (for Server verification)
	Box8Right           int    `csv:"box8right"`           // Belt slot 8 right side coordinate (for Server verification)
	Box9Right           int    `csv:"box9right"`           // Belt slot 9 right side coordinate (for Server verification)
	Box10Right          int    `csv:"box10right"`          // Belt slot 10 right side coordinate (for Server verification)
	Box11Right          int    `csv:"box11right"`          // Belt slot 11 right side coordinate (for Server verification)
	Box12Right          int    `csv:"box12right"`          // Belt slot 12 right side coordinate (for Server verification)
	Box13Right          int    `csv:"box13right"`          // Belt slot 13 right side coordinate (for Server verification)
	Box14Right          int    `csv:"box14right"`          // Belt slot 14 right side coordinate (for Server verification)
	Box15Right          int    `csv:"box15right"`          // Belt slot 15 right side coordinate (for Server verification)
	Box16Right          int    `csv:"box16right"`          // Belt slot 16 right side coordinate (for Server verification)
	Box1Top             int    `csv:"box1top"`             // Belt slot 1 top coordinate (for Server verification)
	Box2Top             int    `csv:"box2top"`             // Belt slot 2 top coordinate (for Server verification)
	Box3Top             int    `csv:"box3top"`             // Belt slot 3 top coordinate (for Server verification)
	Box4Top             int    `csv:"box4top"`             // Belt slot 4 top coordinate (for Server verification)
	Box5Top             int    `csv:"box5top"`             // Belt slot 5 top coordinate (for Server verification)
	Box6Top             int    `csv:"box6top"`             // Belt slot 6 top coordinate (for Server verification)
	Box7Top             int    `csv:"box7top"`             // Belt slot 7 top coordinate (for Server verification)
	Box8Top             int    `csv:"box8top"`             // Belt slot 8 top coordinate (for Server verification)
	Box9Top             int    `csv:"box9top"`             // Belt slot 9 top coordinate (for Server verification)
	Box10Top            int    `csv:"box10top"`            // Belt slot 10 top coordinate (for Server verification)
	Box11Top            int    `csv:"box11top"`            // Belt slot 11 top coordinate (for Server verification)
	Box12Top            int    `csv:"box12top"`            // Belt slot 12 top coordinate (for Server verification)
	Box13Top            int    `csv:"box13top"`            // Belt slot 13 top coordinate (for Server verification)
	Box14Top            int    `csv:"box14top"`            // Belt slot 14 top coordinate (for Server verification)
	Box15Top            int    `csv:"box15top"`            // Belt slot 15 top coordinate (for Server verification)
	Box16Top            int    `csv:"box16top"`            // Belt slot 16 top coordinate (for Server verification)
	Box1Bottom          int    `csv:"box1bottom"`          // Belt slot 1 bottom coordinate (for Server verification)
	Box2Bottom          int    `csv:"box2bottom"`          // Belt slot 2 bottom coordinate (for Server verification)
	Box3Bottom          int    `csv:"box3bottom"`          // Belt slot 3 bottom coordinate (for Server verification)
	Box4Bottom          int    `csv:"box4bottom"`          // Belt slot 4 bottom coordinate (for Server verification)
	Box5Bottom          int    `csv:"box5bottom"`          // Belt slot 5 bottom coordinate (for Server verification)
	Box6Bottom          int    `csv:"box6bottom"`          // Belt slot 6 bottom coordinate (for Server verification)
	Box7Bottom          int    `csv:"box7bottom"`          // Belt slot 7 bottom coordinate (for Server verification)
	Box8Bottom          int    `csv:"box8bottom"`          // Belt slot 8 bottom coordinate (for Server verification)
	Box9Bottom          int    `csv:"box9bottom"`          // Belt slot 9 bottom coordinate (for Server verification)
	Box10Bottom         int    `csv:"box10bottom"`         // Belt slot 10 bottom coordinate (for Server verification)
	Box11Bottom         int    `csv:"box11bottom"`         // Belt slot 11 bottom coordinate (for Server verification)
	Box12Bottom         int    `csv:"box12bottom"`         // Belt slot 12 bottom coordinate (for Server verification)
	Box13Bottom         int    `csv:"box13bottom"`         // Belt slot 13 bottom coordinate (for Server verification)
	Box14Bottom         int    `csv:"box14bottom"`         // Belt slot 14 bottom coordinate (for Server verification)
	Box15Bottom         int    `csv:"box15bottom"`         // Belt slot 15 bottom coordinate (for Server verification)
	Box16Bottom         int    `csv:"box16bottom"`         // Belt slot 16 bottom coordinate (for Server verification)
	DefaultItemTypeCol1 string `csv:"defaultItemTypeCol1"` // Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemTypeCol2 string `csv:"defaultItemTypeCol2"` // Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemTypeCol3 string `csv:"defaultItemTypeCol3"` // Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemTypeCol4 string `csv:"defaultItemTypeCol4"` // Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemCodeCol1 string `csv:"defaultItemCodeCol1"` // Default item code used for the populate belt and auto-use functionality on the controller
	DefaultItemCodeCol2 string `csv:"defaultItemCodeCol2"` // Default item code used for the populate belt and auto-use functionality on the controller
	DefaultItemCodeCol3 string `csv:"defaultItemCodeCol3"` // Default item code used for the populate belt and auto-use functionality on the controller
	DefaultItemCodeCol4 string `csv:"defaultItemCodeCol4"` // Default item code used for the populate belt and auto-use functionality on the controller
}
