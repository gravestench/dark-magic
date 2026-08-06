package models

// MonsterSounds represents the data for each monster's sounds
type MonsterSounds struct {
	ID                  string `csv:"Id"`            // Defines the unique name ID for the monster sound.
	Attack1             string `csv:"Attack1"`       // Play this sound when the monster performs Attack 1. Points to a "Sound" value in the sounds.txt file.
	Attack2             string `csv:"Attack2"`       // Play this sound when the monster performs Attack 2. Points to a "Sound" value in the sounds.txt file.
	Weapon1             string `csv:"Weapon1"`       // Play this sound when the monster performs Attack 1. Points to a "Sound" value in the sounds.txt file.
	Weapon2             string `csv:"Weapon2"`       // Play this sound when the monster performs Attack 2. Points to a "Sound" value in the sounds.txt file.
	Att1Delay           int    `csv:"Att1Del"`       // Controls the amount of game frames to delay playing the "Attack1" sound.
	Att2Delay           int    `csv:"Att2Del"`       // Controls the amount of game frames to delay playing the "Attack2" sound.
	Wea1Delay           int    `csv:"Wea1Del"`       // Controls the amount of game frames to delay playing the "Weapon1" sound.
	Wea2Delay           int    `csv:"Wea2Del"`       // Controls the amount of game frames to delay playing the "Weapon2" sound.
	Att1Probability     int    `csv:"Att1Prb"`       // Controls the percent chance (out of 100) to play the "Attack1" sound.
	Att2Probability     int    `csv:"Att2Prb"`       // Controls the percent chance (out of 100) to play the "Attack2" sound.
	Wea1Volume          int    `csv:"Wea1Vol"`       // Controls the volume of the "Weapon1" sound. Uses a range between 0 to 255, where 255 is the maximum volume.
	Wea2Volume          int    `csv:"Wea2Vol"`       // Controls the volume of the "Weapon2" sound. Uses a range between 0 to 255, where 255 is the maximum volume.
	HitSound            string `csv:"HitSound"`      // Play this sound when the monster gets hit or knocked back. Points to a "Sound" value in the sounds.txt file.
	DeathSound          string `csv:"DeathSound"`    // Play this sound when the monster dies. Points to a "Sound" value in the sounds.txt file.
	HitDelay            int    `csv:"HitDelay"`      // Controls the amount of game frames to delay playing the "HitSound" sound.
	DeathDelay          int    `csv:"DeaDelay"`      // Controls the amount of game frames to delay playing the "DeathSound" sound.
	Skill1              string `csv:"Skill1"`        // Play this sound when the monster uses Skill 1. Points to a "Sound" value in the sounds.txt file.
	Skill2              string `csv:"Skill2"`        // Play this sound when the monster uses Skill 2. Points to a "Sound" value in the sounds.txt file.
	Skill3              string `csv:"Skill3"`        // Play this sound when the monster uses Skill 3. Points to a "Sound" value in the sounds.txt file.
	Skill4              string `csv:"Skill4"`        // Play this sound when the monster uses Skill 4. Points to a "Sound" value in the sounds.txt file.
	Footstep            string `csv:"Footstep"`      // Play this sound while the monster is walking or running. Points to a "Sound" value in the sounds.txt file.
	FootstepLayer       string `csv:"FootstepLayer"` // Play this sound while the monster is walking or running. Points to a "Sound" value in the sounds.txt file.
	FootstepCount       int    `csv:"FsCnt"`         // Controls the footstep count which is used to determine how often to play the "Footstep" and "FootstepLayer" sound.
	FootstepOffset      int    `csv:"FsOff"`         // Controls the footstep offset which is used for calculating when to play the next "Footstep" and "FootstepLayer" sound, based on the current animation frame and the animation rate.
	FootstepProbability int    `csv:"FsPrb"`         // Controls the probability to play the "Footstep" and "FootstepLayer" sound, with a random chance out of 100.
	NeutralSound        string `csv:"Neutral"`       // Play this sound while the monster is in Neutral, Walk, or Run mode. Points to a "Sound" value in the sounds.txt file.
	NeutralTime         int    `csv:"NeuTime"`       // Controls the amount of game frames to delay between re-playing the "Neutral" sound after it finishes.
	InitSound           string `csv:"Init"`          // Play this sound when the monster spawns and is not dead and is not playing its Neutral sound. Points to a "Sound" value in the sounds.txt file.
	TauntSound          string `csv:"Taunt"`         // Play this sound when the server requests that the monster should play its Taunt. Points to a "Sound" value in the sounds.txt file.
	FleeSound           string `csv:"Flee"`          // Play this sound when the monster is told to flee. Points to a "Sound" value in the sounds.txt file.
	ConvertMode1        string `csv:"CvtMo1"`        // This is used to convert the mode for playing the sound. Defines the original mode that the monster is using.
	ConvertMode2        string `csv:"CvtMo2"`        // This is used to convert the mode for playing the sound. Defines the original mode that the monster is using.
	ConvertMode3        string `csv:"CvtMo3"`        // This is used to convert the mode for playing the sound. Defines the original mode that the monster is using.
	ConvertSkill1       string `csv:"CvtSk1"`        // Defines the skill that the monster is using. If the monster uses a specific skill, then the game can change the monster's mode for sound functionalities to another mode to change how sounds are generally handled. Points to a "Skill" in the skills.txt file.
	ConvertSkill2       string `csv:"CvtSk2"`        // Defines the skill that the monster is using. If the monster uses a specific skill, then the game can change the monster's mode for sound functionalities to another mode to change how sounds are generally handled. Points to a "Skill" in the skills.txt file.
	ConvertSkill3       string `csv:"CvtSk3"`        // Defines the skill that the monster is using. If the monster uses a specific skill, then the game can change the monster's mode for sound functionalities to another mode to change how sounds are generally handled. Points to a "Skill" in the skills.txt file.
	ConvertTarget1      string `csv:"CvtTgt1"`       // Defines the mode to convert the sound to when the monster is using the relative skill from the "ConvertSkill1" field.
	ConvertTarget2      string `csv:"CvtTgt2"`       // Defines the mode to convert the sound to when the monster is using the relative skill from the "ConvertSkill2" field.
	ConvertTarget3      string `csv:"CvtTgt3"`       // Defines the mode to convert the sound to when the monster is using the relative skill from the "ConvertSkill3" field.
}
