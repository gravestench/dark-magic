package models

// MonsterProp represents special properties that can be added to a monster based on the game difficulty in an older version of the game.
type MonsterProp struct {
	ID       string `csv:"Id"`          // Defines the monster that should gain the Property (points to the matching "Id" value in the monstats.txt file).
	Prop1    string `csv:"prop1"`       // Defines the first Property to apply to the monster (Uses the "code" field from Properties.txt).
	Chance1  int    `csv:"chance1"`     // The percent chance that the first related property (prop1) will be assigned. If this value equals 0, then the Property will always be applied.
	Param1   string `csv:"par1"`        // The "parameter" value associated with the first related property (prop1). Usage depends on the property function (See the "func" field on Properties.txt).
	Min1     int    `csv:"min1"`        // The "min" value to assign to the first related property (prop1). Usage depends on the property function (See the "func" field on Properties.txt).
	Max1     int    `csv:"max1"`        // The "max" value to assign to the first related property (prop1). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop2    string `csv:"prop2"`       // Defines the second Property to apply to the monster (Uses the "code" field from Properties.txt).
	Chance2  int    `csv:"chance2"`     // The percent chance that the second related property (prop2) will be assigned. If this value equals 0, then the Property will always be applied.
	Param2   string `csv:"par2"`        // The "parameter" value associated with the second related property (prop2). Usage depends on the property function (See the "func" field on Properties.txt).
	Min2     int    `csv:"min2"`        // The "min" value to assign to the second related property (prop2). Usage depends on the property function (See the "func" field on Properties.txt).
	Max2     int    `csv:"max2"`        // The "max" value to assign to the second related property (prop2). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop3    string `csv:"prop3"`       // Defines the third Property to apply to the monster (Uses the "code" field from Properties.txt).
	Chance3  int    `csv:"chance3"`     // The percent chance that the third related property (prop3) will be assigned. If this value equals 0, then the Property will always be applied.
	Param3   string `csv:"par3"`        // The "parameter" value associated with the third related property (prop3). Usage depends on the property function (See the "func" field on Properties.txt).
	Min3     int    `csv:"min3"`        // The "min" value to assign to the third related property (prop3). Usage depends on the property function (See the "func" field on Properties.txt).
	Max3     int    `csv:"max3"`        // The "max" value to assign to the third related property (prop3). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop4    string `csv:"prop4"`       // Defines the fourth Property to apply to the monster (Uses the "code" field from Properties.txt).
	Chance4  int    `csv:"chance4"`     // The percent chance that the fourth related property (prop4) will be assigned. If this value equals 0, then the Property will always be applied.
	Param4   string `csv:"par4"`        // The "parameter" value associated with the fourth related property (prop4). Usage depends on the property function (See the "func" field on Properties.txt).
	Min4     int    `csv:"min4"`        // The "min" value to assign to the fourth related property (prop4). Usage depends on the property function (See the "func" field on Properties.txt).
	Max4     int    `csv:"max4"`        // The "max" value to assign to the fourth related property (prop4). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop5    string `csv:"prop5"`       // Defines the fifth Property to apply to the monster (Uses the "code" field from Properties.txt).
	Chance5  int    `csv:"chance5"`     // The percent chance that the fifth related property (prop5) will be assigned. If this value equals 0, then the Property will always be applied.
	Param5   string `csv:"par5"`        // The "parameter" value associated with the fifth related property (prop5). Usage depends on the property function (See the "func" field on Properties.txt).
	Min5     int    `csv:"min5"`        // The "min" value to assign to the fifth related property (prop5). Usage depends on the property function (See the "func" field on Properties.txt).
	Max5     int    `csv:"max5"`        // The "max" value to assign to the fifth related property (prop5). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop6    string `csv:"prop6"`       // Defines the sixth Property to apply to the monster (Uses the "code" field from Properties.txt).
	Chance6  int    `csv:"chance6"`     // The percent chance that the sixth related property (prop6) will be assigned. If this value equals 0, then the Property will always be applied.
	Param6   string `csv:"par6"`        // The "parameter" value associated with the sixth related property (prop6). Usage depends on the property function (See the "func" field on Properties.txt).
	Min6     int    `csv:"min6"`        // The "min" value to assign to the sixth related property (prop6). Usage depends on the property function (See the "func" field on Properties.txt).
	Max6     int    `csv:"max6"`        // The "max" value to assign to the sixth related property (prop6). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop1N   string `csv:"prop1 (N)"`   // Defines the first Property for Nightmare difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance1N int    `csv:"chance1 (N)"` // The percent chance that the first related property for Nightmare difficulty (prop1 (N)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param1N  string `csv:"par1 (N)"`    // The "parameter" value associated with the first related property for Nightmare difficulty (prop1 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min1N    int    `csv:"min1 (N)"`    // The "min" value to assign to the first related property for Nightmare difficulty (prop1 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max1N    int    `csv:"max1 (N)"`    // The "max" value to assign to the first related property for Nightmare difficulty (prop1 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop2N   string `csv:"prop2 (N)"`   // Defines the second Property for Nightmare difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance2N int    `csv:"chance2 (N)"` // The percent chance that the second related property for Nightmare difficulty (prop2 (N)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param2N  string `csv:"par2 (N)"`    // The "parameter" value associated with the second related property for Nightmare difficulty (prop2 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min2N    int    `csv:"min2 (N)"`    // The "min" value to assign to the second related property for Nightmare difficulty (prop2 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max2N    int    `csv:"max2 (N)"`    // The "max" value to assign to the second related property for Nightmare difficulty (prop2 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop3N   string `csv:"prop3 (N)"`   // Defines the third Property for Nightmare difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance3N int    `csv:"chance3 (N)"` // The percent chance that the third related property for Nightmare difficulty (prop3 (N)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param3N  string `csv:"par3 (N)"`    // The "parameter" value associated with the third related property for Nightmare difficulty (prop3 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min3N    int    `csv:"min3 (N)"`    // The "min" value to assign to the third related property for Nightmare difficulty (prop3 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max3N    int    `csv:"max3 (N)"`    // The "max" value to assign to the third related property for Nightmare difficulty (prop3 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop4N   string `csv:"prop4 (N)"`   // Defines the fourth Property for Nightmare difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance4N int    `csv:"chance4 (N)"` // The percent chance that the fourth related property for Nightmare difficulty (prop4 (N)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param4N  string `csv:"par4 (N)"`    // The "parameter" value associated with the fourth related property for Nightmare difficulty (prop4 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min4N    int    `csv:"min4 (N)"`    // The "min" value to assign to the fourth related property for Nightmare difficulty (prop4 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max4N    int    `csv:"max4 (N)"`    // The "max" value to assign to the fourth related property for Nightmare difficulty (prop4 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop5N   string `csv:"prop5 (N)"`   // Defines the fifth Property for Nightmare difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance5N int    `csv:"chance5 (N)"` // The percent chance that the fifth related property for Nightmare difficulty (prop5 (N)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param5N  string `csv:"par5 (N)"`    // The "parameter" value associated with the fifth related property for Nightmare difficulty (prop5 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min5N    int    `csv:"min5 (N)"`    // The "min" value to assign to the fifth related property for Nightmare difficulty (prop5 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max5N    int    `csv:"max5 (N)"`    // The "max" value to assign to the fifth related property for Nightmare difficulty (prop5 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop6N   string `csv:"prop6 (N)"`   // Defines the sixth Property for Nightmare difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance6N int    `csv:"chance6 (N)"` // The percent chance that the sixth related property for Nightmare difficulty (prop6 (N)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param6N  string `csv:"par6 (N)"`    // The "parameter" value associated with the sixth related property for Nightmare difficulty (prop6 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min6N    int    `csv:"min6 (N)"`    // The "min" value to assign to the sixth related property for Nightmare difficulty (prop6 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max6N    int    `csv:"max6 (N)"`    // The "max" value to assign to the sixth related property for Nightmare difficulty (prop6 (N)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop1H   string `csv:"prop1 (H)"`   // Defines the first Property for Hell difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance1H int    `csv:"chance1 (H)"` // The percent chance that the first related property for Hell difficulty (prop1 (H)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param1H  string `csv:"par1 (H)"`    // The "parameter" value associated with the first related property for Hell difficulty (prop1 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min1H    int    `csv:"min1 (H)"`    // The "min" value to assign to the first related property for Hell difficulty (prop1 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max1H    int    `csv:"max1 (H)"`    // The "max" value to assign to the first related property for Hell difficulty (prop1 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop2H   string `csv:"prop2 (H)"`   // Defines the second Property for Hell difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance2H int    `csv:"chance2 (H)"` // The percent chance that the second related property for Hell difficulty (prop2 (H)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param2H  string `csv:"par2 (H)"`    // The "parameter" value associated with the second related property for Hell difficulty (prop2 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min2H    int    `csv:"min2 (H)"`    // The "min" value to assign to the second related property for Hell difficulty (prop2 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max2H    int    `csv:"max2 (H)"`    // The "max" value to assign to the second related property for Hell difficulty (prop2 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop3H   string `csv:"prop3 (H)"`   // Defines the third Property for Hell difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance3H int    `csv:"chance3 (H)"` // The percent chance that the third related property for Hell difficulty (prop3 (H)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param3H  string `csv:"par3 (H)"`    // The "parameter" value associated with the third related property for Hell difficulty (prop3 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min3H    int    `csv:"min3 (H)"`    // The "min" value to assign to the third related property for Hell difficulty (prop3 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max3H    int    `csv:"max3 (H)"`    // The "max" value to assign to the third related property for Hell difficulty (prop3 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop4H   string `csv:"prop4 (H)"`   // Defines the fourth Property for Hell difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance4H int    `csv:"chance4 (H)"` // The percent chance that the fourth related property for Hell difficulty (prop4 (H)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param4H  string `csv:"par4 (H)"`    // The "parameter" value associated with the fourth related property for Hell difficulty (prop4 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min4H    int    `csv:"min4 (H)"`    // The "min" value to assign to the fourth related property for Hell difficulty (prop4 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max4H    int    `csv:"max4 (H)"`    // The "max" value to assign to the fourth related property for Hell difficulty (prop4 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop5H   string `csv:"prop5 (H)"`   // Defines the fifth Property for Hell difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance5H int    `csv:"chance5 (H)"` // The percent chance that the fifth related property for Hell difficulty (prop5 (H)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param5H  string `csv:"par5 (H)"`    // The "parameter" value associated with the fifth related property for Hell difficulty (prop5 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min5H    int    `csv:"min5 (H)"`    // The "min" value to assign to the fifth related property for Hell difficulty (prop5 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max5H    int    `csv:"max5 (H)"`    // The "max" value to assign to the fifth related property for Hell difficulty (prop5 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Prop6H   string `csv:"prop6 (H)"`   // Defines the sixth Property for Hell difficulty to apply to the monster (Uses the "code" field from Properties.txt).
	Chance6H int    `csv:"chance6 (H)"` // The percent chance that the sixth related property for Hell difficulty (prop6 (H)) will be assigned. If this value equals 0, then the Property will always be applied.
	Param6H  string `csv:"par6 (H)"`    // The "parameter" value associated with the sixth related property for Hell difficulty (prop6 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Min6H    int    `csv:"min6 (H)"`    // The "min" value to assign to the sixth related property for Hell difficulty (prop6 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
	Max6H    int    `csv:"max6 (H)"`    // The "max" value to assign to the sixth related property for Hell difficulty (prop6 (H)). Usage depends on the property function (See the "func" field on Properties.txt).
}

