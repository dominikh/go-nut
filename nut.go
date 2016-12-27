package nut_exporter

import (
	"bytes"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"github.com/prometheus/client_golang/prometheus"
)

var descriptions = map[string]struct {
	name string
	desc string
}{
	"device.uptime": {"ups_uptime_seconds", "Device uptime"},

	"ups.temperature":       {"ups_temperature_celsius", "UPS temperature"},
	"ups.load":              {"ups_load_percent", "Load on UPS"},
	"ups.load.high":         {"ups_load_high_percent", "Load when UPS switches to overload condition"},
	"ups.efficiency":        {"ups_efficiency", "Efficiency of the UPS (ratio of the output current on the input current)"},
	"ups.power":             {"ups_power_voltamperes", "Current value of apparent power"},
	"ups.power.nominal":     {"ups_power_nominal_voltamperes", "Nominal value of apparent power"},
	"ups.realpower":         {"ups_realpower_watts", "Current value of real power"},
	"ups.realpower.nominal": {"ups_realpower_nominal_watts", "Nominal value of real power"},
	"ups.beeper.status":     {"ups_beeper_status", "UPS beeper status (enabled = 0, disabled = 1, muted = 2)"},

	"input.voltage":               {"input_voltage_volts", "Input voltage"},
	"input.voltage.maximum":       {"input_voltage_maximum_volts", "Maximum incoming voltage seen"},
	"input.voltage.minimum":       {"input_voltage_minimum_volts", "Minimum incoming voltage seen"},
	"input.voltage.low.warning":   {"input_voltage_low_warning_volts", "Low warning threshold"},
	"input.voltage.low.critical":  {"input_voltage_low_critical_volts", "Low critical threshold"},
	"input.voltage.high.warning":  {"input_voltage_high_warning_volts", "High warning threshold"},
	"input.voltage.high.critical": {"input_voltage_high_critical_volts", "High critical threshold"},
	"input.voltage.nominal":       {"input_voltage_nominal_volts", "Nominal input voltage"},
	"input.transfer.delay":        {"input_transfer_delay_seconds", "Delay before transfer to mains"},
	"input.transfer.low":          {"input_transfer_low_volts", "Low voltage transfer point"},
	"input.transfer.high":         {"input_transfer_high_volts", "High voltage transfer point"},
	"input.transfer.low.min":      {"input_transfer_low_min_volts", "smallest settable low voltage transfer point"},
	"input.transfer.low.max":      {"input_transfer_low_max_volts", "greatest settable low voltage transfer point"},
	"input.transfer.high.min":     {"input_transfer_high_min_volts", "smallest settable high voltage transfer point"},
	"input.transfer.high.max":     {"input_transfer_high_max_volts", "greatest settable high voltage transfer point"},
	"input.current":               {"input_current_amperes", "Input current"},
	"input.current.nominal":       {"input_current_nominal_amperes", "Nominal input current"},
	"input.current.low.warning":   {"input_current_low_warning_amperes", "Low warning threshold"},
	"input.current.low.critical":  {"input_current_low_critical_amperes", "Low critical threshold"},
	"input.current.high.warning":  {"input_current_high_warning_amperes", "High warning threshold"},
	"input.current.high.critical": {"input_current_high_critical_amperes", "High critical threshold"},
	"input.frequency":             {"input_frequency_hertz", "Input line frequency"},
	"input.frequency.nominal":     {"input_frequency_nominal_hertz", "Nominal input line frequency"},
	"input.frequency.low":         {"input_frequency_low_hertz", "Input line frequency low"},
	"input.frequency.high":        {"input_frequency_high_hertz", "Input line frequency high"},
	"input.transfer.boost.low":    {"input_transfer_boost_low_hertz", "Low voltage boosting transfer point"},
	"input.transfer.boost.high":   {"input_transfer_boost_high_hertz", "High voltage boosting transfer point"},
	"input.transfer.trim.low":     {"input_transfer_trim_low_hertz", "Low voltage trimming transfer point"},
	"input.transfer.trim.high":    {"input_transfer_trim_high_hertz", "High voltage trimming transfer point"},
	"input.load":                  {"input_load_percent", "Load on (ePDU) input"},
	"input.realpower":             {"input_realpower_watts", "Current sum value of all (ePDU) phases real power"},
	"input.power":                 {"input_power_voltamperes", "Current sum value of all (ePDU) phases apparent power"},

	"output.voltage":           {"output_voltage_volts", "Output voltage"},
	"output.voltage.nominal":   {"output_voltage_nominal_volts", "Nominal output voltage"},
	"output.frequency":         {"output_frequency_hertz", "Output frequency"},
	"output.frequency.nominal": {"output_frequency_nominal_hertz", "Nominal output frequency"},
	"output.current":           {"output_current_amperes", "Output current"},
	"output.current.nominal":   {"output_current_nominal_amperes", "Nominal output current"},

	"battery.charge":          {"battery_charge_percent", "Battery charge"},
	"battery.charge.low":      {"battery_charge_low_percent", "Remaining battery level when UPS switches to LB"},
	"battery.charge.restart":  {"battery_charge_restart_percent", "Minimum battery level for UPS restart after power-off"},
	"battery.charge.warning":  {"battery_charge_warning_percent", "Battery level when UPS switches to \"Warning\" state"},
	"battery.charger.status":  {"battery_charger_status", "Status of the battery charger (charging = 0, discharging = 1, floating = 2, resting = 3)"},
	"battery.voltage":         {"battery_voltage_volts", "Battery voltage"},
	"battery.voltage.nominal": {"battery_voltage_nominal_volts", "Nominal battery voltage"},
	"battery.voltage.low":     {"battery_voltage_low_volts", "Minimum battery voltage, that triggers FSD status"},
	"battery.voltage.high":    {"battery_voltage_high_volts", "Maximum battery voltage (i.e. battery.charge = 100)"},
	"battery.capacity":        {"battery_capacity_amperehours", "Battery capacity"},
	"battery.current":         {"battery_current_amperes", "Battery current"},
	"battery.current.total":   {"battery_current_total_amperes", "Total battery current"},
	"battery.temperature":     {"battery_temperature_celsius", "Battery temperature"},
	"battery.runtime":         {"battery_runtime_seconds", "Battery runtime"},
	"battery.runtime.low":     {"battery_runtime_low_seconds", "Remaining battery runtime when UPS switches to LB"},
	"battery.runtime.restart": {"battery_runtime_restart_seconds", "Minimum battery runtime for UPS restart after power-off"},
	"battery.packs":           {"battery_packs", "Number of battery packs"},
	"battery.packs.bad":       {"battery_packs_bad", "Number of bad battery packs"},
}

