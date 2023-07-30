package models

// SkillDescData represents a skill's tooltip description and how it is displayed on the Skill Tree.
type SkillDescData struct {
	SkillDesc        string `csv:"skilldesc"`                // Skill description name.
	SkillPage        int    `csv:"SkillPage"`                // Skill tree page.
	SkillRow         int    `csv:"SkillRow"`                 // Skill tree row.
	SkillColumn      int    `csv:"SkillColumn"`              // Skill tree column.
	ListRow          int    `csv:"ListRow"`                  // Skill listing row in the skill selection UI.
	IconCel          string `csv:"IconCel"`                  // Icon asset for displaying the skill.
	HireableIconCel  string `csv:"HireableIconCel"`          // Icon asset for displaying the skill on a hireable (mercenary).
	StrName          string `csv:"str name"`                 // Skill name.
	StrShort         string `csv:"str short"`                // Short skill description.
	StrLong          string `csv:"str long"`                 // Long skill description on the Skill Tree.
	StrAlt           string `csv:"str alt"`                  // Skill name on the Character Screen when selected.
	DescDam          string `csv:"descdam"`                  // Skill's damage calculation function.
	DDamCalc1        string `csv:"ddam calc1"`               // Integer calc value used as a parameter for DescDam function.
	DDamCalc2        string `csv:"ddam calc2"`               // Integer calc value used as a parameter for DescDam function.
	P1DmElem         string `csv:"p1dmelem"`                 // Elemental type for charge-ups on the Character Screen.
	P2DmElem         string `csv:"p2dmelem"`                 // Elemental type for charge-ups on the Character Screen.
	P3DmElem         string `csv:"p3dmelem"`                 // Elemental type for charge-ups on the Character Screen.
	P1DmMin          string `csv:"p1dmmin"`                  // Minimum damage for charge-ups on the Character Screen.
	P2DmMin          string `csv:"p2dmmin"`                  // Minimum damage for charge-ups on the Character Screen.
	P3DmMin          string `csv:"p3dmmin"`                  // Minimum damage for charge-ups on the Character Screen.
	P1DmMax          string `csv:"p1dmmax"`                  // Maximum damage for charge-ups on the Character Screen.
	P2DmMax          string `csv:"p2dmmax"`                  // Maximum damage for charge-ups on the Character Screen.
	P3DmMax          string `csv:"p3dmmax"`                  // Maximum damage for charge-ups on the Character Screen.
	DescAtt          string `csv:"descatt"`                  // Overall Attack Rating of the skill in the Character Screen.
	DescMissile1     string `csv:"descmissile1"`             // Linked missile from Missiles.txt for reference.
	DescMissile2     string `csv:"descmissile2"`             // Linked missile from Missiles.txt for reference.
	DescMissile3     string `csv:"descmissile3"`             // Linked missile from Missiles.txt for reference.
	DescLine1        string `csv:"descline1"`                // Description function ID for level and next level tooltip lines.
	DescLine2        string `csv:"descline2"`                // Description function ID for level and next level tooltip lines.
	DescLine3        string `csv:"descline3"`                // Description function ID for level and next level tooltip lines.
	DescLine4        string `csv:"descline4"`                // Description function ID for level and next level tooltip lines.
	DescLine5        string `csv:"descline5"`                // Description function ID for level and next level tooltip lines.
	DescLine6        string `csv:"descline6"`                // Description function ID for level and next level tooltip lines.
	DescTextA1       string `csv:"desctexta1"`               // First string parameter for DescLine function.
	DescTextA2       string `csv:"desctexta2"`               // First string parameter for DescLine function.
	DescTextA3       string `csv:"desctexta3"`               // First string parameter for DescLine function.
	DescTextA4       string `csv:"desctexta4"`               // First string parameter for DescLine function.
	DescTextA5       string `csv:"desctexta5"`               // First string parameter for DescLine function.
	DescTextA6       string `csv:"desctexta6"`               // First string parameter for DescLine function.
	DescTextB1       string `csv:"desctextb1"`               // Second string parameter for DescLine function.
	DescTextB2       string `csv:"desctextb2"`               // Second string parameter for DescLine function.
	DescTextB3       string `csv:"desctextb3"`               // Second string parameter for DescLine function.
	DescTextB4       string `csv:"desctextb4"`               // Second string parameter for DescLine function.
	DescTextB5       string `csv:"desctextb5"`               // Second string parameter for DescLine function.
	DescTextB6       string `csv:"desctextb6"`               // Second string parameter for DescLine function.
	DescCalca1       string `csv:"desccalca1"`               // First numeric parameter for DescLine function.
	DescCalca2       string `csv:"desccalca2"`               // First numeric parameter for DescLine function.
	DescCalca3       string `csv:"desccalca3"`               // First numeric parameter for DescLine function.
	DescCalca4       string `csv:"desccalca4"`               // First numeric parameter for DescLine function.
	DescCalca5       string `csv:"desccalca5"`               // First numeric parameter for DescLine function.
	DescCalca6       string `csv:"desccalca6"`               // First numeric parameter for DescLine function.
	DescCalcb1       string `csv:"desccalcb1"`               // Second numeric parameter for DescLine function.
	DescCalcb2       string `csv:"desccalcb2"`               // Second numeric parameter for DescLine function.
	DescCalcb3       string `csv:"desccalcb3"`               // Second numeric parameter for DescLine function.
	DescCalcb4       string `csv:"desccalcb4"`               // Second numeric parameter for DescLine function.
	DescCalcb5       string `csv:"desccalcb5"`               // Second numeric parameter for DescLine function.
	DescCalcb6       string `csv:"desccalcb6"`               // Second numeric parameter for DescLine function.
	Dsc2Line1        string `csv:"dsc2line1"`                // Description function ID for pinned lines after skill description.
	Dsc2Line2        string `csv:"dsc2line2"`                // Description function ID for pinned lines after skill description.
	Dsc2Line3        string `csv:"dsc2line3"`                // Description function ID for pinned lines after skill description.
	Dsc2Line4        string `csv:"dsc2line4"`                // Description function ID for pinned lines after skill description.
	Dsc2Line5        string `csv:"dsc2line5"`                // Description function ID for pinned lines after skill description.
	Dsc2TextA1       string `csv:"dsc2texta1"`               // First string parameter for Dsc2Line function.
	Dsc2TextA2       string `csv:"dsc2texta2"`               // First string parameter for Dsc2Line function.
	Dsc2TextA3       string `csv:"dsc2texta3"`               // First string parameter for Dsc2Line function.
	Dsc2TextA4       string `csv:"dsc2texta4"`               // First string parameter for Dsc2Line function.
	Dsc2TextA5       string `csv:"dsc2texta5"`               // First string parameter for Dsc2Line function.
	Dsc2TextB1       string `csv:"dsc2textb1"`               // Second string parameter for Dsc2Line function.
	Dsc2TextB2       string `csv:"dsc2textb2"`               // Second string parameter for Dsc2Line function.
	Dsc2TextB3       string `csv:"dsc2textb3"`               // Second string parameter for Dsc2Line function.
	Dsc2TextB4       string `csv:"dsc2textb4"`               // Second string parameter for Dsc2Line function.
	Dsc2TextB5       string `csv:"dsc2textb5"`               // Second string parameter for Dsc2Line function.
	Dsc2Calca1       string `csv:"dsc2calca1"`               // First numeric parameter for Dsc2Line function.
	Dsc2Calca2       string `csv:"dsc2calca2"`               // First numeric parameter for Dsc2Line function.
	Dsc2Calca3       string `csv:"dsc2calca3"`               // First numeric parameter for Dsc2Line function.
	Dsc2Calca4       string `csv:"dsc2calca4"`               // First numeric parameter for Dsc2Line function.
	Dsc2Calca5       string `csv:"dsc2calca5"`               // First numeric parameter for Dsc2Line function.
	Dsc2Calca6       string `csv:"dsc2calca6"`               // First numeric parameter for Dsc2Line function.
	Dsc2Calcb1       string `csv:"dsc2calcb1"`               // Second numeric parameter for Dsc2Line function.
	Dsc2Calcb2       string `csv:"dsc2calcb2"`               // Second numeric parameter for Dsc2Line function.
	Dsc2Calcb3       string `csv:"dsc2calcb3"`               // Second numeric parameter for Dsc2Line function.
	Dsc2Calcb4       string `csv:"dsc2calcb4"`               // Second numeric parameter for Dsc2Line function.
	Dsc2Calcb5       string `csv:"dsc2calcb5"`               // Second numeric parameter for Dsc2Line function.
	Dsc2Calcb6       string `csv:"dsc2calcb6"`               // Second numeric parameter for Dsc2Line function.
	Dsc3Line1        string `csv:"dsc3line1"`                // Description function ID for pinned lines at the bottom of the skill tooltip.
	Dsc3Line2        string `csv:"dsc3line2"`                // Description function ID for pinned lines at the bottom of the skill tooltip.
	Dsc3Line3        string `csv:"dsc3line3"`                // Description function ID for pinned lines at the bottom of the skill tooltip.
	Dsc3Line4        string `csv:"dsc3line4"`                // Description function ID for pinned lines at the bottom of the skill tooltip.
	Dsc3Line5        string `csv:"dsc3line5"`                // Description function ID for pinned lines at the bottom of the skill tooltip.
	Dsc3Line6        string `csv:"dsc3line6"`                // Description function ID for pinned lines at the bottom of the skill tooltip.
	Dsc3Line7        string `csv:"dsc3line7"`                // Description function ID for pinned lines at the bottom of the skill tooltip.
	Dsc3TextA1       string `csv:"dsc3texta1"`               // First string parameter for Dsc3Line function.
	Dsc3TextA2       string `csv:"dsc3texta2"`               // First string parameter for Dsc3Line function.
	Dsc3TextA3       string `csv:"dsc3texta3"`               // First string parameter for Dsc3Line function.
	Dsc3TextA4       string `csv:"dsc3texta4"`               // First string parameter for Dsc3Line function.
	Dsc3TextA5       string `csv:"dsc3texta5"`               // First string parameter for Dsc3Line function.
	Dsc3TextA6       string `csv:"dsc3texta6"`               // First string parameter for Dsc3Line function.
	Dsc3TextA7       string `csv:"dsc3texta7"`               // First string parameter for Dsc3Line function.
	Dsc3TextB1       string `csv:"dsc3textb1"`               // Second string parameter for Dsc3Line function.
	Dsc3TextB2       string `csv:"dsc3textb2"`               // Second string parameter for Dsc3Line function.
	Dsc3TextB3       string `csv:"dsc3textb3"`               // Second string parameter for Dsc3Line function.
	Dsc3TextB4       string `csv:"dsc3textb4"`               // Second string parameter for Dsc3Line function.
	Dsc3TextB5       string `csv:"dsc3textb5"`               // Second string parameter for Dsc3Line function.
	Dsc3TextB6       string `csv:"dsc3textb6"`               // Second string parameter for Dsc3Line function.
	Dsc3TextB7       string `csv:"dsc3textb7"`               // Second string parameter for Dsc3Line function.
	ItemProcText     string `csv:"item proc text"`           // Override format for skill appearing as "chance to cast" on an item.
	ItemProcDescLine string `csv:"item proc descline count"` // Number of descline entries formatted into ItemProcText string.
}