// this might be for the d2 resurrected version...
//// MonsterProp represents special properties that can be added to a monster based on the game difficulty.
//type MonsterProp struct {
//	Id      string `csv:"Id"`      // Defines the monster that should gain the Property (points to the matching "Id" value in the monstats.txt file).
//	Prop1   string `csv:"prop1"`   // Defines the first Property to apply to the monster (Uses the "code" field from Properties.txt).
//	Chance1 int    `csv:"chance1"` // The percent chance that the first related property (prop1) will be assigned. If this value equals 0, then the Property will always be applied.
//	Param1  string `csv:"par1"`    // The "parameter" value associated with the first related property (prop1). Usage depends on the property function (See the "func" field on Properties.txt).
//	Min1    int    `csv:"min1"`    // The "min" value to assign to the first related property (prop1). Usage depends on the property function (See the "func" field on Properties.txt).
//	Max1    int    `csv:"max1"`    // The "max" value to assign to the first related property (prop1). Usage depends on the property function (See the "func" field on Properties.txt).
//	Prop2   string `csv:"prop2"`   // Defines the second Property to apply to the monster (Uses the "code" field from Properties.txt).
//	Chance2 int    `csv:"chance2"` // The percent chance that the second related property (prop2) will be assigned. If this value equals 0, then the Property will always be applied.
//	Param2  string `csv:"par2"`    // The "parameter" value associated with the second related property (prop2). Usage depends on the property function (See the "func" field on Properties.txt).
//	Min2    int    `csv:"min2"`    // The "min" value to assign to the second related property (prop2). Usage depends on the property function (See the "func" field on Properties.txt).
//	Max2    int    `csv:"max2"`    // The "max" value to assign to the second related property (prop2). Usage depends on the property function (See the "func" field on Properties.txt).
//	Prop3   string `csv:"prop3"`   // Defines the third Property to apply to the monster (Uses the "code" field from Properties.txt).
//	Chance3 int    `csv:"chance3"` // The percent chance that the third related property (prop3) will be assigned. If this value equals 0, then the Property will always be applied.
//	Param3  string `csv:"par3"`    // The "parameter" value associated with the third related property (prop3). Usage depends on the property function (See the "func" field on Properties.txt).
//	Min3    int    `csv:"min3"`    // The "min" value to assign to the third related property (prop3). Usage depends on the property function (See the "func" field on Properties.txt).
//	Max3    int    `csv:"max3"`    // The "max" value to assign to the third related property (prop3). Usage depends on the property function (See the "func" field on Properties.txt).
//	Prop4   string `csv:"prop4"`   // Defines the fourth Property to apply to the monster (Uses the "code" field from Properties.txt).
//	Chance4 int    `csv:"chance4"` // The percent chance that the fourth related property (prop4) will be assigned. If this value equals 0, then the Property will always be applied.
//	Param4  string `csv:"par4"`    // The "parameter" value associated with the fourth related property (prop4). Usage depends on the property function (See the "func" field on Properties.txt).
//	Min4    int    `csv:"min4"`    // The "min" value to assign to the fourth related property (prop4). Usage depends on the property function (See the "func" field on Properties.txt).
//	Max4    int    `csv:"max4"`    // The "max" value to assign to the fourth related property (prop4). Usage depends on the property function (See the "func" field on Properties.txt).
//	Prop5   string `csv:"prop5"`   // Defines the fifth Property to apply to the monster (Uses the "code" field from Properties.txt).
//	Chance5 int    `csv:"chance5"` // The percent chance that the fifth related property (prop5) will be assigned. If this value equals 0, then the Property will always be applied.
//	Param5  string `csv:"par5"`    // The "parameter" value associated with the fifth related property (prop5). Usage depends on the property function (See the "func" field on Properties.txt).
//	Min5    int    `csv:"min5"`    // The "min" value to assign to the fifth related property (prop5). Usage depends on the property function (See the "func" field on Properties.txt).
//	Max5    int    `csv:"max5"`    // The "max" value to assign to the fifth related property (prop5). Usage depends on the property function (See the "func" field on Properties.txt).
//	Prop6   string `csv:"prop6"`   // Defines the sixth Property to apply to the monster (Uses the "code" field from Properties.txt).
//	Chance6 int    `csv:"chance6"` // The percent chance that the sixth related property (prop6) will be assigned. If this value equals 0, then the Property will always be applied.
//	Param6  string `csv:"par6"`    // The "parameter" value associated with the sixth related property (prop6). Usage depends on the property function (See the "func" field on Properties.txt).
//	Min6    int    `csv:"min6"`    // The "min" value to assign to the sixth related property (prop6). Usage depends on the property function (See the "func" field on Properties.txt).
//	Max6    int    `csv:"max6"`    // The "max" value to assign to the sixth related property (prop6). Usage depends on the property function (See the "func" field on Properties.txt).
//}
