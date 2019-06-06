package storehelper

// import (
//     "strings"
//     "time"
//     "vcs.taiyouxi.net/platform/planx/util/logs"
//     gossdb "vcs.taiyouxi.net/platform/planx/util/storehelper/ssdb"
// )

// //YZH 这个StoreSSDB没有使用, 因此注释掉
// type StoreSSDB struct {
//     pool *gossdb.Connectors

//     server string
//     port   int
//     format string
//     seq    string
// }

// func NewStoreSSDB(server string, port int, format, seq string) *StoreSSDB {
//     return &StoreSSDB{
//         server: server,
//         port:   port,
//         format: format,
//         seq:    seq,
//     }
// }

// func (s *StoreSSDB) Open() error {
//     pool, err := gossdb.NewPool(&gossdb.Config{
//         Host:             s.server,
//         Port:             s.port,
//         MinPoolSize:      32,
//         MaxPoolSize:      64,
//         AcquireIncrement: 8,
//     })
//     if err != nil {
//         return err
//     }
//     s.pool = pool
//     return nil
// }

// func (s *StoreSSDB) Clone() (IStore, error) {
//     ns := NewStoreSSDB(
//         s.server,
//         s.port,
//         s.format,
//         s.seq,
//     )
//     if err := ns.Open(); err != nil {
//         return nil, err
//     }
//     return s, nil
// }

// func (s *StoreSSDB) GetConn() (*gossdb.Client, error) {
//     cli, err := s.pool.NewClient()
//     if err != nil {
//         logs.Critical("StoreSSDB GetConn Err By %s", err.Error())
//     }
//     return cli, err
// }

// func (s *StoreSSDB) Close() error {
//     s.pool.Close()
//     return nil
// }

// func (s *StoreSSDB) Put(key string, val []byte, rh ReadHandler) error {
//     cli, err := s.GetConn()
//     defer cli.Close()
//     if err != nil {
//         return err
//     }
//     return cli.Set(key, val)
// }
// func (s *StoreSSDB) Get(key string) ([]byte, error) {
//     cli, err := s.GetConn()
//     defer cli.Close()
//     if err != nil {
//         return []byte{}, err
//     }
//     re, err := cli.Get(key)
//     return re.Bytes(), err
// }

// func (s *StoreSSDB) Del(key string) error {
//     cli, err := s.GetConn()
//     defer cli.Close()
//     if err != nil {
//         return err
//     }
//     return cli.Del(key)
// }

// func (s *StoreSSDB) ListObject(last_idx string, scan_len int64) ([]string, error) {
//     cli, err := s.GetConn()
//     defer cli.Close()
//     if err != nil {
//         return []string{}, err
//     }

//     res, err := cli.Keys(last_idx, "", scan_len)

//     return res[:], err
// }

// func (s *StoreSSDB) StoreKey(key string) string {
//     now_time := time.Now()
//     if s.format == "" {
//         return key
//     }
//     return now_time.Format(s.format) + s.seq + key
// }

// func (s *StoreSSDB) RedisKey(key_in_store string) (string, bool) {
//     // Redis存储路径
//     if s.seq == "" {
//         return key_in_store, true
//     }

//     d := strings.Split(key_in_store, s.seq)
//     if len(d) < 2 {
//         logs.Error("key_in_store err : %s in %s", key_in_store, s.seq)
//         return "", false
//     }
//     return d[len(d)-1], true
// }
