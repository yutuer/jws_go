package hero_diff

import (
	"math/rand"
	"vcs.taiyouxi.net/jws/gamex/models/gamedata"
)

func GetStageIDSeq(lastMaxIndex int) ([]int, int) {
	stageID := make([]int, gamedata.GetHeroDiffStageCount())
	for i := 0; i < len(stageID); i++ {
		stageID[i] = i + 1
	}
	stageCount, newMaxIndex := getStageID(lastMaxIndex)
	seqID := getRandomStageSeq(stageID, stageCount)
	return seqID, newMaxIndex
}

func getStageID(lastMaxIndex int) ([]int, int) {
	stageCount := make([]int, gamedata.GetHeroDiffStageCount())

	count := gamedata.GetHeroDiffTypeCount()
	for i := 0; i < len(stageCount) && i < len(count); i++ {
		stageCount[i] = count[i]
	}
	return getRandomStage(stageCount, lastMaxIndex)
}

func getRandomStageSeq(stageID, stageCount []int) []int {
	seqIDCount := 0
	for _, count := range stageCount {
		seqIDCount += count
	}
	seqID := make([]int, seqIDCount)
	index := 0
	for i, id := range stageID {
		for j := stageCount[i]; j > 0; j-- {
			seqID[index] = id
			index++
		}
	}

	for i := 0; i < seqIDCount; i++ {
		a := rand.Intn(seqIDCount)
		b := rand.Intn(seqIDCount)
		seqID[a], seqID[b] = seqID[b], seqID[a]
	}
	return seqID
}

func getRandomStage(count []int, lastMaxIndex int) ([]int, int) {
	if len(count) < 2 || lastMaxIndex < 0 {
		return count, lastMaxIndex
	}
	// random
	for i := 0; i < len(count); i++ {
		a := rand.Intn(len(count))
		b := rand.Intn(len(count))
		count[a], count[b] = count[b], count[a]
	}

	maxIndex := 0
	for i := 1; i < len(count); i++ {
		if count[i] > count[maxIndex] {
			maxIndex = i
		}
	}
	if maxIndex == lastMaxIndex {
		if maxIndex > 0 {
			count[0], count[maxIndex] = count[maxIndex], count[0]
			maxIndex = 0
		} else {
			count[0], count[1] = count[1], count[0]
			maxIndex = 1
		}
	}
	return count, maxIndex
}

func HeroDiffID2Index(id int) int {
	return id - 1
}
