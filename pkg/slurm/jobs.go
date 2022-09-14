package slurm

import (
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/prometheus/client_golang/prometheus"
)

const (
	showJobsCommand       = "squeue -a --json"
	showJobsTestDataInput = "./test_data/jobs.json"
	showJobsTestDataProm  = "./test_data/jobs.prom"

	minHistogramBucketRange  = 1              // 1 second
	maxHistogramBucketRange  = 3600 * 24 * 14 //14 days
	numberOfHistogramBuckets = 15
)

var (
	jobLabels       = []string{"name", "job_id", "state", "state_reason", "partition", "user", "node"}
	durationBuckets = prometheus.ExponentialBucketsRange(minHistogramBucketRange, maxHistogramBucketRange, numberOfHistogramBuckets)
)

type jobsCollector struct {
	jobsInfo             *prometheus.GaugeVec
	jobsReqCPU           *prometheus.GaugeVec
	jobsReqMemory        *prometheus.GaugeVec
	jobsReqBilling       *prometheus.GaugeVec
	jobsReqNodes         *prometheus.GaugeVec
	jobsRestartCount     *prometheus.GaugeVec
	jobExecDuration      *prometheus.HistogramVec
	jobSchedlingDuration *prometheus.HistogramVec
	isTest               bool
}

func NewJobsCollector(isTest bool) *jobsCollector {
	return &jobsCollector{
		isTest: isTest,
		jobsInfo: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_job_info",
				Help:      "General informations about slurm jobs.",
			}, jobLabels),
		jobExecDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: "",
				Name:      "slurm_job_exec_duration",
				Help:      "Slurm job execution duration only for COMPLETED jobs.",
				Buckets:   durationBuckets,
			}, jobLabels),
		jobSchedlingDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Subsystem: "",
				Name:      "slurm_job_scheduling_duration",
				Help:      "Slurm job scheduling duration only for COMPLETED or RUNNING jobs.",
				Buckets:   durationBuckets,
			}, jobLabels),
		jobsReqCPU: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_job_req_cpu",
				Help:      "Requested CPU per job.",
			}, jobLabels),
		jobsReqMemory: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_job_req_memory_bytes",
				Help:      "Requested Memory per job.",
			}, jobLabels),
		jobsReqBilling: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_job_req_billing",
				Help:      "Requested billing per job.",
			}, jobLabels),
		jobsReqNodes: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_job_req_nodes",
				Help:      "Requested Nodes per job.",
			}, jobLabels),
		jobsRestartCount: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Subsystem: "",
				Name:      "slurm_job_restart_count",
				Help:      "Requested Restart count per job.",
			}, jobLabels),
	}
}

func (s *jobsCollector) getJobsMetrics() {
	// Get json data
	data := getData(s.isTest, showJobsCommand, showJobsTestDataInput)

	// Parse json
	squeueJson := &SqueuOutput{}
	err := json.Unmarshal([]byte(data), squeueJson)
	if err != nil {
		ExporterErrors.WithLabelValues("json-encoding-sqeueu", err.Error()).Inc()
		fmt.Println(err)
	}
	// create metrics from json object
	for _, job := range squeueJson.Jobs {
		user := job.UserName
		if user == "" {
			user = strconv.Itoa(job.UserID)
		}
		labelValues := []string{job.Name, strconv.Itoa(job.JobID), job.JobState, job.StateReason, job.Partition, user, job.Nodes}
		s.jobsInfo.WithLabelValues(labelValues...).Set(1)
		s.jobsRestartCount.WithLabelValues(labelValues...).Set(float64(job.RestartCnt))
		s.jobsReqCPU.WithLabelValues(labelValues...).Set(float64(job.Cpus))
		s.jobsReqMemory.WithLabelValues(labelValues...).Set(float64(job.MemoryPerCPU * job.Cpus))
		s.jobsReqNodes.WithLabelValues(labelValues...).Set(float64(job.NodeCount))
		s.jobsReqBilling.WithLabelValues(labelValues...).Set(job.BillableTres)
		if job.StartTime != 0 {
			schedulDuration := float64(job.StartTime - job.SubmitTime)
			s.jobSchedlingDuration.WithLabelValues(labelValues...).Observe(schedulDuration)
		}
		if job.EndTime != 0 {
			execDuration := float64(job.EndTime - job.StartTime)
			s.jobExecDuration.WithLabelValues(labelValues...).Observe(execDuration)
		}
	}
}

