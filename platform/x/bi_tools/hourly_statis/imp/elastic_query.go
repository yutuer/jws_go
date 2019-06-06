package imp

import (
	"fmt"

	"time"

	"golang.org/x/net/context"
	"gopkg.in/olivere/elastic.v5"
	"vcs.taiyouxi.net/platform/planx/util/logs"
)

const timeout = 60

func collectLoginAndDevice() error {
	typ := elastic.NewMatchQuery("type_name", "Login")
	date := elastic.NewRangeQuery("@timestamp").
		Gte(data_time_begin).
		Lte(data_time_end).
		TimeZone("+08:00")
	ac := elastic.NewMatchQuery("info.LoginTimes", 1)
	bl := elastic.NewBoolQuery().Must(typ, date, ac)

	builder := elasticClient.Search().Index(Cfg.Index).Query(bl)

	accountIdAgg := elastic.NewCardinalityAggregation().Field("accountid").PrecisionThreshold(1000)
	deviceAgg := elastic.NewCardinalityAggregation().Field("info.DeviceId").PrecisionThreshold(1000)
	sidAgg := elastic.NewTermsAggregation().Field("sid").Size(200).OrderByCountDesc()
	sidAgg.SubAggregation("1", accountIdAgg)
	sidAgg.SubAggregation("2", deviceAgg)
	channelAgg := elastic.NewTermsAggregation().Field("info.ProfileInfo.channelId").Size(200).OrderByCountDesc()
	channelAgg.SubAggregation("sid", sidAgg)

	builder = builder.Aggregation("agg", channelAgg)

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	searchResult, err := builder.Do(ctx)
	if err != nil {
		logs.Error("elastic.Do err %s", err.Error())
		return err
	}
	if searchResult.Hits == nil {
		logs.Error("elastic.searchResult Hits nil")
		return err
	}
	logs.Info("collectLoginAndDevice TotalHits %v", searchResult.Hits.TotalHits)

	resAgg := searchResult.Aggregations
	if resAgg == nil {
		logs.Error("unexpected Aggregations == nil")
		return fmt.Errorf("unexpected Aggregations == nil")
	}
	termsAggRes, found := resAgg.Terms("agg")
	if !found {
		logs.Error("unexpected Aggregations agg not found")
		return fmt.Errorf("unexpected Aggregations agg not found")
	}
	for _, bucket := range termsAggRes.Buckets {
		sidData, ok := channel2Sid2Data[string(bucket.KeyNumber)]
		if !ok {
			sidData = make(map[string]*data, 5)
			channel2Sid2Data[string(bucket.KeyNumber)] = sidData
		}

		_pm, ok := bucket.Aggregations["sid"]
		if ok {
			term := elastic.AggregationBucketKeyItems{}
			if err := term.UnmarshalJSON(*_pm); err != nil {
				logs.Error("term.UnmarshalJSON(*_pm) err %s", err.Error())
				return err
			}
			for _, sidBucket := range term.Buckets {
				sid := string(sidBucket.KeyNumber)
				d, ok := sidData[sid]
				if !ok {
					d = &data{}
					sidData[sid] = d
				}
				_pm, ok := sidBucket.Aggregations["1"]
				if ok {
					metric := elastic.AggregationValueMetric{}
					if err := metric.UnmarshalJSON(*_pm); err != nil {
						logs.Error("metric.UnmarshalJSON(*_pm) err %s", err.Error())
						return err
					}
					d.registerCount = int(*metric.Value)
				}
				_pm, ok = sidBucket.Aggregations["2"]
				if ok {
					metric := elastic.AggregationValueMetric{}
					if err := metric.UnmarshalJSON(*_pm); err != nil {
						logs.Error("metric.UnmarshalJSON(*_pm) err %s", err.Error())
						return err
					}
					d.deviceCount = int(*metric.Value)
				}
			}
		}
	}
	logs.Info("collectLoginAndDevice %v", channel2Sid2Data)
	return nil
}

