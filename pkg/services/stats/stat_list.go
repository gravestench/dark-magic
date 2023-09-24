package stats

// StatList is useful for reducing stats.
// They provide a context for stats to alter other stats or infer values
// during stat assignment/calculation
type StatList struct {
	stats []Stat
}

func (s *StatList) Index(idx int) Stat {
	//TODO implement me
	panic("implement me")
}

func (s *StatList) Stats() []Stat {
	//TODO implement me
	panic("implement me")
}

func (s *StatList) SetStats(stats []Stat) StatList {
	//TODO implement me
	panic("implement me")
}

func (s *StatList) Clone() StatList {
	//TODO implement me
	panic("implement me")
}

func (s *StatList) ReduceStats() StatList {
	//TODO implement me
	panic("implement me")
}

func (s *StatList) RemoveStatAtIndex(idx int) Stat {
	//TODO implement me
	panic("implement me")
}

func (s *StatList) AppendStatList(other StatList) StatList {
	//TODO implement me
	panic("implement me")
}

func (s *StatList) Pop() Stat {
	//TODO implement me
	panic("implement me")
}

func (s *StatList) Push(stat Stat) StatList {
	//TODO implement me
	panic("implement me")
}
