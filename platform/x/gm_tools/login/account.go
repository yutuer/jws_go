package login

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/planx/util/uuid"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/store"
)

const MaxGrantNum = 100

var (
	time_begin_unix = time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC).Unix()
	id_count        = 0
)

func getMd5String(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

type Account struct {
	Id         string           `json:"id"`
	Name       string           `json:"Name"`
	Pass       string           `json:"pass"`
	Typ        string           `json:"typ"`
	Grants     [MaxGrantNum]int `json:"grants"`
	Cookie     string           `json:"cookie"`
	CookieTime int64            `json:"cookietime"`
}

func (a *Account) IsPass(pass string) (bool, string) {
	p := getMd5String(getMd5String(a.Pass) + "ta1hehud0ng")

	logs.Trace("Pass : %v", pass)
	logs.Trace("True : %v", p)

	if p != pass {
		return false, ""
	}

	a.Cookie = getMd5String(getMd5String(uuid.NewV4().String()))
	a.CookieTime = time.Now().Unix() + 36000

	return true, a.Cookie
}

func (a *Account) IsCookiePass(cookie string) bool {
	now := time.Now().Unix()
	if now >= a.CookieTime {
		return false
	}

	return cookie == a.Cookie
}

//判断账号是否拥有权限
func (a *Account) IsGrant(gid, p int) bool {
	if gid < 0 || gid >= len(a.Grants) {
		return false
	}
	return a.Grants[gid] >= p
}

//添加权限
func (a *Account) Grant(gid int) {
	if gid < 0 || gid >= len(a.Grants) {
		return
	}
	a.Grants[gid] = 1
}

//删除权限
func (a *Account) DisGrant(gid int) {
	if gid < 0 || gid >= len(a.Grants) {
		return
	}
	a.Grants[gid] = 0
}

//创建账号，name：账号名、typ：账户类型、pass：密码
func NewAccount(name, typ, pass string) Account {
	new_id := time_begin_unix*1000 + int64(id_count)
	id_count++
	if id_count == 1000 {
		id_count = 0
	}
	return Account{
		Id:   fmt.Sprintf("%d", new_id),
		Name: name,
		Typ:  typ,
		Pass: pass,
	}
}

var AccountManager AccountMng

func Init() {
	AccountManager.Load()
}

func Save() {
	AccountManager.Save()
}

type AccountMng struct {
	Accounts map[string]Account `json:"Accounts"`
}

func (s *AccountMng) GetAll() []byte {
	j, err := json.Marshal(s.Accounts)
	if err != nil {
		logs.Error("[AccountMng] GetAll Marshal Err by %s", err.Error())
		return nil
	}
	return j[:]
}

func (s *AccountMng) Add(n Account) {
	s.Accounts[n.Name] = n
	s.Save()
}

func (s *AccountMng) Get(name string) *Account {
	a, ok := s.Accounts[name]
	if !ok {
		return nil
	} else {
		return &a
	}
}

func (s *AccountMng) GetByKey(key string) *Account {
	for _, a := range s.Accounts {
		if a.Cookie == key {
			return &a
		}
	}

	return nil
}

func (s *AccountMng) IsPass(name, pass string) (bool, string) {
	acc := s.Get(name)
	if acc == nil {
		return false, ""
	}

	is_pass, key := acc.IsPass(pass)
	if is_pass {
		s.Accounts[name] = *acc
	}

	return is_pass, key
}

func (s *AccountMng) Del(id string) {
	delete(s.Accounts, id)
	s.Save()
}

func (s *AccountMng) AddGrant(name string, ids ...int) {
	acc := s.Get(name)
	if acc == nil {
		logs.Error("未查找到账号:%v", name)
		return
	}

	for _, id := range ids {
		acc.Grant(id)
	}

	s.Accounts[name] = *acc
}

func (s *AccountMng) SetGrant(name string, ids ...int) {
	logs.Trace("SetGrant get name : %v,grants:%v", name, ids)
	acc := s.Get(name)
	if acc == nil {
		logs.Error("未查找到账号:%v", name)
		return
	}
	//先清空权限再赋值，如果顺序颠倒会导致固定清空权限bug
	acc = cleanGrant(acc)
	logs.Trace("after clean grants:%v", acc.Grants)
	for _, id := range ids {
		acc.Grant(id)
	}

	s.Accounts[name] = *acc
}

//清空账号name的权限
func cleanGrant(acc *Account) *Account {
	for i := range acc.Grants {
		acc.DisGrant(i)
	}
	return acc
}

//查询账号name的权限
func (s *AccountMng) GetGrant(name string) (error, []int) {
	acc := s.Get(name)
	if acc == nil {
		logs.Error("未查找到账号:%v", name)
		return fmt.Errorf("未查找到账号:%v", name), nil
	}

	return nil, acc.Grants[:]
}

//  权限的值的意义请参考 grants.go
func (s *AccountMng) Load() {
	err := store.GetFromJson(store.NormalBucket, "Accounts", s)
	logs.Trace("[AccountMng]Load %v", *s)
	if s.Accounts == nil {
		s.Accounts = make(map[string]Account, 128)
	}

	_, ok := s.Accounts["fanyang5d5e38v2d96ff3"]
	if !ok {
		s.Add(NewAccount("fanyang5d5e38v2d96ff3", "admin", "vca8##923&&56ak62erqrw"))
	}

	s.Add(NewAccount("guoxizheng", "admin", "23bf45ty789h7y89bjkl@bndf"))
	s.Add(NewAccount("lichong", "admin", "ny7890234tvnS$$uy89-3dx23"))
	s.Add(NewAccount("shaojiakun", "admin", "zPvo9VT#HF6IzIuH"))
	s.Add(NewAccount("wangchenyang", "admin", "U1qCS7Sj7fAKCkyt"))
	s.Add(NewAccount("wangliang", "admin", "dGMN&*(G@#F2f873n9r78ethgfmm"))
	s.Add(NewAccount("yxadminZKYB7SS9", "admin", "VYWd*9}%8cQ;3t^32EzP2y}jEa2o[X=3"))
	s.Add(NewAccount("hero", "admin", "vca8#8cQ;8ethgfmkl@bn"))
	s.Add(NewAccount("zhangzhen", "admin", "vca8##923&&56ak62erqrw"))
	s.Add(NewAccount("zhangdongping", "admin", "vca8##923&&56ak62erqrw"))
	s.Add(NewAccount("liujingzhao", "admin", "vca8##923&&56ak62erqrw"))
	s.Add(NewAccount("libingbing", "admin", "vca8##923&&56ak62erqrw"))
	s.Add(NewAccount("huquanqi", "admin", "vca8##923&&56ak62erqrw"))
	s.Add(NewAccount("baihao", "admin", "vca8##923&&56ak62erqrw"))
	s.Add(NewAccount("zhangwenfang", "admin", "SYvTP2@6yw$CI1cn"))
	s.Add(NewAccount("liangrui", "admin", "SYvTP2@6yw$CI1cn"))
	s.Add(NewAccount("lixiao", "admin", "qza&tlgeoj8eAiSb"))
	s.Add(NewAccount("liujunyan", "admin", "SYvTP2@6yw$CI1cn"))
	s.Add(NewAccount("shaoqi", "admin", "SYvTP2@6yw$CI1cn"))
	s.Add(NewAccount("guoyafeng", "admin", "vca8##923&&56ak62erqrw"))
	s.Add(NewAccount("zhaodan", "admin", "vca8##923&&56ak62erqrw"))
	s.Add(NewAccount("jpadmin", "admin", "SYvTP2@6yw$CI1cn"))
	s.Add(NewAccount("vnadmin", "admin", "SYvTP2@6yw$CI1cn"))
	s.Add(NewAccount("cuinanji", "admin", "2NQp9p^2eyCiPKJc"))
	s.Add(NewAccount("lichengmin", "admin", "2NQp9p^2eyCiPKJc"))
	s.AddQAAccount()
	s.AddGrant("fanyang5d5e38v2d96ff3", 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24, 25, 26,
		27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45,
		46, 47, 48, 49, 50, 51, 52, 53, 54, 61)
	s.AddGrant("guoxizheng", 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		30, 31, 32, 36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 55, 61)

	// 港澳台
	s.AddGrant("zhangwenfang", 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		30, 31, 32, 36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 59, 60, 61)
	s.AddGrant("liangrui", 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		30, 31, 32, 36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	s.AddGrant("liujunyan", 20, 21, 23, 26, 27, 30, 31, 32, 33, 40, 53, 61)
	s.AddGrant("shaoqi", 0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13,
		14, 15, 16, 17, 18, 19, 20, 23, 24, 25, 26,
		27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42, 43, 44, 45,
		46, 47, 48, 49, 50, 51, 52, 53, 54, 61)

	// 大陆
	s.AddGrant("shaojiakun", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	s.AddGrant("yxadminZKYB7SS9", 20, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 20, 19, 18, 38, 36,
		37, 39, 40, 41, 42, 43, 44, 45, 46, 47, 48, 61)
	s.AddGrant("hero", 15)
	//越南
	s.AddGrant("vnadmin", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	// 日本
	s.AddGrant("jpadmin", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	s.AddGrant("cuinanji", 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35,
		36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 64) // 韩国

	s.AddGrant("lichengmin", 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22,
		23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35,
		36, 37, 38, 39, 40, 41, 42, 43, 44, 45, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 64) // 韩国
	s.AddAllGrant("wangchenyang")
	s.AddAllGrant("wangliang")
	s.AddAllGrant("lichong")
	s.AddAllGrant("zhangzhen")
	s.AddAllGrant("zhangdongping")
	s.AddAllGrant("libingbing")
	s.AddAllGrant("huquanqi")
	s.AddAllGrant("baihao")
	s.AddAllGrant("lixiao")
	s.AddAllGrant("liujingzhao")
	s.AddAllGrant("guoyafeng")
	s.AddAllGrant("zhaodan")

	s.Add(NewAccount("vn-kefu", "admin", "#-ay0[F8CPWN9s=f"))
	s.AddGrant("vn-kefu", 20, 15, 51)

	if err == store.ErrNoKey {
		return
	}

	if err != nil {
		logs.Error("[AccountMng]Load Err by %s", err.Error())
	}
	return
}

func (s *AccountMng) AddQAAccount() {
	s.Add(NewAccount("qa1", "admin", "123456"))
	s.Add(NewAccount("qa2", "admin", "123456"))
	s.Add(NewAccount("qa3", "admin", "123456"))
	s.Add(NewAccount("qa4", "admin", "123456"))
	s.Add(NewAccount("qa5", "admin", "123456"))
	s.Add(NewAccount("qa6", "admin", "123456"))
	s.AddGrant("qa1", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	s.AddGrant("qa2", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	s.AddGrant("qa3", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	s.AddGrant("qa4", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	s.AddGrant("qa5", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)
	s.AddGrant("qa6", 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20, 21, 22, 23, 24,
		36, 37, 38, 39, 40, 41, 43, 44, 42, 45, 51, 52, 53, 54, 61)

}

func (s *AccountMng) Save() {
	store.SetIntoJson(store.NormalBucket, "Accounts", *s)
}

func (s *AccountMng) AddAllGrant(name string) {
	allGrants := make([]int, MaxGrantNum)
	for i := 0; i < MaxGrantNum; i++ {
		allGrants[i] = i
	}
	s.AddGrant(name, allGrants...)
}
