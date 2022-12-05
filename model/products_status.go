package model

type ServiceClassDefinition struct {
	Name                         string `db:"name" json:"name"`
	MaxDowntimeFrequencyAllowed  string `db:"max_downtime_frequency_allowed" json:"maxDowntimeFrequencyAllowed"`
	MaxContinuousDowntimeAllowed int64  `db:"max_continuous_downtime_allowed" json:"maxContinuousDowntimeAllowed"`
	AvailabilityUpperThreshold   string `db:"availability_upper_threshold" json:"availabilityUpperThreshold"`
	MinimumAvailability          string `db:"minimum_availability" json:"minimumAvailability"`
	ServiceTimeHours             int64  `db:"service_time_hours" json:"serviceTimeHours"`
	ServiceTimeDays              int64  `db:"service_time_days" json:"serviceTimeDays"`
}
type ProductStatus struct {
	MonitorID                   string                  `db:"monitor_id" json:"monitorId"`
	MarcID                      string                  `db:"marc_id" json:"marcId"`
	ProductName                 string                  `db:"product_name" json:"productName"`
	ServiceClass                string                  `db:"service_class" json:"-"`
	Level                       string                  `db:"level" json:"level"`
	ClassDefinition             *ServiceClassDefinition `json:"classDefinition"`
	Availability                string                  `db:"availability" json:"availability"`
	NumberOfDowntimes           int64                   `db:"number_of_downtimes" json:"numberOfDowntimes"`
	MaxContinuousDowntime       int64                   `db:"max_continuous_downtime" json:"maxContinuousDowntime"`
	AvailabilityStatus          string                  `db:"availability_status" json:"availabilityStatus"`
	NumberOfDowntimesStatus     string                  `db:"number_of_downtimes_status" json:"numberOfDowntimesStatus"`
	MaxContinuousDowntimeStatus string                  `db:"max_continuous_downtime_status" json:"maxContinuousDowntimeStatus"`
	MonitorStatus               string                  `db:"monitor_status" json:"monitorStatus"`
	FromDate                    int64                   `db:"from_date" json:"fromDate"`
	ToDate                      int64                   `db:"to_date" json:"toDate"`
}
