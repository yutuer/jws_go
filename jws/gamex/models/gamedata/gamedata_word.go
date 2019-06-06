package gamedata

import (
	"math/rand"
	"regexp"
	"strings"
	"time"

	"sort"

	"fmt"

	"github.com/golang/protobuf/proto"
	"vcs.taiyouxi.net/jws/gamex/protogen"
	"vcs.taiyouxi.net/jws/gamex/uutil"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const (
	//允许存在多个不连续空格
	sym_reg_vn_spaces = "\\s{2,}"
	sym_reg_vn        = "[^a-zA-Z0-9\\sĂÂÀẰẦÁẮẤẠẶẬÃẴẪẢẲẨÊÈỀẾÉẸỆỄẼẺỂÌÍỊĨỈÔƠÒỒỜÓỐỚỌỘỢÕỖÕỖỠỎỔỞƯÙỪÚỨỤỰŨỮỦỦỬỲÝỴỸỶăâàằầầáắấạặậãẵẫảẳẩêèềếéẹệễẽẻểìíịĩỉôơòồờóốớọộợõỗỡỏổởưùừúứụựũữủửỳýỵỹỷ]"
	sym_reg           = "[`~!@#\\$%\\^&\\*\\(\\)_\\-\\+=\\|\\\\{}\\[\\]\\:;\"'/\\?,<>？，。·！￥……（）+｛｝【】、|《》\\s]"
)

type INameRand interface {
	GetIndex() uint32
	GetHomeTown() string
	GetFamilyName() string
	GetFirstName() string
}
type NameRandInfo struct {
	HomeTownNameCount int32
	FamilyNameCount   int32
	FirstNameCount    int32
	NameItems         []INameRand
}

type NameRandForForeign struct {
	FamilyNameCounts []int
	FirstNameCounts  []int
	FamilyNameMap    map[int][]string
	FirstNameMap     map[int][]string
}

var (
	gdSenReg           *regexp.Regexp
	gdSymReg           *regexp.Regexp
	gdVnReg            *regexp.Regexp
	gdLangNameRandInfo = map[string]NameRandInfo{
		uutil.Lang_HANS: NameRandInfo{},
		uutil.Lang_HMT:  NameRandInfo{},
		uutil.Lang_EN:   NameRandInfo{},
		uutil.Lang_VN:   NameRandInfo{},
		uutil.Lang_KO:   NameRandInfo{},
		uutil.Lang_JA:   NameRandInfo{},
		uutil.Lang_TH:   NameRandInfo{},
	}
	EnRand NameRand
	VnRand NameRand
	KoRand NameRand
	ThRand NameRand
)

type NameRand interface {
	InitFromLoad(buffer []byte)
	Rand(count int) []string
}

type NameRandImpl struct {
	FamilyNameCounts []int
	FirstNameCounts  []int
	FamilyNameMap    map[int][]string
	FirstNameMap     map[int][]string
	MidName          string
	FamilyMaxLen     int
	AllMaxLen        int
}

type NameRandEn struct {
	NameRandImpl
}

type NameRandVn struct {
	NameRandImpl
}

type NameRandKo struct {
	NameRandImpl
}

type NameRandTh struct {
	NameRandImpl
}

func (nr *NameRandVn) Rand(count int) []string {
	return nr.randForeignNames(count)
}

func (nr *NameRandVn) InitFromLoad(buffer []byte) {
	dataListVn := &ProtobufGen.NAMEVI_ARRAY{}
	err := proto.Unmarshal(buffer, dataListVn)
	panicIfErr(err)
	nr.init()
	dataListVn.GetItems()
	for _, item := range dataListVn.GetItems() {
		nr.addNewNameForForeign(item.GetFamilyName(), item.GetFirstName())
	}
	nr.addLangForeignCount()
	nr.calcForeignCount()
	nr.MidName = " "
	nr.FamilyMaxLen = 9
	nr.AllMaxLen = 11
}

func (nr *NameRandTh) Rand(count int) []string {
	return nr.randForeignNames(count)
}

func (nr *NameRandTh) InitFromLoad(buffer []byte) {
	dataListTh := &ProtobufGen.NAMETH_ARRAY{}
	err := proto.Unmarshal(buffer, dataListTh)
	panicIfErr(err)
	nr.init()
	dataListTh.GetItems()
	for _, item := range dataListTh.GetItems() {
		nr.addNewNameForForeign(item.GetHomeTown(), item.GetFamilyName())
	}
	nr.addLangForeignCount()
	nr.calcForeignCount()
	nr.MidName = ""
	nr.FamilyMaxLen = 27
	nr.AllMaxLen = 33
}

func (nr *NameRandEn) Rand(count int) []string {
	return nr.randForeignNames(count)
}

func (nr *NameRandEn) InitFromLoad(buffer []byte) {
	dataListEn := &ProtobufGen.NAMEEN_ARRAY{}
	err := proto.Unmarshal(buffer, dataListEn)
	panicIfErr(err)
	nr.init()
	dataListEn.GetItems()
	for _, item := range dataListEn.GetItems() {
		nr.addNewNameForForeign(item.GetHomeTown(), item.GetFirstName())
	}
	nr.addLangForeignCount()
	nr.calcForeignCount()
	nr.MidName = "."
	nr.FamilyMaxLen = 9
	nr.AllMaxLen = 11
}

func (nr *NameRandKo) Rand(count int) []string {
	return nr.randForeignNames(count)
}

func (nr *NameRandKo) InitFromLoad(buffer []byte) {
	dataListKo := &ProtobufGen.NAMEKO_ARRAY{}
	err := proto.Unmarshal(buffer, dataListKo)
	panicIfErr(err)
	nr.init()
	dataListKo.GetItems()
	for _, item := range dataListKo.GetItems() {
		nr.addNewNameForForeign(item.GetFamilyName(), item.GetFirstName())
	}
	nr.addLangForeignCount()
	nr.calcForeignCount()
	nr.MidName = ""
	nr.FamilyMaxLen = 12
	nr.AllMaxLen = 18
}

func loadSensitiveWordConfig(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.ZHHANS_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	senWords := make([]string, 0, len(dataList.GetItems()))
	for _, item := range dataList.GetItems() {
		itemStr := strings.ToLower(item.GetWords())
		itemStr = strings.Replace(itemStr, "\\", "\\\\", -1)
		itemStr = strings.Replace(itemStr, "*", "\\*", -1)
		itemStr = strings.Replace(itemStr, ".", "\\.", -1)
		itemStr = strings.Replace(itemStr, "?", "\\?", -1)
		itemStr = strings.Replace(itemStr, "(", "\\(", -1)
		itemStr = strings.Replace(itemStr, ")", "\\)", -1)
		itemStr = strings.Replace(itemStr, "-", "\\-", -1)
		itemStr = strings.Replace(itemStr, "[", "\\[", -1)
		itemStr = strings.Replace(itemStr, "]", "\\]", -1)
		itemStr = strings.Replace(itemStr, "{", "\\{", -1)
		itemStr = strings.Replace(itemStr, "}", "\\}", -1)
		senWords = append(senWords, itemStr)
	}
	//logs.Debug("senWords %v", senWords)
	if len(senWords) != 0 {
		reg := strings.Join(senWords, "|")
		gdSenReg, err = regexp.Compile(reg)
		if err != nil {
			logs.Error("load sensitive err %v", err)
			return
		}
	}

	if uutil.IsVNVer() {
		gdSymReg, err = regexp.Compile(sym_reg_vn)
		gdVnReg, err = regexp.Compile(sym_reg_vn_spaces)
		if err != nil {
			logs.Error("load VN symbol regexp err %v", err)
			return
		}
	} else {
		gdSymReg, err = regexp.Compile(sym_reg)
		if err != nil {
			logs.Error("load symbol regexp err %v", err)
			return
		}
	}

}

func loadNameZhans(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.NAMEZHHANS_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	info := gdLangNameRandInfo[uutil.Lang_HANS]
	info.NameItems = make([]INameRand, 0, len(dataList.GetItems()))
	for _, item := range dataList.GetItems() {
		if item.GetHomeTown() != "" {
			info.HomeTownNameCount++
		}
		if item.GetFamilyName() != "" {
			info.FamilyNameCount++
		}
		if item.GetFirstName() != "" {
			info.FirstNameCount++
		}
		info.NameItems = append(info.NameItems, item)
	}
	gdLangNameRandInfo[uutil.Lang_HANS] = info
}

func loadNameJapan(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.NAMEJA_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	info := gdLangNameRandInfo[uutil.Lang_JA]
	info.NameItems = make([]INameRand, 0, len(dataList.GetItems()))
	for _, item := range dataList.GetItems() {
		if item.GetHomeTown() != "" {
			info.HomeTownNameCount++
		}
		if item.GetFamilyName() != "" {
			info.FamilyNameCount++
		}
		if item.GetFirstName() != "" {
			info.FirstNameCount++
		}
		info.NameItems = append(info.NameItems, item)
	}
	gdLangNameRandInfo[uutil.Lang_JA] = info
}

func LoadNameHmt(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	dataList := &ProtobufGen.NAMEHMT_ARRAY{}
	err = proto.Unmarshal(buffer, dataList)
	errcheck(err)

	info := gdLangNameRandInfo[uutil.Lang_HMT]
	info.NameItems = make([]INameRand, 0, len(dataList.GetItems()))
	for _, item := range dataList.GetItems() {
		if item.GetHomeTown() != "" {
			info.HomeTownNameCount++
		}
		if item.GetFamilyName() != "" {
			info.FamilyNameCount++
		}
		if item.GetFirstName() != "" {
			info.FirstNameCount++
		}
		info.NameItems = append(info.NameItems, item)
	}
	gdLangNameRandInfo[uutil.Lang_HMT] = info
}

func LoadNameEN(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	EnRand = &NameRandEn{}
	EnRand.InitFromLoad(buffer)
}

func LoadNameVN(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	VnRand = &NameRandVn{}
	VnRand.InitFromLoad(buffer)
}

func LoadNameKO(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	KoRand = &NameRandKo{}
	KoRand.InitFromLoad(buffer)
}

func LoadNameTH(filepath string) {
	errcheck := func(err error) {
		if err != nil {
			panic(err)
		}
	}
	buffer, err := loadBin(filepath)
	errcheck(err)

	ThRand = &NameRandTh{}
	ThRand.InitFromLoad(buffer)
}

func (nri *NameRandImpl) addNewNameForForeign(familyName, firstName string) {
	if familyName != "" {
		familyLen := len(familyName)
		if list, ok := nri.FamilyNameMap[familyLen]; ok {
			nri.FamilyNameMap[familyLen] = append(list, familyName)
		} else {
			nri.FamilyNameMap[familyLen] = append([]string{}, familyName)
			logs.Debug("family new count %d", familyLen)
		}
	}
	if firstName != "" {
		firstLen := len(firstName)
		if list, ok := nri.FirstNameMap[firstLen]; ok {
			nri.FirstNameMap[firstLen] = append(list, firstName)
		} else {
			nri.FirstNameMap[firstLen] = append([]string{}, firstName)
			logs.Debug("first new count %d", firstLen)
		}
	}
}

func (nri *NameRandImpl) init() {
	nri.FamilyNameMap = make(map[int][]string)
	nri.FirstNameMap = make(map[int][]string)
	nri.FirstNameCounts = make([]int, 0)
	nri.FamilyNameCounts = make([]int, 0)
}

func (nri *NameRandImpl) addLangForeignCount() {
	for key := range nri.FamilyNameMap {
		nri.FamilyNameCounts = append(nri.FamilyNameCounts, key)
	}
	for key := range nri.FirstNameMap {
		nri.FirstNameCounts = append(nri.FirstNameCounts, key)
	}
	sort.Ints(nri.FamilyNameCounts)
	sort.Ints(nri.FirstNameCounts)
	logs.Info("FamilyNameCounts %v", nri.FamilyNameCounts)
	logs.Info("FirstNameCounts %v", nri.FirstNameCounts)
}

func (nri *NameRandImpl) calcForeignCount() {
	familyLen := make([]int, 0)
	for _, family := range nri.FamilyNameCounts {
		familyLen = append(familyLen, len(nri.FamilyNameMap[family]))
	}
	firstLen := make([]int, 0)
	for _, first := range nri.FirstNameCounts {
		firstLen = append(firstLen, len(nri.FirstNameMap[first]))
	}
	if len(familyLen) == 0 || len(firstLen) == 0 {
		return
	}
	logs.Info("familyLen %v", familyLen)
	logs.Info("firstLen %v", firstLen)
	// >=[i] 的总数
	familyTotal := 0
	for _, faLen := range familyLen {
		familyTotal += faLen
	}
	for i := len(firstLen) - 2; i >= 0; i-- {
		firstLen[i] += firstLen[i+1]
	}
	logs.Info("familyLen total %v", familyTotal)
	logs.Info("firstLen total %v", firstLen)
	totalCount := familyTotal * firstLen[0]
	for _, count := range nri.FamilyNameCounts { //姓太长
		if count > 9 {
			totalCount++
		}
	}

	notMatchCount := 0 //不匹配的数量（名字太长）
	for i, family := range nri.FamilyNameCounts {
		for j, first := range nri.FirstNameCounts {
			if family+first > 11 {
				notMatchCount += familyLen[i] * firstLen[j]
				break
			}
		}
	}
	logs.Info("total Combat count total %d - notmatch %d = %d", totalCount, notMatchCount, totalCount-notMatchCount)
}

func CheckSymbol(str string) bool {
	if uutil.IsVNVer() {
		return gdSymReg.Find([]byte(strings.ToLower(str))) != nil || gdVnReg.Find([]byte(strings.ToLower(str))) != nil
	} else {
		return gdSymReg.Find([]byte(strings.ToLower(str))) != nil
	}
}

func CheckSensitive(str string) bool {
	if gdSenReg == nil {
		return false
	}
	return gdSenReg.Find([]byte(strings.ToLower(str))) != nil
}

func RandNames(count int, langToken string) []string {
	if langToken == uutil.Lang_EN {
		return EnRand.Rand(count)
	}

	if langToken == uutil.Lang_VN {
		return VnRand.Rand(count)
	}

	if langToken == uutil.Lang_KO {
		return KoRand.Rand(count)
	}

	if langToken == uutil.Lang_TH {
		return ThRand.Rand(count)
	}

	ni, ok := gdLangNameRandInfo[langToken]
	if !ok {
		if uutil.IsHMTVer() {
			ni = gdLangNameRandInfo[uutil.Lang_HMT]
		} else if uutil.IsJAVer() {
			ni = gdLangNameRandInfo[uutil.Lang_JA]
		} else {
			ni = gdLangNameRandInfo[uutil.Lang_HANS]
		}
	}
	ret := make([]string, 0, count)
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := 0; i < count; i++ {
		familyName := ni.NameItems[r.Int31n(ni.FamilyNameCount)].GetFamilyName()
		firstName := ni.NameItems[r.Int31n(ni.FirstNameCount)].GetFirstName()
		j := 1
		for familyName == firstName && j <= 3 {
			firstName = ni.NameItems[r.Int31n(ni.FirstNameCount)].GetFirstName()
			j++
		}
		if j == 3 {
			logs.Warn("randnames firstname may have duplicated word, %s", firstName)
		}
		ret = append(ret, ni.NameItems[r.Int31n(ni.HomeTownNameCount)].GetHomeTown()+familyName+firstName)
	}
	return ret
}

func RandRobotNames(count int) []string {
	if uutil.IsVNVer() {
		return randRobotNamesByLimit(VnRand, count, true)
	}
	if uutil.IsKOVer() {
		return randRobotNamesByLimit(KoRand, count, false)
	}
	if uutil.IsThVer() {
		return randRobotNamesByLimit(ThRand, count, false)
	}
	ni := gdLangNameRandInfo[uutil.Lang_HANS]
	if uutil.IsJAVer() {
		ni = gdLangNameRandInfo[uutil.Lang_JA]
	}
	if uutil.IsHMTVer() {
		ni = gdLangNameRandInfo[uutil.Lang_HMT]
	}
	ret := make([]string, 0, count)
	bHome := rand.Intn(int(ni.HomeTownNameCount))
	bFamily := rand.Intn(int(ni.FamilyNameCount))
	bFirst := rand.Intn(int(ni.FirstNameCount))
	for i := 0; i < count; i++ {
		iHome := (bHome + i) % int(ni.HomeTownNameCount)
		iFamiley := (bFamily + i) % int(ni.FamilyNameCount)
		iFirst := (bFirst + i) % int(ni.FirstNameCount)
		ret = append(ret, ni.NameItems[iHome].GetHomeTown()+
			ni.NameItems[iFamiley].GetFamilyName()+
			ni.NameItems[iFirst].GetFirstName())
	}
	return ret
}

// 随机机器人的名字， 名字受长度限制， 目前用于ko, vn
func randRobotNamesByLimit(nr NameRand, count int, isSurfix bool) []string {
	nameSet := make(map[string]struct{}, count)
	maxTime := 300
	loopTime := 0
	leftCount := count
	for {
		loopTime++
		tempCount := leftCount
		if leftCount < 1000 {
			tempCount = tempCount * 2
		}
		tempName := randVnRobotNames2(nr, tempCount, isSurfix)
		for _, name := range tempName {
			nameSet[name] = struct{}{}
		}
		leftCount = count - len(nameSet)
		if leftCount <= 0 {
			break
		}
		if loopTime > maxTime {
			logs.Info("rand robot names reach max time %d", loopTime)
			break
		}
	}
	if len(nameSet) < count {
		logs.Info("rand robot names not enough, %d", len(nameSet))
	}
	retNames := make([]string, 0)
	for name := range nameSet {
		retNames = append(retNames, name)
		if len(retNames) == count {
			break
		}
	}
	// 没有随到相应的人数的处理方案
	if leftCount > 0 {
		tempName := nr.Rand(leftCount)
		retNames = append(retNames, tempName...)
	}
	return retNames
}

func randVnRobotNames2(nr NameRand, tempCount int, isNumSurfix bool) []string {
	tempName := nr.Rand(tempCount)
	if !isNumSurfix {
		return tempName
	}
	for index, name := range tempName {
		randNum := rand.Intn(100)
		tempName[index] = name + fmt.Sprintf("%d", randNum)
	}
	return tempName
}

// 该实现方法不是等概率的
func (nri *NameRandImpl) randForeignNames(count int) []string {
	ret := make([]string, 0, count)
	for i := 0; i < count; i++ {
		familyLen, firstLen := nri.randNamePairForForeign()
		familyList := nri.FamilyNameMap[familyLen]
		familyName := familyList[rand.Intn(len(familyList))]
		midName := ""
		firstName := ""
		if firstLen != -1 {
			midName = nri.MidName
			firstList := nri.FirstNameMap[firstLen]
			firstName = firstList[rand.Intn(len(firstList))]
		}
		ret = append(ret, familyName+midName+firstName)
	}
	return ret
}

func (nri *NameRandImpl) randFirstNameLen(r *rand.Rand, maxLen int) int {
	maxIndex := len(nri.FirstNameCounts)
	for i, v := range nri.FirstNameCounts {
		if v > maxLen {
			maxIndex = i
			break
		}
	}
	if maxIndex == 0 {
		return -1
	}
	return nri.FirstNameCounts[r.Int31n(int32(maxIndex))]
}

// 随机一对长度, 加起来不超过MaxLen
func (nri *NameRandImpl) randNamePairForForeign() (int, int) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	lenFamilyIndex := r.Intn(len(nri.FamilyNameCounts))
	randFamilyLen := nri.FamilyNameCounts[lenFamilyIndex]
	if randFamilyLen >= nri.FamilyMaxLen {
		return randFamilyLen, -1
	}
	randFirstLen := nri.randFirstNameLen(r, nri.AllMaxLen-randFamilyLen)
	if randFirstLen == -1 {
		return randFamilyLen, -1
	}
	return randFamilyLen, randFirstLen
}
