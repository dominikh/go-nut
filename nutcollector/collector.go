// Package nutcollector implements a Prometheus collector for Network
// UPS Tools.
package nutcollector

import (
	"log"
	"strconv"

	"honnef.co/go/nut"

	"github.com/prometheus/client_golang/prometheus"
)

var descriptions = map[string]struct {
	name string
	desc string
	enum string
}{
	"device.uptime": {name: "ups_uptime_seconds", desc: "Device uptime"},

	"ups.temperature":       {name: "ups_temperature_celsius", desc: "UPS temperature"},
	"ups.load":              {name: "ups_load_percent", desc: "Load on UPS"},
	"ups.load.high":         {name: "ups_load_high_percent", desc: "Load when UPS switches to overload condition"},
	"ups.efficiency":        {name: "ups_efficiency", desc: "Efficiency of the UPS (ratio of the output current on the input current)"},
	"ups.power":             {name: "ups_power_voltamperes", desc: "Current value of apparent power"},
	"ups.power.nominal":     {name: "ups_power_nominal_voltamperes", desc: "Nominal value of apparent power"},
	"ups.realpower":         {name: "ups_realpower_watts", desc: "Current value of real power"},
	"ups.realpower.nominal": {name: "ups_realpower_nominal_watts", desc: "Nominal value of real power"},
	"ups.beeper.status":     {name: "ups_beeper_status", desc: "UPS beeper status", enum: "status"},
	"ups.status":            {name: "ups_status", desc: "UPS status", enum: "status"},

	"input.voltage":               {name: "input_voltage_volts", desc: "Input voltage"},
	"input.voltage.maximum":       {name: "input_voltage_maximum_volts", desc: "Maximum incoming voltage seen"},
	"input.voltage.minimum":       {name: "input_voltage_minimum_volts", desc: "Minimum incoming voltage seen"},
	"input.voltage.low.warning":   {name: "input_voltage_low_warning_volts", desc: "Low warning threshold"},
	"input.voltage.low.critical":  {name: "input_voltage_low_critical_volts", desc: "Low critical threshold"},
	"input.voltage.high.warning":  {name: "input_voltage_high_warning_volts", desc: "High warning threshold"},
	"input.voltage.high.critical": {name: "input_voltage_high_critical_volts", desc: "High critical threshold"},
	"input.voltage.nominal":       {name: "input_voltage_nominal_volts", desc: "Nominal input voltage"},
	"input.transfer.delay":        {name: "input_transfer_delay_seconds", desc: "Delay before transfer to mains"},
	"input.transfer.low":          {name: "input_transfer_low_volts", desc: "Low voltage transfer point"},
	"input.transfer.high":         {name: "input_transfer_high_volts", desc: "High voltage transfer point"},
	"input.transfer.low.min":      {name: "input_transfer_low_min_volts", desc: "smallest settable low voltage transfer point"},
	"input.transfer.low.max":      {name: "input_transfer_low_max_volts", desc: "greatest settable low voltage transfer point"},
	"input.transfer.high.min":     {name: "input_transfer_high_min_volts", desc: "smallest settable high voltage transfer point"},
	"input.transfer.high.max":     {name: "input_transfer_high_max_volts", desc: "greatest settable high voltage transfer point"},
	"input.current":               {name: "input_current_amperes", desc: "Input current"},
	"input.current.nominal":       {name: "input_current_nominal_amperes", desc: "Nominal input current"},
	"input.current.low.warning":   {name: "input_current_low_warning_amperes", desc: "Low warning threshold"},
	"input.current.low.critical":  {name: "input_current_low_critical_amperes", desc: "Low critical threshold"},
	"input.current.high.warning":  {name: "input_current_high_warning_amperes", desc: "High warning threshold"},
	"input.current.high.critical": {name: "input_current_high_critical_amperes", desc: "High critical threshold"},
	"input.frequency":             {name: "input_frequency_hertz", desc: "Input line frequency"},
	"input.frequency.nominal":     {name: "input_frequency_nominal_hertz", desc: "Nominal input line frequency"},
	"input.frequency.low":         {name: "input_frequency_low_hertz", desc: "Input line frequency low"},
	"input.frequency.high":        {name: "input_frequency_high_hertz", desc: "Input line frequency high"},
	"input.transfer.boost.low":    {name: "input_transfer_boost_low_hertz", desc: "Low voltage boosting transfer point"},
	"input.transfer.boost.high":   {name: "input_transfer_boost_high_hertz", desc: "High voltage boosting transfer point"},
	"input.transfer.trim.low":     {name: "input_transfer_trim_low_hertz", desc: "Low voltage trimming transfer point"},
	"input.transfer.trim.high":    {name: "input_transfer_trim_high_hertz", desc: "High voltage trimming transfer point"},
	"input.load":                  {name: "input_load_percent", desc: "Load on (ePDU) input"},
	"input.realpower":             {name: "input_realpower_watts", desc: "Current sum value of all (ePDU) phases real power"},
	"input.power":                 {name: "input_power_voltamperes", desc: "Current sum value of all (ePDU) phases apparent power"},

	"output.voltage":           {name: "output_voltage_volts", desc: "Output voltage"},
	"output.voltage.nominal":   {name: "output_voltage_nominal_volts", desc: "Nominal output voltage"},
	"output.frequency":         {name: "output_frequency_hertz", desc: "Output frequency"},
	"output.frequency.nominal": {name: "output_frequency_nominal_hertz", desc: "Nominal output frequency"},
	"output.current":           {name: "output_current_amperes", desc: "Output current"},
	"output.current.nominal":   {name: "output_current_nominal_amperes", desc: "Nominal output current"},

	"battery.charge":          {name: "battery_charge_percent", desc: "Battery charge"},
	"battery.charge.low":      {name: "battery_charge_low_percent", desc: "Remaining battery level when UPS switches to LB"},
	"battery.charge.restart":  {name: "battery_charge_restart_percent", desc: "Minimum battery level for UPS restart after power-off"},
	"battery.charge.warning":  {name: "battery_charge_warning_percent", desc: "Battery level when UPS switches to \"Warning\" state"},
	"battery.charger.status":  {name: "battery_charger_status", desc: "Status of the battery charger", enum: "status"},
	"battery.voltage":         {name: "battery_voltage_volts", desc: "Battery voltage"},
	"battery.voltage.nominal": {name: "battery_voltage_nominal_volts", desc: "Nominal battery voltage"},
	"battery.voltage.low":     {name: "battery_voltage_low_volts", desc: "Minimum battery voltage, desc: that triggers FSD status"},
	"battery.voltage.high":    {name: "battery_voltage_high_volts", desc: "Maximum battery voltage (i.e. battery.charge = 100)"},
	"battery.capacity":        {name: "battery_capacity_amperehours", desc: "Battery capacity"},
	"battery.current":         {name: "battery_current_amperes", desc: "Battery current"},
	"battery.current.total":   {name: "battery_current_total_amperes", desc: "Total battery current"},
	"battery.temperature":     {name: "battery_temperature_celsius", desc: "Battery temperature"},
	"battery.runtime":         {name: "battery_runtime_seconds", desc: "Battery runtime"},
	"battery.runtime.low":     {name: "battery_runtime_low_seconds", desc: "Remaining battery runtime when UPS switches to LB"},
	"battery.runtime.restart": {name: "battery_runtime_restart_seconds", desc: "Minimum battery runtime for UPS restart after power-off"},
	"battery.packs":           {name: "battery_packs", desc: "Number of battery packs"},
	"battery.packs.bad":       {name: "battery_packs_bad", desc: "Number of bad battery packs"},
}

