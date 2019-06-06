[ // 多个消息，json数组
  {
    "name": "GetTBTeamInfo",                    //消息名字
    "title": "获取组队BOSS的队伍列表",            // 消息title, 注释用
    "path": "Attr",                             // 消息路径
    "comment": "获取组队BOSS的队伍信息",     // 消息备注，注释用
    "req": {                                // 请求
      "params": [                           // 参数列表， 数组
        [
          "DifficultyId",                   // 代码中的变量
          "long",                           // 变量类型， 只支持bool, long, string, 自定义的struct  以及对应的数组
          "请求的难度ID",                    // 注释
          "diff_id"                         // 传输用的变量， 越短越好， 但是不能重复
        ]
      ]
    },
    "rsp": {                        // 返回
      "base": "",       // 是否带奖励， WithRewards
      "params": [
        [
          "TeamList",
          "TBTeam[]",
          "队伍列表",
          "team_list"
        ]
      ]
    },
    "objects": [{    // 自定义结构体， 数组
      "name": "TBTeam",   // 结构体名称
      "params": [    // 结构体成员变量
        [
          "TeamMemberCount",
          "long",
          "队伍人数",
          "tm_c"
        ],
        [
          "LeaderServerName",
          "string",
          "队长服务器名字",
          "leader_sn"
        ],
        [
          "LeaderPlayerName",
          "string",
          "队长名字",
          "leader_pn"
        ],
        [
          "FightAvatarIds",
          "long[]",
          "出战的角色avatarId",
          "f_ava_ids"
        ],
        [
          "TeamState",
          "long",
          "队伍状态 1=开放;2=无法加入；3=满员",
          "t_state"
        ]
      ]
    }]
  }
]