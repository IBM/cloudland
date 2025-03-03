package apis

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"web/src/routes"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/gorm"
	"github.com/spf13/viper"
)

type MonitorAPI struct{}

var monitorAPI = &MonitorAPI{}

var (

	// Prometheus API path
	PrometheusQueryPath      = "/api/v1/query"
	PrometheusQueryRangePath = "/api/v1/query_range"

	// build full Prometheus URL
	PrometheusURL      string
	PrometheusRangeURL string

	prometheusIP   string
	prometheusPort int

	// query range metrics
	rangeQueries = map[string]string{
		"cpu": `100 * rate(libvirt_domain_info_cpu_time_seconds_total{domain=~"%s"}[2m]) / (2 * 5 * 60)`,
		//"memory_used":      `guest_kvm_vm_memory_usage_bytes{domain=~"%s"} / 1024 / 1024`,
		"memory_unused":    `libvirt_domain_memory_stats_unused_bytes{domain=~"%s"} / 1024 / 1024`,
		"memory_total":     `libvirt_domain_info_maximum_memory_bytes{domain=~"%s"} / 1024 / 1024`,
		"disk_read":        `rate(libvirt_domain_block_stats_read_bytes_total{domain=~"%s",target_device=~"%s"}[2m]) / 1024`,
		"disk_write":       `rate(libvirt_domain_block_stats_write_bytes_total{domain=~"%s",target_device=~"%s"}[2m]) / 1024`,
		"network_receive":  `rate(libvirt_domain_interface_stats_receive_bytes_total{domain=~"%s",target_device=~"%s"}[1m]) * 8 / 1024`,
		"network_transmit": `rate(libvirt_domain_interface_stats_transmit_bytes_total{domain=~"%s",target_device=~"%s"}[1m]) * 8 / 1024`,
		"traffic":          `(rate(libvirt_domain_interface_stats_receive_bytes_total{domain=~"%s",target_device=~"%s"}[1m]) * 1440) / (1024 * 1024)`, // ingress only
	}
)

func init() {
	viper.SetConfigFile("conf/config.toml")
	if err := viper.ReadInConfig(); err != nil {
		fmt.Printf("Failed to load configuration file %+v", err)
		prometheusIP = "localhost"
		prometheusPort = 9090
	} else {
		prometheusIP = viper.GetString("monitor.host")
		prometheusPort = viper.GetInt("monitor.port")
		logger.Info("prometheusIP: %s,  prometheusPort: %d", prometheusIP, prometheusPort)
	}
	if prometheusIP == "" {
		prometheusIP = "localhost"
	}
	if prometheusPort == 0 {
		prometheusPort = 9090
	}

	// init Prometheus URL
	PrometheusURL = fmt.Sprintf("http://%s:%d%s", prometheusIP, prometheusPort, PrometheusQueryPath)
	PrometheusRangeURL = fmt.Sprintf("http://%s:%d%s", prometheusIP, prometheusPort, PrometheusQueryRangePath)
}

type PrometheusResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Result     []struct {
			Metric map[string]string `json:"metric"`
			Values [][]interface{}   `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

type MetricsRequest struct {
	Start   string   `json:"start" binding:"required"`
	End     string   `json:"end" binding:"required"`
	Step    string   `json:"step" binding:"required"`
	ID      []string `json:"id"`
	Disk    []string `json:"disk"`
	Network []string `json:"network"`
}

// 1. CPU monitor - single metric
type CPUResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Label      string `json:"label"` // "CPU Utilization Rate (%)"
		Unit       string `json:"unit"`  // "%"
		Result     []struct {
			Metric struct {
				Domain   string `json:"domain"`
				Instance string `json:"instance"`
				Job      string `json:"job"`
			} `json:"metric"`
			Values []struct { // one-dimensional array
				Time  string `json:"time"`
				Value string `json:"value"`
			} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// 2. memory monitor - double metrics (total and used)
type MemoryResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string   `json:"resultType"`
		ChartType  string   `json:"chart_type"` // "bar"
		Label      []string `json:"label"`      // ["total(MB)", "used(MB)"]
		Unit       string   `json:"unit"`       // "MB"
		Result     []struct {
			Metric struct {
				Domain   string `json:"domain"`
				Instance string `json:"instance"`
				Job      string `json:"job"`
			} `json:"metric"`
			Values [][]struct { // two-dimensional array [total, used]
				Time  string `json:"time"`
				Value string `json:"value"`
			} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// 3. disk monitor - double metrics (read and write speed)
type DiskResponse struct {
	Status string `json:"status"`
	Data   struct {
		ChartType string   `json:"chart_type"` // "line"
		Label     []string `json:"label"`      // ["read(KB/s)", "write(KB/s)"]
		Unit      string   `json:"unit"`       // "KB/s"
		Result    []struct {
			Metric struct {
				Domain   string `json:"domain"`
				Instance string `json:"instance"`
				Job      string `json:"job"`
			} `json:"metric"`
			Values [][]struct { // two-dimensional array [read speed, write speed]
				Time  string `json:"time"`
				Value string `json:"value"`
			} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// 4. network monitor - double metrics (inbound and outbound bandwidth)
type NetworkResponse struct {
	Status string `json:"status"`
	Data   struct {
		ChartType  string   `json:"chart_type"` // "line"
		Label      []string `json:"label"`      // ["receive(KB/s)", "transmit(KB/s)"]
		Unit       string   `json:"unit"`       // "KB/s"
		ResultType string   `json:"resultType"`
		Result     []struct {
			Metric struct {
				Domain       string `json:"domain"`
				Instance     string `json:"instance"`
				Job          string `json:"job"`
				TargetDevice string `json:"target_device"`
			} `json:"metric"`
			Values [][]struct { // two-dimensional array [receive speed, transmit speed]
				Time  string `json:"time"`
				Value string `json:"value"`
			} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

// 5. traffic monitor - single metric
type TrafficResponse struct {
	Status string `json:"status"`
	Data   struct {
		ResultType string `json:"resultType"`
		Unit       string `json:"unit"` // "KB/s"
		Result     []struct {
			Metric struct {
				Domain       string `json:"domain"`
				Instance     string `json:"instance"`
				Job          string `json:"job"`
				TargetDevice string `json:"target_device"`
			} `json:"metric"`
			Values []struct { // one-dimensional array
				Time  string `json:"time"`
				Value string `json:"value"`
			} `json:"values"`
		} `json:"result"`
	} `json:"data"`
}

func parseUnixTime(timeStr string) (int64, error) {
	if timestamp, err := strconv.ParseInt(timeStr, 10, 64); err == nil {
		return timestamp, nil
	}
	return 0, fmt.Errorf("invalid time format: %s. Please use Unix timestamp (e.g., 1740382609)", timeStr)
}

const (
	QueryTypeInstant = "instant"
	QueryTypeRange   = "range"
)

func (api *MonitorAPI) getRangeQuery(metricType string, instanceIDs []string, deviceIDs []string) string {
	logger.Info("Building range query - metric type: %s, instance IDs: %v, device IDs: %v",
		metricType, instanceIDs, deviceIDs)

	if len(instanceIDs) == 0 {
		logger.Warning("No instance IDs provided")
		return ""
	}

	uuidFilter := strings.Join(instanceIDs, "|")
	logger.Info("UUID filter: %s", uuidFilter)

	query, ok := rangeQueries[metricType]
	if !ok {
		logger.Error("Unknown range metric type: %s", metricType)
		return ""
	}

	var finalQuery string
	if metricType == "disk_read" || metricType == "disk_write" ||
		metricType == "network_receive" || metricType == "network_transmit" || metricType == "traffic" {
		if len(deviceIDs) == 0 {
			logger.Warning("No device IDs provided for %s metrics", metricType)
			return ""
		}
		deviceFilter := strings.Join(deviceIDs, "|")
		logger.Info("Device filter: %s", deviceFilter)
		finalQuery = fmt.Sprintf(query, uuidFilter, deviceFilter)
	} else {
		finalQuery = fmt.Sprintf(query, uuidFilter)
	}

	logger.Info("Generated query: %s", finalQuery)
	return finalQuery

}

func (api *MonitorAPI) GetTraffic(c *gin.Context) {
	var request MetricsRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	if len(request.ID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Instance ID is required"})
		return
	}

	if len(request.Network) == 0 {
		logger.Warning("Network interface not provided in request")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Network interface is required"})
		return
	}

	if len(request.ID) != len(request.Network) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "The number of instance IDs and network interfaces must be the same"})
		return
	}

	var instanceIDs []string
	for _, uuid := range request.ID {
		fmt.Printf("Attempting to convert UUID: %s\n", uuid)
		instanceID, err := routes.GetDBIndexByInstanceUUID(c, uuid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			logger.Errorf("failed to get instance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		fmt.Printf("Successfully converted UUID %s to instanceID %d\n", uuid, instanceID)
		instanceIDs = append(instanceIDs, "inst-"+strconv.Itoa(instanceID))
	}

	// validate time params
	start, end, err := validateAndParseTimeParams(request.Start, request.End, request.Step)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	var allResults []interface{}
	for i := range instanceIDs {
		convertedID := instanceIDs[i]
		networkDevice := request.Network[i]

		// generate query for the instance
		query := api.getRangeQuery("traffic", []string{convertedID}, []string{networkDevice})
		if query == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid metric type"})
			return
		}

		// execute Prometheus query
		result, err := queryPrometheus(query, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), request.Step)
		if err != nil {
			logger.Error("Failed to query traffic for %s (%s): %v", convertedID, networkDevice, err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to query metrics"})
			return
		}

		// format return result
		formattedResult := formatResponse(result, "traffic")
		if trafficResult, ok := formattedResult.(*TrafficResponse); ok {
			for i := range trafficResult.Data.Result {
				trafficResult.Data.Result[i].Metric = struct {
					Domain       string `json:"domain"`
					Instance     string `json:"instance"`
					Job          string `json:"job"`
					TargetDevice string `json:"target_device"`
				}{
					Domain:       trafficResult.Data.Result[i].Metric.Domain,
					Instance:     trafficResult.Data.Result[i].Metric.Instance,
					Job:          trafficResult.Data.Result[i].Metric.Job,
					TargetDevice: trafficResult.Data.Result[i].Metric.TargetDevice,
				}
			}
		}
		allResults = append(allResults, formattedResult)
	}
	c.JSON(http.StatusOK, allResults)
}

func (api *MonitorAPI) GetCPU(c *gin.Context) {
	var request MetricsRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	if len(request.ID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Instance ID is required"})
		return
	}

	// convert UUID to index ID
	var instanceIDs []string
	for _, uuid := range request.ID {
		fmt.Printf("Attempting to convert UUID: %s\n", uuid)
		instanceID, err := routes.GetDBIndexByInstanceUUID(c, uuid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			logger.Errorf("failed to get instance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		fmt.Printf("Successfully converted UUID %s to instanceID %d\n", uuid, instanceID)
		instanceIDs = append(instanceIDs, "inst-"+strconv.Itoa(instanceID))
	}

	// validate time params
	start, end, err := validateAndParseTimeParams(request.Start, request.End, request.Step)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// build query
	query := api.getRangeQuery("cpu", instanceIDs, nil)
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid metric type"})
		return
	}

	// execute query
	result, err := queryPrometheus(query, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), request.Step)
	if err != nil {
		logger.Error("Failed to query CPU: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to query metrics"})
		return
	}

	// format result
	formattedResult := formatResponse(result, "cpu")
	if formattedResult != nil {
		if cpuResp, ok := formattedResult.(*CPUResponse); ok {
			for i := range cpuResp.Data.Result {
				cpuResp.Data.Result[i].Metric = struct {
					Domain   string `json:"domain"`
					Instance string `json:"instance"`
					Job      string `json:"job"`
				}{
					Domain:   cpuResp.Data.Result[i].Metric.Domain,
					Instance: cpuResp.Data.Result[i].Metric.Instance,
					Job:      cpuResp.Data.Result[i].Metric.Job,
				}
			}
		}
	}
	c.JSON(http.StatusOK, formattedResult)
}

func (api *MonitorAPI) GetMemory(c *gin.Context) {
	var request MetricsRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	if len(request.ID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Instance ID is required"})
		return
	}

	var instanceIDs []string
	for _, uuid := range request.ID {
		fmt.Printf("Attempting to convert UUID: %s\n", uuid)
		instanceID, err := routes.GetDBIndexByInstanceUUID(c, uuid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			logger.Errorf("failed to get instance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		fmt.Printf("Successfully converted UUID %s to instanceID %d\n", uuid, instanceID)
		instanceIDs = append(instanceIDs, "inst-"+strconv.Itoa(instanceID))
	}

	// validate time params
	start, end, err := validateAndParseTimeParams(request.Start, request.End, request.Step)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// build memory unused and total query
	unusedQuery := api.getRangeQuery("memory_unused", instanceIDs, nil)
	totalQuery := api.getRangeQuery("memory_total", instanceIDs, nil)
	if unusedQuery == "" || totalQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid metric type"})
		return
	}

	// execute query
	unusedResult, err := queryPrometheus(unusedQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), request.Step)
	if err != nil {
		logger.Error("Failed to query memory unused: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to query metrics"})
		return
	}

	totalResult, err := queryPrometheus(totalQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), request.Step)
	if err != nil {
		logger.Error("Failed to query memory total: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to query metrics"})
		return
	}

	// merge results
	result := mergeMemoryResults(unusedResult, totalResult)
	for i := range result.Data.Result {
		result.Data.Result[i].Metric = struct {
			Domain   string `json:"domain"`
			Instance string `json:"instance"`
			Job      string `json:"job"`
		}{
			Domain:   result.Data.Result[i].Metric.Domain,
			Instance: result.Data.Result[i].Metric.Instance,
			Job:      result.Data.Result[i].Metric.Job,
		}
	}
	c.JSON(http.StatusOK, result)
}

func (api *MonitorAPI) GetDisk(c *gin.Context) {
	var request MetricsRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	if len(request.ID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Instance ID is required"})
		return
	}

	if len(request.Disk) == 0 {
		logger.Warning("Disk ID not provided in request")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Disk ID is required"})
		return
	}

	var instanceIDs []string
	for _, uuid := range request.ID {
		fmt.Printf("Attempting to convert UUID: %s\n", uuid)
		instanceID, err := routes.GetDBIndexByInstanceUUID(c, uuid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			logger.Errorf("failed to get instance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		fmt.Printf("Successfully converted UUID %s to instanceID %d\n", uuid, instanceID)
		instanceIDs = append(instanceIDs, "inst-"+strconv.Itoa(instanceID))
	}

	// validate time params
	start, end, err := validateAndParseTimeParams(request.Start, request.End, request.Step)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	// build read and write query
	readQuery := api.getRangeQuery("disk_read", instanceIDs, request.Disk)
	writeQuery := api.getRangeQuery("disk_write", instanceIDs, request.Disk)
	fmt.Println("Read Query:", readQuery)
	fmt.Println("Write Query:", writeQuery)
	if readQuery == "" || writeQuery == "" {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid metric type"})
		return
	}

	// execute query
	readResult, err := queryPrometheus(readQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), request.Step)
	if err != nil {
		logger.Error("Failed to query disk read metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to query metrics"})
		return
	}

	writeResult, err := queryPrometheus(writeQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), request.Step)
	if err != nil {
		logger.Error("Failed to query disk write metrics: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to query metrics"})
		return
	}

	// merge results
	result := mergeDiskResults(readResult, writeResult)
	for i := range result.Data.Result {
		result.Data.Result[i].Metric = struct {
			Domain   string `json:"domain"`
			Instance string `json:"instance"`
			Job      string `json:"job"`
		}{
			Domain:   result.Data.Result[i].Metric.Domain,
			Instance: result.Data.Result[i].Metric.Instance,
			Job:      result.Data.Result[i].Metric.Job,
		}
	}
	c.JSON(http.StatusOK, result)
}

func (api *MonitorAPI) GetNetwork(c *gin.Context) {
	var request MetricsRequest
	if err := c.BindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid request body"})
		return
	}

	if len(request.ID) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Instance ID is required"})
		return
	}

	if len(request.Network) == 0 {
		logger.Warning("Network interface not provided in request")
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Network interface is required"})
		return
	}
	if len(request.ID) != len(request.Network) {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "The number of instance IDs and network interfaces must be the same"})
		return
	}

	var instanceIDs []string
	for _, uuid := range request.ID {
		fmt.Printf("Attempting to convert UUID: %s\n", uuid)
		instanceID, err := routes.GetDBIndexByInstanceUUID(c, uuid)
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
				return
			}
			logger.Errorf("failed to get instance: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Internal Server Error"})
			return
		}
		fmt.Printf("Successfully converted UUID %s to instanceID %d\n", uuid, instanceID)
		instanceIDs = append(instanceIDs, "inst-"+strconv.Itoa(instanceID))
	}

	// validate time params
	start, end, err := validateAndParseTimeParams(request.Start, request.End, request.Step)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": err.Error()})
		return
	}

	var allResults []interface{}
	for i, id := range instanceIDs {
		network := request.Network[i]

		// build inbound and outbound bandwidth query
		receiveQuery := api.getRangeQuery("network_receive", []string{id}, []string{network})
		transmitQuery := api.getRangeQuery("network_transmit", []string{id}, []string{network})
		if receiveQuery == "" || transmitQuery == "" {
			c.JSON(http.StatusBadRequest, gin.H{"status": "error", "message": "Invalid metric type"})
			return
		}

		// execute query
		receiveResult, err := queryPrometheus(receiveQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), request.Step)
		if err != nil {
			logger.Error("Failed to query network receive: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to query metrics"})
			return
		}

		transmitResult, err := queryPrometheus(transmitQuery, fmt.Sprintf("%d", start), fmt.Sprintf("%d", end), request.Step)
		if err != nil {
			logger.Error("Failed to query network transmit: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"status": "error", "message": "Failed to query metrics"})
			return
		}

		// merge results
		result := mergeNetworkResults(receiveResult, transmitResult)
		for i := range result.Data.Result {
			result.Data.Result[i].Metric = struct {
				Domain       string `json:"domain"`
				Instance     string `json:"instance"`
				Job          string `json:"job"`
				TargetDevice string `json:"target_device"`
			}{
				Domain:       result.Data.Result[i].Metric.Domain,
				Instance:     result.Data.Result[i].Metric.Instance,
				Job:          result.Data.Result[i].Metric.Job,
				TargetDevice: result.Data.Result[i].Metric.TargetDevice,
			}
		}
		allResults = append(allResults, result)
	}

	c.JSON(http.StatusOK, allResults)
}

func validateTimeParams(start, end int64, step string) error {
	if start >= end {
		return fmt.Errorf("start time must be less than end time")
	}
	duration, err := time.ParseDuration(step)
	if err != nil {
		return fmt.Errorf("invalid step format: %s", err)
	}
	if duration < time.Second {
		return fmt.Errorf("step must be at least 1 second")
	}
	return nil
}

func validateAndParseTimeParams(startStr, endStr, step string) (int64, int64, error) {
	start, err := parseUnixTime(startStr)
	if err != nil {
		return 0, 0, err
	}
	end, err := parseUnixTime(endStr)
	if err != nil {
		return 0, 0, err
	}
	if err := validateTimeParams(start, end, step); err != nil {
		return 0, 0, err
	}
	return start, end, nil
}

func queryPrometheus(query string, start, end, step string) (*PrometheusResponse, error) {
	// record query params
	logger.Info("Prometheus query: %s, start: %s, end: %s, step: %s", query, start, end, step)

	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", PrometheusRangeURL, nil)
	if err != nil {
		return nil, err
	}
	q := req.URL.Query()
	q.Add("query", query)
	q.Add("start", start)
	q.Add("end", end)
	q.Add("step", step)
	req.URL.RawQuery = q.Encode()

	// record full request URL
	logger.Info("Prometheus request URL: %s", req.URL.String())

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var prometheusResp PrometheusResponse
	if err := json.NewDecoder(resp.Body).Decode(&prometheusResp); err != nil {
		return nil, err
	}

	// record query result
	resultBytes, _ := json.MarshalIndent(prometheusResp, "", "  ")
	logger.Debug("Prometheus response: %s", string(resultBytes))

	return &prometheusResp, nil
}

// merge memory query results
func mergeMemoryResults(unused, total *PrometheusResponse) *MemoryResponse {
	var memResp MemoryResponse
	memResp.Status = "success"
	memResp.Data.ResultType = "matrix"
	memResp.Data.ChartType = "bar"
	memResp.Data.Label = []string{"Total(MB)", "Used(MB)"}
	memResp.Data.Unit = "MB"

	// Use map to match results with the same domain
	resultMap := make(map[string]struct {
		metric struct {
			Domain   string
			Instance string
			Job      string
		}
		unusedValues [][]interface{}
		totalValues  [][]interface{}
	})

	// Collect unused data
	for _, r := range unused.Data.Result {
		domain := r.Metric["domain"]
		resultMap[domain] = struct {
			metric struct {
				Domain   string
				Instance string
				Job      string
			}
			unusedValues [][]interface{}
			totalValues  [][]interface{}
		}{
			metric: struct {
				Domain   string
				Instance string
				Job      string
			}{
				Domain:   r.Metric["domain"],
				Instance: r.Metric["instance"],
				Job:      r.Metric["job"],
			},
			unusedValues: r.Values,
		}
	}

	// Merge total data
	for _, r := range total.Data.Result {
		domain := r.Metric["domain"]
		if item, exists := resultMap[domain]; exists {
			item.totalValues = r.Values
			resultMap[domain] = item
		}
	}

	// Build final result
	for _, item := range resultMap {
		var result struct {
			Metric struct {
				Domain   string `json:"domain"`
				Instance string `json:"instance"`
				Job      string `json:"job"`
			} `json:"metric"`
			Values [][]struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			} `json:"values"`
		}

		result.Metric = struct {
			Domain   string `json:"domain"`
			Instance string `json:"instance"`
			Job      string `json:"job"`
		}{
			Domain:   item.metric.Domain,
			Instance: item.metric.Instance,
			Job:      item.metric.Job,
		}

		// Process total and unused values to calculate used memory
		var totalValues, usedValues []struct {
			Time  string `json:"time"`
			Value string `json:"value"`
		}

		for i := 0; i < len(item.totalValues) && i < len(item.unusedValues); i++ {
			totalTime := item.totalValues[i][0].(float64)
			totalVal, _ := strconv.ParseFloat(item.totalValues[i][1].(string), 64)
			unusedVal, _ := strconv.ParseFloat(item.unusedValues[i][1].(string), 64)
			usedVal := totalVal - unusedVal

			totalValues = append(totalValues, struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{
				Time:  time.Unix(int64(totalTime), 0).Format("2006-01-02 15:04:05"),
				Value: fmt.Sprintf("%.2f", totalVal),
			})

			usedValues = append(usedValues, struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{
				Time:  time.Unix(int64(totalTime), 0).Format("2006-01-02 15:04:05"),
				Value: fmt.Sprintf("%.2f", usedVal),
			})
		}

		result.Values = [][]struct {
			Time  string `json:"time"`
			Value string `json:"value"`
		}{totalValues, usedValues}

		memResp.Data.Result = append(memResp.Data.Result, result)
	}

	return &memResp
}

func mergeDiskResults(read, write *PrometheusResponse) *DiskResponse {
	var diskResp DiskResponse
	diskResp.Status = "success"
	diskResp.Data.ChartType = "line"
	diskResp.Data.Label = []string{"读取速度 (KB/s)", "写入速度 (KB/s)"}
	diskResp.Data.Unit = "KB/s"

	resultMap := make(map[string]struct {
		metric struct {
			Domain   string
			Instance string
			Job      string
		}
		readValues  [][]interface{}
		writeValues [][]interface{}
	})

	// collect read data
	for _, r := range read.Data.Result {
		uuid := r.Metric["uuid"]
		resultMap[uuid] = struct {
			metric struct {
				Domain   string
				Instance string
				Job      string
			}
			readValues  [][]interface{}
			writeValues [][]interface{}
		}{
			metric: struct {
				Domain   string
				Instance string
				Job      string
			}{
				Domain:   r.Metric["domain"],
				Instance: r.Metric["instance"],
				Job:      r.Metric["job"],
			},
			readValues: r.Values,
		}
	}

	// merge write data
	for _, r := range write.Data.Result {
		uuid := r.Metric["uuid"]
		if item, exists := resultMap[uuid]; exists {
			item.writeValues = r.Values
			resultMap[uuid] = item
		}
	}

	// build final result
	for _, item := range resultMap {
		var result struct {
			Metric struct {
				Domain   string `json:"domain"`
				Instance string `json:"instance"`
				Job      string `json:"job"`
			} `json:"metric"`
			Values [][]struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			} `json:"values"`
		}

		result.Metric = struct {
			Domain   string `json:"domain"`
			Instance string `json:"instance"`
			Job      string `json:"job"`
		}{
			Domain:   item.metric.Domain,
			Instance: item.metric.Instance,
			Job:      item.metric.Job,
		}

		// process read data
		var readValues []struct {
			Time  string `json:"time"`
			Value string `json:"value"`
		}
		for _, v := range item.readValues {
			timestamp := v[0].(float64)
			value := v[1].(string)
			readValues = append(readValues, struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{
				Time:  time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05"),
				Value: value,
			})
		}

		// process write data
		var writeValues []struct {
			Time  string `json:"time"`
			Value string `json:"value"`
		}
		for _, v := range item.writeValues {
			timestamp := v[0].(float64)
			value := v[1].(string)
			writeValues = append(writeValues, struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{
				Time:  time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05"),
				Value: value,
			})
		}

		result.Values = [][]struct {
			Time  string `json:"time"`
			Value string `json:"value"`
		}{readValues, writeValues}

		diskResp.Data.Result = append(diskResp.Data.Result, result)
	}

	return &diskResp
}

// merge network query results
func mergeNetworkResults(receive, transmit *PrometheusResponse) *NetworkResponse {
	var netResp NetworkResponse
	netResp.Status = "success"
	netResp.Data.ChartType = "line"
	netResp.Data.Label = []string{"Receive Speed (KB/s)", "Transmit Speed (KB/s)"}
	netResp.Data.Unit = "KB/s"
	netResp.Data.ResultType = receive.Data.ResultType

	// use map to match results with the same UUID
	resultMap := make(map[string]struct {
		metric struct {
			Domain       string
			Instance     string
			Job          string
			TargetDevice string
		}
		receiveValues  [][]interface{}
		transmitValues [][]interface{}
	})

	// collect receive data
	for _, r := range receive.Data.Result {
		uuid := r.Metric["uuid"]
		resultMap[uuid] = struct {
			metric struct {
				Domain       string
				Instance     string
				Job          string
				TargetDevice string
			}
			receiveValues  [][]interface{}
			transmitValues [][]interface{}
		}{
			metric: struct {
				Domain       string
				Instance     string
				Job          string
				TargetDevice string
			}{
				Domain:       r.Metric["domain"],
				Instance:     r.Metric["instance"],
				Job:          r.Metric["job"],
				TargetDevice: r.Metric["target_device"],
			},
			receiveValues: r.Values,
		}
	}

	// merge transmit data
	for _, r := range transmit.Data.Result {
		uuid := r.Metric["uuid"]
		if item, exists := resultMap[uuid]; exists {
			item.transmitValues = r.Values
			resultMap[uuid] = item
		}
	}

	// build final result
	for _, item := range resultMap {
		var result struct {
			Metric struct {
				Domain       string `json:"domain"`
				Instance     string `json:"instance"`
				Job          string `json:"job"`
				TargetDevice string `json:"target_device"`
			} `json:"metric"`
			Values [][]struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			} `json:"values"`
		}

		result.Metric = struct {
			Domain       string `json:"domain"`
			Instance     string `json:"instance"`
			Job          string `json:"job"`
			TargetDevice string `json:"target_device"`
		}{
			Domain:       item.metric.Domain,
			Instance:     item.metric.Instance,
			Job:          item.metric.Job,
			TargetDevice: item.metric.TargetDevice,
		}

		// process receive data
		var receiveValues []struct {
			Time  string `json:"time"`
			Value string `json:"value"`
		}
		for _, v := range item.receiveValues {
			timestamp := v[0].(float64)
			value := v[1].(string)
			timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")

			receiveValues = append(receiveValues, struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{
				Time:  timeStr,
				Value: value,
			})
		}

		// process transmit data
		var transmitValues []struct {
			Time  string `json:"time"`
			Value string `json:"value"`
		}
		for _, v := range item.transmitValues {
			timestamp := v[0].(float64)
			value := v[1].(string)
			timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")

			transmitValues = append(transmitValues, struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{
				Time:  timeStr,
				Value: value,
			})
		}

		result.Values = [][]struct {
			Time  string `json:"time"`
			Value string `json:"value"`
		}{receiveValues, transmitValues}

		netResp.Data.Result = append(netResp.Data.Result, result)
	}

	return &netResp
}

// add response format conversion function
func formatResponse(resp *PrometheusResponse, metricType string) interface{} {
	switch metricType {
	case "cpu":
		// CPU monitoring - single metric, use one-dimensional array
		var cpuResp CPUResponse
		cpuResp.Status = resp.Status
		cpuResp.Data.ResultType = resp.Data.ResultType
		cpuResp.Data.Label = "CPU Utilization Rate (%)"
		cpuResp.Data.Unit = "%"

		for _, r := range resp.Data.Result {
			var result struct {
				Metric struct {
					Domain   string `json:"domain"`
					Instance string `json:"instance"`
					Job      string `json:"job"`
				} `json:"metric"`
				Values []struct { // one-dimensional array
					Time  string `json:"time"`
					Value string `json:"value"`
				} `json:"values"`
			}

			result.Metric.Domain = r.Metric["domain"]
			result.Metric.Instance = r.Metric["instance"]
			result.Metric.Job = r.Metric["job"]

			var values []struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}

			for _, v := range r.Values {
				timestamp := v[0].(float64)
				value := v[1].(string)
				timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
				values = append(values, struct {
					Time  string `json:"time"`
					Value string `json:"value"`
				}{
					Time:  timeStr,
					Value: value,
				})
			}

			result.Values = values // directly assign one-dimensional array
			cpuResp.Data.Result = append(cpuResp.Data.Result, result)
		}
		return cpuResp

	case "disk_read":
		var diskResp DiskResponse
		diskResp.Status = resp.Status
		diskResp.Data.ChartType = "line"
		diskResp.Data.Label = []string{"Read Speed (KB/s)", "Write Speed (KB/s)"}
		diskResp.Data.Unit = "KB/s"

		for _, r := range resp.Data.Result {
			var result struct {
				Metric struct {
					Domain   string `json:"domain"`
					Instance string `json:"instance"`
					Job      string `json:"job"`
				} `json:"metric"`
				Values [][]struct { // two-dimensional array
					Time  string `json:"time"`
					Value string `json:"value"`
				} `json:"values"`
			}

			result.Metric.Domain = r.Metric["domain"]
			result.Metric.Instance = r.Metric["instance"]
			result.Metric.Job = r.Metric["job"]

			var readValues []struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}

			for _, v := range r.Values {
				timestamp := v[0].(float64)
				value := v[1].(string)
				timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
				readValues = append(readValues, struct {
					Time  string `json:"time"`
					Value string `json:"value"`
				}{
					Time:  timeStr,
					Value: value,
				})
			}

			result.Values = [][]struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{readValues} // wrap as two-dimensional array
			diskResp.Data.Result = append(diskResp.Data.Result, result)
		}
		return diskResp

	case "disk_write":
		var diskResp DiskResponse
		diskResp.Status = resp.Status
		diskResp.Data.ChartType = "line"
		diskResp.Data.Label = []string{"读取速度 (KB/s)", "写入速度 (KB/s)"}
		diskResp.Data.Unit = "KB/s"

		for _, r := range resp.Data.Result {
			var result struct {
				Metric struct {
					Domain   string `json:"domain"`
					Instance string `json:"instance"`
					Job      string `json:"job"`
				} `json:"metric"`
				Values [][]struct { // two-dimensional array
					Time  string `json:"time"`
					Value string `json:"value"`
				} `json:"values"`
			}

			result.Metric.Domain = r.Metric["domain"]
			result.Metric.Instance = r.Metric["instance"]
			result.Metric.Job = r.Metric["job"]

			var writeValues []struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}

			for _, v := range r.Values {
				timestamp := v[0].(float64)
				value := v[1].(string)
				timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
				writeValues = append(writeValues, struct {
					Time  string `json:"time"`
					Value string `json:"value"`
				}{
					Time:  timeStr,
					Value: value,
				})
			}

			result.Values = [][]struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{writeValues} // wrap as two-dimensional array
			diskResp.Data.Result = append(diskResp.Data.Result, result)
		}
		return diskResp

	case "network_receive":
		var networkResp NetworkResponse
		networkResp.Status = resp.Status
		networkResp.Data.ChartType = "line"
		networkResp.Data.Label = []string{"接收速度 (KB/s)", "发送速度 (KB/s)"}
		networkResp.Data.Unit = "KB/s"
		networkResp.Data.ResultType = resp.Data.ResultType

		for _, r := range resp.Data.Result {
			var result struct {
				Metric struct {
					Domain       string `json:"domain"`
					Instance     string `json:"instance"`
					Job          string `json:"job"`
					TargetDevice string `json:"target_device"`
				} `json:"metric"`
				Values [][]struct { // two-dimensional array
					Time  string `json:"time"`
					Value string `json:"value"`
				} `json:"values"`
			}

			result.Metric.Domain = r.Metric["domain"]
			result.Metric.Instance = r.Metric["instance"]
			result.Metric.Job = r.Metric["job"]
			result.Metric.TargetDevice = r.Metric["target_device"]

			var receiveValues []struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}

			for _, v := range r.Values {
				timestamp := v[0].(float64)
				value := v[1].(string)
				timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
				receiveValues = append(receiveValues, struct {
					Time  string `json:"time"`
					Value string `json:"value"`
				}{
					Time:  timeStr,
					Value: value,
				})
			}

			result.Values = [][]struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{receiveValues} // wrap as two-dimensional array
			networkResp.Data.Result = append(networkResp.Data.Result, result)
		}
		return networkResp

	case "network_transmit":
		var networkResp NetworkResponse
		networkResp.Status = resp.Status
		networkResp.Data.ChartType = "line"
		networkResp.Data.Label = []string{"接收速度 (KB/s)", "发送速度 (KB/s)"}
		networkResp.Data.Unit = "KB/s"
		networkResp.Data.ResultType = resp.Data.ResultType

		for _, r := range resp.Data.Result {
			var result struct {
				Metric struct {
					Domain       string `json:"domain"`
					Instance     string `json:"instance"`
					Job          string `json:"job"`
					TargetDevice string `json:"target_device"`
				} `json:"metric"`
				Values [][]struct { // two-dimensional array
					Time  string `json:"time"`
					Value string `json:"value"`
				} `json:"values"`
			}

			result.Metric.Domain = r.Metric["domain"]
			result.Metric.Instance = r.Metric["instance"]
			result.Metric.Job = r.Metric["job"]
			result.Metric.TargetDevice = r.Metric["target_device"]

			var transmitValues []struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}

			for _, v := range r.Values {
				timestamp := v[0].(float64)
				value := v[1].(string)
				timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
				transmitValues = append(transmitValues, struct {
					Time  string `json:"time"`
					Value string `json:"value"`
				}{
					Time:  timeStr,
					Value: value,
				})
			}

			result.Values = [][]struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}{transmitValues} // wrap as two-dimensional array
			networkResp.Data.Result = append(networkResp.Data.Result, result)
		}
		return networkResp

	case "traffic":
		// traffic monitoring - single metric, use one-dimensional array
		var trafficResp TrafficResponse
		trafficResp.Status = resp.Status
		trafficResp.Data.ResultType = resp.Data.ResultType
		trafficResp.Data.Unit = "KB/s"

		for _, r := range resp.Data.Result {
			var result struct {
				Metric struct {
					Domain       string `json:"domain"`
					Instance     string `json:"instance"`
					Job          string `json:"job"`
					TargetDevice string `json:"target_device"`
				} `json:"metric"`
				Values []struct { // one-dimensional array
					Time  string `json:"time"`
					Value string `json:"value"`
				} `json:"values"`
			}

			result.Metric.Domain = r.Metric["domain"]
			result.Metric.Instance = r.Metric["instance"]
			result.Metric.Job = r.Metric["job"]
			result.Metric.TargetDevice = r.Metric["target_device"]

			var values []struct {
				Time  string `json:"time"`
				Value string `json:"value"`
			}

			for _, v := range r.Values {
				timestamp := v[0].(float64)
				value := v[1].(string)
				timeStr := time.Unix(int64(timestamp), 0).Format("2006-01-02 15:04:05")
				values = append(values, struct {
					Time  string `json:"time"`
					Value string `json:"value"`
				}{
					Time:  timeStr,
					Value: value,
				})
			}

			result.Values = values // directly assign one-dimensional array
			trafficResp.Data.Result = append(trafficResp.Data.Result, result)
		}
		return trafficResp

	default:
		logger.Error("Unknown metric type: %s", metricType)
		return nil
	}
}
