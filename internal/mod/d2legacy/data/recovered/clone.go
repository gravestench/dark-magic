package recovered

// clone deep-copies every mutable slice and lookup map in a snapshot. Callers
// may therefore edit returned generations without corrupting cached provenance.
func clone(source Snapshot) Snapshot {
	result := Snapshot{
		Quests:           make([]Quest, len(source.Quests)),
		QuestsByID:       make(map[int]Quest, len(source.QuestsByID)),
		Speech:           append([]Speech(nil), source.Speech...),
		SpeechByName:     make(map[string]Speech, len(source.SpeechByName)),
		DS1Types:         append([]DS1Type(nil), source.DS1Types...),
		DS1TypeByDef:     make(map[int]DS1Type, len(source.DS1TypeByDef)),
		MapObjects:       append([]MapObject(nil), source.MapObjects...),
		MapObjectByActID: make(map[string]MapObject, len(source.MapObjectByActID)),
	}

	for index, quest := range source.Quests {
		quest.Stages = append([]QuestStage(nil), quest.Stages...)
		for stage := range quest.Stages {
			quest.Stages[stage].Alternates = append(
				[]string(nil),
				quest.Stages[stage].Alternates...,
			)
		}

		result.Quests[index] = quest
		result.QuestsByID[quest.ID] = quest
	}

	for key, value := range source.SpeechByName {
		result.SpeechByName[key] = value
	}

	for key, value := range source.DS1TypeByDef {
		result.DS1TypeByDef[key] = value
	}

	for key, value := range source.MapObjectByActID {
		result.MapObjectByActID[key] = value
	}

	return result
}
