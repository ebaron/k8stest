package main

import (
	"fmt"
	"strings"
	"time"

	hawkular "github.com/hawkular/hawkular-client-go/metrics"
	v1 "k8s.io/client-go/pkg/api/v1"
)

type metricsClient struct {
	hawkularClient *hawkular.Client
}

const (
	descriptorTag string = "descriptor_name"
	cpuDesc       string = "cpu/usage_rate"
	memDesc       string = "memory/usage"
	netSent       string = "network/tx_rate"
	netRecv       string = "network/rx_rate"
	typeTag       string = "type"
	typePod       string = "pod"
	podIDTag      string = "pod_id"
)

func newMetricsClient(metricsURL string, token string) (*metricsClient, error) {
	params := hawkular.Parameters{
		Url:   metricsURL,
		Token: token,
	}
	client, err := hawkular.NewHawkularClient(params)
	if err != nil {
		return nil, err
	}

	mc := new(metricsClient)
	mc.hawkularClient = client

	return mc, nil
}

func (mc *metricsClient) getCPUMetrics(pods []v1.Pod, namespace string) (float64, error) {
	return mc.getBucketAverage(pods, namespace, cpuDesc)
}

func (mc *metricsClient) getMemoryMetrics(pods []v1.Pod, namespace string) (float64, error) {
	return mc.getBucketAverage(pods, namespace, memDesc)
}

func (mc *metricsClient) getBucketAverage(pods []v1.Pod, namespace, descTag string) (float64, error) {
	result, err := mc.readBuckets(pods, namespace, descTag)
	if err != nil {
		return -1, err
	} else if result == nil {
		return -1, nil
	}

	// Return average from bucket
	return result.Avg, err
}

func (mc *metricsClient) getMetricsForPods(pods []v1.Pod, namespace string, descTag sting) (float64, int64, error) {
	// Get most recent sample from each pod's gauge
	samples, err := mc.readRaw(pods, namespace, descTag)
	if err != nil {
		return -1, -1, err
	}

	// Return sum of metrics for each pod, and average of timestamp
	var totalValue float64
	var avgTimestamp int64
}

func (mc *metricsClient) readBuckets(pods []v1.Pod, namespace string, descTag string) (*hawkular.Bucketpoint, error) {
	numPods := len(pods)
	if numPods == 0 {
		return nil, nil
	}

	// Extract UIDs from pods
	podUIDs := make([]string, numPods)
	for idx, pod := range pods {
		podUIDs[idx] = string(pod.UID)
	}
	// Build Hawkular tags for query
	podsForTag := strings.Join(podUIDs, "|")
	tags := map[string]string{
		descriptorTag: descTag,
		typeTag:       typePod,
		podIDTag:      podsForTag,
	}

	// Tenant should be set to OSO project name
	mc.hawkularClient.Tenant = namespace
	// Get a bucket for the last 2 minutes (OSO's bucket duration for last hour shown)
	startTime := time.Now().Add(-120000 * time.Millisecond)
	// FIXME Get most recent bucket for now since we don't have time-series based API yet
	buckets, err := mc.hawkularClient.ReadBuckets(hawkular.Gauge, hawkular.Filters(hawkular.TagsFilter(tags),
		hawkular.BucketsFilter(1), hawkular.StackedFilter() /* Sum of each pod */, hawkular.StartTimeFilter(startTime)))
	//	hawkular.BucketsDurationFilter(120000*time.Millisecond), hawkular.StartTimeFilter(time.Now().Add(-60*time.Minute)))) What OSO uses
	if err != nil {
		return nil, err
	}

	// XXX Raw requests:
	// {"tags":"descriptor_name:memory/usage|cpu/usage_rate,type:pod_container,pod_id:7e61b8f4-ca45-11e7-b904-02e52a0be43d|871b6d6f-ca4e-11e7-b904-02e52a0be43d,container_name:vertx","bucketDuration":"120000ms","start":"-60mn"}
	// {"tags":"descriptor_name:network/tx_rate|network/rx_rate,type:pod,pod_id:7e61b8f4-ca45-11e7-b904-02e52a0be43d|871b6d6f-ca4e-11e7-b904-02e52a0be43d","bucketDuration":"120000ms","start":1511293645209}

	// Should have gotten at most one bucket
	if len(buckets) == 0 {
		return nil, nil
	}
	return buckets[0], nil
}

func (mc *metricsClient) readRaw(pods []v1.Pod, namespace string, descTag string) ([]*hawkular.Datapoint, error) {
	numPods := len(pods)
	if numPods == 0 {
		return nil, nil
	}

	// Tenant should be set to OSO project name
	mc.hawkularClient.Tenant = namespace
	result := make([]*hawkular.Datapoint, 0, len(pods))
	for _, pod := range pods {
		// Gauge ID is "pod/<pod UID>/<descriptor>"
		gaugeID := typePod + "/" + string(pod.UID) + "/" + descTag
		// Get most recent sample from gauge
		points, err := mc.hawkularClient.ReadRaw(hawkular.Gauge, gaugeID, hawkular.Filters(hawkular.LimitFilter(1),
			hawkular.OrderFilter(hawkular.DESC)))
		if err != nil {
			return nil, err
		}

		// We should have received at most one datapoint
		if len(points) > 0 {
			result = append(result, points[0])
		}
	}
	return result, nil
}

func (mc *metricsClient) getMetrics(tags map[string]string) error {
	definitions, err := mc.hawkularClient.Definitions(hawkular.Filters(hawkular.TagsFilter(tags)))
	if err != nil {
		return err
	}

	for _, def := range definitions {
		fmt.Println(def.ID)
	}
	return nil
}