// New returns a Prometheus collector, collecting statistics
// from all UPSs on the hosts.
func New(hosts []string) prometheus.Collector {
	const namespace = "nut"

	descs := map[string]*prometheus.Desc{}
	for k, v := range descriptions {
		labels := []string{"name", "model", "mfr", "serial", "type"}
		if v.enum != "" {
			labels = append(labels, v.enum)
		}
		descs[k] = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", v.name),
			v.desc,
			labels,
			nil,
		)
	}

	return &nutCollector{
		hosts: hosts,
		descs: descs,
	}
}

type nutCollector struct {
	hosts []string
	descs map[string]*prometheus.Desc
}

func (c *nutCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range c.descs {
		ch <- v
	}
}

func (c *nutCollector) Collect(ch chan<- prometheus.Metric) {
	for _, host := range c.hosts {
		conn, err := nut.Dial(host)
		if err != nil {
			log.Printf("error connecting to NUT server: %s", err)
			continue
		}
		upss, err := conn.UPSs()
		if err != nil {
			log.Printf("error getting list of UPSs: %s", err)
			_ = conn.Close()
			continue
		}
		for _, ups := range upss {
			if err := c.readNUT(conn, ups, ch); err != nil {
				log.Printf("error reading UPS values: %s", err)
			}
		}
		_ = conn.Close()
	}
}

func (c *nutCollector) readNUT(conn *nut.Client, name string, ch chan<- prometheus.Metric) error {
	vars, err := conn.Variables(name)
	if err != nil {
		return err
	}
	labelValues := []string{
		name, vars["device.model"], vars["device.mfr"], vars["device.serial"], vars["device.type"],
	}
	for k, v := range vars {
		desc, ok := c.descs[k]
		if !ok {
			continue
		}

		if descriptions[k].enum == "" {
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				continue
			}
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, f, labelValues...)
		} else {
			labels := append(labelValues, v)
			ch <- prometheus.MustNewConstMetric(desc, prometheus.GaugeValue, 1.0, labels...)
		}
	}

	return nil
}
