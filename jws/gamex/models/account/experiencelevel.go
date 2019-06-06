package account

type ExperienceLevel struct {
	LevelId []string
}

func (el *ExperienceLevel) AddExperiendeLevel(levelId string) {
	el.LevelId = append(el.LevelId, levelId)
}

func (el *ExperienceLevel) GetExperiendeLevel() []string {
	return el.LevelId
}