func (s *jobsCollector) Describe(ch chan<- *prometheus.Desc) {
	s.jobsInfo.Describe(ch)
	s.jobExecDuration.Describe(ch)
	s.jobSchedlingDuration.Describe(ch)
	s.jobsReqCPU.Describe(ch)
	s.jobsReqMemory.Describe(ch)
	s.jobsReqBilling.Describe(ch)
	s.jobsReqNodes.Describe(ch)
	s.jobsRestartCount.Describe(ch)
}

func (s *jobsCollector) Collect(ch chan<- prometheus.Metric) {
	s.jobsInfo.Reset()
	s.jobExecDuration.Reset()
	s.jobSchedlingDuration.Reset()
	s.jobsReqCPU.Reset()
	s.jobsReqMemory.Reset()
	s.jobsReqBilling.Reset()
	s.jobsReqNodes.Reset()
	s.jobsRestartCount.Reset()
	s.getJobsMetrics()
	s.jobsInfo.Collect(ch)
	s.jobExecDuration.Collect(ch)
	s.jobSchedlingDuration.Collect(ch)
	s.jobsReqCPU.Collect(ch)
	s.jobsReqMemory.Collect(ch)
	s.jobsReqBilling.Collect(ch)
	s.jobsReqNodes.Collect(ch)
	s.jobsRestartCount.Collect(ch)
}

