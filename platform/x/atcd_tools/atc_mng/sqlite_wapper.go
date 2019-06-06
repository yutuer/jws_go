package atc_mng

import ()

var ips_addon []string
var names_addon []string

func init() {
	ips_addon = make([]string, 0, 64)
	ips_addon = append(ips_addon, "10.0.1.250")
	names_addon = make([]string, 0, 64)
	names_addon = append(names_addon, "For Test")
}

type AtcdClientInfo struct {
	Idx      int    `json:"idx"`
	Ip       string `json:"ip"`
	Name     string `json:"name"`
	Tc_obj   string `json:"tc_obj"`
	Time_out int64  `json:"time_out"`
}

func GetAllAtcdClient(db_address string) ([]AtcdClientInfo, error) {
	res := make([]AtcdClientInfo, 0, 32)

	for idx, ip := range ips_addon {
		name := "Unknown"
		if len(names_addon) > idx {
			name = names_addon[idx]
		}
		res = append(res, AtcdClientInfo{
			0, ip, name, "", 0,
		})
	}
	return res[:], nil
}

func AddAddonIp(addon_ip string, name string) {
	for _, ip := range ips_addon {
		if ip == addon_ip {
			return
		}
	}
	ips_addon = append(ips_addon, addon_ip)
	names_addon = append(names_addon, name)
}
