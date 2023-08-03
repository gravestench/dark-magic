package models

// BeltData holds the statistics for belts and their various item slots
type BeltData struct {
	Name                string `csv:"name" lua:"name"`                               // Reference field to define the belt type
	NumBoxes            int    `csv:"numboxes" lua:"numboxes"`                       // Number of item slots in the belt
	Box1Left            int    `csv:"box1left" lua:"box1left"`                       // Belt slot 1 left side coordinate (for Server verification)
	Box2Left            int    `csv:"box2left" lua:"box2left"`                       // Belt slot 2 left side coordinate (for Server verification)
	Box3Left            int    `csv:"box3left" lua:"box3left"`                       // Belt slot 3 left side coordinate (for Server verification)
	Box4Left            int    `csv:"box4left" lua:"box4left"`                       // Belt slot 4 left side coordinate (for Server verification)
	Box5Left            int    `csv:"box5left" lua:"box5left"`                       // Belt slot 5 left side coordinate (for Server verification)
	Box6Left            int    `csv:"box6left" lua:"box6left"`                       // Belt slot 6 left side coordinate (for Server verification)
	Box7Left            int    `csv:"box7left" lua:"box7left"`                       // Belt slot 7 left side coordinate (for Server verification)
	Box8Left            int    `csv:"box8left" lua:"box8left"`                       // Belt slot 8 left side coordinate (for Server verification)
	Box9Left            int    `csv:"box9left" lua:"box9left"`                       // Belt slot 9 left side coordinate (for Server verification)
	Box10Left           int    `csv:"box10left" lua:"box10left"`                     // Belt slot 10 left side coordinate (for Server verification)
	Box11Left           int    `csv:"box11left" lua:"box11left"`                     // Belt slot 11 left side coordinate (for Server verification)
	Box12Left           int    `csv:"box12left" lua:"box12left"`                     // Belt slot 12 left side coordinate (for Server verification)
	Box13Left           int    `csv:"box13left" lua:"box13left"`                     // Belt slot 13 left side coordinate (for Server verification)
	Box14Left           int    `csv:"box14left" lua:"box14left"`                     // Belt slot 14 left side coordinate (for Server verification)
	Box15Left           int    `csv:"box15left" lua:"box15left"`                     // Belt slot 15 left side coordinate (for Server verification)
	Box16Left           int    `csv:"box16left" lua:"box16left"`                     // Belt slot 16 left side coordinate (for Server verification)
	Box1Right           int    `csv:"box1right" lua:"box1right"`                     // Belt slot 1 right side coordinate (for Server verification)
	Box2Right           int    `csv:"box2right" lua:"box2right"`                     // Belt slot 2 right side coordinate (for Server verification)
	Box3Right           int    `csv:"box3right" lua:"box3right"`                     // Belt slot 3 right side coordinate (for Server verification)
	Box4Right           int    `csv:"box4right" lua:"box4right"`                     // Belt slot 4 right side coordinate (for Server verification)
	Box5Right           int    `csv:"box5right" lua:"box5right"`                     // Belt slot 5 right side coordinate (for Server verification)
	Box6Right           int    `csv:"box6right" lua:"box6right"`                     // Belt slot 6 right side coordinate (for Server verification)
	Box7Right           int    `csv:"box7right" lua:"box7right"`                     // Belt slot 7 right side coordinate (for Server verification)
	Box8Right           int    `csv:"box8right" lua:"box8right"`                     // Belt slot 8 right side coordinate (for Server verification)
	Box9Right           int    `csv:"box9right" lua:"box9right"`                     // Belt slot 9 right side coordinate (for Server verification)
	Box10Right          int    `csv:"box10right" lua:"box10right"`                   // Belt slot 10 right side coordinate (for Server verification)
	Box11Right          int    `csv:"box11right" lua:"box11right"`                   // Belt slot 11 right side coordinate (for Server verification)
	Box12Right          int    `csv:"box12right" lua:"box12right"`                   // Belt slot 12 right side coordinate (for Server verification)
	Box13Right          int    `csv:"box13right" lua:"box13right"`                   // Belt slot 13 right side coordinate (for Server verification)
	Box14Right          int    `csv:"box14right" lua:"box14right"`                   // Belt slot 14 right side coordinate (for Server verification)
	Box15Right          int    `csv:"box15right" lua:"box15right"`                   // Belt slot 15 right side coordinate (for Server verification)
	Box16Right          int    `csv:"box16right" lua:"box16right"`                   // Belt slot 16 right side coordinate (for Server verification)
	Box1Top             int    `csv:"box1top" lua:"box1top"`                         // Belt slot 1 top coordinate (for Server verification)
	Box2Top             int    `csv:"box2top" lua:"box2top"`                         // Belt slot 2 top coordinate (for Server verification)
	Box3Top             int    `csv:"box3top" lua:"box3top"`                         // Belt slot 3 top coordinate (for Server verification)
	Box4Top             int    `csv:"box4top" lua:"box4top"`                         // Belt slot 4 top coordinate (for Server verification)
	Box5Top             int    `csv:"box5top" lua:"box5top"`                         // Belt slot 5 top coordinate (for Server verification)
	Box6Top             int    `csv:"box6top" lua:"box6top"`                         // Belt slot 6 top coordinate (for Server verification)
	Box7Top             int    `csv:"box7top" lua:"box7top"`                         // Belt slot 7 top coordinate (for Server verification)
	Box8Top             int    `csv:"box8top" lua:"box8top"`                         // Belt slot 8 top coordinate (for Server verification)
	Box9Top             int    `csv:"box9top" lua:"box9top"`                         // Belt slot 9 top coordinate (for Server verification)
	Box10Top            int    `csv:"box10top" lua:"box10top"`                       // Belt slot 10 top coordinate (for Server verification)
	Box11Top            int    `csv:"box11top" lua:"box11top"`                       // Belt slot 11 top coordinate (for Server verification)
	Box12Top            int    `csv:"box12top" lua:"box12top"`                       // Belt slot 12 top coordinate (for Server verification)
	Box13Top            int    `csv:"box13top" lua:"box13top"`                       // Belt slot 13 top coordinate (for Server verification)
	Box14Top            int    `csv:"box14top" lua:"box14top"`                       // Belt slot 14 top coordinate (for Server verification)
	Box15Top            int    `csv:"box15top" lua:"box15top"`                       // Belt slot 15 top coordinate (for Server verification)
	Box16Top            int    `csv:"box16top" lua:"box16top"`                       // Belt slot 16 top coordinate (for Server verification)
	Box1Bottom          int    `csv:"box1bottom" lua:"box1bottom"`                   // Belt slot 1 bottom coordinate (for Server verification)
	Box2Bottom          int    `csv:"box2bottom" lua:"box2bottom"`                   // Belt slot 2 bottom coordinate (for Server verification)
	Box3Bottom          int    `csv:"box3bottom" lua:"box3bottom"`                   // Belt slot 3 bottom coordinate (for Server verification)
	Box4Bottom          int    `csv:"box4bottom" lua:"box4bottom"`                   // Belt slot 4 bottom coordinate (for Server verification)
	Box5Bottom          int    `csv:"box5bottom" lua:"box5bottom"`                   // Belt slot 5 bottom coordinate (for Server verification)
	Box6Bottom          int    `csv:"box6bottom" lua:"box6bottom"`                   // Belt slot 6 bottom coordinate (for Server verification)
	Box7Bottom          int    `csv:"box7bottom" lua:"box7bottom"`                   // Belt slot 7 bottom coordinate (for Server verification)
	Box8Bottom          int    `csv:"box8bottom" lua:"box8bottom"`                   // Belt slot 8 bottom coordinate (for Server verification)
	Box9Bottom          int    `csv:"box9bottom" lua:"box9bottom"`                   // Belt slot 9 bottom coordinate (for Server verification)
	Box10Bottom         int    `csv:"box10bottom" lua:"box10bottom"`                 // Belt slot 10 bottom coordinate (for Server verification)
	Box11Bottom         int    `csv:"box11bottom" lua:"box11bottom"`                 // Belt slot 11 bottom coordinate (for Server verification)
	Box12Bottom         int    `csv:"box12bottom" lua:"box12bottom"`                 // Belt slot 12 bottom coordinate (for Server verification)
	Box13Bottom         int    `csv:"box13bottom" lua:"box13bottom"`                 // Belt slot 13 bottom coordinate (for Server verification)
	Box14Bottom         int    `csv:"box14bottom" lua:"box14bottom"`                 // Belt slot 14 bottom coordinate (for Server verification)
	Box15Bottom         int    `csv:"box15bottom" lua:"box15bottom"`                 // Belt slot 15 bottom coordinate (for Server verification)
	Box16Bottom         int    `csv:"box16bottom" lua:"box16bottom"`                 // Belt slot 16 bottom coordinate (for Server verification)
	DefaultItemTypeCol1 string `csv:"defaultItemTypeCol1" lua:"defaultItemTypeCol1"` // Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemTypeCol2 string `csv:"defaultItemTypeCol2" lua:"defaultItemTypeCol2"` // Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemTypeCol3 string `csv:"defaultItemTypeCol3" lua:"defaultItemTypeCol3"` // Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemTypeCol4 string `csv:"defaultItemTypeCol4" lua:"defaultItemTypeCol4"` // Default item type used for the populate belt and auto-use functionality on the controller
	DefaultItemCodeCol1 string `csv:"defaultItemCodeCol1" lua:"defaultItemCodeCol1"` // Default item code used for the populate belt and auto-use functionality on the controller
	DefaultItemCodeCol2 string `csv:"defaultItemCodeCol2" lua:"defaultItemCodeCol2"` // Default item code used for the populate belt and auto-use functionality on the controller
	DefaultItemCodeCol3 string `csv:"defaultItemCodeCol3" lua:"defaultItemCodeCol3"` // Default item code used for the populate belt and auto-use functionality on the controller
	DefaultItemCodeCol4 string `csv:"defaultItemCodeCol4" lua:"defaultItemCodeCol4"` // Default item code used for the populate belt and auto-use functionality on the controller
}