func NewNUTCollector(names []string, hosts []string) prometheus.Collector {
	const namespace = "nut"

	descs := map[string]*prometheus.Desc{}
	for k, v := range descriptions {
		descs[k] = prometheus.NewDesc(
			prometheus.BuildFQName(namespace, "", v.name),
			v.desc,
			[]string{"model", "mfr", "serial", "type"},
			nil,
		)
	}

	return &nutCollector{
		names: names,
		hosts: hosts,
		descs: descs,
	}
}

type nutCollector struct {
	names []string
	hosts []string
	descs map[string]*prometheus.Desc
}

func (c *nutCollector) Describe(ch chan<- *prometheus.Desc) {
	for _, v := range c.descs {
		ch <- v
	}
}

func (c *nutCollector) Collect(ch chan<- prometheus.Metric) {
	names := append(([]string)(nil), c.names...)
	for _, host := range c.hosts {
		hn, err := c.getNUTNames(host)
		if err != nil {
			continue
		}
		names = append(names, hn...)
	}
	for _, name := range names {
		if err := c.readNUT(name, ch); err != nil {
			log.Printf("error reading UPS values: %s", err)
		}
	}
}

func (c *nutCollector) getNUTNames(host string) ([]string, error) {
	cmd := exec.Command("upsc", "-l", host)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	names := strings.Split(string(out), "\n")
	if len(names) > 0 {
		names = names[:len(names)-1]
	}
	return names, nil
}

func (c *nutCollector) readNUT(name string, ch chan<- prometheus.Metric) error {
	cmd := exec.Command("upsc", name)
	out, err := cmd.Output()
	if err != nil {
		return err
	}

	labels := map[string]string{}
	values := map[string]float64{}
	for k := range descriptions {
		values[k] = 0
	}
	for _, l := range bytes.Split(out, []byte("\n")) {
		parts := bytes.SplitN(l, []byte(": "), 2)
		if len(parts) != 2 {
			continue
		}
		k, v := string(parts[0]), string(parts[1])

		switch k {
		case "device.model", "device.mfr", "device.serial", "device.type":
			labels[k] = v
		case "ups.beeper.status":
			f := float64(-1)
			switch v {
			case "enabled":
				f = 9
			case "disabled":
				f = 1
			case "muted":
				f = 2
			}
			values[k] = f
		case "battery.charger.status":
			f := float64(-1)
			switch v {
			case "charging":
				f = 0
			case "discharging":
				f = 1
			case "floating":
				f = 2
			case "resting":
				f = 3
			}
			values[k] = f
		default:
			if _, ok := descriptions[k]; !ok {
				continue
			}
			f, err := strconv.ParseFloat(v, 64)
			if err != nil {
				continue
			}
			values[k] = f
		}
	}

	labelValues := []string{
		labels["device.model"], labels["device.mfr"], labels["device.serial"], labels["device.type"],
	}

	for k, v := range values {
		ch <- prometheus.MustNewConstMetric(c.descs[k], prometheus.GaugeValue, v, labelValues...)
	}
	return nil
}
