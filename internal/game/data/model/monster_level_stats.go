package models

// MonsterLevelStats is one integer MonLvl.txt baseline row. MonStats supplies
// percentage ratios for ordinary monsters; the two legacy tables were split by
// the old Excel column limit and are joined by the monster domain.
type MonsterLevelStats struct {
	Level int `csv:"Level"`

	Defense       int `csv:"AC"`
	DefenseN      int `csv:"AC(N)"`
	DefenseH      int `csv:"AC(H)"`
	AttackRating  int `csv:"TH"`
	AttackRatingN int `csv:"TH(N)"`
	AttackRatingH int `csv:"TH(H)"`
	Life          int `csv:"HP"`
	LifeN         int `csv:"HP(N)"`
	LifeH         int `csv:"HP(H)"`
	Damage        int `csv:"DM"`
	DamageN       int `csv:"DM(N)"`
	DamageH       int `csv:"DM(H)"`
	Experience    int `csv:"XP"`
	ExperienceN   int `csv:"XP(N)"`
	ExperienceH   int `csv:"XP(H)"`

	LadderDefense     int `csv:"L-AC"`
	LadderDefenseN    int `csv:"L-AC(N)"`
	LadderDefenseH    int `csv:"L-AC(H)"`
	LadderLife        int `csv:"L-HP"`
	LadderLifeN       int `csv:"L-HP(N)"`
	LadderLifeH       int `csv:"L-HP(H)"`
	LadderDamage      int `csv:"L-DM"`
	LadderDamageN     int `csv:"L-DM(N)"`
	LadderDamageH     int `csv:"L-DM(H)"`
	LadderExperience  int `csv:"L-XP"`
	LadderExperienceN int `csv:"L-XP(N)"`
	LadderExperienceH int `csv:"L-XP(H)"`
}
