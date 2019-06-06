package csrob

type utilCache struct {
	res *resources

	guild *cacheGuild
}

func newUtilCache(res *resources) *utilCache {
	c := &utilCache{
		res: res,
	}
	c.guild = newCacheGuild(c)

	c.res.ticker.regTickerToList(10, 1, c.guild.refresh)
	return c
}