type SqueuOutput struct {
	Meta struct {
		Plugin struct {
			Type string `json:"type"`
			Name string `json:"name"`
		} `json:"plugin"`
		Slurm struct {
			Version struct {
				Major int `json:"major"`
				Micro int `json:"micro"`
				Minor int `json:"minor"`
			} `json:"version"`
			Release string `json:"release"`
		} `json:"Slurm"`
	} `json:"meta"`
	Errors []interface{} `json:"errors"`
	Jobs   []struct {
		Account                  string        `json:"account"`
		AccrueTime               int           `json:"accrue_time"`
		AdminComment             string        `json:"admin_comment"`
		ArrayJobID               int           `json:"array_job_id"`
		ArrayTaskID              interface{}   `json:"array_task_id"`
		ArrayMaxTasks            int           `json:"array_max_tasks"`
		ArrayTaskString          string        `json:"array_task_string"`
		AssociationID            int           `json:"association_id"`
		BatchFeatures            string        `json:"batch_features"`
		BatchFlag                bool          `json:"batch_flag"`
		BatchHost                string        `json:"batch_host"`
		Flags                    []string      `json:"flags"`
		BurstBuffer              string        `json:"burst_buffer"`
		BurstBufferState         string        `json:"burst_buffer_state"`
		Cluster                  string        `json:"cluster"`
		ClusterFeatures          string        `json:"cluster_features"`
		Command                  string        `json:"command"`
		Comment                  string        `json:"comment"`
		Contiguous               bool          `json:"contiguous"`
		CoreSpec                 interface{}   `json:"core_spec"`
		ThreadSpec               interface{}   `json:"thread_spec"`
		CoresPerSocket           interface{}   `json:"cores_per_socket"`
		BillableTres             float64       `json:"billable_tres"`
		CpusPerTask              interface{}   `json:"cpus_per_task"`
		CPUFrequencyMinimum      interface{}   `json:"cpu_frequency_minimum"`
		CPUFrequencyMaximum      interface{}   `json:"cpu_frequency_maximum"`
		CPUFrequencyGovernor     interface{}   `json:"cpu_frequency_governor"`
		CpusPerTres              string        `json:"cpus_per_tres"`
		Deadline                 int           `json:"deadline"`
		DelayBoot                int           `json:"delay_boot"`
		Dependency               string        `json:"dependency"`
		DerivedExitCode          int           `json:"derived_exit_code"`
		EligibleTime             int           `json:"eligible_time"`
		EndTime                  int           `json:"end_time"`
		ExcludedNodes            string        `json:"excluded_nodes"`
		ExitCode                 int           `json:"exit_code"`
		Features                 string        `json:"features"`
		FederationOrigin         string        `json:"federation_origin"`
		FederationSiblingsActive string        `json:"federation_siblings_active"`
		FederationSiblingsViable string        `json:"federation_siblings_viable"`
		GresDetail               []interface{} `json:"gres_detail"`
		GroupID                  int           `json:"group_id"`
		JobID                    int           `json:"job_id"`
		JobResources             struct {
		} `json:"job_resources"`
		JobState                string      `json:"job_state"`
		LastSchedEvaluation     int         `json:"last_sched_evaluation"`
		Licenses                string      `json:"licenses"`
		MaxCpus                 int         `json:"max_cpus"`
		MaxNodes                int         `json:"max_nodes"`
		McsLabel                string      `json:"mcs_label"`
		MemoryPerTres           string      `json:"memory_per_tres"`
		Name                    string      `json:"name"`
		Nodes                   string      `json:"nodes"`
		Nice                    int         `json:"nice"`
		TasksPerCore            interface{} `json:"tasks_per_core"`
		TasksPerNode            int         `json:"tasks_per_node"`
		TasksPerSocket          interface{} `json:"tasks_per_socket"`
		TasksPerBoard           int         `json:"tasks_per_board"`
		Cpus                    int         `json:"cpus"`
		NodeCount               int         `json:"node_count"`
		Tasks                   int         `json:"tasks"`
		HetJobID                int         `json:"het_job_id"`
		HetJobIDSet             string      `json:"het_job_id_set"`
		HetJobOffset            int         `json:"het_job_offset"`
		Partition               string      `json:"partition"`
		MemoryPerNode           interface{} `json:"memory_per_node"`
		MemoryPerCPU            int         `json:"memory_per_cpu"`
		MinimumCpusPerNode      int         `json:"minimum_cpus_per_node"`
		MinimumTmpDiskPerNode   int         `json:"minimum_tmp_disk_per_node"`
		PreemptTime             int         `json:"preempt_time"`
		PreSusTime              int         `json:"pre_sus_time"`
		Priority                int         `json:"priority"`
		Profile                 interface{} `json:"profile"`
		Qos                     string      `json:"qos"`
		Reboot                  bool        `json:"reboot"`
		RequiredNodes           string      `json:"required_nodes"`
		Requeue                 bool        `json:"requeue"`
		ResizeTime              int         `json:"resize_time"`
		RestartCnt              int         `json:"restart_cnt"`
		ResvName                string      `json:"resv_name"`
		Shared                  string      `json:"shared"`
		ShowFlags               []string    `json:"show_flags"`
		SocketsPerBoard         int         `json:"sockets_per_board"`
		SocketsPerNode          interface{} `json:"sockets_per_node"`
		StartTime               int         `json:"start_time"`
		StateDescription        string      `json:"state_description"`
		StateReason             string      `json:"state_reason"`
		StandardError           string      `json:"standard_error"`
		StandardInput           string      `json:"standard_input"`
		StandardOutput          string      `json:"standard_output"`
		SubmitTime              int         `json:"submit_time"`
		SuspendTime             int         `json:"suspend_time"`
		SystemComment           string      `json:"system_comment"`
		TimeLimit               int         `json:"time_limit"`
		TimeMinimum             int         `json:"time_minimum"`
		ThreadsPerCore          interface{} `json:"threads_per_core"`
		TresBind                string      `json:"tres_bind"`
		TresFreq                string      `json:"tres_freq"`
		TresPerJob              string      `json:"tres_per_job"`
		TresPerNode             string      `json:"tres_per_node"`
		TresPerSocket           string      `json:"tres_per_socket"`
		TresPerTask             string      `json:"tres_per_task"`
		TresReqStr              string      `json:"tres_req_str"`
		TresAllocStr            string      `json:"tres_alloc_str"`
		UserID                  int         `json:"user_id"`
		UserName                string      `json:"user_name"`
		Wckey                   string      `json:"wckey"`
		CurrentWorkingDirectory string      `json:"current_working_directory"`
	} `json:"jobs"`
}
