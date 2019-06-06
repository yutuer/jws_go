package discover

import (
	"golang.org/x/net/context"
	"vcs.taiyouxi.net/platform/planx/util/etcdClient"
)

//RegService ..
func RegService(s *Service) error {
	path := makeServicePath(s)
	val, err := encodeValue(s)
	if nil != err {
		return err
	}
	call := func(cli client.Client) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRequest)
		api := client.NewKeysAPI(cli)
		_, err := api.Set(ctx, path, val, nil)
		cancel()
		if nil != err {
			return err
		}
		return nil
	}
	return callWithClient(call)
}

//UnRegService ..
func UnRegService(s *Service) error {
	path := makeServicePath(s)
	call := func(cli client.Client) error {
		ctx, cancel := context.WithTimeout(context.Background(), defaultTimeoutRequest)
		api := client.NewKeysAPI(cli)
		_, err := api.Delete(ctx, path, nil)
		cancel()
		if nil != err {
			return err
		}
		return nil
	}
	return callWithClient(call)
}