func collectActivePlayer() error {
	all := elastic.NewMatchAllQuery()
	date := elastic.NewRangeQuery("@timestamp").
		Gte(data_time_begin).
		Lte(data_time_end).
		TimeZone("+08:00")
	bl := elastic.NewBoolQuery().Must(all, date)

	builderQ := elasticClient.Search().Index(Cfg.Index).Query(bl)
	accountIdAgg := elastic.NewCardinalityAggregation().Field("accountid").PrecisionThreshold(1000)
	sidAgg := elastic.NewTermsAggregation().Field("sid").Size(200).OrderByCountDesc()
	sidAgg.SubAggregation("1", accountIdAgg)
	channelAgg := elastic.NewTermsAggregation().Field("channel").Size(200).OrderByCountDesc()
	channelAgg.SubAggregation("sid", sidAgg)

	builder := builderQ.Aggregation("agg", channelAgg)

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	searchResult, err := builder.Do(ctx)
	if err != nil {
		logs.Error("elastic.Do err %s", err.Error())
		return err
	}
	if searchResult.Hits == nil {
		logs.Error("elastic.searchResult Hits nil")
		return err
	}
	logs.Info("collectActivePlayer TotalHits %v", searchResult.Hits.TotalHits)

	resAgg := searchResult.Aggregations
	if resAgg == nil {
		logs.Error("unexpected Aggregations == nil")
		return fmt.Errorf("unexpected Aggregations == nil")
	}
	termsAggRes, found := resAgg.Terms("agg")
	if !found {
		logs.Error("unexpected Aggregations agg not found")
		return fmt.Errorf("unexpected Aggregations agg not found")
	}

	for _, bucket := range termsAggRes.Buckets {
		sidData, ok := channel2Sid2Data[string(bucket.KeyNumber)]
		if !ok {
			sidData = make(map[string]*data, 5)
			channel2Sid2Data[string(bucket.KeyNumber)] = sidData
		}

		_pm, ok := bucket.Aggregations["sid"]
		if ok {
			term := elastic.AggregationBucketKeyItems{}
			if err := term.UnmarshalJSON(*_pm); err != nil {
				logs.Error("term.UnmarshalJSON(*_pm) err %s", err.Error())
				return err
			}
			for _, sidBucket := range term.Buckets {
				sid := string(sidBucket.KeyNumber)
				d, ok := sidData[sid]
				if !ok {
					d = &data{}
					sidData[sid] = d
				}
				_pm, ok := sidBucket.Aggregations["1"]
				if ok {
					metric := elastic.AggregationValueMetric{}
					if err := metric.UnmarshalJSON(*_pm); err != nil {
						logs.Error("metric.UnmarshalJSON(*_pm) err %s", err.Error())
						return err
					}
					d.activeCount = int(*metric.Value)
				}
			}
		}
	}
	logs.Info("collectActivePlayer %v", channel2Sid2Data)

	return nil
}

func collectChargeAccountCount() error {
	typ := elastic.NewMatchQuery("type_name", "IAP")
	date := elastic.NewRangeQuery("@timestamp").
		Gte(data_time_begin).
		Lte(data_time_end).
		TimeZone("+08:00")
	bl := elastic.NewBoolQuery().Must(typ, date)

	builder := elasticClient.Search().Index(Cfg.Index).Query(bl)

	accountIdAgg := elastic.NewCardinalityAggregation().Field("accountid").PrecisionThreshold(1000)
	sidAgg := elastic.NewTermsAggregation().Field("sid").Size(200).OrderByCountDesc()
	sidAgg.SubAggregation("1", accountIdAgg)
	channelAgg := elastic.NewTermsAggregation().Field("channel").Size(200).OrderByCountDesc()
	channelAgg.SubAggregation("sid", sidAgg)

	builder = builder.Aggregation("agg", channelAgg)

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	searchResult, err := builder.Do(ctx)
	if err != nil {
		logs.Error("elastic.Do err %s", err.Error())
		return err
	}
	if searchResult.Hits == nil {
		logs.Error("elastic.searchResult Hits nil")
		return err
	}
	logs.Info("collectChargeAccountCount TotalHits %v", searchResult.Hits.TotalHits)

	resAgg := searchResult.Aggregations
	if resAgg == nil {
		logs.Error("unexpected Aggregations == nil")
		return fmt.Errorf("unexpected Aggregations == nil")
	}
	termsAggRes, found := resAgg.Terms("agg")
	if !found {
		logs.Error("unexpected Aggregations agg not found")
		return fmt.Errorf("unexpected Aggregations agg not found")
	}
	for _, bucket := range termsAggRes.Buckets {
		sidData, ok := channel2Sid2Data[string(bucket.KeyNumber)]
		if !ok {
			sidData = make(map[string]*data, 5)
			channel2Sid2Data[string(bucket.KeyNumber)] = sidData
		}

		_pm, ok := bucket.Aggregations["sid"]
		if ok {
			term := elastic.AggregationBucketKeyItems{}
			if err := term.UnmarshalJSON(*_pm); err != nil {
				logs.Error("term.UnmarshalJSON(*_pm) err %s", err.Error())
				return err
			}
			for _, sidBucket := range term.Buckets {
				sid := string(sidBucket.KeyNumber)
				d, ok := sidData[sid]
				if !ok {
					d = &data{}
					sidData[sid] = d
				}
				_pm, ok := sidBucket.Aggregations["1"]
				if ok {
					metric := elastic.AggregationValueMetric{}
					if err := metric.UnmarshalJSON(*_pm); err != nil {
						logs.Error("metric.UnmarshalJSON(*_pm) err %s", err.Error())
						return err
					}
					d.chargeAcidCount = int(*metric.Value)
				}
			}
		}
	}
	logs.Info("collectChargeAccountCount %v", channel2Sid2Data)
	return nil
}

