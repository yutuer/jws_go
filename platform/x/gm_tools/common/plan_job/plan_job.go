package plan_job

import (
	"encoding/json"
	"sync"
	"time"
	"vcs.taiyouxi.net/platform/planx/util/logs"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/gm_command"
	"vcs.taiyouxi.net/platform/x/gm_tools/common/store"
)

var (
	time_begin_unix = time.Date(2015, 9, 1, 0, 0, 0, 0, time.UTC).Unix()
)

type planJob struct {
	ID           int64                    `json:"id"`
	TimeBegin    int64                    `json:"begin"`
	TimeEnd      int64                    `json:"end"`
	TimeInterval int64                    `json:"interval"`
	MaxSend      int                      `json:"max"`
	Command      gm_command.CommandRecord `json:"command"`
	CurrSended   int                      `json:"curr"`
	LastSendTime int64                    `json:"last"`
}

func NewPlanJob(id, begin, end, in int64, max int, command gm_command.CommandRecord) planJob {
	return planJob{
		ID:           id,
		TimeBegin:    begin,
		TimeEnd:      end,
		TimeInterval: in,
		MaxSend:      max,
		Command:      command,
		CurrSended:   0,
		LastSendTime: 0,
	}
}

const (
	planJobCommand_Typ_Add = iota
	planJobCommand_Typ_Update
	planJobCommand_Typ_Del
	planJobCommand_Typ_Get
	planJobCommand_Typ_Get_By_Serv
)

type planJobCommand struct {
	typ      int
	job      planJob
	id       int64
	server   string
	res_chan chan []byte
}

type PlanJobManager struct {
	Plans []planJob `json:"jobs"`
	Name  string    `json:"N"`
	Count int64     `json:"c"`

	isNeedSave bool
	waitter    sync.WaitGroup
	command    chan planJobCommand
}

func (p *PlanJobManager) Load() error {
	err := store.GetFromJson(store.NormalBucket, p.Name+"planJob", p)
	logs.Trace("[PlanJobManager]Load %v", *p)
	if err == store.ErrNoKey {
		return nil
	}
	return err
}

func (p *PlanJobManager) Save() error {
	return store.SetIntoJson(store.NormalBucket, p.Name+"planJob", *p)
}

func (p *PlanJobManager) Start() error {
	err := p.Load()
	if err != nil {
		return err
	}

	p.command = make(chan planJobCommand, 128)

	jobTimer := time.NewTicker(1 * time.Second)

	p.waitter.Add(1)
	go func() {
		defer p.waitter.Done()
		for {
			select {
			case c, ok := <-p.command:
				logs.Trace("planJobCommand %v", c)
				if !ok {
					logs.Warn("command close")
					return
				}
				switch c.typ {
				case planJobCommand_Typ_Add:
					p.Plans = append(p.Plans, c.job)
					p.isNeedSave = true
					c.res_chan <- []byte{}
					break
				case planJobCommand_Typ_Update:
					for idx, i := range p.Plans {
						logs.Info("c %v -> %v", i.ID, c.id)
						if i.ID == c.id {
							logs.Info("update %v -> %v", p.Plans[idx], c.job)
							p.Plans[idx] = c.job
						}
					}
					c.res_chan <- []byte{}
					break
				case planJobCommand_Typ_Del:
					l := len(p.Plans)
					logs.Trace("del %v", p.Plans)
					for i := 0; i < l; i++ {
						if p.Plans[i].ID == c.id {
							p.Plans[i] = p.Plans[l-1]
							p.Plans = p.Plans[:l-1]
							break
						}
					}
					logs.Trace("del a %v", p.Plans)
					c.res_chan <- []byte{}
					break
				case planJobCommand_Typ_Get:
					c.res_chan <- p.getJobs()
					break
				case planJobCommand_Typ_Get_By_Serv:
					c.res_chan <- p.getJobByServer(c.server)
				default:
					logs.Error("Error typ %v", c)
				}
			case <-jobTimer.C:
				p.waitter.Add(1)
				func() {
					defer p.waitter.Done()
					p.Tick()
				}()
			}
		}
	}()

	return nil
}

func (p *PlanJobManager) Stop() {
	close(p.command)
	err := p.Save()
	if err != nil {
		logs.Error("Save Err by %s --- %v",
			err.Error(), p.Plans)
	}
	p.waitter.Wait()
	logs.Warn("PlanJobManager %s Stop", p.Name)
}

func (p *PlanJobManager) ActivityJob(pl *planJob) {
	go func() {
		defer func() {
			if err := recover(); err != nil {
				logs.Error("[ActivityJob] recover error %v", err)
			}
		}()
		gm_command.OnCommandByRecord(&pl.Command)
	}()
}

func (p *PlanJobManager) Tick() {
	now_t := time.Now().Unix()
	for i := 0; i < len(p.Plans); i++ {
		plan := &p.Plans[i]
		if plan.TimeBegin < now_t && now_t <= plan.TimeEnd {
			if now_t-plan.LastSendTime > plan.TimeInterval {
				plan.LastSendTime = now_t
				p.ActivityJob(plan)
			}
		}
	}
}

func (p *PlanJobManager) mkJobID() int64 {
	now_t := time.Now().Unix() - time_begin_unix
	p.Count += 1
	if p.Count >= 10000 {
		p.Count = 0
	}
	return now_t*10000 + p.Count
}

func (p *PlanJobManager) getJobs() []byte {
	res, err := json.Marshal(p.Plans)
	if err != nil {
		logs.Error("PlanJobManager getJobs Er by %s", err.Error())
		return []byte{}
	}
	return res
}
func (p *PlanJobManager) getJobByServer(serverName string) []byte {
	plans := make([]planJob, 0, 10)
	for _, item := range p.Plans {
		if item.Command.ServerName == serverName {
			plans = append(plans, item)
		}
	}
	res, err := json.Marshal(plans)
	if err != nil {
		logs.Error("PlanJobManager getJobs Er by %s", err.Error())
		return []byte{}
	}
	return res
}

func (p *PlanJobManager) AddJob(begin, end, in int64, max int, command gm_command.CommandRecord) {
	res_chan := make(chan []byte, 1)
	p.command <- planJobCommand{
		typ:      planJobCommand_Typ_Add,
		job:      NewPlanJob(p.mkJobID(), begin, end, in, max, command),
		res_chan: res_chan,
	}
	<-res_chan
}

func (p *PlanJobManager) UpdateJob(id, begin, end, in int64, max int, command gm_command.CommandRecord) {
	res_chan := make(chan []byte, 1)
	p.command <- planJobCommand{
		id:       id,
		typ:      planJobCommand_Typ_Update,
		job:      NewPlanJob(id, begin, end, in, max, command),
		res_chan: res_chan,
	}
	<-res_chan
}

func (p *PlanJobManager) DelJob(id int64) {
	res_chan := make(chan []byte, 1)
	p.command <- planJobCommand{
		typ:      planJobCommand_Typ_Del,
		id:       id,
		res_chan: res_chan,
	}
	<-res_chan
}

func (p *PlanJobManager) GetAllJob() []byte {
	res_chan := make(chan []byte, 1)
	p.command <- planJobCommand{
		typ:      planJobCommand_Typ_Get,
		res_chan: res_chan,
	}

	res := <-res_chan

	return res
}

func (p *PlanJobManager) GetJobByServer(server string) []byte {
	res_chan := make(chan []byte, 1)
	p.command <- planJobCommand{
		typ:      planJobCommand_Typ_Get_By_Serv,
		res_chan: res_chan,
		server:   server,
	}

	res := <-res_chan

	return res
}
