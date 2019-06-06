package csrob

//NamePoolPlayer 玩家缓存
type NamePoolPlayer struct {
	Acid     string `json:"acid,omitempty"`
	Sid      uint   `json:"sid,omitempty"`
	Name     string `json:"name,omitempty"`
	GuildID  string `json:"guild_id,omitempty"`
	GuildPos int    `json:"guild_pos,omitempty"`
}

//NamePoolGuild 公会缓存
type NamePoolGuild struct {
	GuildID     string `json:"guild_id,omitempty"`
	Sid         uint   `json:"sid,omitempty"`
	GuildName   string `json:"guild_name,omitempty"`
	GuildMaster string `json:"guild_master,omitempty"`
	Dismissed   bool   `json:"dismissed,omitempty"`
}

//NamePoolServer 服务器信息缓存
type NamePoolServer struct {
	Sid        uint   `json:"sid,omitempty"`
	ServerName string `json:"server_name,omitempty"`
}