func collectChargeSum() error {
	typ := elastic.NewMatchQuery("type_name", "IAP")
	date := elastic.NewRangeQuery("@timestamp").
		Gte(data_time_begin).
		Lte(data_time_end).
		TimeZone("+08:00")
	bl := elastic.NewBoolQuery().Must(typ, date)

	builder := elasticClient.Search().Index(Cfg.Index).Query(bl)

	accountIdAgg := elastic.NewSumAggregation().Field("info.Money")
	sidAgg := elastic.NewTermsAggregation().Field("sid").Size(200).OrderByCountDesc()
	sidAgg.SubAggregation("1", accountIdAgg)
	channelAgg := elastic.NewTermsAggregation().Field("channel").Size(200).OrderByCountDesc()
	channelAgg.SubAggregation("sid", sidAgg)

	builder = builder.Aggregation("agg", channelAgg)

	ctx, cancel := context.WithTimeout(context.Background(), timeout*time.Second)
	defer cancel()
	searchResult, err := builder.Do(ctx)
	if err != nil {
		logs.Error("elastic.Do err %s", err.Error())
		return err
	}
	if searchResult.Hits == nil {
		logs.Error("elastic.searchResult Hits nil")
		return err
	}
	logs.Info("collectChargeSum TotalHits %v", searchResult.Hits.TotalHits)

	resAgg := searchResult.Aggregations
	if resAgg == nil {
		logs.Error("unexpected Aggregations == nil")
		return fmt.Errorf("unexpected Aggregations == nil")
	}
	termsAggRes, found := resAgg.Terms("agg")
	if !found {
		logs.Error("unexpected Aggregations agg not found")
		return fmt.Errorf("unexpected Aggregations agg not found")
	}
	for _, bucket := range termsAggRes.Buckets {
		sidData, ok := channel2Sid2Data[string(bucket.KeyNumber)]
		if !ok {
			sidData = make(map[string]*data, 5)
			channel2Sid2Data[string(bucket.KeyNumber)] = sidData
		}

		_pm, ok := bucket.Aggregations["sid"]
		if ok {
			term := elastic.AggregationBucketKeyItems{}
			if err := term.UnmarshalJSON(*_pm); err != nil {
				logs.Error("term.UnmarshalJSON(*_pm) err %s", err.Error())
				return err
			}
			for _, sidBucket := range term.Buckets {
				sid := string(sidBucket.KeyNumber)
				d, ok := sidData[sid]
				if !ok {
					d = &data{}
					sidData[sid] = d
				}
				_pm, ok := sidBucket.Aggregations["1"]
				if ok {
					metric := elastic.AggregationValueMetric{}
					if err := metric.UnmarshalJSON(*_pm); err != nil {
						logs.Error("metric.UnmarshalJSON(*_pm) err %s", err.Error())
						return err
					}
					d.chargeSum = int(*metric.Value) * 100
				}
			}
		}
	}
	logs.Info("collectChargeSum %v", channel2Sid2Data)
	return nil
}
